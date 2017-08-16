package main

import (
	"fmt"
	"net/http"
	"net/url"
	"io/ioutil"
	"encoding/json"
	"io"
	"log"
	"gopkg.in/yaml.v2"
	"path/filepath"
)

type TranslationResponse struct {
	Lang string
	Text []string
}

type Config struct {
	ApiKey string `yaml:"ApiKey"`
	Hits int64 `yaml:"Hits"`
}

func (c *Config) readConfig() *Config {
	filename, _ := filepath.Abs("./resources/config.yaml")
	yamlFile, err := ioutil.ReadFile(filename)

	if err != nil {
		fmt.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("unable to unmarshal yaml: %v", err)
	}

	return c
}

func getTranslation(response io.ReadCloser) (*TranslationResponse, error) {
	body, err := ioutil.ReadAll(response)
	if err != nil {
		fmt.Println("unable to read response body:", err)
	}

	data := TranslationResponse{}
	err = json.Unmarshal(body, &data);
	if err != nil {
		fmt.Println("unable to get translation from json:", err)
	}
	var translation = new(TranslationResponse)
	translation.Lang = data.Lang
	translation.Text = data.Text
	return translation, err
}

func main() {

	translatorUrl := "https://translate.yandex.net"
	resource := "/api/v1.5/tr.json/translate"
	data := url.Values{}
	
	config := Config{}
	config.readConfig()
	fmt.Println(config)
	data.Set("text", "Hello world!")
	data.Set("lang", "en-ru")
	data.Set("key", config.ApiKey)

	u, _ := url.ParseRequestURI(translatorUrl)
	u.Path = resource
	u.RawQuery = data.Encode()

	urlStr := fmt.Sprintf("%v", u)

	client := &http.Client{}
	r, _ := http.NewRequest("POST", urlStr, nil)//bytes.NewBufferString(data.Encode()))
	r.Header.Add("Content-Type", "application/json")

	response, _ := client.Do(r)
	defer response.Body.Close()
	fmt.Println(response.Status)

	s, err := getTranslation(response.Body)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(s.Lang)
	fmt.Println(s.Text)
}