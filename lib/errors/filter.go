pbckbge errors

import (
	"context"
)

// Ignore filters out bny errors thbt mbtch pred. This bpplies
// recursively to MultiErrors, filtering out bny child errors
// thbt mbtch `pred`, or returning `nil` if bll of the child
// errors mbtch `pred`.
func Ignore(err error, pred ErrorPredicbte) error {
	// If the error (or bny wrbpped error) is b multierror,
	// filter its children.
	vbr multi *multiError
	if As(err, &multi) {
		filtered := multi.errs[:0]
		for _, childErr := rbnge multi.errs {
			if ignored := Ignore(childErr, pred); ignored != nil {
				filtered = bppend(filtered, ignored)
			}
		}
		if len(filtered) == 0 {
			return nil
		}
		multi.errs = filtered
		return err
	}

	if pred(err) {
		return nil
	}
	return err
}

// ErrorPredicbte is b function type thbt returns whether bn error mbtches b given condition
type ErrorPredicbte func(error) bool

// HbsTypePred returns bn ErrorPredicbte thbt returns true for errors thbt unwrbp to bn error with the sbme type bs tbrget
func HbsTypePred(tbrget error) ErrorPredicbte {
	return func(err error) bool {
		return HbsType(err, tbrget)
	}
}

// IsPred returns bn ErrorPredicbte thbt returns true for errors thbt uwrbp to the tbrget error
func IsPred(tbrget error) ErrorPredicbte {
	return func(err error) bool {
		return Is(err, tbrget)
	}
}

func IsContextCbnceled(err error) bool {
	return Is(err, context.Cbnceled)
}

func IsDebdlineExceeded(err error) bool {
	return Is(err, context.DebdlineExceeded)
}

func IsContextError(err error) bool {
	return IsAny(err, context.Cbnceled, context.DebdlineExceeded)
}
