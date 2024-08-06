package codenav

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/scip/bindings/go/scip"

	lsifstoremocks "github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/internal/lsifstore/mocks"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/core"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search/client"
)

const sampleFile1 = `package food

type banana struct{}`

func TestSnapshotForDocument(t *testing.T) {
	// Set up mocks
	fakeRepoStore := AllPresentFakeRepoStore{}
	mockLsifStore := lsifstoremocks.NewMockLsifStore()
	mockUploadSvc := NewMockUploadService()
	mockGitserverClient := gitserver.NewMockClient()
	mockGitserverClient.DiffFunc.SetDefaultReturn(gitserver.NewDiffFileIterator(io.NopCloser(strings.NewReader(""))), nil)
	mockSearchClient := client.NewMockSearchClient()

	// Init service
	svc := newService(observation.TestContextTB(t), fakeRepoStore, mockLsifStore, mockUploadSvc, mockGitserverClient, mockSearchClient, log.NoOp())

	mockUploadSvc.GetCompletedUploadsByIDsFunc.SetDefaultReturn([]shared.CompletedUpload{{}}, nil)
	mockGitserverClient.NewFileReaderFunc.SetDefaultReturn(io.NopCloser(bytes.NewReader([]byte(sampleFile1))), nil)
	mockLsifStore.SCIPDocumentFunc.SetDefaultReturn(core.Some(&scip.Document{
		RelativePath: "burger.go",
		Occurrences: []*scip.Occurrence{{
			Range:       []int32{2, 4, 9},
			Symbol:      "scip-go gomod github.com/sourcegraph/banter v4.2.0 github.com/sourcegraph/banter/food/banana#",
			SymbolRoles: int32(scip.SymbolRole_Definition),
		}},
		Symbols: []*scip.SymbolInformation{{
			Symbol: "scip-go gomod github.com/sourcegraph/banter v4.2.0 github.com/sourcegraph/banter/food/banana#",
			Relationships: []*scip.Relationship{{
				Symbol:           "scip-go gomod github.com/golang/go go1.18 fmt/Banterer#",
				IsImplementation: true,
			}},
		}},
	}), nil)

	data, err := svc.SnapshotForDocument(context.Background(), 0, "deadbeef", repoRelPath("burger.go"), 0)
	if err != nil {
		t.Fatal(err)
	}

	if len(data) != 1 {
		t.Fatal("no snapshot data returned")
	}

	if data[0].DocumentOffset != 35 {
		t.Fatalf("unexpected document offset (want=%d,got=%d)", 35, data[0].DocumentOffset)
	}
}
