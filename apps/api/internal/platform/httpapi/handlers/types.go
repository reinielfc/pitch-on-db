package handlers

import (
	"net/http"
	"reflect"

	"github.com/danielgtaylor/huma/v2"
)

// I is a placeholder struct for endpoints that do not require any input parameters.
type I struct{}

// ResourceIDInput is a struct for endpoints that require a resource ID as input.
type ResourceIDInput struct {
	ID string `path:"id"`
}

// NoContentOutput is a struct for endpoints that return no content in the response body.
type NoContentOutput struct {
	Status int `json:"-"`
}

// NoContent returns a NoContentOutput with the appropriate HTTP status code for a successful operation that does not return any content.
func NoContent() *NoContentOutput {
	return &NoContentOutput{Status: http.StatusNoContent}
}

// List is a generic slice type that implements huma.SchemaProvider to ensure
// the OpenAPI schema is generated as `array` rather than `array or null`.
type List[T any] []T

func (List[T]) Schema(r huma.Registry) *huma.Schema {
	t := reflect.TypeFor[T]()
	return &huma.Schema{
		Type:  "array",
		Items: r.Schema(t, true, t.Name()),
	}
}

func stringToAny(v string) any { return v }
