package errors

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Enforce some invariants with our error libraries.

func TestMultipleErrorPrinting(t *testing.T) {
	// Make sure all our ways of combining errors actually print them.

	errFoo := New("foo")
	errBar := New("bar")

	for fn, str := range map[string]string{
		"Append":            Append(errFoo, errBar).Error(),
		"Append Sprintf %s": fmt.Sprintf("%s", Append(errFoo, errBar)),
		"Append Sprintf %v": fmt.Sprintf("%v", Append(errFoo, errBar)),

		"CombineErrors":            CombineErrors(errFoo, errBar).Error(),
		"CombineErrors Sprintf %s": fmt.Sprintf("%s", CombineErrors(errFoo, errBar)),
		"CombineErrors Sprintf %v": fmt.Sprintf("%v", CombineErrors(errFoo, errBar)),

		"Wrap Append":            Wrap(Append(errFoo, errBar), "hello world").Error(),
		"Wrap Append Sprintf %s": fmt.Sprintf("%s", Wrap(Append(errFoo, errBar), "hello world")),
		"Wrap Append Sprintf %v": fmt.Sprintf("%v", Wrap(Append(errFoo, errBar), "hello world")),

		"Fancy stack %+v": fmt.Sprintf("%+v", Wrap(Wrap(Append(errFoo, errBar), "hello world"), "deep!")),
	} {
		t.Run(fn, func(t *testing.T) {
			t.Log(str)
			assert.Contains(t, str, "foo", fn)
			assert.Contains(t, str, "bar", fn)
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
