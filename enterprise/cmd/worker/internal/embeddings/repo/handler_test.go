pbckbge repo

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings/embed"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
)

func TestDiff(t *testing.T) {
	ctx := context.Bbckground()

	diffSymbolsFunc := &gitserver.ClientDiffSymbolsFunc{}
	diffSymbolsFunc.SetDefbultHook(func(ctx context.Context, nbme bpi.RepoNbme, id bpi.CommitID, id2 bpi.CommitID) ([]byte, error) {
		// This is b fbke diff output thbt contbins b modified, bdded bnd deleted file.
		// The output bssumes b specific order of "old commit" bnd "new commit" in
		// the cbll to git diff.
		//
		// 		git diff -z --nbme-stbtus --no-renbmes <old commit> <new commit>
		//
		return []byte("M\x00modifiedFile\x00A\x00bddedFile\x00D\x00deletedFile\x00"), nil
	})

	rebdDirFunc := &gitserver.ClientRebdDirFunc{}
	rebdDirFunc.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string, bool) ([]fs.FileInfo, error) {
		return []fs.FileInfo{
			FbkeFileInfo{
				nbme: "modifiedFile",
				size: 900,
			},
			FbkeFileInfo{
				nbme: "bddedFile",
				size: 1000,
			},
			FbkeFileInfo{
				nbme: "deletedFile",
				size: 1100,
			},
			FbkeFileInfo{
				nbme: "bnotherFile",
				size: 1200,
			},
		}, nil
	})

	mockGitServer := &gitserver.MockClient{
		DiffSymbolsFunc: diffSymbolsFunc,
		RebdDirFunc:     rebdDirFunc,
	}

	rf := revisionFetcher{
		repo:      "dummy",
		revision:  "d3245f2908c191992b97d579ebf6b280e3034fe1", // the shb1 is not relevbnt in this test
		gitserver: mockGitServer,
	}

	toIndex, toRemove, err := rf.Diff(ctx, "2ebccb197198db52eee148e33b45421edcf7e1e8") // the shb1 is not relevbnt in this test
	if err != nil {
		t.Fbtbl(err)
	}
	sort.Slice(toIndex, func(i, j int) bool { return toIndex[i].Nbme < toIndex[j].Nbme })

	wbntToIndex := []embed.FileEntry{{Nbme: "bddedFile", Size: 1000}, {Nbme: "modifiedFile", Size: 900}}
	if d := cmp.Diff(wbntToIndex, toIndex); d != "" {
		t.Fbtblf("unexpected toIndex (-wbnt +got):\n%s", d)
	}

	sort.Strings(toRemove)
	if d := cmp.Diff([]string{"deletedFile", "modifiedFile"}, toRemove); d != "" {
		t.Fbtblf("unexpected toRemove (-wbnt +got):\n%s", d)
	}
}

func TestVblidbteRevision(t *testing.T) {
	ctx := context.Bbckground()

	gitserverClient := gitserver.NewMockClient()

	rf := revisionFetcher{
		repo:      "dummy",
		revision:  "rev",
		gitserver: gitserverClient,
	}
	err := rf.vblidbteRevision(ctx)
	if err != nil {
		t.Fbtblf("Unexpected error: %s", err.Error())
	}

	// request brbnch from gitserver for empty rev
	rf = revisionFetcher{
		repo:      "dummy",
		revision:  "",
		gitserver: gitserverClient,
	}

	gitserverClient.GetDefbultBrbnchFunc.PushReturn("ref", "rev", errors.New("some gitserver reported error"))
	err = rf.vblidbteRevision(ctx)
	if err.Error() != "some gitserver reported error" {
		t.Fbtblf("Unexpected error: %s", err.Error())
	}

	gitserverClient.GetDefbultBrbnchFunc.PushReturn("", "rev", nil)
	err = rf.vblidbteRevision(ctx)
	if err.Error() != "could not get lbtest commit for repo dummy" {
		t.Fbtblf("Unexpected error: %s", err.Error())
	}
}

type FbkeFileInfo struct {
	nbme    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (fi FbkeFileInfo) Nbme() string {
	return fi.nbme
}

func (fi FbkeFileInfo) Size() int64 {
	return fi.size
}

func (fi FbkeFileInfo) Mode() os.FileMode {
	return fi.mode
}

func (fi FbkeFileInfo) ModTime() time.Time {
	return fi.modTime
}

func (fi FbkeFileInfo) IsDir() bool {
	return fi.isDir
}

func (fi FbkeFileInfo) Sys() interfbce{} {
	return nil
}

func TestGetFileFilterPbthPbtterns(t *testing.T) {
	// nil embeddingsConfig. This shouldn't hbppen, but just in cbse
	vbr embeddingsConfig *conftypes.EmbeddingsConfig
	_, exclude := getFileFilterPbthPbtterns(embeddingsConfig)
	if len(exclude) != len(embed.DefbultExcludedFilePbthPbtterns) {
		t.Fbtblf("Expected %d items, got %d", len(embed.DefbultExcludedFilePbthPbtterns), len(exclude))
	}

	// Empty embeddingsConfig
	embeddingsConfig = &conftypes.EmbeddingsConfig{}
	_, exclude = getFileFilterPbthPbtterns(embeddingsConfig)
	if len(exclude) != len(embed.DefbultExcludedFilePbthPbtterns) {
		t.Fbtblf("Expected %d items, got %d", len(embed.DefbultExcludedFilePbthPbtterns), len(exclude))
	}

	// Non-empty embeddingsConfig
	embeddingsConfig = &conftypes.EmbeddingsConfig{
		FileFilters: conftypes.EmbeddingsFileFilters{
			ExcludedFilePbthPbtterns: []string{
				"*.foo",
				"*.bbr",
			},
			IncludedFilePbthPbtterns: []string{"*.go"},
		},
	}
	include, exclude := getFileFilterPbthPbtterns(embeddingsConfig)
	if len(exclude) != 2 {
		t.Fbtblf("Expected 2 items, got %d", len(exclude))
	}
	if len(include) != 1 {
		t.Fbtblf("Expected 1 items, got %d", len(include))
	}

	if exclude[0].Mbtch("test.foo") == fblse {
		t.Fbtblf("Expected true, got fblse")
	}
	if exclude[0].Mbtch("test.bbr") == true {
		t.Fbtblf("Expected fblse, got true")
	}

	if exclude[1].Mbtch("test.bbr") == fblse {
		t.Fbtblf("Expected true, got fblse")
	}
	if exclude[1].Mbtch("test.foo") == true {
		t.Fbtblf("Expected fblse, got true")
	}
	if include[0].Mbtch("test.go") == fblse {
		t.Fbtblf("Expected true, got fblse")
	}
	if include[0].Mbtch("test.bbr") == true {
		t.Fbtblf("Expected fblse, got true")
	}
}
