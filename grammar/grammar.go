package grammar

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const (
	IgnoreUppercase      = 1    // Пропускать слова, написанные заглавными буквами, например, "ВПК".
	IgnoreDigits         = 2    // Пропускать слова с цифрами, например, "авп17х4534".
	IgnoreUrls           = 4    // Пропускать интернет-адреса, почтовые адреса и имена файлов.
	FindRepeatWords      = 8    // Подсвечивать повторы слов, идущие подряд. Например, "я полетел на на Кипр".
	IgnoreLatin          = 16   // Пропускать слова, написанные латиницей, например, "madrid".
	NoSuggest            = 32   // Только проверять текст, не выдавая вариантов для замены.
	FlagLatin            = 128  // Отмечать слова, написанные латиницей, как ошибочные.
	ByWords              = 256  // Не использовать словарное окружение (контекст) при проверке. Опция полезна в случа    ях, когда на вход сервиса передается список отдельных слов.
	IgnoreCapitalization = 512  // Игнорировать неверное употребление ПРОПИСНЫХ/строчных букв, например, в слове "мос    ква".
	IgnoreRomanNumerals  = 2048 // Игнорировать римские цифры ("I, II, III, ...").
)

type GrammarChecker struct {
	ServerUrl   string
	ResourceUrl string
	Client      *http.Client
	options     int
}

type Word struct {
	Code int
	Word string
	S    []string
}

func CreateGrammarChecker(serverUrl string, resourceUrl string) GrammarChecker {
	options := IgnoreUrls + IgnoreUppercase
	return GrammarChecker{serverUrl, resourceUrl, &http.Client{}, options}
}

func (gChecker *GrammarChecker) CheckPhrase(text string, options ...int) (string, error) {
	data := url.Values{}
	data.Set("text", text)
	data.Set("options", string(gChecker.options))

	urlPath, err := url.ParseRequestURI(gChecker.ServerUrl)
	if err != nil {
		fmt.Println("unable to parse request url:", err)
		return "", err
	}
	urlPath.Path = gChecker.ResourceUrl
	urlPath.RawQuery = data.Encode()

	grammarRequest, err := http.NewRequest("GET", fmt.Sprintf("%v", urlPath), nil)
	if err != nil {
		fmt.Println("unable to create request:", err)
		return "", err
	}
	grammarResponse, err := gChecker.Client.Do(grammarRequest)
	if err != nil {
		fmt.Println("unable to execute http request:", err)
		return "", err
	}
	defer grammarResponse.Body.Close()

	body, err := ioutil.ReadAll(grammarResponse.Body)
	if err != nil {
		fmt.Println("unable to read response body:", err)
		return "", err
	}
	var words []Word
	err = json.Unmarshal(body, &words)
	if err != nil {
		fmt.Println("unable to get phrase from response:", err)
		return "", err
	}
	text = replaceWords(text, words)
	dumpInfo(words)
	return text, err
}

func replaceWords(text string, words []Word) string {
	for _, word := range words {
		if len(word.S) > 0 {
			text = strings.Replace(text, word.Word, word.S[0], 1)
		}
	}
	return text
}

func dumpInfo(words []Word) {
	log.Println("------------------------------------")
	log.Println("Следующие слова были изменены:")
	for _, word := range words {
		if len(word.S) > 0 {
			log.Println(word.Word, " -> ", word.S[0])
		}
	}
	log.Println("------------------------------------")
}
