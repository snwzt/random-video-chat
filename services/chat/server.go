package chat

import (
	"net/http"
	"snwzt/rvc/data/interfaces"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
)

type ChatServer struct {
	port     string
	engine   *echo.Echo
	handlers interfaces.ChatHandler
	logger   *zerolog.Logger
}

func NewChatServer(port string, engine *echo.Echo, handlers interfaces.ChatHandler, logger *zerolog.Logger) *ChatServer {
	return &ChatServer{
		port:     port,
		engine:   engine,
		handlers: handlers,
		logger:   logger,
	}
}

func (svc *ChatServer) Run() {
	svc.engine.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			svc.logger.Info().
				Str("URI", v.URI).
				Int("status", v.Status).
				Msg("request")

			return nil
		},
	}))
	svc.engine.Use(middleware.Recover())
	svc.engine.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	svc.engine.GET("/health", svc.handlers.CheckHealth)
	svc.engine.GET("/chat/:id", svc.handlers.Chat)

	svc.engine.Logger.Fatal(svc.engine.Start(svc.port))
}
