package ranking

import (
	"context"
	"io"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/unpack/unpacktest"
)

func TestIndexRepository(t *testing.T) {
	ctx := context.Background()
	mockStore := NewMockStore()
	gitserverClient := NewMockGitserverClient()
	svc := newService(mockStore, nil, gitserverClient, siteConfigQuerier{}, &observation.TestContext)

	repositoryContents := map[string]string{
		"foo.go": "func Foo()",
		"bar.go": "func Bar() { Foo() }",
		"baz.go": "func Baz() { Bar(); Foo() }",
	}

	gitserverClient.ArchiveReaderFunc.SetDefaultHook(func(_ context.Context, _ authz.SubRepoPermissionChecker, repoName api.RepoName, opts gitserver.ArchiveOptions) (io.ReadCloser, error) {
		return unpacktest.CreateTarArchive(t, repositoryContents), nil
	})

	if err := svc.indexRepository(ctx, api.RepoName("foo")); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// TODO - add an assertion here to prevent regressions
	t.Logf("ranks=%#v", mockStore.SetDocumentRanksFunc.History()[0])
}
