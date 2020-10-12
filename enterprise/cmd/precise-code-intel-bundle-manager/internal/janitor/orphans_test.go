package janitor

import (
	"context"
	"testing"

	lsifstoremocks "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/lsifstore/mocks"
	storemocks "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store/mocks"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
)

func TestRemoveOrphanedData(t *testing.T) {
	mockStore := storemocks.NewMockStore()
	mockLSIFStore := lsifstoremocks.NewMockStore()

	j := &Janitor{
		store:     mockStore,
		lsifStore: mockLSIFStore,
		metrics:   NewJanitorMetrics(metrics.TestRegisterer),
	}

	if err := j.removeOrphanedData(context.Background()); err != nil {
		t.Fatalf("unexpected error removing orphaned data: %s", err)
	}

	// TODO
}
