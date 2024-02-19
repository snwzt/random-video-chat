package main

import (
	"os"
	"snwzt/rvc/internal/handlers"
	"snwzt/rvc/pkg/logger"
	"snwzt/rvc/services/db"
	"snwzt/rvc/services/forwarder"

	"github.com/joho/godotenv"
)

func main() {
	logger := logger.NewLogger()
	err := godotenv.Load("config/.env")
	if err != nil {
		logger.Err(err).Msg("unable to load .env")
	}

	cancelChan := make(chan string)
	defer close(cancelChan)

	redis, err := db.NewRedisStore(os.Getenv("REDIS_URI"))
	if err != nil {
		logger.Err(err).Msg("unable to connect to redis")
	}

	forwarderOperationsHandle := &handlers.ForwarderOperationsHandle{
		Redis:           redis,
		CancelForwarder: cancelChan,
		Logger:          logger,
	}
	forwarder := forwarder.NewForwarder(forwarderOperationsHandle)

	forwarder.Run()
}
