package task

import "sync"

type RecognizeTask struct {
	RequestId  string
	Wg         sync.WaitGroup
	PictureUrl string
	LangFrom   string
	LangTo     string
}

type GrammarCheckTask struct {
	RequestId      string
	Wg             sync.WaitGroup
	RecognizedText string
	LangFrom       string
	LangTo         string
}

type TranslateTask struct {
	RequestId   string
	Wg          sync.WaitGroup
	CheckedText string
	LangFrom    string
	LangTo      string
}
