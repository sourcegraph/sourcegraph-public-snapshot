package graphqlbackend

import (
	"bytes"
	"context"
	"io"
	"io/fs"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/fileutil"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/inventory"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestSearchResultsStatsLanguages(t *testing.T) {
	logger := logtest.Scoped(t)
	wantCommitID := api.CommitID(strings.Repeat("a", 40))
	rcache.SetupForTest(t)

	gsClient := gitserver.NewMockClient()
	gsClient.NewFileReaderFunc.SetDefaultHook(func(_ context.Context, _ api.RepoName, commit api.CommitID, name string) (io.ReadCloser, error) {
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
		return io.NopCloser(bytes.NewReader(data)), nil
	})
	const wantDefaultBranchRef = "refs/heads/foo"
	gsClient.GetDefaultBranchFunc.SetDefaultHook(func(context.Context, api.RepoName, bool) (string, api.CommitID, error) {
		// Mock default branch lookup in (*RepositoryResolver).DefaultBranch.
		return wantDefaultBranchRef, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", nil
	})
	gsClient.ResolveRevisionFunc.SetDefaultHook(func(_ context.Context, _ api.RepoName, spec string) (api.CommitID, error) {
		if want := "HEAD"; spec != want {
			t.Errorf("got spec %q, want %q", spec, want)
		}
		return wantCommitID, nil
	})

	gsClient.StatFunc.SetDefaultHook(func(_ context.Context, _ api.RepoName, _ api.CommitID, path string) (fs.FileInfo, error) {
		return &fileutil.FileInfo{Name_: path, Mode_: os.ModeDir}, nil
	})

	mkResult := func(path string, lineNumbers ...int) *result.FileMatch {
		rn := types.MinimalRepo{
			Name: "r",
		}
		fm := mkFileMatch(rn, path, lineNumbers...)
		fm.CommitID = wantCommitID
		return fm
	}

	tests := map[string]struct {
		results  []result.Match
		getFiles []fs.FileInfo
		want     []inventory.Lang // TotalBytes values are incorrect (known issue doc'd in GraphQL schema)
	}{
		"empty": {
			results: nil,
			want:    []inventory.Lang{},
		},
		"1 entire file": {
			results: []result.Match{
				mkResult("three.go"),
			},
			want: []inventory.Lang{{Name: "Go", TotalBytes: 6, TotalLines: 3}},
		},
		"line matches in 1 file": {
			results: []result.Match{
				mkResult("three.go", 1),
			},
			want: []inventory.Lang{{Name: "Go", TotalBytes: 6, TotalLines: 1}},
		},
		"line matches in 2 files": {
			results: []result.Match{
				mkResult("two.go", 1, 2),
				mkResult("three.go", 1),
			},
			want: []inventory.Lang{{Name: "Go", TotalBytes: 10, TotalLines: 3}},
		},
		"1 entire repo": {
			results: []result.Match{
				&result.RepoMatch{Name: "r"},
			},
			getFiles: []fs.FileInfo{
				fileInfo{path: "two.go"},
				fileInfo{path: "three.go"},
			},
			want: []inventory.Lang{{Name: "Go", TotalBytes: 10, TotalLines: 5}},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			gsClient.ReadDirFunc.SetDefaultHook(func(context.Context, api.RepoName, api.CommitID, string, bool) ([]fs.FileInfo, error) {
				return test.getFiles, nil
			})

			langs, err := searchResultsStatsLanguages(context.Background(), logger, dbmocks.NewMockDB(), gsClient, test.results)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(langs, test.want) {
				t.Errorf("got %+v, want %+v", langs, test.want)
			}
		})
	}
}
