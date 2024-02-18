package server

import (
	"net/http"
	"snwzt/random-video-chat/data/interfaces"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type ChatHTTPServer struct {
	port     string
	engine   *echo.Echo
	handlers interfaces.ChatHandler
}

func NewChatHTTPServer(port string, engine *echo.Echo, handlers interfaces.ChatHandler) *ChatHTTPServer {
	return &ChatHTTPServer{
		port:     port,
		engine:   engine,
		handlers: handlers,
	}
}

func (s *ChatHTTPServer) Run() {
	s.engine.Use(middleware.Logger())
	s.engine.Use(middleware.Recover())
	s.engine.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	s.engine.GET("/health", s.handlers.CheckHealth)
	s.engine.GET("/chat/:id", s.handlers.Chat)

	s.engine.Logger.Fatal(s.engine.Start(s.port))
}
