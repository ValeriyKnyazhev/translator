package translator

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type Translation struct {
	Status   int
	LangFrom string
	LangTo   string
	Text     []string
}

type translationData struct {
	Code int
	Lang string
	Text []string
}

func checkLanguage(lang string) (string, error) {
	//TODO create global map of supported languages and check if this lang exists there
	return lang, nil
}

func (data *translationData) parseData() (*Translation, error) {
	delimeterPosition := strings.Index(data.Lang, "-")
	langFrom, err := checkLanguage(data.Lang[:delimeterPosition])
	if err != nil {
		fmt.Println("source language not found in the list of supported languages:", err)
		return nil, err
	}
	langTo, err := checkLanguage(data.Lang[delimeterPosition+1:])
	if err != nil {
		fmt.Println("translation language not found in the list of supported languages:", err)
		return nil, err
	}
	return &Translation{data.Code, langFrom, langTo, data.Text}, err
}

type Translator struct {
	ServerUrl   string
	ResourceUrl string
	ApiKey      string
	Client      *http.Client
}

func CreateTranslator(serverUrl string, resourceUrl string, apiKey string) Translator {
	return Translator{serverUrl, resourceUrl, apiKey, &http.Client{}}
}

func (translator *Translator) Translate(lang string, text string) (*Translation, error) {

	data := url.Values{}
	data.Set("text", text)
	data.Set("lang", lang)
	data.Set("key", translator.ApiKey)

	urlPath, err := url.ParseRequestURI(translator.ServerUrl)
	if err != nil {
		fmt.Println("unable to parse request url:", err)
		return nil, err
	}
	urlPath.Path = translator.ResourceUrl
	urlPath.RawQuery = data.Encode()

	translatorRequest, err := http.NewRequest("POST", fmt.Sprintf("%v", urlPath), nil)
	if err != nil {
		fmt.Println("unable to create request:", err)
		return nil, err
	}
	translatorRequest.Header.Add("Content-Type", "application/json")

	translatorResponse, err := translator.Client.Do(translatorRequest)
	if err != nil {
		fmt.Println("unable to execute http request:", err)
		return nil, err
	}
	defer translatorResponse.Body.Close()

	translation, err := getTranslation(translatorResponse.Body)
	if err != nil {
		fmt.Println("unable to parse translation:", err)
		return nil, err
	}
	return translation, err
}

func getTranslation(response io.ReadCloser) (*Translation, error) {
	body, err := ioutil.ReadAll(response)
	if err != nil {
		fmt.Println("unable to read response body:", err)
		return nil, err
	}
	data := translationData{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println("unable to get translation from response:", err)
	}

	translation, err := data.parseData()
	if err != nil {
		fmt.Println("unable to parse translation data:", err)
	}
	return translation, err
}
