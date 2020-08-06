package indexer

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	indexmanager "github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-indexer-vm/internal/index_manager"
	queuemocks "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/queue/client/mocks"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
)

var testHandlerOptions = HandlerOptions{
	FrontendURL:           "https://sourcegraph.test:1234",
	FrontendURLFromDocker: "https://sourcegraph.test:5432",
	AuthToken:             "hunter2",
}

func init() {
	makeTempDir = func() (string, error) { return "/tmp", nil }
}

func TestHandle(t *testing.T) {
	queueClient := queuemocks.NewMockClient()
	indexManager := indexmanager.New()
	commander := NewMockCommander()

	handler := &Handler{
		queueClient:  queueClient,
		indexManager: indexManager,
		commander:    commander,
		options:      testHandlerOptions,
	}

	index := store.Index{
		ID:             42,
		RepositoryName: "github.com/sourcegraph/sourcegraph",
		Commit:         "e2249f2173e8ca0c8c2541644847e7bf01aaef4a",
	}

	if err := handler.Handle(context.Background(), nil, index); err != nil {
		t.Fatalf("unexpected error handling index: %s", err)
	}

	if callCount := len(commander.RunFunc.History()); callCount != 4 {
		t.Errorf("unexpected run call count. want=%d have=%d", 4, callCount)
	} else {
		expectedCalls := []string{
			"git -C /tmp init",
			"git -C /tmp -c protocol.version=2 fetch https://indexer:hunter2@sourcegraph.test:1234/.internal-code-intel/git/github.com/sourcegraph/sourcegraph e2249f2173e8ca0c8c2541644847e7bf01aaef4a",
			"git -C /tmp checkout e2249f2173e8ca0c8c2541644847e7bf01aaef4a",
			"docker run --rm -v /tmp:/data -w /data sourcegraph/lsif-go:latest bash -c lsif-go && src -endpoint https://sourcegraph.test:5432 lsif upload -repo github.com/sourcegraph/sourcegraph -commit e2249f2173e8ca0c8c2541644847e7bf01aaef4a",
		}

		calls := commander.RunFunc.History()

		for i, expectedCall := range expectedCalls {
			if diff := cmp.Diff(expectedCall, fmt.Sprintf("%s %s", calls[i].Arg1, strings.Join(calls[i].Arg2, " "))); diff != "" {
				t.Errorf("unexpected command (-want +got):\n%s", diff)
			}
		}
	}
}
