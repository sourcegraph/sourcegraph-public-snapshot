package resolvers

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBatchSpecWorkspaceOutputLinesResolver(t *testing.T) {
	var lines = make([]string, 100)
	for i := range lines {
		lines[i] = fmt.Sprintf("Hello world: %d", i+1)
	}
	totalCount := int32(len(lines))

	t.Run("with paginated output lines", func(t *testing.T) {
		noOfLines := 50
		resolver := &batchSpecWorkspaceOutputLinesResolver{
			lines: lines,
			first: int32(noOfLines),
		}

		tc, err := resolver.TotalCount()
		assert.NoError(t, err)
		assert.Equal(t, tc, totalCount)

		nodes, err := resolver.Nodes()
		assert.NoError(t, err)
		assert.Len(t, nodes, 50)

		pi, err := resolver.PageInfo()
		assert.NoError(t, err)
		assert.Equal(t, pi.HasNextPage(), true)
		assert.Equal(t, *pi.EndCursor(), "50")
	})

	t.Run("cursor used to access paginated lines", func(t *testing.T) {
		noOfLines := 50
		endCursor := "50"

		resolver := &batchSpecWorkspaceOutputLinesResolver{
			lines: lines,
			first: int32(noOfLines),
			after: &endCursor,
		}

		tc, err := resolver.TotalCount()
		assert.NoError(t, err)
		assert.Equal(t, tc, totalCount)

		nodes, err := resolver.Nodes()
		assert.NoError(t, err)
		assert.Len(t, nodes, 50)

		pi, err := resolver.PageInfo()
		assert.NoError(t, err)
		assert.Equal(t, pi.HasNextPage(), false)
		if pi.EndCursor() != nil {
			t.Fatal("expected cursor to be nil")
		}
	})

	t.Run("offset greater than length of lines", func(t *testing.T) {
		noOfLines := 150
		endCursor := "50"

		resolver := &batchSpecWorkspaceOutputLinesResolver{
			lines: lines,
			first: int32(noOfLines),
			after: &endCursor,
		}

		tc, err := resolver.TotalCount()
		assert.NoError(t, err)
		assert.Equal(t, tc, totalCount)

		nodes, err := resolver.Nodes()
		assert.NoError(t, err)
		assert.Len(t, nodes, 50)

		pi, err := resolver.PageInfo()
		assert.NoError(t, err)
		assert.Equal(t, pi.HasNextPage(), false)
		if pi.EndCursor() != nil {
			t.Fatal("expected cursor to be nil")
		}
	})
}
