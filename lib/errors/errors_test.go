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
		"Append.Error":          Append(errFoo, errBar).Error(),
		"Append.Sprintf":        fmt.Sprintf("%s", Append(errFoo, errBar)),
		"CombineErrors.Error":   CombineErrors(errFoo, errBar).Error(),
		"CombineErrors.Sprintf": fmt.Sprintf("%s", CombineErrors(errFoo, errBar)),
	} {
		assert.Contains(t, str, "foo", fn)
		assert.Contains(t, str, "bar", fn)
	}
}
