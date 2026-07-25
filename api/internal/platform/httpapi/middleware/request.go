package middleware

import (
	"context"
	"net/http"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

const (
	ContextKeyRequestID = "requestID"
)

func BridgeRequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := requestid.Get(c)
		c.Request = setRequestValue(c.Request, ContextKeyRequestID, requestID)
		c.Next()
	}
}

func setRequestValue(req *http.Request, key, value any) *http.Request {
	return req.WithContext(context.WithValue(req.Context(), key, value))
}
