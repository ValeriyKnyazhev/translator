package aitserver

import (
	"../configuration"
	"../grammar"
	"../translator"
	"../vision"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
)

type aitServerLog struct {
	logFileName string
	logFile     *os.File
	logObj      *log.Logger
}

func newLogger(logName string) *aitServerLog {
	return &aitServerLog{logName, nil, nil}
}

func (logServ *aitServerLog) creatLogger() error {
	var err error
	logServ.logFile, err = os.OpenFile(logServ.logFileName, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Println("creat logger Error: ", err)
		return err
	}

	logServ.logObj = log.New()

	logServ.logObj.Out = logServ.logFile
	logServ.logObj.SetLevel(log.DebugLevel)

	return nil
}

func (logServ *aitServerLog) closeLogger() error {
	return logServ.logFile.Close()
}

type AitHTTPServer struct {
	ServerConfig  *configuration.Config
	ServerVision  vision.Vision
	ServerGrammar grammar.GrammarChecker
	ServerTrans   translator.Translator
	ServerHost    string
	ServerPort    string
	ServerLogger  *aitServerLog
}

func NewHTTPServer() *AitHTTPServer {
	return &AitHTTPServer{&configuration.Config{},
		vision.Vision{},
		grammar.GrammarChecker{},
		translator.Translator{},
		"",
		"",
		&aitServerLog{}}
}

func (servConf *AitHTTPServer) initServer() error {

	var err error
	servConf.ServerConfig, err = configuration.ReadConfig()
	if err != nil {
		log.Println("Init server Error: ", err)
		return err
	}

	servConf.ServerLogger = newLogger(servConf.ServerConfig.HTTPServerLogFile)

	err = servConf.ServerLogger.creatLogger()
	if err != nil {
		log.Println("Init server Error: ", err)
		return err
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

	servConf.ServerHost = servConf.ServerConfig.HTTPServerHost
	servConf.ServerPort = servConf.ServerConfig.HTTPServerPort

	return nil
}

func (servConf *AitHTTPServer) stopServer() error {
	err := servConf.ServerLogger.closeLogger()
	if err != nil {
		log.Println("Stop server Error: ", err)
		return err
	}
	return nil
}

type serverErrorResponse struct {
	Code string `json:"code"`
	Info string `json:"info"`
}

type serverAcceptedResponse struct {
	Info string `json:"info"`
	Id   string `json:"id"`
}

type serverRequsetURLPost struct {
	PictureUrl string `json:"pictureUrl"`
}

const (
	HeaderLangFrom = "Img-Lang-From"
	HeaderLangTo   = "Img-Lang-To"
	HeaderContType = "Content-Type"
	JsonContType   = "application/json"
	OctetsContType = "multipart/"
)

func (servConf AitHTTPServer) makeResponse(w http.ResponseWriter, code int, body interface{}) {
	w.Header().Set("Content-Type", JsonContType)
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(body)
	if err != nil {
		servConf.ServerLogger.logObj.Println(err)
	}
}

func (servConf AitHTTPServer) CreatNewTranslationTask(w http.ResponseWriter, req *http.Request) {

	log.Println("New connect with POST")
	uuid := uuid.NewV4()

	langFrom := req.Header.Get(HeaderLangFrom)

	if langFrom == "" {
		servConf.makeResponse(w, http.StatusBadRequest,
			serverErrorResponse{Code: http.StatusText(http.StatusBadRequest),
				Info: "Header Img-Lang-From wasm't found"})
		return
	}

	langTo := req.Header.Get(HeaderLangTo)
	if langTo == "" {
		servConf.makeResponse(w, http.StatusBadRequest,
			serverErrorResponse{Code: http.StatusText(http.StatusBadRequest),
				Info: "Header Img-Lang-To wasm't found"})
		return
	}
	contType, params, err := mime.ParseMediaType(req.Header.Get(HeaderContType))
	if err != nil {
		servConf.ServerLogger.logObj.Println(err)
		servConf.makeResponse(w, http.StatusInternalServerError,
			serverErrorResponse{Code: http.StatusText(http.StatusInternalServerError),
				Info: fmt.Sprint(err)})
		return
	}
	if contType == "" {
		servConf.makeResponse(w, http.StatusBadRequest,
			serverErrorResponse{Code: http.StatusText(http.StatusBadRequest),
				Info: "Header Content-Type wasm't found"})
		return
	}

	var pathToImg string

	if strings.HasPrefix(contType, OctetsContType) {
		pathToImg = fmt.Sprintf("%s.img", uuid)
		var imgFile *os.File
		imgFile, err = os.OpenFile(pathToImg, os.O_WRONLY|os.O_CREATE, 0666)
		defer imgFile.Close()
		if err != nil {
			servConf.ServerLogger.logObj.Println(err)
			servConf.makeResponse(w, http.StatusInternalServerError,
				serverErrorResponse{Code: http.StatusText(http.StatusInternalServerError),
					Info: fmt.Sprint(err)})
			return
		}
		pathToImg = "file://" + pathToImg

		mr := multipart.NewReader(req.Body, params["boundary"])

		for {
			npart, err := mr.NextPart()
			if err == io.EOF {

				err = nil

				break
			}
			if err != nil {
				servConf.ServerLogger.logObj.Println(err)
				servConf.makeResponse(w, http.StatusInternalServerError,
					serverErrorResponse{Code: http.StatusText(http.StatusInternalServerError),
						Info: fmt.Sprint(err)})
				return
			}
			slurp, err := ioutil.ReadAll(npart)
			if err != nil {
				servConf.ServerLogger.logObj.Println(err)
				servConf.makeResponse(w, http.StatusInternalServerError,
					serverErrorResponse{Code: http.StatusText(http.StatusInternalServerError),
						Info: fmt.Sprint(err)})
				return
			}
			_, err = imgFile.Write(slurp)
			if err != nil {
				servConf.ServerLogger.logObj.Println(err)
				servConf.makeResponse(w, http.StatusInternalServerError,
					serverErrorResponse{Code: http.StatusText(http.StatusInternalServerError),
						Info: fmt.Sprint(err)})
				return
			}
		}

	} else if contType == JsonContType {

		var reqURLimg serverRequsetURLPost

		reqBody, err := ioutil.ReadAll(req.Body)
		if err != nil {
			servConf.ServerLogger.logObj.Println(err)
			servConf.makeResponse(w, http.StatusInternalServerError,
				serverErrorResponse{Code: http.StatusText(http.StatusInternalServerError),
					Info: fmt.Sprint(err)})
			return
		}
		err = json.Unmarshal(reqBody, &reqURLimg)
		if err != nil {
			servConf.ServerLogger.logObj.Println(err)
			servConf.makeResponse(w, http.StatusInternalServerError,
				serverErrorResponse{Code: http.StatusText(http.StatusInternalServerError),
					Info: fmt.Sprint(err)})
			return
		}
		pathToImg = reqURLimg.PictureUrl
	} else {
		servConf.makeResponse(w, http.StatusBadRequest,
			serverErrorResponse{Code: http.StatusText(http.StatusBadRequest),
				Info: "Invalide content-type"})
		return
	}

	/*
	   insert to DB
	*/

	/*

	   Start Task

	*/
	//test { "pictureUrl": "https://www.w3.org/TR/SVGTiny12/examples/textArea01.png", "langFrom": "en", "langTo" : "ru" }
	TEXT, err := servConf.ServerVision.GetTextFromImg(pathToImg, langFrom)

	if err != nil {
		servConf.makeResponse(w, http.StatusBadRequest,
			serverErrorResponse{Code: http.StatusText(http.StatusBadRequest),
				Info: fmt.Sprint(err)})
		return
	}
	servConf.makeResponse(w, http.StatusAccepted, serverAcceptedResponse{Info: TEXT.Text, Id: string(uuid[:])})
	//	servConf.makeResponse(w, http.StatusAccepted, serverAcceptedResponse{Info: "For future translation use request Id", Id: string(uuid[:])})

}

type serverOKResponseForGet struct {
	LangFrom string `json:"langFrom"`
	LangTo   string `json:"langTo"`
	Text     string `json:"text"`
}

func (servConf AitHTTPServer) GetTranslationResult(w http.ResponseWriter, req *http.Request) {

	log.Println("New connect for GET")
	params := mux.Vars(req)
	taskID := params["id"]

	/*
	   insert to DB
	*/
	/* if not found
		       servConf.ServerLogger.Println(<<err msg>>)
	               makeResponse(w, http.StatusNotFound, &serverErrorResponse{Code: http.StatusText(http.StatusNotFound), Info: fmt.Sprint(<<err msg>>)})
		       return
		   else if task run or die
		       if initTime > 5 (min)
		          restart task

		       servConf.ServerLogger.Println(<<err msg>>)
		       makeResponse(w, http.StatusAccepted, &serverAcceptedResponse {Info: "For future translation use request Id", Id: taskID})
		       return

		   else if task is complite
		   makeResponse(w, http.StatusOK, &serverOKResponseForGet {LangFrom: "unk", LangTo: "unk", Text: "text"})
		       return
	*/
	//test
	servConf.makeResponse(w, http.StatusOK, serverOKResponseForGet{LangFrom: "unk", LangTo: "unk", Text: taskID})

}

func (servConf *AitHTTPServer) RunHTTPServer() error {

	err := servConf.initServer()
	if err != nil {
		log.Println("Run server Error: ", err)
		return err
	}
	defer servConf.stopServer()

	log.Println("Server run!")
	log.Println("Host: ", servConf.ServerHost)
	log.Println("Port: ", servConf.ServerPort)

	router := mux.NewRouter()
	router.HandleFunc("/translation", servConf.CreatNewTranslationTask).Methods(http.MethodPost)
	router.HandleFunc("/translation/{id}", servConf.GetTranslationResult).Methods(http.MethodGet)

	err = http.ListenAndServe(servConf.ServerHost+":"+servConf.ServerPort, router)
	if err != nil {
		log.Println("Run server Error: ", err)
		return err
	}
	return nil
}
