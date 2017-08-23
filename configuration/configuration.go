package configuration

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/go-yaml/yaml"
)

type Config struct {
	VisionApiKey          string `yaml:"visionApiKey"`
	VisionServerUrl       string `yaml:"visionServerUrl"`
	GrammarApiKey         string `yaml:"grammarApiKey"`
	GrammarServerUrl      string `yaml:"grammarServerUrl"`
	GrammarResourceUrl    string `yaml:"grammarResourceUrl"`
	TranslatorApiKey      string `yaml:"translatorApiKey"`
	TranslatorServerUrl   string `yaml:"translatorServerUrl"`
	TranslatorResourceUrl string `yaml:"translatorResourceUrl"`
}

type DBConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml":port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBname   string `yaml:"dbname"`
}

func read(config interface{}, path string) (interface{}, error) {
	filename, err := filepath.Abs(path)
	if err != nil {
		fmt.Println("unable to find absolute configuration file path:", err)
		return nil, err
	}
	configFile, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("unable to read configuration file:", err)
		return nil, err
	}
	err = yaml.Unmarshal(configFile, config)
	if err != nil {
		fmt.Println("unable to unmarshal yaml:", err)
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

func ReadDBConfig(path string) (*DBConfig, error) {
	conf, err := read(&DBConfig{}, path) // "./resources/dbconfig.yaml"
	return conf.(*DBConfig), err
}

func ReadDBConfigDefault() (*DBConfig, error) {
	conf, err := read(&DBConfig{}, "./resources/dbconfig.yaml")
	return conf.(*DBConfig), err
}
