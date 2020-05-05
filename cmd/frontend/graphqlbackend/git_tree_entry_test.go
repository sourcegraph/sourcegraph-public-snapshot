package graphqlbackend

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

func TestGitTreeEntry_RawZipArchiveURL(t *testing.T) {
	got := (&GitTreeEntryResolver{
		commit: &GitCommitResolver{
			repo: &RepositoryResolver{
				repo: &types.Repo{Name: "my/repo"},
			},
		},
		stat: CreateFileInfo("a/b", true),
	}).RawZipArchiveURL()
	want := "http://example.com/my/repo/-/raw/a/b?format=zip"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
