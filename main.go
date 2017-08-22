package main

import (
	"fmt"
	"github.com/ValeriyKnyazhev/translator/aitserver"
	"github.com/ValeriyKnyazhev/translator/database"
	"log"
)

func main() {
	testingDB()
	testServer := aitserver.NewServer()

	err := testServer.InitServer("", "2345")
	if err != nil {
		log.Fatal("[MAIN] Init Server Error:", err)
	}

	err = testServer.StartServer()

	if err != nil {
		log.Fatal("[MAIN] Start server Error:", err)
	}
}

func testingDB() {
	err := database.Manager.CreateTable()
	if err != nil {
		log.Println("[MAIN] can't create table")
	}

	d := database.Data{Id: "74b64702-8608-11e7-bb31-be2e44b06b34",
		UserId: 15, PictureUrl: "picture2", RecognizedText: "qwerty",
		RecognizedLang: "en", CheckedText: "qwertyu",
		TranslatedText: "йцукенг", TranslatedLang: "ru", Error: "none"}
	err = database.Manager.SetData(&d)
	if err != nil {
		log.Println("can't set data to database")
	}

	d = database.Data{Id: "74b64702-8608-11e7-bb31-be2e44b06b34",
		UserId: 999, PictureUrl: "picture15", RecognizedText: "anothertext",
		RecognizedLang: "uk", CheckedText: "texttext",
		TranslatedText: "тексттекст", TranslatedLang: "ua", Error: "something wrong"}
	err = database.Manager.UpdateData(&d)
	if err != nil {
		log.Println("can't set data to database")
	}

	data, err := database.Manager.GetData("74b64702-8608-11e7-bb31-be2e44b06b34")
	if err != nil {
		log.Println("[MAIN] can't get data")
	}
	fmt.Printf("%+v", data)
}
