package executor

import (
	"./pool"

	"github.com/ValeriyKnyazhev/translator/grammar"
	"github.com/ValeriyKnyazhev/translator/translator"
	"github.com/ValeriyKnyazhev/translator/vision"
)

type IPool interface {
	Size() int
	Run()
	AddRecognizeTask(requestId int, pictureUrl string, langFrom string, langTo string) (string, error)
	AddCheckTask(requestId int, text string, langFrom string, langTo string) (string, error)
	AddTranslateTask(requestId int, text string, langFrom string, langTo string) (string, error)
}

type Executor struct {
	taskPool    IPool
	recognizer  vision.Vision
	checker     grammar.GrammarChecker
	interpreter translator.Translator
}

func CreateExecutor(recognizer vision.Vision, checker grammar.GrammarChecker, interpreter translator.Translator) *Executor {
	taskPool := pool.NewTaskPool(10, recognizer, checker, interpreter)
	taskPool.Run()
	return &Executor{
		recognizer:  recognizer,
		checker:     checker,
		interpreter: interpreter,
		taskPool:    taskPool,
	}
}

func (e *Executor) AddRecognizeTask(requestId int, pictureUrl string, langFrom string, langTo string) (string, error) {
	e.taskPool.AddRecognizeTask(requestId, pictureUrl, langFrom, langTo)
	return "Recognizing...", nil
}

func (e *Executor) AddCheckTask(requestId int, text string, langFrom string, langTo string) (string, error) {
	e.taskPool.AddCheckTask(requestId, text, langFrom, langTo)
	return "Checking...", nil
}
func (e *Executor) AddTranslateTask(requestId int, text string, langFrom string, langTo string) (string, error) {
	e.taskPool.AddTranslateTask(requestId, text, langFrom, langTo)
	return "Translating...", nil
}
