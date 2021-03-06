package main

import (
	"github.com/ValeriyKnyazhev/translator/aitserver"
	log "github.com/sirupsen/logrus"
)

func main() {

	testServer := aitserver.NewHTTPServer()

	err := testServer.RunHTTPServer()
	if err != nil {
		log.Println("[MAIN] Init Server Error:", err)
	}
}
