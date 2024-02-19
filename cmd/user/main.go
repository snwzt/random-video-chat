package main

import (
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
	logger := logger.NewLogger()
	var wg sync.WaitGroup
	err := godotenv.Load("config/.env")
	if err != nil {
		logger.Err(err).Msg("unable to load .env")
	}

	instance := echo.New()

	instance.Renderer, err = common.NewTemplate("web/*.html")
	if err != nil {
		logger.Err(err).Msg("unable to load templates")
	}

	redis, err := db.NewRedisStore(os.Getenv("REDIS_URI"))
	if err != nil {
		logger.Err(err).Msg("unable to connect to redis")
	}
	userHttpHandle := &handlers.UserServerHandle{
		Redis: redis,
	}

	s := user.NewUserServer(":5000", instance, userHttpHandle, logger)

	wg.Add(1)
	go func() {
		defer wg.Done()

		s.Run()
	}()

	userOperationsHandle := &handlers.UserOperationsHandle{
		Redis:  redis,
		Logger: logger,
	}
	userOperations := user.NewUserOperations(userOperationsHandle)

	wg.Add(1)
	go func() {
		defer wg.Done()

		userOperations.Run()
	}()

	wg.Wait()
}
