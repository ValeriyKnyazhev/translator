package grammar

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type GrammarChecker struct {
	ServerUrl   string
	ResourceUrl string
	Client      *http.Client
}

type Word struct {
	Code int
	Word string
	S    []string
}

func CreateGrammarChecker(serverUrl string, resourceUrl string) GrammarChecker {
	return GrammarChecker{serverUrl, resourceUrl, &http.Client{}}
}

func (gChecker *GrammarChecker) CheckPhrase(text string, options ...int) (string, error) {
	data := url.Values{}
	data.Set("text", text)
	if len(options) > 0 {
		data.Set("options", string(options[0]))
	}

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
		fmt.Println("unable to read responce body:", err)
		return "", err
	}
	var words []Word
	err = json.Unmarshal(body, &words)
	if err != nil {
		fmt.Println("unable to get phrase from responce:", err)
	}

	text = replaceWords(text, words)
	dumpInfo(words)
	return text, err
}

func replaceWords(text string, words []Word) string {
	for _, word := range words {
		if len(word.S[0]) > 0 {
			text = strings.Replace(text, word.Word, word.S[0], 1)
		}
	}
	return text
}

func dumpInfo(words []Word) {
	fmt.Println("------------------------------------")
	fmt.Println("Следующие слова были изменены:\n")
	for _, word := range words {
		if len(word.S[0]) > 0 {
			fmt.Println(word.Word, " -> ", word.S[0])
		}
	}
	fmt.Println("------------------------------------")
}
