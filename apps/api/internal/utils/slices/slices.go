package slices

func Map[F any, T any](input []F, f func(F) T) []T {
	result := make([]T, len(input))
	for i, v := range input {
		result[i] = f(v)
	}
	return result
}
