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
			// Is implies HasType and As for errors without data
			if Is(err, payloadLessStructError{}) {
				require.True(t, HasType[payloadLessStructError](err))
				var check payloadLessStructError
				require.True(t, As(err, &check))
			}
			// HasType implies Is and As for errors without data
			if HasType[payloadLessStructError](err) {
				require.True(t, Is(err, payloadLessStructError{}))
				var check payloadLessStructError
				require.True(t, As(err, &check))
			}
			var check payloadLessStructError
			// As implies Is and HasType for errors without data
			if As(err, &check) {
				require.True(t, Is(err, payloadLessStructError{}))
				require.True(t, HasType[payloadLessStructError](err))
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

			// Is implies HasType and As for errors with data
			if Is(err, errorOfInterest) {
				// This is false, see Counter-example 2
				//require.False(t, Is(err, errorWithOtherData))
				require.False(t, Is(err, withPayloadStructError{}))
				require.True(t, HasType[withPayloadStructError](err))
				var check withPayloadStructError
				require.True(t, As(err, &check))
				// This can be false, see Counter-example 3
				//require.Equal(t, errorOfInterest, check)
			}

			// HasType implies As for errors with data
			if HasType[withPayloadStructError](err) {
				// This can be false, see Counter-example 1
				//require.True(t, Is(err, errorOfInterest))
				var check withPayloadStructError
				require.True(t, As(err, &check))
			}

			// As implies a limited form of Is for errors with data
			var check withPayloadStructError
			if As(err, &check) {
				require.True(t, check == errorOfInterest || check == errorWithOtherData)
				require.True(t, Is(err, errorOfInterest) || Is(err, errorWithOtherData))
				require.True(t, HasType[withPayloadStructError](err))
			}
		})
	})

	t.Run("payloadLessPtrError", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			err := errorTree(rapid.OneOf(
				rapid.Just(error(&notTheErrorOfInterest{})),
				rapid.Just(error(&payloadLessPtrError{})),
			)).Draw(t, "err")
			// Is implies HasType and As for errors without data
			if Is(err, &payloadLessPtrError{}) {
				require.True(t, HasType[*payloadLessPtrError](err))
				var check *payloadLessPtrError
				require.True(t, As(err, &check))
			}
			// HasType implies Is and As for errors without data
			if HasType[*payloadLessPtrError](err) {
				require.True(t, Is(err, &payloadLessPtrError{}))
				var check *payloadLessPtrError
				require.True(t, As(err, &check))
			}
			var check *payloadLessPtrError
			// As implies Is and HasType for errors without data
			if As(err, &check) {
				require.True(t, Is(err, &payloadLessPtrError{}))
				require.True(t, HasType[*payloadLessPtrError](err))
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

			// Is implies HasType and As for errors with data
			if Is(err, errorOfInterest) {
				// This is false, see Counter-example 2
				//require.False(t, Is(err, errorWithOtherData))
				require.False(t, Is(err, &withPayloadPtrError{}))
				require.True(t, HasType[*withPayloadPtrError](err))
				var check *withPayloadPtrError
				require.True(t, As(err, &check))
				// This can be false, see Counter-example 3
				//require.Equal(t, *errorOfInterest, *check)
			}

			// HasType implies As for errors with data
			if HasType[*withPayloadPtrError](err) {
				//This can be false, see Counter-example 1
				//require.True(t, Is(err, errorOfInterest))
				var check *withPayloadPtrError
				require.True(t, As(err, &check))
				require.True(t, *check == *errorOfInterest || *check == *errorWithOtherData)
			}

			// As implies HasType and a limited form of Is for errors with data
			var check *withPayloadPtrError
			if As(err, &check) {
				require.True(t, *check == *errorOfInterest || *check == *errorWithOtherData)
				require.True(t, Is(err, errorOfInterest) || Is(err, errorWithOtherData))
				require.True(t, HasType[*withPayloadPtrError](err))
			}
		})
	})

	t.Run("Counter-examples", func(t *testing.T) {
		// Counter-example 1. HasType does not imply Is
		{
			err := error(withPayloadStructError{data: 3})
			require.True(t, HasType[withPayloadStructError](err))
			require.False(t, Is(err, withPayloadStructError{data: 1}))
		}
		// Counter-example 2. Is can return true for distinct values
		{
			err := Append(withPayloadStructError{data: 3}, withPayloadStructError{data: 1})
			require.True(t, Is(err, withPayloadStructError{data: 3}))
			require.True(t, Is(err, withPayloadStructError{data: 1}))
		}
		// Counter-example 3. As can return a different value than the one passed to Is
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
