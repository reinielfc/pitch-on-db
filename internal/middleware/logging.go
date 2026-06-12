package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()

		c.Next()

		slog.Info("request",
			"status", c.Writer.Status(),
			"latency", fmt.Sprintf("%v", time.Since(t)),
			"client_ip", c.ClientIP(),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
		)
	}
}

func VerboseRequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()

		var reqBody []byte
		if c.Request.Body != nil {
			reqBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBody))
		}

		respBuf := &bytes.Buffer{}
		c.Writer = &bodyLogWriter{ResponseWriter: c.Writer, body: respBuf}

		c.Next()

		slog.Info("request",
			"status", c.Writer.Status(),
			"latency", fmt.Sprintf("%v", time.Since(t)),
			"client_ip", c.ClientIP(),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			slog.Any("request", requestLog{headers: c.Request.Header, body: reqBody}),
			slog.Any("response", responseLog{headers: c.Writer.Header(), body: respBuf}),
		)
	}
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

type requestLog struct {
	headers http.Header
	body    []byte
}

func (r requestLog) LogValue() slog.Value {
	attrs := []slog.Attr{slog.Any("headers", r.headers)}
	if len(r.body) > 0 {
		attrs = append(attrs, bodyAttr(r.body))
	}
	return slog.GroupValue(attrs...)
}

type responseLog struct {
	headers http.Header
	body    *bytes.Buffer
}

func (r responseLog) LogValue() slog.Value {
	attrs := []slog.Attr{slog.Any("headers", r.headers)}
	if r.body.Len() > 0 {
		attrs = append(attrs, bodyAttr(r.body.Bytes()))
	}
	return slog.GroupValue(attrs...)
}

func bodyAttr(b []byte) slog.Attr {
	var v any
	if err := json.Unmarshal(b, &v); err == nil {
		return slog.Any("body", v)
	}
	return slog.String("body", string(b))
}
