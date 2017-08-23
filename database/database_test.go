package database

import (
	"fmt"
	"log"
	"testing"
)

func TestDatabase(t *testing.T) {
	var manager DBManager = CreateFromConfig("../resources/dbconfig.yaml")
	err := manager.CreateTable()
	if err != nil {
		log.Println("[MAIN] can't create table")
	}

	d := Data{Id: "74b64702-8608-11e7-bb31-be2e44b06b34",
		UserId: 15, PictureUrl: "picture2", RecognizedText: "qwerty",
		RecognizedLang: "en", CheckedText: "qwertyu",
		TranslatedText: "йцукенг", TranslatedLang: "ru", Error: "none"}
	err = manager.SetData(&d)
	if err != nil {
		t.Errorf("can't set data to database")
	}

	d = Data{Id: "74b64702-8608-11e7-bb31-be2e44b06b34",
		UserId: 999, PictureUrl: "picture15", RecognizedText: "anothertext",
		RecognizedLang: "uk", CheckedText: "texttext",
		TranslatedText: "тексттекст", TranslatedLang: "ua", Error: "something wrong"}
	err = manager.UpdateData(&d)
	if err != nil {
		t.Errorf("can't set data to database")
	}

	data, err := manager.GetData("74b64702-8608-11e7-bb31-be2e44b06b34")
	if err != nil {
		t.Errorf("[MAIN] can't get data")
	}
	fmt.Printf("%+v", data)
}
