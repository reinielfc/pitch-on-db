package logging

import (
	"fmt"
	"log/slog"
	"os"
)

func Init(level string) (*slog.Logger, error) {
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}

	l := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl}))
	slog.SetDefault(l)
	l.Info("starting logger", "level", lvl)
	return l, nil
}
