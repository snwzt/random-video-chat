package main

import (
	"log"
	"os"
	"snwzt/rvc/internal/handlers"
	"snwzt/rvc/pkg/common"
	"snwzt/rvc/pkg/logger"
	"snwzt/rvc/services/db"
	"snwzt/rvc/services/user"
	"sync"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

func main() {
	var wg sync.WaitGroup
	err := godotenv.Load("config/.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	instance := echo.New()

	instance.Renderer, err = common.NewTemplate("web/*.html")
	if err != nil {
		log.Println(err)
	}

	redis, err := db.NewRedisStore(os.Getenv("REDIS_URI"))
	if err != nil {
		log.Println(err)
	}
	userHttpHandle := &handlers.UserServerHandle{
		Redis: redis,
	}

	logger := logger.NewLogger()

	s := user.NewUserServer(":5000", instance, userHttpHandle, logger)

	wg.Add(1)
	go func() {
		defer wg.Done()

		s.Run()
	}()

	userOperationsHandle := &handlers.UserOperationsHandle{
		Redis: redis,
	}
	userOperations := user.NewUserOperations(userOperationsHandle)

	wg.Add(1)
	go func() {
		defer wg.Done()

		userOperations.Run()
	}()

	wg.Wait()
}
