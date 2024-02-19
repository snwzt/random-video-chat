package main

import (
	"log"
	"os"
	"snwzt/rvc/internal/handlers"
	"snwzt/rvc/services/db"
	"snwzt/rvc/services/forwarder"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load("config/.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	cancelChan := make(chan string)
	defer close(cancelChan)

	redis, err := db.NewRedisStore(os.Getenv("REDIS_URI"))
	if err != nil {
		log.Println(err)
	}

	forwarderOperationsHandle := &handlers.ForwarderOperationsHandle{
		Redis:           redis,
		CancelForwarder: cancelChan,
	}
	forwarder := forwarder.NewForwarder(forwarderOperationsHandle)

	forwarder.Run()
}
