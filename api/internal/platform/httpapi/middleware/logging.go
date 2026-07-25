package middleware

import (
	"context"
	"log/slog"
	"time"

	"github.com/danielgtaylor/huma/v2"
)

const (
	headerAccept      = "Accept"
	headerContentType = "Content-Type"
	headerRequestID   = "X-Request-ID"
	headerTraceparent = "Traceparent"
	headerUserAgent   = "User-Agent"
)

func LogRequests() func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		start := time.Now()
		reqProps := getRequestAttrs(ctx)
		next(ctx)
		resProps := getResponseAttrs(ctx, start)

		slog.Info("request", "request", reqProps, "response", resProps)
	}
}

var requestHeaders = []string{
	headerAccept,
	headerContentType,
	headerTraceparent,
	headerUserAgent,
}

func getRequestAttrs(ctx huma.Context) []slog.Attr {
	headers := getHeaderAttrs(ctx, requestHeaders...)

	attrs := []slog.Attr{
		slog.String("method", ctx.Method()),
		slog.String("path", ctx.URL().Path),
		slog.String("client", ctx.RemoteAddr()),
		slog.GroupAttrs("headers", headers...),
	}

	if requestID, ok := getRequestID(ctx.Context()); ok {
		attrs = append(attrs, slog.String("id", requestID))
	}

	return attrs
}

func getRequestID(ctx context.Context) (string, bool) {
	requestID, ok := ctx.Value(ContextKeyRequestID).(string)
	return requestID, ok
}

func getHeaderAttrs(ctx huma.Context, headerNames ...string) []slog.Attr {
	attrs := []slog.Attr{}
	for _, headerName := range headerNames {
		if headerValue := ctx.Header(headerName); headerValue != "" {
			attrs = append(attrs, slog.String(headerName, headerValue))
		}
	}
	return attrs
}

func getResponseAttrs(ctx huma.Context, startTime time.Time) []slog.Attr {
	return []slog.Attr{
		slog.Int("status", ctx.Status()),
		slog.String("latency", time.Since(startTime).String()),
	}
}
