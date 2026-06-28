// Package logging initialises the application-wide structured logger.
// It configures a JSON-format [log/slog] handler at the requested level and
// sets it as the default logger via [slog.SetDefault].
package logging
