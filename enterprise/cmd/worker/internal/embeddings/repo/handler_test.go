package repo

import (
	"context"
	"io/fs"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/embeddings/embed"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

func TestDiff(t *testing.T) {
	ctx := context.Background()

	diffSymbolsFunc := &gitserver.ClientDiffSymbolsFunc{}
	diffSymbolsFunc.SetDefaultHook(func(ctx context.Context, name api.RepoName, id api.CommitID, id2 api.CommitID) ([]byte, error) {
		// This is a fake diff output that contains a modified, added and deleted file.
		// The output assumes a specific order of "old commit" and "new commit" in
		// the call to git diff.
		//
		// 		git diff -z --name-status --no-renames <old commit> <new commit>
		//
		return []byte("M\x00modifiedFile\x00A\x00addedFile\x00D\x00deletedFile\x00"), nil
	})

	readDirFunc := &gitserver.ClientReadDirFunc{}
	readDirFunc.SetDefaultHook(func(context.Context, authz.SubRepoPermissionChecker, api.RepoName, api.CommitID, string, bool) ([]fs.FileInfo, error) {
		return []fs.FileInfo{
			FakeFileInfo{
				name: "modifiedFile",
				size: 900,
			},
			FakeFileInfo{
				name: "addedFile",
				size: 1000,
			},
			FakeFileInfo{
				name: "deletedFile",
				size: 1100,
			},
			FakeFileInfo{
				name: "anotherFile",
				size: 1200,
			},
		}, nil
	})

	mockGitServer := &gitserver.MockClient{
		DiffSymbolsFunc: diffSymbolsFunc,
		ReadDirFunc:     readDirFunc,
	}

	rf := revisionFetcher{
		repo:      "dummy",
		revision:  "d3245f2908c191992b97d579eaf6a280e3034fe1", // the sha1 is not relevant in this test
		gitserver: mockGitServer,
	}

	toIndex, toRemove, err := rf.Diff(ctx, "2ebccb197198da52eee148e33a45421edcf7e1e8") // the sha1 is not relevant in this test
	if err != nil {
		t.Fatal(err)
	}
	sort.Slice(toIndex, func(i, j int) bool { return toIndex[i].Name < toIndex[j].Name })

	wantToIndex := []embed.FileEntry{{Name: "addedFile", Size: 1000}, {Name: "modifiedFile", Size: 900}}
	if d := cmp.Diff(wantToIndex, toIndex); d != "" {
		t.Fatalf("unexpected toIndex (-want +got):\n%s", d)
	}

	sort.Strings(toRemove)
	if d := cmp.Diff([]string{"deletedFile", "modifiedFile"}, toRemove); d != "" {
		t.Fatalf("unexpected toRemove (-want +got):\n%s", d)
	}
}

type FakeFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (fi FakeFileInfo) Name() string {
	return fi.name
}

func (fi FakeFileInfo) Size() int64 {
	return fi.size
}

func (fi FakeFileInfo) Mode() os.FileMode {
	return fi.mode
}

func (fi FakeFileInfo) ModTime() time.Time {
	return fi.modTime
}

func (fi FakeFileInfo) IsDir() bool {
	return fi.isDir
}

func (fi FakeFileInfo) Sys() interface{} {
	return nil
}

func TestGetExcludedFilePathPatterns(t *testing.T) {
	// nil embeddingsConfig. This shouldn't happen, but just in case
	var embeddingsConfig *conftypes.EmbeddingsConfig
	result := getExcludedFilePathPatterns(embeddingsConfig)
	if len(result) != len(embed.DefaultExcludedFilePathPatterns) {
		t.Fatalf("Expected %d items, got %d", len(embed.DefaultExcludedFilePathPatterns), len(result))
	}

	// Empty embeddingsConfig
	embeddingsConfig = &conftypes.EmbeddingsConfig{}
	result = getExcludedFilePathPatterns(embeddingsConfig)
	if len(result) != len(embed.DefaultExcludedFilePathPatterns) {
		t.Fatalf("Expected %d items, got %d", len(embed.DefaultExcludedFilePathPatterns), len(result))
	}

	// Non-empty embeddingsConfig
	embeddingsConfig = &conftypes.EmbeddingsConfig{
		ExcludedFilePathPatterns: []string{
			"*.foo",
			"*.bar",
		},
	}
	result = getExcludedFilePathPatterns(embeddingsConfig)
	if len(result) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(result))
	}

	if result[0].Match("test.foo") == false {
		t.Fatalf("Expected true, got false")
	}
	if result[0].Match("test.bar") == true {
		t.Fatalf("Expected false, got true")
	}

	if result[1].Match("test.bar") == false {
		t.Fatalf("Expected true, got false")
	}
	if result[1].Match("test.foo") == true {
		t.Fatalf("Expected false, got true")
	}
}
