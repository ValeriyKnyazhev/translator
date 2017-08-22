package main

import (
	"./aitserver"
	"log"
)

func main() {

	testServer := aitserver.NewHTTPServer()

	err := testServer.RunHTTPServer()
	if err != nil {
		log.Println("[MAIN] Init Server Error:", err)
	}

}
