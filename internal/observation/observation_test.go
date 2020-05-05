package observation

import (
	"context"
	"testing"
)

// Ensure we don't have a nil panic when some optional value is not supplied.
func TestWithMissingItems(t *testing.T) {
	_, finish := With(context.Background(), Args{})
	finish(0)
}
