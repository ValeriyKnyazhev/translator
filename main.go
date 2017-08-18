package main

import (
	"fmt"
	"github.com/ValeriyKnyazhev/translator/configuration"
	"github.com/ValeriyKnyazhev/translator/grammar"
	"github.com/ValeriyKnyazhev/translator/translator"
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

	gChecker := grammar.CreateGrammarChecker(
		config.GrammarServerUrl,
		config.GrammarResourceUrl)

	text, err := gChecker.CheckPhrase("Професор Немур и доктар Штраус прихадили ко мне в комнату штобы узнать почиму я не пришол в лабалаторию.")
	if err != nil {
		fmt.Println("[MAIN] unable to parse request url:", err)
		os.Exit(1)
	}
	fmt.Println(text)

	translation, err := interpreter.Translate("en-ru", "Hello world!\nMy name is Valeriy.")
	if err != nil {
		fmt.Println("[MAIN] unable to parse request url:", err)
		os.Exit(1)
	}
	fmt.Println(translation)
}
