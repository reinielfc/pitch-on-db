package telemetry

import (
	"log/slog"
	"os"
)

func SetupLogging(debug bool) {
	var lvl slog.Level
	if debug {
		lvl = slog.LevelDebug
	} else {
		lvl = slog.LevelInfo
	}

	jsonLogger := slog.New(slog.NewJSONHandler(
		os.Stdout, &slog.HandlerOptions{Level: lvl},
	))

	slog.SetDefault(jsonLogger)
	jsonLogger.Info("starting logger", "level", lvl)
}
