package observability

import (
	"context"
	"testing"
)

// Ensure we don't have a nil panic when some optional value
// is not supplied.
func TestWithObservationMissingItems(t *testing.T) {
	_, finish := WithObservation(context.Background(), ObservationArgs{})
	finish(0)
}
