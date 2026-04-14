package logger

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

type Logger struct {
	*slog.Logger
}

var L Logger

func Init(serviceName string) {
	var handler slog.Handler

	if os.Getenv("GO_ENV") == "development" {
		handler = tint.NewHandler(os.Stdout, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, nil)
	}

	logger := slog.New(handler).With("service", serviceName)
	slog.SetDefault(logger)
	L = Logger{logger}
}
