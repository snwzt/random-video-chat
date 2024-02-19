package main

import (
	"log"
	"os"
	"snwzt/rvc/internal/handlers"
	"snwzt/rvc/pkg/logger"
	"snwzt/rvc/services/chat"
	"snwzt/rvc/services/db"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

func main() {
	err := godotenv.Load("config/.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	redis, err := db.NewRedisStore(os.Getenv("REDIS_URI"))
	if err != nil {
		log.Println(err)
	}

	logger := logger.NewLogger()

	instance := echo.New()
	chatServerHandle := &handlers.ChatServerHandle{
		Redis: redis,
	}
	s := chat.NewChatServer(":5001", instance, chatServerHandle, logger)

	s.Run()
}
