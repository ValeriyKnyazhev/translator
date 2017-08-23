package main

import (
	"context"

	"github.com/ValeriyKnyazhev/translator/aitserver"
	log "github.com/sirupsen/logrus"
)

func main() {
	//testingDB()
	logger, err := CreateLogger()
	if err != nil {
		log.Panic("Could not create logging facility", err)
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, "logger", logger)
	testServer := aitserver.NewServer()

	err = testServer.InitServer("", "2345", ctx)
	if err != nil {
		logger.Panic("[MAIN] Init Server Error:", err)
	}

	err = testServer.StartServer()

	if err != nil {
		logger.Panic("[MAIN] Start server Error:", err)
	}
}
