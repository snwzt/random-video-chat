package user

import (
	"snwzt/rvc/data/interfaces"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
)

type UserServer struct {
	port     string
	engine   *echo.Echo
	handlers interfaces.UserHTTPHandler
	logger   *zerolog.Logger
}

func NewUserServer(port string, engine *echo.Echo, handlers interfaces.UserHTTPHandler, logger *zerolog.Logger) *UserServer {
	return &UserServer{
		port:     port,
		engine:   engine,
		handlers: handlers,
		logger:   logger,
	}
}

func (svc *UserServer) Run() {
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

	svc.engine.GET("/health", svc.handlers.CheckHealth)
	svc.engine.GET("/", svc.handlers.Home)
	svc.engine.POST("/register", svc.handlers.RegisterUser)
	svc.engine.GET("/match", svc.handlers.MatchUser)

	svc.engine.Logger.Fatal(svc.engine.Start(svc.port))
}
