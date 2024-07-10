package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModelRef(t *testing.T) {
	t.Run("WellFormed", func(t *testing.T) {
		testMRef := ModelRef("foo::bar::baz")
		assert.EqualValues(t, "foo", testMRef.ProviderID())
		assert.EqualValues(t, "bar", testMRef.APIVersionID())
		assert.EqualValues(t, "baz", testMRef.ModelID())
	})

	// If given a malformed ModelRef, we want to fail loudly so it is
	// obvious what the problem is. (Returning an `error` would be even
	// better, but would probably just be too onerous.)
	t.Run("Malformed", func(t *testing.T) {
		oldStyleMRef := ModelRef("foo/baz")
		assert.EqualValues(t, "error-invalid-modelref", oldStyleMRef.ProviderID())
		assert.EqualValues(t, "error-invalid-modelref", oldStyleMRef.APIVersionID())
		assert.EqualValues(t, "error-invalid-modelref", oldStyleMRef.ModelID())
	})
}
