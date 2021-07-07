// Package errcode maps Go errors to HTTP status codes as well as other useful
// functions for inspecting errors.
package errcode

import "errors"

// IsNotFound will check if err or one of its causes is a not found
// error. Note: This will not check os.IsNotExist, but rather is returned by
// methods like Repo.Get when the repo is not found. It will also *not* map
// HTTPStatusCode into not found.
func IsNotFound(err error) bool {
	var e interface{ NotFound() bool }
	return errors.As(err, &e) && e.NotFound()
}

// IsUnauthorized will check if err or one of its causes is an unauthorized
// error.
func IsUnauthorized(err error) bool {
	var e interface{ Unauthorized() bool }
	return errors.As(err, &e) && e.Unauthorized()
}

// IsForbidden will check if err or one of its causes is a forbidden error.
func IsForbidden(err error) bool {
	var e interface{ Forbidden() bool }
	return errors.As(err, &e) && e.Forbidden()
}

// IsAccountSuspended will check if err or one of its causes was due to the
// account being suspended
func IsAccountSuspended(err error) bool {
	var e interface{ AccountSuspended() bool }
	return errors.As(err, &e) && e.AccountSuspended()
}

// IsBadRequest will check if err or one of its causes is a bad request.
func IsBadRequest(err error) bool {
	var e interface{ BadRequest() bool }
	return errors.As(err, &e) && e.BadRequest()
}

// IsTemporary will check if err or one of its causes is temporary. A
// temporary error can be retried. Many errors in the go stdlib implement the
// temporary interface.
func IsTemporary(err error) bool {
	var e interface{ Temporary() bool }
	return errors.As(err, &e) && e.Temporary()
}

// IsBlocked will check if err or one of its causes is a blocked error.
func IsBlocked(err error) bool {
	var e interface{ Blocked() bool }
	return errors.As(err, &e) && e.Blocked()
}

// IsTimeout will check if err or one of its causes is a timeout. Many errors
// in the go stdlib implement the timeout interface.
func IsTimeout(err error) bool {
	var e interface{ Timeout() bool }
	return errors.As(err, &e) && e.Timeout()
}
