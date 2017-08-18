package aitserver

import (
	"../configuration"
	"../grammar"
	"../translator"
	"../vision"
	"fmt"
	"net"
)

type AitServer struct {
	ServerConfig  configuration.Config
	ServerVision  vision.Visoin
	ServerGrammar grammar.GrammarChecker
	ServerTrans   translator.Translator
	Host          string
	Port          string
}

func NewServer() AitServer {
	return AitServer{configuration.Config{}, vision.Visoin{}, grammar.GrammarChecker{}, translator.Translator{}, "", ""}
}

func (servConf *AitServer) InitServer(host string, port string) (err error) {
	servConf.Host = host
	servConf.Port = port

	servConf.ServerConfig = configuration.ReadConfig()
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
		go handleServerConnection(c)
	}

}

func (servConf AitServer) connectionHandler(connect net.Conn) (err error) {
	defer connect.Close()

	const buffMaxSize = 1024

	textBuff := make([]byte, buffMaxSize)
	reqStr := ""
	bitNum := 0

	for bitNum, err = connect.Read(textBuff); bitNum > 0; bitNum, err = file.Read(buff) {
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
		if butNum != itArr {
			break
		}
	}

	imgDes, err := visioonn.GetTextFromImg(reqStr, vision.UrlPathType, vision.OcrImgType, "en")

	if err != nil {
		return
	}

	textGram, err := servConf.ServerGrammar.CheckPhrase(imgDes.Text)
	if err != nil {
		return
	}

	translation, err := servConf.ServerTrans.Translate("en-ru", textGram)
	if err != nil {
		return
	}

	connect.Write([]byte(reqStr))
}
