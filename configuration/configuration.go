package configuration

import (
	"github.com/go-yaml/yaml"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"path/filepath"
)

type Config struct {
	Api struct {
		VisionApiKey          string `yaml:"visionApiKey"`
		VisionServerUrl       string `yaml:"visionServerUrl"`
		GrammarApiKey         string `yaml:"grammarApiKey"`
		GrammarServerUrl      string `yaml:"grammarServerUrl"`
		GrammarResourceUrl    string `yaml:"grammarResourceUrl"`
		TranslatorApiKey      string `yaml:"translatorApiKey"`
		TranslatorServerUrl   string `yaml:"translatorServerUrl"`
		TranslatorResourceUrl string `yaml:"translatorResourceUrl"`
	} `ymal:"api"`
	Server struct {
		HTTPServerHost    string `yaml:"httpServerHost"`
		HTTPServerPort    string `yaml:"httpServerPort"`
		HTTPServerLogFile string `yaml:"httpServerLogFile"`
	} `ymal:"server"`
	DB struct {
		Host     string `yaml:"host"`
		Port     int    `yaml":port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DBname   string `yaml:"dbname"`
	} `ymal:"db"`
}

func read(config interface{}, path string) (interface{}, error) {
	filename, err := filepath.Abs(path)
	if err != nil {
		log.Println("unable to find absolute configuration file path:", err)
		return nil, err
	}
	configFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Println("unable to read configuration file:", err)
		return nil, err
	}
	err = yaml.Unmarshal(configFile, config)
	if err != nil {
		log.Println("unable to unmarshal yaml:", err)
		return nil, err
	}
	return config, nil
}

func ReadConfig(path string) (*Config, error) {
	conf, err := read(&Config{}, path) // "./resources/config.yaml"
	return conf.(*Config), err
}

func ReadConfigDefault() (*Config, error) {
	conf, err := read(&Config{}, "./resources/config.yaml")
	return conf.(*Config), err
}
