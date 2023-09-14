package errors

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

type errBazType struct{}

func (e *errBazType) Error() string { return "baz" }

type errZooType struct{}

func (e *errZooType) Error() string { return "zoo" }

// Enforce some invariants with our error libraries.

func TestMultiError(t *testing.T) {
	errFoo := New("foo")
	errBar := New("bar")
	// Tests using errBaz also make a As test
	errBaz := &errBazType{}
	// Tests using errZoo also make a As test
	errZoo := &errZooType{}
	formattingDirectives := []string{"", "%s", "%v", "%+v"}
	tests := []struct {
		name string
		err  error
		// Make sure all our ways of combining errors actually print them.
		wantStrings []string
		// Make sure all our ways of combining errors retains our ability to assert
		// against them.
		wantIs []error
	}{
		{
			name:        "Append",
			err:         Append(errFoo, errBar),
			wantStrings: []string{"foo", "bar"},
			wantIs:      []error{errFoo, errBar},
		},
		{
			name:        "CombineErrors",
			err:         CombineErrors(errFoo, errZoo),
			wantStrings: []string{"foo", "zoo"},
			wantIs:      []error{errFoo, errZoo},
		},
		{
			name:        "Wrap(Append)",
			err:         Wrap(Append(errFoo, errBar), "hello world"),
			wantStrings: []string{"hello world", "foo", "bar"},
			wantIs:      []error{errFoo, errBar},
		},
		{
			name:        "Wrap(Wrap(Append))",
			err:         Wrap(Wrap(Append(errFoo, errZoo), "hello world"), "deep!"),
			wantStrings: []string{"deep", "hello world", "foo", "zoo"},
			wantIs:      []error{errFoo, errZoo},
		},
		{
			name:        "Append(Wrap(Append))",
			err:         Append(Wrap(Append(errFoo, errBar), "hello world"), errZoo),
			wantStrings: []string{"hello world", "foo", "bar"},
			wantIs:      []error{errFoo, errBar, errZoo},
		},
		{
			name:        "Append(Append(Append(Append)))",
			err:         Append(Append(Append(errFoo, errBar), errBaz), errZoo),
			wantStrings: []string{"zoo", "baz", "foo", "bar"},
			wantIs:      []error{errFoo, errBar, errBaz, errZoo},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, directive := range formattingDirectives {
				var str string
				if directive == "" {
					str = tt.err.Error()
				} else {
					str = fmt.Sprintf(directive, tt.err)
				}

				if directive == "" || directive == "%+v" {
					// Run tests with -v to see what the error output looks like
					t.Log(str)
				}

				for _, contains := range tt.wantStrings {
					assert.Contains(t, str, contains)
				}
			}
			for _, isErr := range tt.wantIs {
				assert.ErrorIs(t, tt.err, isErr)
				if isErr.Error() == "baz" {
					var errBaz *errBazType
					assert.ErrorAs(t, tt.err, &errBaz, "Want "+isErr.Error())
				}
				if isErr.Error() == "zoo" {
					var errZoo *errZooType
					assert.ErrorAs(t, tt.err, &errZoo, "Want "+isErr.Error())
				}
			}
			// We always want to be able to extract a MultiError from this error, because
			// all the test cases test appends. We don't assert against its contents, but
			// to see how we unwrap errors you can add:
			//
			//   t.Log("Extracted multi-error:\n", multi.Error())
			//
			var multi MultiError
			assert.ErrorAs(t, tt.err, &multi)
		})
	}
}

func TestCombineNil(t *testing.T) {
	assert.Nil(t, Append(nil, nil))
	assert.Nil(t, CombineErrors(nil, nil))
}

func TestCombineSingle(t *testing.T) {
	errFoo := New("foo")

	assert.ErrorIs(t, Append(errFoo, nil), errFoo)
	assert.ErrorIs(t, CombineErrors(errFoo, nil), errFoo)
	assert.ErrorIs(t, Append(nil, errFoo), errFoo)
	assert.ErrorIs(t, CombineErrors(nil, errFoo), errFoo)
}

// TestRepeatedCombine tests the most common patterns of instantiate + append
func TestRepeatedCombine(t *testing.T) {
	t.Run("mixed append with typed nil", func(t *testing.T) {
		var errs MultiError
		for i := 1; i < 10; i++ {
			if i%2 == 0 {
				errs = Append(errs, New(strconv.Itoa(i)))
			} else {
				errs = Append(errs, nil)
			}
		}
		assert.NotNil(t, errs)
		assert.Equal(t, 4, len(errs.Errors()))
		assert.Equal(t, `4 errors occurred:
	* 2
	* 4
	* 6
	* 8`, errs.Error())
	})
	t.Run("mixed append with untyped nil", func(t *testing.T) {
		var errs error
		for i := 1; i < 10; i++ {
			if i%2 == 0 {
				errs = Append(errs, New(strconv.Itoa(i)))
			} else {
				errs = Append(errs, nil)
			}
		}
		assert.NotNil(t, errs)
		assert.Equal(t, `4 errors occurred:
	* 2
	* 4
	* 6
	* 8`, errs.Error())
		// try casting
		var multi MultiError
		assert.True(t, As(errs, &multi))
		assert.Equal(t, 4, len(multi.Errors()))
	})
	t.Run("all nil append with typed nil", func(t *testing.T) {
		var errs MultiError
		for i := 1; i < 10; i++ {
			errs = Append(errs, nil)
		}
		assert.Nil(t, errs)
	})
	t.Run("all nil append with untyped nil", func(t *testing.T) {
		var errs error
		for i := 1; i < 10; i++ {
			errs = Append(errs, nil)
		}
		assert.Nil(t, errs)
		// try casting
		var multi MultiError
		assert.False(t, As(errs, &multi))
	})
}

func TestNotRedacted(t *testing.T) {
	err := Newf("foo: %s", "bar")
	assert.Equal(t, "foo: bar", err.Error())
}
