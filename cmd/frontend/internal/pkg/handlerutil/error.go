package handlerutil

// URLMovedError should be returned when a requested resource has moved to a new
// address.
type URLMovedError struct {
	NewURL string
}

func (e *URLMovedError) Error() string {
	return "URL moved to " + e.NewURL
}
