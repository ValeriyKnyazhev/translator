package aitserver

import (
	"fmt"
	"github.com/ValeriyKnyazhev/translator/configuration"
	"github.com/ValeriyKnyazhev/translator/grammar"
	"github.com/ValeriyKnyazhev/translator/translator"
	"github.com/ValeriyKnyazhev/translator/vision"
	"net"
)

type AitServer struct {
	ServerConfig  *configuration.Config
	ServerVision  vision.Vision
	ServerGrammar grammar.GrammarChecker
	ServerTrans   translator.Translator
	Host          string
	Port          string
}

func NewServer() AitServer {
	return AitServer{&configuration.Config{}, vision.Vision{}, grammar.GrammarChecker{}, translator.Translator{}, "", ""}
}

func (servConf *AitServer) InitServer(host string, port string) (err error) {
	servConf.Host = host
	servConf.Port = port

	servConf.ServerConfig, err = configuration.ReadConfig()
	if err != nil {
		return
	}

	servConf.ServerVision = vision.CreateVisoin(
		servConf.ServerConfig.VisionServerUrl,
		servConf.ServerConfig.VisionApiKey)

	servConf.ServerGrammar = grammar.CreateGrammarChecker(
		servConf.ServerConfig.GrammarServerUrl,
		servConf.ServerConfig.GrammarResourceUrl)

	servConf.ServerTrans = translator.CreateTranslator(
		servConf.ServerConfig.TranslatorServerUrl,
		servConf.ServerConfig.TranslatorResourceUrl,
		servConf.ServerConfig.TranslatorApiKey)

	return
}

func (servConf *AitServer) StartServer() (err error) {
	lnistenServ, err := net.Listen("tcp", servConf.Host+":"+servConf.Port)
	if err != nil {
		fmt.Println(err)
		return
	}
	for {
		connection, err := lnistenServ.Accept()
		if err != nil {
			continue
		}
		go servConf.connectionHandler(connection)
	}

}

func (servConf AitServer) connectionHandler(connect net.Conn) (err error) {
	defer connect.Close()

	const buffMaxSize = 1024

	textBuff := make([]byte, buffMaxSize)
	reqStr := ""
	bitNum := 0

	for bitNum, err = connect.Read(textBuff); bitNum > 0; bitNum, err = connect.Read(textBuff) {
		var itArr int

		if err != nil {
			return
		}

		for itArr = 0; itArr < bitNum; itArr++ {
			if textBuff[itArr] == '\n' {
				break
			}
		}
		reqStr += string(textBuff[:itArr])
		if bitNum != itArr {
			break
		}
	}

	fmt.Println(reqStr)

	imgDes, err := servConf.ServerVision.GetTextFromImg("http://5klass.net/datas/russkij-jazyk/Prichastie-10-klass/0008-008-Rabota-s-tekstom-variant-A-zadanie-v-kakoj-posledovatelnosti.jpg", vision.UrlPathType, vision.OcrImgType, "en")

	if err != nil {

		fmt.Println(err)

		return
	}

	fmt.Println(imgDes.Text)

	textGram, err := servConf.ServerGrammar.CheckPhrase(imgDes.Text)
	if err != nil {
		return
	}

	translation, err := servConf.ServerTrans.Translate("en-ru", textGram)
	if err != nil {
		return
	}

	connect.Write([]byte(translation.Text[0]))

	return
}
