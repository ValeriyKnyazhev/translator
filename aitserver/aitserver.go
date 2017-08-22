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
	"io/ioutil"
	"log"
	"net/http"
	"os"
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
	logServ.logFile, err = os.OpenFile(logServ.logFileName, os.O_WRONLY | os.O_CREATE, 0666)
	if err != nil {
		return err
	}

	logServ.logObj = log.New(logServ.logFile, "AitHTTPServer: ", log.Lshortfile)

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
		log.Println("Config Error: ", err)
		return err
	}

	servConf.ServerLogger = newLogger(servConf.ServerConfig.HTTPServerLogFile)

	err = servConf.ServerLogger.creatLogger()
	if err != nil {
		log.Println("newLogger Error: ", err)
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
	LangFrom   string `json:"langFrom"`
	LangTo     string `json:"langTo"`
}

func (servConf AitHTTPServer) makeResponse(w http.ResponseWriter, code int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(body)
	if err != nil {
		servConf.ServerLogger.logObj.Println(err)

	}
}

func (servConf AitHTTPServer) CreatNewTranslationTask(w http.ResponseWriter, req *http.Request) {

        log.Println("New connect for POST")

	reqBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		servConf.ServerLogger.logObj.Println(err)
		servConf.makeResponse(w, http.StatusInternalServerError, serverErrorResponse{Code: http.StatusText(http.StatusInternalServerError), Info: fmt.Sprint(err)})
		return
	}

	var reqURLimg serverRequsetURLPost
	err = json.Unmarshal(reqBody, &reqURLimg)
	if err != nil {
		servConf.ServerLogger.logObj.Println(err)
		servConf.makeResponse(w, http.StatusInternalServerError, serverErrorResponse{Code: http.StatusText(http.StatusInternalServerError), Info: fmt.Sprint(err)})
		return
	}

	uuid := uuid.NewV4()
	/*
	   insert to DB
	*/

	/*

	   Start Task

	*/
	//test { "pictureUrl": "https://www.w3.org/TR/SVGTiny12/examples/textArea01.png", "langFrom": "en", "langTo" : "ru" } 
	TEXT, _ := servConf.ServerVision.GetTextFromImg(reqURLimg.PictureUrl, vision.UrlPathType, vision.OcrImgType, reqURLimg.LangFrom)

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
		log.Println("Init Error: ", err)
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
		return err
	}
	return nil
}
