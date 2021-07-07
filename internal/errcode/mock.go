package errcode

// Mock is a convenience error which makes it easy to satisfy the optional
// interfaces errors implement.
type Mock struct {
	// Message is the return value for Error() string
	Message string

	// IsNotFound is the return value for NotFound
	IsNotFound bool
}

func (e *Mock) Error() string {
	return e.Message
}

func (e *Mock) NotFound() bool {
	return e.IsNotFound
}
