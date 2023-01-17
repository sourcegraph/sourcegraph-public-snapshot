package resolvers

import (
	"context"
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
	ctx := context.Background()

	t.Run("with paginated output lines", func(t *testing.T) {
		noOfLines := 50
		resolver := &batchSpecWorkspaceOutputLinesResolver{
			lines: lines,
			first: int32(noOfLines),
		}

		assert.Equal(t, resolver.TotalCount(ctx), totalCount)
		assert.Len(t, resolver.Nodes(ctx), 50)

		pi := resolver.PageInfo(ctx)
		assert.Equal(t, pi.HasNextPage(), true)
		assert.Equal(t, *pi.EndCursor(), "50")
	})

	t.Run("cursor used to access paginated lines", func(t *testing.T) {
		noOfLines := 50
		endCursor := int32(50)

		resolver := &batchSpecWorkspaceOutputLinesResolver{
			lines: lines,
			first: int32(noOfLines),
			after: &endCursor,
		}

		assert.Equal(t, resolver.TotalCount(ctx), totalCount)
		assert.Len(t, resolver.Nodes(ctx), 50)

		pi := resolver.PageInfo(ctx)
		assert.Equal(t, pi.HasNextPage(), false)
		if pi.EndCursor() != nil {
			t.Fatal("expected cursor to be nil")
		}
	})

}
