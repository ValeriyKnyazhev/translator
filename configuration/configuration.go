package configuration

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
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

func ReadConfig() (*Config, error) {
	filename, err := filepath.Abs("./resources/config.yaml")
	if err != nil {
		fmt.Println("unable to find absolute configuration file path:", err)
		return nil, err
	}
	configFile, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("unable to read configuration file:", err)
		return nil, err
	}

	config := Config{}
	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		fmt.Println("unable to unmarshal yaml:", err)
		return nil, err
	}

	return &config, err
}

func ReadDBConfig() (*DBConfig, error) {
	filename, err := filepath.Abs("./resources/dbconfig.yaml")
	if err != nil {
		fmt.Println("unable to find absolute configuration file path:", err)
		return nil, err
	}
	configFile, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("unable to read configuration file:", err)
		return nil, err
	}

	config := DBConfig{}
	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		fmt.Println("unable to unmarshal yaml:", err)
		return nil, err
	}

	return &config, err
}
