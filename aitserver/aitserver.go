package aitserver

import (
	"encoding/json"
	"fmt"
	"github.com/ValeriyKnyazhev/translator/configuration"
	"github.com/ValeriyKnyazhev/translator/database"
	"github.com/ValeriyKnyazhev/translator/grammar"
	"github.com/ValeriyKnyazhev/translator/translator"
	"github.com/ValeriyKnyazhev/translator/vision"
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
	"time"
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
		log.Println(logServ.logFileName)
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
	DataBase      *database.Dbmanager
	ServerHost    string
	ServerPort    string
	ServerLogger  *aitServerLog
}

func NewHTTPServer() *AitHTTPServer {
	return &AitHTTPServer{ServerConfig: &configuration.Config{},
		ServerVision:  vision.Vision{},
		ServerGrammar: grammar.GrammarChecker{},
		ServerTrans:   translator.Translator{},
		DataBase:      &database.Dbmanager{},
		ServerHost:    "",
		ServerPort:    "",
		ServerLogger:  &aitServerLog{}}
}

func (servConf *AitHTTPServer) initServer() error {

	var err error
	servConf.ServerConfig, err = configuration.ReadConfigDefault()
	if err != nil {
		log.Println("Init server Error: ", err)
		return err
	}

	servConf.ServerLogger = newLogger(servConf.ServerConfig.Server.HTTPServerLogFile)

	err = servConf.ServerLogger.creatLogger()
	if err != nil {
		log.Println("Init server Error: ", err)
		return err
	}

	servConf.DataBase, err = database.CreateDB(servConf.ServerConfig.DB.Host,
		servConf.ServerConfig.DB.Port,
		servConf.ServerConfig.DB.User,
		servConf.ServerConfig.DB.Password,
		servConf.ServerConfig.DB.DBname)

	if err != nil {
		log.Println("Init server Error: ", err)
		return err
	}

	servConf.DataBase.CreateTable()
	if err != nil {
		servConf.ServerLogger.logObj.Println("Init server Error: ", err)
	}

	servConf.ServerVision = vision.CreateVision(
		servConf.ServerConfig.Api.VisionServerUrl,
		servConf.ServerConfig.Api.VisionApiKey)

	servConf.ServerGrammar = grammar.CreateGrammarChecker(
		servConf.ServerConfig.Api.GrammarServerUrl,
		servConf.ServerConfig.Api.GrammarResourceUrl)

	servConf.ServerTrans = translator.CreateTranslator(
		servConf.ServerConfig.Api.TranslatorServerUrl,
		servConf.ServerConfig.Api.TranslatorResourceUrl,
		servConf.ServerConfig.Api.TranslatorApiKey)

	servConf.ServerHost = servConf.ServerConfig.Server.HTTPServerHost
	servConf.ServerPort = servConf.ServerConfig.Server.HTTPServerPort

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

	insertedData := database.Data{Id: fmt.Sprintf("%s", uuid),
		UserId: 1, CurrTaskId: 1, PictureUrl: pathToImg, RecognizedText: "",
		RecognizedLang: langFrom, CheckedText: "",
		TranslatedText: "", TranslatedLang: langTo, Status: database.TaskStatusRun,
		Error: database.TaskErrNone}

	err = servConf.DataBase.SetData(&insertedData)
	if err != nil {
		servConf.ServerLogger.logObj.Println(err)
		servConf.makeResponse(w, http.StatusInternalServerError,
			serverErrorResponse{Code: http.StatusText(http.StatusInternalServerError),
				Info: fmt.Sprint("can't set data to database", err)})
		return
	}
	/*

	   Start New Task

	*/
	servConf.makeResponse(w, http.StatusAccepted, serverAcceptedResponse{Info: "For future translation use request Id", Id: fmt.Sprintf("%s", uuid)})

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

	dbData, err := servConf.DataBase.GetData(taskID)
	if err != nil {
		servConf.makeResponse(w, http.StatusInternalServerError,
			serverErrorResponse{Code: http.StatusText(http.StatusInternalServerError),
				Info: fmt.Sprint(err)})
		return
	}
	if dbData == nil {
		servConf.makeResponse(w, http.StatusBadRequest,
			serverErrorResponse{Code: http.StatusText(http.StatusBadRequest),
				Info: "No rows were returned!"})
		return
	}

	currentTime := (dbData.Timestamp.Unix() - time.Now().Unix()) / 60.0

	if dbData.CurrTaskId == 3 && dbData.Status == database.TaskStatusComplete {
		servConf.makeResponse(w, http.StatusOK, serverOKResponseForGet{LangFrom: dbData.RecognizedLang,
			LangTo: dbData.TranslatedLang,
			Text:   dbData.TranslatedText})
		return
	} else if dbData.Status == database.TaskStatusStop || currentTime >= 5.0 {
		updateData := database.Data{Id: dbData.Id,
			UserId: 1, CurrTaskId: dbData.CurrTaskId, PictureUrl: dbData.PictureUrl, RecognizedText: dbData.RecognizedText,
			RecognizedLang: dbData.RecognizedLang, CheckedText: dbData.CheckedText,
			TranslatedText: dbData.TranslatedText, TranslatedLang: dbData.TranslatedLang, Status: database.TaskStatusRun,
			Error: database.TaskErrNone}

		err = servConf.DataBase.UpdateData(&updateData)
		if err != nil {
			servConf.ServerLogger.logObj.Println(err)
			servConf.makeResponse(w, http.StatusInternalServerError,
				serverErrorResponse{Code: http.StatusText(http.StatusInternalServerError),
					Info: fmt.Sprint(err)})
			return
		}
		/*
		   restart task
		*/
		servConf.makeResponse(w, http.StatusAccepted, serverAcceptedResponse{Info: "For future translation use request Id", Id: dbData.Id})
		return
	}

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
