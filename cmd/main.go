package main

import (
	"os"

	"snwzt/rvc/cmd/sugar"
	"snwzt/rvc/pkg/logger"
	"snwzt/rvc/services/db"

	"github.com/joho/godotenv"
)

func main() {
	logger := logger.NewLogger()

	if err := godotenv.Load("config/.env"); err != nil {
		logger.Err(err).Msg("unable to load env file")
	}

	redis, err := db.NewRedisStore(os.Getenv("REDIS_URI"))
	if err != nil {
		logger.Err(err).Msg("unable to connect to redis")
	}

	sugar.Execute(os.Exit, os.Args[1:], redis, logger)
}
