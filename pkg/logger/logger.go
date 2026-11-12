package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

func NewLogger(level zerolog.Level, pretty bool) zerolog.Logger {
	var writer io.Writer = os.Stdout
	if pretty {
		writer = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339Nano,
		}
	}

	logger := zerolog.New(writer).
		Level(level).
		With().
		Timestamp().
		Logger()

	return logger
}

func New(level string, pretty bool) zerolog.Logger {
	loggerLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		loggerLevel = zerolog.DebugLevel
	}

	log := NewLogger(
		loggerLevel, pretty,
	)

	if err != nil {
		log.Warn().Msg(
			"Config: logger level wasn't parsed. Set default level = DEBUG",
		)
	}

	return log
}
