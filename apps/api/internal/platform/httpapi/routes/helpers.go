package routes

import "github.com/danielgtaylor/huma/v2"

func WithSummary(summary string) func(o *huma.Operation) {
	return func(o *huma.Operation) {
		o.Summary = summary
	}
}
