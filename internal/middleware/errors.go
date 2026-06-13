package middleware

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"pitch-on-db/internal/errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last()

		switch e := err.Err.(type) {
		case *errors.AppError:
			if e.Status >= 500 {
				slog.Error("internal error",
					"error", e.Cause,
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
				)
			}
			c.JSON(e.Status, gin.H{
				"code":    e.Code,
				"message": e.Message,
			})
		case validator.ValidationErrors:
			fields := make(map[string]string, len(e))
			for _, fe := range e {
				fields[strings.ToLower(fe.Field())] = fe.Tag()
			}
			c.JSON(http.StatusBadRequest, gin.H{
				"code":   http.StatusText(http.StatusBadRequest),
				"fields": fields,
			})
		case *json.SyntaxError, *json.UnmarshalTypeError:
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusText(http.StatusBadRequest),
				"message": "invalid request body",
			})
		default:
			slog.Error("unhandled error",
				"error", err.Err,
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
