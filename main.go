package main

import (
	"./configuration"
	"./translator"
	"fmt"
	"os"
)

func main() {
	config, err := configuration.ReadConfig()
	if err != nil {
		fmt.Println("[MAIN] unable to read configuration:", err)
		os.Exit(1)
	}

	interpreter := translator.CreateTranslator(
		config.TranslatorServerUrl,
		config.TranslatorResourceUrl,
		config.TranslatorApiKey)

	translation, err := interpreter.Translate("en-ru", "Hello world!\nMy name is Valeriy.")
	if err != nil {
		fmt.Println("[MAIN] unable to parse request url:", err)
		os.Exit(1)
	}
	fmt.Println(translation)

}
