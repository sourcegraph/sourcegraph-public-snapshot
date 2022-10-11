package ranking

import (
	"context"
	"io"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/unpack/unpacktest"
)

func TestIndexRepository(t *testing.T) {
	ctx := context.Background()
	mockStore := NewMockStore()
	gitserverClient := NewMockGitserverClient()
	symbolsClient := NewMockSymbolsClient()
	svc := newService(mockStore, nil, gitserverClient, symbolsClient, siteConfigQuerier{}, &observation.TestContext)

	repositoryContents := map[string]string{
		"foo.go": "func Foo()",
		"bar.go": "func Bar() { Foo() }",
		"baz.go": "func Baz() { Bar(); Foo() }",
	}

	gitserverClient.HeadFromNameFunc.SetDefaultReturn("deadebeef", true, nil)
	gitserverClient.ArchiveReaderFunc.SetDefaultHook(func(_ context.Context, _ authz.SubRepoPermissionChecker, repoName api.RepoName, opts gitserver.ArchiveOptions) (io.ReadCloser, error) {
		return unpacktest.CreateTarArchive(t, repositoryContents), nil
	})

	symbolsClient.SearchFunc.SetDefaultReturn([]result.Symbol{
		{Name: "Foo", Path: "foo.go"},
		{Name: "Bar", Path: "bar.go"},
		{Name: "Baz", Path: "baz.go"},
	}, nil)

	if err := svc.indexRepository(ctx, api.RepoName("foo")); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if calls := mockStore.SetDocumentRanksFunc.History(); len(calls) != 1 {
		t.Fatalf("unexpected call count. want=%d have=%d", 1, len(calls))
	} else {
		ranks := calls[0].Arg2
		if !(ranks["foo.go"][0] < ranks["bar.go"][0] && ranks["bar.go"][0] < ranks["baz.go"][0]) {
			t.Fatalf("unexpected ordering. want=%v have=%v", []string{"foo.go", "bar.go", "baz.go"}, ranks)
		}
	}
}
