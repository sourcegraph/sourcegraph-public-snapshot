package errors

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func errorTree(baseError error) *rapid.Generator[error] {
	errorSliceGen := rapid.SliceOfN(rapid.Deferred(func() *rapid.Generator[error] {
		return errorTree(baseError)
	}), 1, 3)
	return rapid.OneOf(
		rapid.Just(error(&notTheErrorOfInterest{})),
		rapid.Just(baseError),
		rapid.Map[[]error, error](errorSliceGen, func(errs []error) error {
			return Append(nil, errs...)
		}),
	)
}

func TestInvariants(t *testing.T) {
	t.Run("payloadLessStruct", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			err := errorTree(payloadLessStructError{}).Draw(t, "err")
			if Is(err, payloadLessStructError{}) {
				// This can be false, see Counter-example 1
				//require.True(t, HasType(err, payloadLessStructError{}))
				var check payloadLessStructError
				require.True(t, As(err, &check))
			}
			if HasType(err, payloadLessStructError{}) {
				require.True(t, Is(err, payloadLessStructError{}))
				var check payloadLessStructError
				require.True(t, As(err, &check))
			}
			var check payloadLessStructError
			if As(err, &check) {
				require.True(t, Is(err, payloadLessStructError{}))
				// This can be false, see Counter-example 2
				//require.True(t, errors.HasType(err, payloadLessStructError{}))
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
