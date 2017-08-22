package pool

import (
	"../task"
	"errors"
	"github.com/ValeriyKnyazhev/translator/grammar"
	"github.com/ValeriyKnyazhev/translator/translator"
	"github.com/ValeriyKnyazhev/translator/vision"
	"sync"
	"time"
)

const timeout time.Duration = time.Second * 2

var (
	TimeoutError = errors.New("request timed out")
)

type TaskPool struct {
	concurrency   int
	recognizeChan chan *task.RecognizeTask
	checkChan     chan *task.GrammarCheckTask
	translateChan chan *task.TranslateTask
	wgRecognize   sync.WaitGroup
	wgCheck       sync.WaitGroup
	wgTranslate   sync.WaitGroup
	recognizer    vision.Vision
	checker       grammar.GrammarChecker
	interpreter   translator.Translator
}

func (p *TaskPool) Size() int {
	return p.concurrency
}

func NewTaskPool(concurrency int,
	recognizer vision.Vision,
	checker grammar.GrammarChecker,
	interpreter translator.Translator) *TaskPool {
	return &TaskPool{
		concurrency:   concurrency,
		recognizeChan: make(chan *task.RecognizeTask),
		checkChan:     make(chan *task.GrammarCheckTask),
		translateChan: make(chan *task.TranslateTask),
		recognizer:    recognizer,
		checker:       checker,
		interpreter:   interpreter,
	}
}

func (p *TaskPool) Run() {
	for i := 0; i < p.concurrency; i++ {
		p.wgRecognize.Add(1)
		p.wgCheck.Add(1)
		p.wgTranslate.Add(1)
		go p.recognize()
		go p.check()
		go p.translate()
	}
}

func (p *TaskPool) Stop() {
	close(p.recognizeChan)
	close(p.checkChan)
	close(p.translateChan)
	p.wgRecognize.Wait()
	p.wgCheck.Wait()
	p.wgTranslate.Wait()
}

func (p *TaskPool) AddRecognizeTask(requestId int, pictureUrl string, langTo string) (string, error) {
	t := task.RecognizeTask{
		RequestId:  requestId,
		Wg:         sync.WaitGroup{},
		PictureUrl: pictureUrl,
		LangTo:     langTo,
	}

	t.Wg.Add(1)
	select {
	case p.recognizeChan <- &t:
		break
	case <-time.After(timeout):
		return "", TimeoutError
	}

	t.Wg.Wait()
	return "Recognize task has been successfully added", nil
}

func (p *TaskPool) AddCheckTask(requestId int, text string, langFrom string, langTo string) (string, error) {
	t := task.GrammarCheckTask{
		RequestId:      requestId,
		Wg:             sync.WaitGroup{},
		RecognizedText: text,
		LangFrom:       langFrom,
		LangTo:         langTo,
	}

	t.Wg.Add(1)
	select {
	case p.checkChan <- &t:
		break
	case <-time.After(timeout):
		return "", TimeoutError
	}

	t.Wg.Wait()
	return "Recognize task has been successfully added", nil
}

func (p *TaskPool) AddTranslateTask(requestId int, text string, langFrom string, langTo string) (string, error) {
	t := task.TranslateTask{
		RequestId:   requestId,
		Wg:          sync.WaitGroup{},
		CheckedText: text,
		LangFrom:    langFrom,
		LangTo:      langTo,
	}

	t.Wg.Add(1)
	select {
	case p.translateChan <- &t:
		break
	case <-time.After(timeout):
		return "", TimeoutError
	}

	t.Wg.Wait()
	return "Recognize task has been successfully added", nil
}

func (p *TaskPool) recognize() {
	for t := range p.recognizeChan {
		p.recognizer.GetTextFromImg(t.PictureUrl, vision.UrlPathType, vision.OcrImgType, "en")
		t.Wg.Done()
	}
	p.wgRecognize.Done()
}

func (p *TaskPool) check() {
	for t := range p.checkChan {
		p.checker.CheckPhrase(t.RecognizedText)
		t.Wg.Done()
	}
	p.wgCheck.Done()
}

func (p *TaskPool) translate() {
	for t := range p.translateChan {
		lang := t.LangFrom + "-" + t.LangTo
		p.interpreter.Translate(lang, t.CheckedText)
		t.Wg.Done()
	}
	p.wgTranslate.Done()
}
