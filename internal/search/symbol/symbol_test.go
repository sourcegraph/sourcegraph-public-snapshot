package symbol

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/stretchr/testify/assert"
)

func TestSearchZoektDoesntPanicWithNilQuery(t *testing.T) {
	// As soon as we reach Streamer.Search function, we can consider test successful,
	// that's why we can just mock it.
	mockStreamer := NewMockStreamer()
	expectedErr := errors.New("short circuit")
	mockStreamer.SearchFunc.SetDefaultReturn(nil, expectedErr)
	search.IndexedMock = mockStreamer
	t.Cleanup(func() {
		search.IndexedMock = nil
	})

	_, err := searchZoekt(context.Background(), types.MinimalRepo{ID: 1}, "commitID", nil, "branch", nil, nil, nil)
	assert.ErrorIs(t, err, expectedErr)
}
