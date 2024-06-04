package errors

// This file contains tests which check the relationships between
// the various error-checking functions such as Is, As and HasType.

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// errorTree generates a tree of errors, where the leaves
// are generated using leafError.
func errorTree(leafError *rapid.Generator[error]) *rapid.Generator[error] {
	errorSliceGen := rapid.SliceOfN(rapid.Deferred(func() *rapid.Generator[error] {
		return errorTree(leafError)
	}), 1, 3)
	return rapid.OneOf(
		leafError,
		rapid.Map[[]error, error](errorSliceGen, func(errs []error) error {
			return Append(nil, errs...)
		}),
	)
}

func TestInvariants(t *testing.T) {
	// Check the behavior of the various error-checking
	// functions for different kinds of error types -
	// value receiver vs pointer receiver,
	// and errors without data vs with data.

	t.Run("payloadLessStruct", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			err := errorTree(rapid.OneOf(
				rapid.Just(error(&notTheErrorOfInterest{})),
				rapid.Just(error(payloadLessStructError{})),
			)).Draw(t, "err")
			// Is implies As for errors without data
			if Is(err, payloadLessStructError{}) {
				// This can be false, see Counter-example 1
				//require.True(t, HasType(err, payloadLessStructError{}))
				var check payloadLessStructError
				require.True(t, As(err, &check))
			}
			// HasType implies Is and As for errors without data
			if HasType(err, payloadLessStructError{}) {
				require.True(t, Is(err, payloadLessStructError{}))
				var check payloadLessStructError
				require.True(t, As(err, &check))
			}
			var check payloadLessStructError
			// As implies Is for errors without data
			if As(err, &check) {
				require.True(t, Is(err, payloadLessStructError{}))
				// This can be false, see Counter-example 2
				//require.True(t, HasType(err, payloadLessStructError{}))
			}
		})
	})

	t.Run("withPayloadStructError", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			errorOfInterest := withPayloadStructError{data: 11}
			errorWithOtherData := withPayloadStructError{data: 24}
			err := errorTree(rapid.OneOf(
				rapid.Just(error(&notTheErrorOfInterest{})),
				rapid.Just(error(errorOfInterest)),
				rapid.Just(error(errorWithOtherData)),
			)).Draw(t, "err")

			// Is implies As for errors with data
			if Is(err, errorOfInterest) {
				// This is false, see Counter-example 5
				//require.False(t, Is(err, errorWithOtherData))
				require.False(t, Is(err, withPayloadStructError{}))
				// These can be false, see Counter-example 1
				//require.True(t, HasType(err, errorOfInterest))
				//require.True(t, HasType(err, errorWithOtherData))
				//require.True(t, HasType(err, withPayloadStructError{}))
				var check withPayloadStructError
				require.True(t, As(err, &check))
				// This can be false, see Counter-example 6
				//require.Equal(t, errorOfInterest, check)
			}

			// HasType implies As for errors with data
			if HasType(err, errorOfInterest) {
				require.True(t, HasType(err, errorWithOtherData))
				require.True(t, HasType(err, withPayloadStructError{}))
				// This can be false, see Counter-example 3
				//require.True(t, Is(err, errorOfInterest))
				var check withPayloadStructError
				require.True(t, As(err, &check))
				// This can be false, see Counter-example 4
				//require.Equal(t, errorOfInterest, check)
			}

			// As implies a limited form of Is for errors with data
			var check withPayloadStructError
			if As(err, &check) {
				require.True(t, check == errorOfInterest || check == errorWithOtherData)
				require.True(t, Is(err, errorOfInterest) || Is(err, errorWithOtherData))
				// These can be false, see Counter-example 2
				//require.True(t, HasType(err, errorOfInterest))
				//require.True(t, HasType(err, errorWithOtherData))
				//require.True(t, HasType(err, withPayloadStructError{}))
			}
		})
	})

	t.Run("payloadLessPtrError", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			err := errorTree(rapid.OneOf(
				rapid.Just(error(&notTheErrorOfInterest{})),
				rapid.Just(error(&payloadLessPtrError{})),
			)).Draw(t, "err")
			// Is implies As for errors without data
			if Is(err, &payloadLessPtrError{}) {
				// This can be false, see Counter-example 1
				//require.True(t, HasType(err, &payloadLessPtrError{}))
				var check *payloadLessPtrError
				require.True(t, As(err, &check))
			}
			// HasType implies Is and As for errors without data
			if HasType(err, &payloadLessPtrError{}) {
				require.True(t, Is(err, &payloadLessPtrError{}))
				var check *payloadLessPtrError
				require.True(t, As(err, &check))
			}
			var check *payloadLessPtrError
			// As implies Is for errors without data
			if As(err, &check) {
				require.True(t, Is(err, &payloadLessPtrError{}))
				// This can be false, see Counter-example 2
				//require.True(t, errors.HasType(err, &payloadLessPtrError{}))
			}
		})
	})

	t.Run("withPayloadPtrError", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			errorOfInterest := &withPayloadPtrError{data: 11}
			errorWithOtherData := &withPayloadPtrError{data: 24}
			err := errorTree(rapid.OneOf(
				rapid.Just(error(&notTheErrorOfInterest{})),
				rapid.Just(error(errorOfInterest)),
				rapid.Just(error(errorWithOtherData)),
			)).Draw(t, "err")

			// Is implies As for errors with data
			if Is(err, errorOfInterest) {
				// This is false, see Counter-example 5
				//require.False(t, Is(err, errorWithOtherData))
				require.False(t, Is(err, &withPayloadPtrError{}))
				// These can be false, see Counter-example 1
				//require.True(t, HasType(err, errorOfInterest))
				//require.True(t, HasType(err, errorWithOtherData))
				//require.True(t, HasType(err, withPayloadStructError{}))
				var check *withPayloadPtrError
				require.True(t, As(err, &check))
				// This can be false, see Counter-example 6
				//require.Equal(t, *errorOfInterest, *check)
			}

			// HasType implies As for errors with data
			if HasType(err, errorOfInterest) {
				require.True(t, HasType(err, errorWithOtherData))
				require.True(t, HasType(err, &withPayloadPtrError{}))
				//This can be false, see Counter-example 3
				//require.True(t, Is(err, errorOfInterest))
				var check *withPayloadPtrError
				require.True(t, As(err, &check))
				require.True(t, *check == *errorOfInterest || *check == *errorWithOtherData)
			}

			// As implies a limited form of Is for errors with data
			var check *withPayloadPtrError
			if As(err, &check) {
				require.True(t, *check == *errorOfInterest || *check == *errorWithOtherData)
				require.True(t, Is(err, errorOfInterest) || Is(err, errorWithOtherData))
				// These can be false, see Counter-example 2
				//require.True(t, HasType(err, errorOfInterest))
				//require.True(t, HasType(err, errorWithOtherData))
			}
		})
	})

	t.Run("Counter-examples", func(t *testing.T) {
		// Counter-example 1. Is does not imply HasType
		{
			err := Append(payloadLessStructError{}, &notTheErrorOfInterest{})
			check := payloadLessStructError{}
			require.True(t, Is(err, check))
			require.False(t, HasType(err, check))
		}
		// Counter-example 2. As does not imply HasType
		{
			err := Append(payloadLessStructError{}, &notTheErrorOfInterest{})
			check := payloadLessStructError{}
			require.True(t, As(err, &check))
			require.False(t, HasType(err, payloadLessStructError{}))
		}
		// Counter-example 3. HasType does not imply Is
		{
			err := error(withPayloadStructError{data: 3})
			require.True(t, HasType(err, withPayloadStructError{}))
			require.False(t, Is(err, withPayloadStructError{data: 1}))
		}
		// Counter-example 4. HasType does not imply As
		{
			err := error(withPayloadStructError{data: 3})
			hasTypeCheck := withPayloadStructError{data: 1}
			require.True(t, HasType(err, hasTypeCheck))
			var valueFromAs withPayloadStructError
			require.True(t, As(err, &valueFromAs))
			require.NotEqual(t, hasTypeCheck, valueFromAs)
		}
		// Counter-example 5. Is can return true for distinct values
		{
			err := Append(withPayloadStructError{data: 3}, withPayloadStructError{data: 1})
			require.True(t, Is(err, withPayloadStructError{data: 3}))
			require.True(t, Is(err, withPayloadStructError{data: 1}))
		}
		// Counter-example 6. As can return a different value than the one passed to Is
		{
			err := Append(withPayloadStructError{data: 3}, withPayloadStructError{data: 1})
			var check withPayloadStructError
			require.True(t, Is(err, withPayloadStructError{data: 1}))
			require.True(t, As(err, &check))
			// 'As' picks the first value it finds.
			require.Equal(t, check, withPayloadStructError{data: 3})
			require.NotEqual(t, check, withPayloadStructError{data: 1})
		}
	})
}

type payloadLessStructError struct{}

func (p payloadLessStructError) Error() string {
	return "payloadLessStructError{}"
}

var _ error = payloadLessStructError{}

type withPayloadStructError struct {
	data int
}

var _ error = withPayloadStructError{}

func (p withPayloadStructError) Error() string {
	return fmt.Sprintf("withPayloadStructError{data: %v}", p.data)
}

type payloadLessPtrError struct{}

var _ error = &payloadLessPtrError{}

func (p *payloadLessPtrError) Error() string {
	return "payloadLessPtrError{}"
}

type withPayloadPtrError struct {
	data int
}

var _ error = &withPayloadPtrError{}

func (p *withPayloadPtrError) Error() string {
	return fmt.Sprintf("withPayloadPtrError{data: %d}", p.data)
}

type notTheErrorOfInterest struct{}

var _ error = &notTheErrorOfInterest{}

func (p *notTheErrorOfInterest) Error() string {
	return "notTheErrorOfInterest{}"
}

func TestAsInterface(t *testing.T) {
	require.Panics(t, func() {
		p := &payloadLessPtrError{}
		err := error(&payloadLessPtrError{})
		AsInterface(err, &p)
	})
	var e error
	require.True(t, AsInterface(error(&payloadLessPtrError{}), &e))
}
