package server

import (
	"snwzt/random-video-chat/data/interfaces"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type UserHTTPServer struct {
	port     string
	engine   *echo.Echo
	handlers interfaces.UserHandler
}

func NewUserHTTPServer(port string, engine *echo.Echo, handlers interfaces.UserHandler) *UserHTTPServer {
	return &UserHTTPServer{
		port:     port,
		engine:   engine,
		handlers: handlers,
	}
}

func (s *UserHTTPServer) Run() {
	s.engine.Use(middleware.Logger())
	s.engine.Use(middleware.Recover())

	s.engine.GET("/health", s.handlers.CheckHealth)
	s.engine.GET("/", s.handlers.Home)
	s.engine.POST("/register", s.handlers.RegisterUser)
	s.engine.GET("/match", s.handlers.MatchUser)

	s.engine.Logger.Fatal(s.engine.Start(s.port))
}

type UserQueue struct {
	handlers interfaces.UserQueueHandler
}

func NewUserQueue(handlers interfaces.UserQueueHandler) *UserQueue {
	return &UserQueue{
		handlers: handlers,
	}
}

func (u *UserQueue) Run() {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		u.handlers.Matcher()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		u.handlers.UserRemove()
	}()

	wg.Wait()
}
