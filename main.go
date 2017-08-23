package main

import (
	//"./aitserver"
	//"log"
	"./executor"
	"log"
)

func main() {
	log.Fatal(executor.RunHTTPServer(":8081"))

	//testServer := aitserver.NewServer()
	//
	//err := testServer.InitServer("", "2345")
	//if err != nil {
	//	log.Fatal("[MAIN] Init Server Error:", err)
	//}
	//
	//err = testServer.StartServer()
	//
	//
	//if err != nil {
	//	log.Fatal("[MAIN] Start server Error:", err)
	//}

}
