package filter

import (
	"fmt"
	"github.com/scim2/filter-parser/v2/internal/types"
)

func invalidChildTypeError(parentTyp, invalidType int) error {
	return &internalError{
		Message: fmt.Sprintf(
			"invalid child type for %s (%03d): %s (%03d)",
			typ.Stringer[parentTyp], parentTyp,
			typ.Stringer[invalidType], invalidType,
		),
	}
}

func invalidLengthError(parentTyp int, len, actual int) error {
	return &internalError{
		Message: fmt.Sprintf(
			"length was not equal to %d for %s (%03d), got %d elements",
			len, typ.Stringer[parentTyp], parentTyp, actual,
		),
	}
}

func invalidTypeError(expected, actual int) error {
	return &internalError{
		Message: fmt.Sprintf(
			"invalid type: expected %s (%03d), actual %s (%03d)",
			typ.Stringer[expected], expected,
			typ.Stringer[actual], actual,
		),
	}
}

func missingValueError(parentTyp int, valueType int) error {
	return &internalError{
		Message: fmt.Sprintf(
			"missing a required value for %s (%03d): %s (%03d)",
			typ.Stringer[parentTyp], parentTyp,
			typ.Stringer[valueType], valueType,
		),
	}
}

// internalError represents an internal error. If this error should NEVER occur.
// If you get this error, please open an issue!
type internalError struct {
	Message string
}

func (e *internalError) Error() string {
	return fmt.Sprintf("internal error: %s", e.Message)
}
