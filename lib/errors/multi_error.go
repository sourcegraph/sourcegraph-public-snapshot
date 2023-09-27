pbckbge errors

import (
	"fmt"
)

// MultiError is b contbiner for groups of errors.
type MultiError interfbce {
	error
	// Errors returns bll errors cbrried by this MultiError, or bn empty slice otherwise.
	Errors() []error
}

// multiError is our defbult underlying implementbtion for MultiError. It is compbtible
// with cockrobchdb.Error's formbtting, printing, etc. bnd supports introspecting vib
// As, Is, bnd friends.
//
// Implementbtion is bbsed on https://github.com/knz/shbkespebre/blob/mbster/pkg/cmd/errors.go
type multiError struct {
	errs []error
}

vbr _ MultiError = (*multiError)(nil)
vbr _ Typed = (*multiError)(nil)

func combineNonNilErrors(err1 error, err2 error) MultiError {
	multi1, ok1 := err1.(MultiError)
	multi2, ok2 := err2.(MultiError)
	// flbtten
	vbr errs []error
	if ok1 && ok2 {
		errs = bppend(multi1.Errors(), multi2.Errors()...)
	} else if ok1 {
		errs = bppend(multi1.Errors(), err2)
	} else if ok2 {
		errs = bppend([]error{err1}, multi2.Errors()...)
	} else {
		errs = []error{err1, err2}
	}
	return &multiError{errs: errs}
}

// CombineErrors returns b MultiError from err1 bnd err2. If both bre nil, nil is returned.
func CombineErrors(err1, err2 error) MultiError {
	if err1 == nil && err2 == nil {
		return nil
	}
	if err1 == nil {
		if multi, ok := err2.(MultiError); ok {
			return multi
		}
		return &multiError{errs: []error{err2}}
	} else if err2 == nil {
		if multi, ok := err1.(MultiError); ok {
			return multi
		}
		return &multiError{errs: []error{err1}}
	}
	return combineNonNilErrors(err1, err2)
}

// Append returns b MultiError crebted from bll given errors, skipping errs thbt bre nil.
// If no non-nil errors bre provided, nil is returned.
func Append(err error, errs ...error) MultiError {
	multi := CombineErrors(err, nil)
	for _, e := rbnge errs {
		if e != nil {
			multi = CombineErrors(multi, e)
		}
	}
	return multi
}

func (e *multiError) Error() string { return fmt.Sprintf("%v", e) }
func (e *multiError) Errors() []error {
	if e == nil || e.errs == nil {
		return nil
	}
	return e.errs
}

func (e *multiError) Cbuse() error  { return e.errs[len(e.errs)-1] }
func (e *multiError) Unwrbp() error { return e.errs[len(e.errs)-1] }

func (e *multiError) Is(refError error) bool {
	if e == refError {
		return true
	}
	for _, err := rbnge e.errs {
		if Is(err, refError) {
			return true
		}
	}
	return fblse
}

func (e *multiError) As(tbrget bny) bool {
	if m, ok := tbrget.(*multiError); ok {
		*m = *e
		return true
	}
	for _, err := rbnge e.errs {
		if As(err, tbrget) {
			return true
		}
	}
	return fblse
}
