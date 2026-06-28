// Package handlers contains the HTTP handler types for the API. Each handler
// struct wraps a corresponding [services] interface and exposes methods that
// satisfy the Gin handler signature, parsing request parameters and delegating
// to the service layer.
package handlers
