package indexer

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	indexmanager "github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-indexer-vm/internal/index_manager"
	queuemocks "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/queue/client/mocks"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
)

func init() {
	makeTempDir = func() (string, error) { return "/tmp/testing", nil }
}

func TestHandleWithDocker(t *testing.T) {
	queueClient := queuemocks.NewMockClient()
	indexManager := indexmanager.New()
	commander := NewMockCommander()

	handler := &Handler{
		queueClient:  queueClient,
		indexManager: indexManager,
		commander:    commander,
		options: HandlerOptions{
			FrontendURL:           "https://sourcegraph.test:1234",
			FrontendURLFromDocker: "https://sourcegraph.test:5432",
			AuthToken:             "hunter2",
			FirecrackerNumCPUs:    8,
			FirecrackerMemory:     "32G",
		},
		uuidGenerator: uuid.NewRandom,
	}

	index := store.Index{
		ID:             42,
		RepositoryName: "github.com/sourcegraph/sourcegraph",
		Commit:         "e2249f2173e8ca0c8c2541644847e7bf01aaef4a",
	}

	if err := handler.Handle(context.Background(), nil, index); err != nil {
		t.Fatalf("unexpected error handling index: %s", err)
	}

	if callCount := len(commander.RunFunc.History()); callCount != 5 {
		t.Errorf("unexpected run call count. want=%d have=%d", 5, callCount)
	} else {
		expectedCalls := []string{
			"git -C /tmp/testing init",
			"git -C /tmp/testing -c protocol.version=2 fetch https://indexer:hunter2@sourcegraph.test:1234/.internal-code-intel/git/github.com/sourcegraph/sourcegraph e2249f2173e8ca0c8c2541644847e7bf01aaef4a",
			"git -C /tmp/testing checkout e2249f2173e8ca0c8c2541644847e7bf01aaef4a",
			"docker run --rm --cpus 8 --memory 32G -v /tmp/testing:/data -w /data sourcegraph/lsif-go:latest lsif-go --noProgress",
			"docker run --rm --cpus 8 --memory 32G -v /tmp/testing:/data -w /data -e SRC_ENDPOINT=https://indexer:hunter2@sourcegraph.test:5432 sourcegraph/src-cli:latest lsif upload -no-progress -repo github.com/sourcegraph/sourcegraph -commit e2249f2173e8ca0c8c2541644847e7bf01aaef4a -upload-route /.internal-code-intel/lsif/upload",
		}

		calls := commander.RunFunc.History()

		for i, expectedCall := range expectedCalls {
			if diff := cmp.Diff(expectedCall, fmt.Sprintf("%s %s", calls[i].Arg1, strings.Join(calls[i].Arg2, " "))); diff != "" {
				t.Errorf("unexpected command (-want +got):\n%s", diff)
			}
		}
	}
}

func TestHandleWithFirecracker(t *testing.T) {
	queueClient := queuemocks.NewMockClient()
	indexManager := indexmanager.New()
	commander := NewMockCommander()

	handler := &Handler{
		queueClient:  queueClient,
		indexManager: indexManager,
		commander:    commander,
		options: HandlerOptions{
			FrontendURL:           "https://sourcegraph.test:1234",
			FrontendURLFromDocker: "https://sourcegraph.test:5432",
			AuthToken:             "hunter2",
			UseFirecracker:        true,
			FirecrackerImage:      "sourcegraph/ignite-ubuntu:latest",
			FirecrackerNumCPUs:    8,
			FirecrackerMemory:     "32G",
		},
		uuidGenerator: func() (uuid.UUID, error) {
			return uuid.MustParse("97b45daf-53d1-48ad-b992-547469d8e438"), nil
		},
	}

	index := store.Index{
		ID:             42,
		RepositoryName: "github.com/sourcegraph/sourcegraph",
		Commit:         "e2249f2173e8ca0c8c2541644847e7bf01aaef4a",
	}

	if err := handler.Handle(context.Background(), nil, index); err != nil {
		t.Fatalf("unexpected error handling index: %s", err)
	}

	if callCount := len(commander.RunFunc.History()); callCount != 8 {
		t.Errorf("unexpected run call count. want=%d have=%d", 8, callCount)
	} else {
		expectedCalls := []string{
			"git -C /tmp/testing init",
			"git -C /tmp/testing -c protocol.version=2 fetch https://indexer:hunter2@sourcegraph.test:1234/.internal-code-intel/git/github.com/sourcegraph/sourcegraph e2249f2173e8ca0c8c2541644847e7bf01aaef4a",
			"git -C /tmp/testing checkout e2249f2173e8ca0c8c2541644847e7bf01aaef4a",
			"ignite run --runtime docker --network-plugin docker-bridge --cpus 8 --memory 32G --copy-files /tmp/testing:/repo-dir --ssh --name 97b45daf-53d1-48ad-b992-547469d8e438 sourcegraph/ignite-ubuntu:latest",
			"ignite exec 97b45daf-53d1-48ad-b992-547469d8e438 -- docker run --rm --cpus 8 --memory 32G -v /repo-dir:/data -w /data sourcegraph/lsif-go:latest lsif-go --noProgress",
			"ignite exec 97b45daf-53d1-48ad-b992-547469d8e438 -- docker run --rm --cpus 8 --memory 32G -v /repo-dir:/data -w /data -e SRC_ENDPOINT=https://indexer:hunter2@sourcegraph.test:5432 sourcegraph/src-cli:latest lsif upload -no-progress -repo github.com/sourcegraph/sourcegraph -commit e2249f2173e8ca0c8c2541644847e7bf01aaef4a -upload-route /.internal-code-intel/lsif/upload",
			"ignite stop --runtime docker --network-plugin docker-bridge 97b45daf-53d1-48ad-b992-547469d8e438",
			"ignite rm -f --runtime docker --network-plugin docker-bridge 97b45daf-53d1-48ad-b992-547469d8e438",
		}

		calls := commander.RunFunc.History()

		for i, expectedCall := range expectedCalls {
			if diff := cmp.Diff(expectedCall, fmt.Sprintf("%s %s", calls[i].Arg1, strings.Join(calls[i].Arg2, " "))); diff != "" {
				t.Errorf("unexpected command (-want +got):\n%s", diff)
			}
		}
	}
}
