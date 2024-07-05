package database

// resourceNotFoundError is an error that indicates that a database resource was not found. It can be
// returned by methods that get a single resource (such as Get or GetByXyz).
//
// errcode.IsNotFound(err) == true for notFoundError values.
type resourceNotFoundError struct {
	noun string
}

func (e resourceNotFoundError) Error() string {
	const notFound = "not found"
	if e.noun == "" {
		return notFound
	}
	return e.noun + " " + notFound
}

func (resourceNotFoundError) NotFound() bool { return true }
