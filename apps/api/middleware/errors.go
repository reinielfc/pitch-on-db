package middleware

import (
	"log/slog"
	"net/http"

	"github.com/reinielfc/pitch-on-db/apps/api/domain"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}
		if c.Writer.Written() {
			slog.Warn("error occurred but response already written",
				"errors", c.Errors.Errors(),
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
			)
			return
		}

		err := c.Errors.Last().Err

		switch e := err.(type) {
		case domain.DomainError:
			status := statusForCode(e.Code())
			c.JSON(status, gin.H{
				"code":    http.StatusText(status),
				"message": e.Error(),
			})

		case validator.ValidationErrors:
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"code":   "Validation Error",
				"fields": formatValidationErrors(e),
			})

		default:
			slog.Error("unhandled error",
				"error", err,
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    http.StatusText(http.StatusInternalServerError),
				"message": "something went wrong",
			})
		}
	}
}

func statusForCode(code domain.ErrorCode) int {
	switch code {
	case domain.ErrNotFound:
		return http.StatusNotFound
	case domain.ErrConflict:
		return http.StatusConflict
	case domain.ErrInvalid:
		return http.StatusBadRequest
	case domain.ErrUnauthorized:
		return http.StatusUnauthorized
	case domain.ErrForbidden:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}
