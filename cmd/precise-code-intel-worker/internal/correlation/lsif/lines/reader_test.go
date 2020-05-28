package lines

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-worker/internal/correlation/lsif"
)

func TestRead(t *testing.T) {
	var content []byte
	var expectedIDs []string
	for i := 0; i < 10000; i++ {
		id := fmt.Sprintf("%d", i)
		content = append(content, id...)
		content = append(content, '\n')
		expectedIDs = append(expectedIDs, id)
	}

	pairs := Read(context.Background(), bytes.NewReader(content), func(line []byte) (lsif.Element, error) {
		return lsif.Element{ID: string(line)}, nil
	})

	var ids []string
	for pair := range pairs {
		if pair.Err != nil {
			t.Fatalf("unexpected error: %s", pair.Err)
		}

		ids = append(ids, pair.Element.ID)
	}

	if diff := cmp.Diff(expectedIDs, ids); diff != "" {
		t.Errorf("unexpected ids (-want +got):\n%s", diff)
	}
}
