package legacyerr

import "fmt"

// LEGACY! Use proper Go errors instead.
type Error struct {
	Code Code
	Desc string
}

func (e Error) Error() string {
	return fmt.Sprintf("error: code = %d desc = %s", e.Code, e.Desc)
}

// Errorf returns an error containing an error code and a description;
// Errorf returns nil if c is OK.
// LEGACY! Use proper Go errors instead.
func Errorf(c Code, format string, a ...interface{}) error {
	return Error{
		Code: c,
		Desc: fmt.Sprintf(format, a...),
	}
}

// ErrCode returns the error code for err if it is an Error.
// Otherwise, it returns codes.Unknown.
func ErrCode(err error) Code {
	if e, ok := err.(Error); ok {
		return e.Code
	}
	return Unknown
}

// ErrorDesc returns the error description of err if it is an Error.
// Otherwise, it returns err.Error() or empty string when err is nil.
func ErrorDesc(err error) string {
	if err == nil {
		return ""
	}
	if e, ok := err.(Error); ok {
		return e.Desc
	}
	return err.Error()
}

// LEGACY! Use proper Go errors instead.
type Code int

const (
	Unknown Code = iota
	InvalidArgument
	NotFound
	AlreadyExists
	PermissionDenied
	Unauthenticated
	ResourceExhausted
	FailedPrecondition
	Unimplemented
	Internal
	Unavailable
)
