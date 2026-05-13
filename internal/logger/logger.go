package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func New(ginMode string) zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339

	if ginMode != "release" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	}

	return log.Logger
}
