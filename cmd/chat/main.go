package main

import (
	"os"
	"snwzt/rvc/internal/handlers"
	"snwzt/rvc/pkg/logger"
	"snwzt/rvc/services/chat"
	"snwzt/rvc/services/db"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

func main() {
	logger := logger.NewLogger()
	err := godotenv.Load("config/.env")
	if err != nil {
		logger.Err(err).Msg("unable to load .env")
	}

	redis, err := db.NewRedisStore(os.Getenv("REDIS_URI"))
	if err != nil {
		logger.Err(err).Msg("unable to connect to redis")
	}

	instance := echo.New()
	chatServerHandle := &handlers.ChatServerHandle{
		Redis:  redis,
		Logger: logger,
	}
	s := chat.NewChatServer(":5001", instance, chatServerHandle, logger)

	s.Run()
}
