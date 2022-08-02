package errors

// Typed is an interface custom error types that want to be checkable with errors.Is,
// errors.As can implement for more predicable behaviour. Learn more about error checking:
// https://pkg.go.dev/errors#pkg-overview
//
// In all implementations, the error should not attempt to unwrap itself or the target.
type Typed interface {
	// As sets the target to the error value of this type if target is of the same type as
	// this error.
	//
	// See https://pkg.go.dev/errors#example-As
	As(target any) bool
	// Is reports whether this error matches the target.
	//
	// See: https://pkg.go.dev/errors#example-Is
	Is(target error) bool
}

// Wrapper is an interface custom error types that carry errors internally should
// implement.
type Wrapper interface {
	Unwrap() error
}
