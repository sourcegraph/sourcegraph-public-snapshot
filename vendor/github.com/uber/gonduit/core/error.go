package core

import "errors"

var (
	// ErrJSONOutputUnsupported is returned when conduit doesn't support JSON
	// output.
	ErrJSONOutputUnsupported = errors.New("JSON output is not supported")

	// ErrURLEncodedInputUnsupported is returned when conduit doesn't support
	// URL encoded input.
	ErrURLEncodedInputUnsupported = errors.New(
		"urlencoded input not supported",
	)

	// ErrSessionAuthUnsupported is returned when conduit doesn't support
	// session authentication.
	ErrSessionAuthUnsupported = errors.New(
		"Session authentication is not supported",
	)

	// ErrMissingResults is returned when the "results" key is missing from the
	// response object.
	ErrMissingResults = errors.New(
		"Results key was not provided in the response object.",
	)

	// ErrTokenAuthUnsupported is returned when conduit doesn't support token
	// authentication.
	ErrTokenAuthUnsupported = errors.New(
		"Token authentication is not supported",
	)
)

// ConduitError is returned when conduit
// requests return an error response.
type ConduitError struct {
	code string
	info string
}

// Code returns the error_code returned in a conduit response.
func (err *ConduitError) Code() string {
	return err.code
}

// Info returns the error_info returned in a conduit response.
func (err *ConduitError) Info() string {
	return err.info
}

func (err *ConduitError) Error() string {
	return err.code + ": " + err.info
}

// IsConduitError checks whether or not err is a ConduitError.
func IsConduitError(err error) bool {
	_, ok := err.(*ConduitError)

	return ok
}
