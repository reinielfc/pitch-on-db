package slicesx

// Map applies f to each element of input and returns a new slice
// containing the results. It returns nil if input is nil.
func Map[F any, T any](input []F, f func(F) T) []T {
	if input == nil {
		return nil
	}

	result := make([]T, len(input))
	for i, v := range input {
		result[i] = f(v)
	}
	return result
}

// AsAny converts a slice of any type F to a slice of any type, returning a new slice.
func AsAny[F any](input []F) []any {
	return Map(input, func(s F) any { return s })
}
