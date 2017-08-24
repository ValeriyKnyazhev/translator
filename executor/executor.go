package executor

import (
	"./pool"

	"../database"
	"../grammar"
	"../translator"
	"../vision"
)

type IPool interface {
	Size() int
	Run()
	AddRecognizeTask(requestId string, pictureUrl string, langFrom string, langTo string) (string, error)
	AddCheckTask(requestId string, text string, langFrom string, langTo string) (string, error)
	AddTranslateTask(requestId string, text string, langFrom string, langTo string) (string, error)
}

type Executor struct {
	taskPool    IPool
	recognizer  vision.Vision
	checker     grammar.GrammarChecker
	interpreter translator.Translator
	base        *database.Dbmanager
}

func CreateExecutor(recognizer vision.Vision, checker grammar.GrammarChecker, interpreter translator.Translator, base *database.Dbmanager) *Executor {
	taskPool := pool.NewTaskPool(10, recognizer, checker, interpreter, base)
	taskPool.Run()
	return &Executor{
		recognizer:  recognizer,
		checker:     checker,
		interpreter: interpreter,
		taskPool:    taskPool,
		base:        base,
	}
}

func (e *Executor) AddRecognizeTask(requestId string, pictureUrl string, langFrom string, langTo string) (string, error) {
	e.taskPool.AddRecognizeTask(requestId, pictureUrl, langFrom, langTo)
	return "Recognizing...", nil
}

func (e *Executor) AddCheckTask(requestId string, text string, langFrom string, langTo string) (string, error) {
	e.taskPool.AddCheckTask(requestId, text, langFrom, langTo)
	return "Checking...", nil
}
func (e *Executor) AddTranslateTask(requestId string, text string, langFrom string, langTo string) (string, error) {
	e.taskPool.AddTranslateTask(requestId, text, langFrom, langTo)
	return "Translating...", nil
}
