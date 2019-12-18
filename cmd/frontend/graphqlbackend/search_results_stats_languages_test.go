package graphqlbackend

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/inventory"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

func TestSearchResultsStatsLanguages(t *testing.T) {
	wantCommitID := api.CommitID(strings.Repeat("a", 40))
	rcache.SetupForTest(t)

	git.Mocks.NewFileReader = func(commit api.CommitID, name string) (io.ReadCloser, error) {
		if commit != wantCommitID {
			t.Errorf("got commit %q, want %q", commit, wantCommitID)
		}
		var data []byte
		switch name {
		case "two.go":
			data = []byte("a\nb\n")
		case "three.go":
			data = []byte("a\nb\nc\n")
		default:
			panic("unhandled mock NewFileReader " + name)
		}
		return ioutil.NopCloser(bytes.NewReader(data)), nil
	}
	const wantDefaultBranchRef = "refs/heads/foo"
	git.Mocks.ExecSafe = func(params []string) (stdout, stderr []byte, exitCode int, err error) {
		// Mock default branch lookup in (*RepsitoryResolver).DefaultBranch.
		return []byte(wantDefaultBranchRef), nil, 0, nil
	}
	git.Mocks.ResolveRevision = func(spec string, opt *git.ResolveRevisionOptions) (api.CommitID, error) {
		if want := "HEAD"; spec != want {
			t.Errorf("got spec %q, want %q", spec, want)
		}
		return wantCommitID, nil
	}
	git.Mocks.GetObject = func(objectName string) (git.OID, git.ObjectType, error) {
		oid := git.OID{} // empty is OK for this test
		copy(oid[:], bytes.Repeat([]byte{0xaa}, 40))
		return oid, git.ObjectTypeTree, nil
	}
	defer git.ResetMocks()

	tests := map[string]struct {
		results  []SearchResultResolver
		getFiles []os.FileInfo
		want     []inventory.Lang // TotalBytes values are incorrect (known issue doc'd in GraphQL schema)
	}{
		"empty": {
			results: nil,
			want:    []inventory.Lang{},
		},
		"1 entire file": {
			results: []SearchResultResolver{
				&FileMatchResolver{
					JPath:    "three.go",
					Repo:     &types.Repo{Name: "r"},
					CommitID: wantCommitID,
				},
			},
			want: []inventory.Lang{{Name: "Go", TotalBytes: 1, TotalLines: 3}},
		},
		"line matches in 1 file": {
			results: []SearchResultResolver{
				&FileMatchResolver{
					JPath:        "three.go",
					Repo:         &types.Repo{Name: "r"},
					CommitID:     wantCommitID,
					JLineMatches: []*lineMatch{{JLineNumber: 1}},
				},
			},
			want: []inventory.Lang{{Name: "Go", TotalBytes: 1, TotalLines: 1}},
		},
		"line matches in 2 files": {
			results: []SearchResultResolver{
				&FileMatchResolver{
					JPath:        "two.go",
					Repo:         &types.Repo{Name: "r"},
					CommitID:     wantCommitID,
					JLineMatches: []*lineMatch{{JLineNumber: 1}, {JLineNumber: 2}},
				},
				&FileMatchResolver{
					JPath:        "three.go",
					Repo:         &types.Repo{Name: "r"},
					CommitID:     wantCommitID,
					JLineMatches: []*lineMatch{{JLineNumber: 1}},
				},
			},
			want: []inventory.Lang{{Name: "Go", TotalBytes: 2, TotalLines: 3}},
		},
		"1 entire repo": {
			results: []SearchResultResolver{
				&RepositoryResolver{
					repo: &types.Repo{Name: "r"},
				},
			},
			getFiles: []os.FileInfo{
				fileInfo{path: "two.go", size: 1},
				fileInfo{path: "three.go", size: 1},
			},
			want: []inventory.Lang{{Name: "Go", TotalBytes: 2, TotalLines: 5}},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			git.Mocks.ReadDir = func(commit api.CommitID, name string, recurse bool) ([]os.FileInfo, error) {
				return test.getFiles, nil
			}

			langs, err := searchResultsStatsLanguages(context.Background(), test.results)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(langs, test.want) {
				t.Errorf("got %+v, want %+v", langs, test.want)
			}
		})
	}
}
