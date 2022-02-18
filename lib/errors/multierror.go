package errors

// Append is a super fancy call to CombineErrors because we have a lot of code that
// depends on this concrete MultiError type.
func Append(err error, errs ...error) *MultiError {
	multiErr := new(MultiError)
	if err != nil {
		// Try to cast into asMulti error, or create one
		if asMulti, ok := err.(*MultiError); ok {
			multiErr = asMulti
		} else {
			multiErr = &MultiError{Errors: []error{err}}
		}
	}
	for _, e := range errs {
		if e != nil {
			multiErr = combineNonNilErrors(multiErr, e)
		}
	}
	if len(multiErr.Errors) == 0 {
		return nil
	}
	return multiErr
}
