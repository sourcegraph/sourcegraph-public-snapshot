package errors

// CombineErrors replaces what used to be usages of cockroachdb/errors.CombineErrors to
// unify multiple-error types with go-multierror.
func CombineErrors(err, other error) error {
	if err == nil {
		return other
	}
	return Append(err, other)
}
