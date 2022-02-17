package errors

import (
	"fmt"
	"strings"
)

func Append(root error, errs ...error) error {
	chain := make(errChain, 0, len(errs)+1)
	if root != nil {
		chain = append(chain, root)
	}
	for _, err := range errs {
		if err != nil {
			chain = append(chain, err)
		}
	}
	if len(chain) == 0 {
		return nil
	}
	if len(chain) == 1 {
		return chain[0]
	}
	return chain
}

func CombineErrors(err error, other error) error {
	if err == nil {
		return other
	}
	return Append(err, other)
}

type errChain []error

// Error implements the error interface
func (e errChain) Error() string {
	if len(e) == 1 {
		return e[0].Error()
	}
	points := make([]string, len(e))
	for i, err := range e {
		points[i] = fmt.Sprintf("* %s", err)
	}
	return fmt.Sprintf(
		"%d errors occurred:\n\t%s\n\n",
		len(e), strings.Join(points, "\n\t"))
}

// Unwrap implements errors.Unwrap by returning the next error in the
// chain or nil if there are no more errors.
func (e errChain) Unwrap() error {
	if len(e) == 1 {
		return nil
	}
	return e[1:]
}

// As implements errors.As by attempting to map to the current value.
func (e errChain) As(target interface{}) bool {
	return As(e[0], target)
}

// Is implements errors.Is by comparing the current value directly.
func (e errChain) Is(target error) bool {
	return Is(e[0], target)
}
