package pool

import (
	"errors"
	"fmt"
	"github.com/ValeriyKnyazhev/translator/database"
	"github.com/ValeriyKnyazhev/translator/executor/task"
	"github.com/ValeriyKnyazhev/translator/grammar"
	"github.com/ValeriyKnyazhev/translator/translator"
	"github.com/ValeriyKnyazhev/translator/vision"
	log "github.com/sirupsen/logrus"
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
	base          *database.Dbmanager
}

func (p *TaskPool) Size() int {
	return p.concurrency
}

func NewTaskPool(concurrency int,
	recognizer vision.Vision,
	checker grammar.GrammarChecker,
	interpreter translator.Translator, base *database.Dbmanager) *TaskPool {
	return &TaskPool{
		concurrency:   concurrency,
		recognizeChan: make(chan *task.RecognizeTask),
		checkChan:     make(chan *task.GrammarCheckTask),
		translateChan: make(chan *task.TranslateTask),
		recognizer:    recognizer,
		checker:       checker,
		interpreter:   interpreter,
		base:          base,
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

func (p *TaskPool) AddRecognizeTask(requestId string, pictureUrl string, langFrom string, langTo string) (string, error) {
	t := task.RecognizeTask{
		RequestId:  requestId,
		Wg:         sync.WaitGroup{},
		PictureUrl: pictureUrl,
		LangFrom:   langFrom,
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

func (p *TaskPool) AddCheckTask(requestId string, text string, langFrom string, langTo string) (string, error) {
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
	return "Check grammatics task has been successfully added", nil
}

func (p *TaskPool) AddTranslateTask(requestId string, text string, langFrom string, langTo string) (string, error) {
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
	return "Translate task has been successfully added", nil
}

func (p *TaskPool) recognize() {
	for t := range p.recognizeChan {
		dbData, err := p.base.GetData(t.RequestId)
		if err != nil {
			log.Println(err)
		}

		text, err := p.recognizer.GetTextFromImg(t.PictureUrl, "en")
		t.Wg.Done()

		if err == nil {
			dbData.CurrTaskId = 2
			dbData.RecognizedText = text.Text
			dbData.Status = database.TaskStatusRun
			dbData.Error = database.TaskErrNone
			err = p.base.UpdateData(dbData)
			if err != nil {
				log.Println(err)
			} else {
				log.WithFields(log.Fields{
					"RequestId": t.RequestId,
					"Text":      text.Text,
					"LangFrom":  t.LangFrom,
					"LangTo":    t.LangTo,
				}).Info("Recognize success")
				p.AddCheckTask(t.RequestId, text.Text, t.LangFrom, t.LangTo)
			}
		} else {
			dbData.Status = database.TaskStatusStop
			dbData.Error = err.Error()
			err = p.base.UpdateData(dbData)
			if err != nil {
				log.Println(err)
			}
		}
	}
	p.wgRecognize.Done()
}

func (p *TaskPool) check() {
	for t := range p.checkChan {
		dbData, err := p.base.GetData(t.RequestId)
		if err != nil {
			log.Println(err)
		}

		text, err := p.checker.CheckPhrase(t.RecognizedText)
		t.Wg.Done()
		if err == nil {
			dbData.CurrTaskId = 3
			dbData.CheckedText = text
			dbData.Status = database.TaskStatusRun
			dbData.Error = database.TaskErrNone
			err = p.base.UpdateData(dbData)
			if err != nil {
				log.Println(err)
			} else {
				log.WithFields(log.Fields{
					"RequestId": t.RequestId,
					"Text":      text,
					"LangFrom":  t.LangFrom,
					"LangTo":    t.LangTo,
				}).Info("Gramma check success")
				p.AddTranslateTask(t.RequestId, text, t.LangFrom, t.LangTo)
			}
		} else {
			dbData.Status = database.TaskStatusStop
			dbData.Error = err.Error()
			err = p.base.UpdateData(dbData)
			if err != nil {
				log.Println(err)
			}

		}
	}
	p.wgCheck.Done()
}

func (p *TaskPool) translate() {
	for t := range p.translateChan {
		dbData, err := p.base.GetData(t.RequestId)
		if err != nil {
			log.Println(err)
		}

		lang := t.LangFrom + "-" + t.LangTo
		text, err := p.interpreter.Translate(lang, t.CheckedText)
		t.Wg.Done()
		if err == nil {
			log.WithFields(log.Fields{
				"Lang": lang,
				"Text": text,
			}).Info("Translate success")
			log.Println("Translated text: %s", text.Text)
			dbData.TranslatedText = fmt.Sprintf("%s", text.Text)
			dbData.Status = database.TaskStatusComplete
			dbData.Error = database.TaskErrNone
			err = p.base.UpdateData(dbData)
			if err != nil {
				log.Println(err)
			}
		} else {
			dbData.Status = database.TaskStatusStop
			dbData.Error = err.Error()
			err = p.base.UpdateData(dbData)
			if err != nil {
				log.Println(err)
			}
		}
	}
	p.wgTranslate.Done()
}
