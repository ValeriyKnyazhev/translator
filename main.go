package main

import (
	"./aitserver"
	"fmt"
	"log"
)

func main() {

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
