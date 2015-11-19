package sourcegraph

import "net/http"

// InvalidOptionsError indicates that the provided XxxOptions
// (RepoListOptions, DefGetOptions, etc.) was invalid.
type InvalidOptionsError struct{ Reason string }

func (e *InvalidOptionsError) Error() string { return e.Reason }

func (e *InvalidOptionsError) HTTPStatusCode() int { return http.StatusBadRequest }

// InvalidSpecError indicates that the provided XxxSpec (RepoSpec,
// UserSpec, etc.) was invalid.
type InvalidSpecError struct{ Reason string }

func (e *InvalidSpecError) Error() string { return e.Reason }

func (e *InvalidSpecError) HTTPStatusCode() int { return http.StatusBadRequest }

// NotImplementedError indicates that a not-yet-implemented method was
// called.
type NotImplementedError struct{ What string }

func (e *NotImplementedError) Error() string { return e.What + " is not implemented" }

func (e *NotImplementedError) HTTPStatusCode() int { return http.StatusNotFound }
