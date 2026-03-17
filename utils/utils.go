package utils

// Must returns the value or panics if there is an error.
func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
