package main

import (
	"fmt"
	"github.com/ValeriyKnyazhev/translator/aitserver"
	"github.com/ValeriyKnyazhev/translator/database"
	"log"
)

func main() {
	err := database.Manager.CreateTable()
	if err != nil {
		log.Println("[MAIN] can't create table")
	}
	data, err := database.Manager.GetData("74b64702-8608-11e7-bb31-be2e44b06b34")
	if err != nil {
		log.Println("[MAIN] can't get data")
	}
	fmt.Printf("%+v", data)

	d := database.Data{Id: "74b64702-8608-11e7-bb31-be2e44b06b34",
		UserId: 15, PictureUrl: "picture2", RecognizedText: "qwerty",
		RecognizedLang: "en", CheckedText: "qwertyu",
		TranslatedText: "йцукенг", TranslatedLang: "ru"}
	err = database.Manager.SetData(&d)
	if err != nil {
		log.Println("can't set data to database")
	}
	testServer := aitserver.NewServer()

	err = testServer.InitServer("", "2345")
	if err != nil {
		log.Fatal("[MAIN] Init Server Error:", err)
	}

	err = testServer.StartServer()

	if err != nil {
		log.Fatal("[MAIN] Start server Error:", err)
	}

}
