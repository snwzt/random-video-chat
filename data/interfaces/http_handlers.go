package interfaces

import "github.com/labstack/echo/v4"

type UserHTTPHandler interface {
	CheckHealth(echo.Context) error
	Home(echo.Context) error
	RegisterUser(echo.Context) error
	MatchUser(echo.Context) error
}

type ChatHandler interface {
	CheckHealth(echo.Context) error
	Chat(echo.Context) error
}
