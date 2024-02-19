package logger

import (
	"os"

	"github.com/rs/zerolog"
)

func NewLogger() *zerolog.Logger {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	return &logger
}
