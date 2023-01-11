package background

import (
	"testing"
)

func TestIndexRepository(t *testing.T) {
	t.Skip() // Flaky

	// TODO: this test needs to be redone, but is currently
	// disabled (via a hard-coded package-level variable).
	// Just commenting this out for now as we should have a
	// similar test once we resurrect the background indexing
	// effort.

	// ctx := context.Background()
	// mockStore := NewMockStore()
	// gitserverClient := NewMockGitserverClient()
	// symbolsClient := NewMockSymbolsClient()
	// svc := newService(mockStore, nil, gitserverClient, symbolsClient, conf.DefaultClient(), nil, &observation.TestContext)

	// repositoryContents := map[string]string{
	// 	"foo.go": "func Foo()",
	// 	"bar.go": "func Bar() { Foo() }",
	// 	"baz.go": "func Baz() { Bar(); Foo() }",
	// }

	// gitserverClient.HeadFromNameFunc.SetDefaultReturn("deadebeef", true, nil)
	// gitserverClient.ArchiveReaderFunc.SetDefaultHook(func(_ context.Context, _ authz.SubRepoPermissionChecker, repoName api.RepoName, opts gitserver.ArchiveOptions) (io.ReadCloser, error) {
	// 	return unpacktest.CreateTarArchive(t, repositoryContents), nil
	// })

	// symbolsClient.SearchFunc.SetDefaultReturn([]result.Symbol{
	// 	{Name: "Foo", Path: "foo.go"},
	// 	{Name: "Bar", Path: "bar.go"},
	// 	{Name: "Baz", Path: "baz.go"},
	// }, nil)

	// if err := svc.indexRepository(ctx, api.RepoName("foo")); err != nil {
	// 	t.Fatalf("unexpected error: %s", err)
	// }

	// if calls := mockStore.SetDocumentRanksFunc.History(); len(calls) != 1 {
	// 	t.Fatalf("unexpected call count. want=%d have=%d", 1, len(calls))
	// } else {
	// 	ranks := calls[0].Arg3
	// 	if !(ranks["foo.go"] < ranks["bar.go"] && ranks["bar.go"] < ranks["baz.go"]) {
	// 		t.Fatalf("unexpected ordering. want=%v have=%v", []string{"foo.go", "bar.go", "baz.go"}, ranks)
	// 	}
	// }
}
