package errors

import (
	"fmt"
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
