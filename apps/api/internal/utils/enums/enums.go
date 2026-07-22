package enums

import "fmt"

// StringPtr converts an enum-like value implementing fmt.Stringer to a *string, nil-safe.
func StringPtr[T fmt.Stringer](v *T) *string {
	if v == nil {
		return nil
	}
	s := (*v).String()
	return &s
}

// ParsePtr parses a *string into a *T using the given parse function, nil-safe.
func ParsePtr[T fmt.Stringer](s *string, parse func(string) (T, error)) (*T, error) {
	if s == nil {
		return nil, nil
	}
	val, err := parse(*s)
	if err != nil {
		return nil, err
	}
	return &val, nil
}

// AsType converts an enum-like value implementing fmt.Stringer to a different type T, which must be a string type.
func AsType[T ~string, F fmt.Stringer](from F) T {
	return T(from.String())
}

// AsPtrType converts a pointer to an enum-like value implementing fmt.Stringer to a pointer to a different type T, which must be a string type.
func AsPtrType[T ~string, F fmt.Stringer](from *F) *T {
	if from == nil {
		return nil
	}
	t := AsType[T](*from)
	return &t
}
