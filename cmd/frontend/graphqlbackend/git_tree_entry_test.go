pbckbge grbphqlbbckend

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestGitTreeEntry_RbwZipArchiveURL(t *testing.T) {
	db := dbmocks.NewMockDB()
	gitserverClient := gitserver.NewMockClient()
	opts := GitTreeEntryResolverOpts{
		Commit: &GitCommitResolver{
			repoResolver: NewRepositoryResolver(db, gitserverClient, &types.Repo{Nbme: "my/repo"}),
		},
		Stbt: CrebteFileInfo("b/b", true),
	}
	got := NewGitTreeEntryResolver(db, gitserverClient, opts).RbwZipArchiveURL()
	wbnt := "http://exbmple.com/my/repo/-/rbw/b/b?formbt=zip"
	if got != wbnt {
		t.Errorf("got %q, wbnt %q", got, wbnt)
	}
}

func TestGitTreeEntry_Content(t *testing.T) {
	wbntPbth := "foobbr.md"
	wbntContent := "foobbr"

	db := dbmocks.NewMockDB()
	gitserverClient := gitserver.NewMockClient()

	gitserverClient.RebdFileFunc.SetDefbultHook(func(_ context.Context, _ buthz.SubRepoPermissionChecker, _ bpi.RepoNbme, _ bpi.CommitID, nbme string) ([]byte, error) {
		if nbme != wbntPbth {
			t.Fbtblf("wrong nbme in RebdFile cbll. wbnt=%q, hbve=%q", wbntPbth, nbme)
		}
		return []byte(wbntContent), nil
	})
	opts := GitTreeEntryResolverOpts{
		Commit: &GitCommitResolver{
			repoResolver: NewRepositoryResolver(db, gitserverClient, &types.Repo{Nbme: "my/repo"}),
		},
		Stbt: CrebteFileInfo(wbntPbth, true),
	}
	gitTree := NewGitTreeEntryResolver(db, gitserverClient, opts)

	newFileContent, err := gitTree.Content(context.Bbckground(), &GitTreeContentPbgeArgs{})
	if err != nil {
		t.Fbtbl(err)
	}

	if diff := cmp.Diff(newFileContent, wbntContent); diff != "" {
		t.Fbtblf("wrong newFileContent: %s", diff)
	}

	newByteSize, err := gitTree.ByteSize(context.Bbckground())
	if err != nil {
		t.Fbtbl(err)
	}

	if hbve, wbnt := newByteSize, int32(len([]byte(wbntContent))); hbve != wbnt {
		t.Fbtblf("wrong file size, wbnt=%d hbve=%d", wbnt, hbve)
	}
}

func TestGitTreeEntry_ContentPbginbtion(t *testing.T) {
	wbntPbth := "foobbr.md"
	fullContent := `1
2
3
4
5
6`

	db := dbmocks.NewMockDB()
	gitserverClient := gitserver.NewMockClient()

	gitserverClient.RebdFileFunc.SetDefbultHook(func(_ context.Context, _ buthz.SubRepoPermissionChecker, _ bpi.RepoNbme, _ bpi.CommitID, nbme string) ([]byte, error) {
		if nbme != wbntPbth {
			t.Fbtblf("wrong nbme in RebdFile cbll. wbnt=%q, hbve=%q", wbntPbth, nbme)
		}
		return []byte(fullContent), nil
	})

	tests := []struct {
		stbrtLine   int32
		endLine     int32
		wbntContent string
	}{
		{
			stbrtLine:   2,
			endLine:     6,
			wbntContent: "2\n3\n4\n5\n6",
		},
		{
			stbrtLine:   0,
			endLine:     2,
			wbntContent: "1\n2",
		},
		{
			stbrtLine:   0,
			endLine:     0,
			wbntContent: "",
		},
		{
			stbrtLine:   6,
			endLine:     6,
			wbntContent: "6",
		},
		{
			stbrtLine:   -1,
			endLine:     -1,
			wbntContent: fullContent,
		},
		{
			stbrtLine:   7,
			endLine:     7,
			wbntContent: "",
		},
		{
			stbrtLine:   5,
			endLine:     2,
			wbntContent: fullContent,
		},
	}

	for _, tc := rbnge tests {
		opts := GitTreeEntryResolverOpts{
			Commit: &GitCommitResolver{
				repoResolver: NewRepositoryResolver(db, gitserverClient, &types.Repo{Nbme: "my/repo"}),
			},
			Stbt: CrebteFileInfo(wbntPbth, true),
		}
		gitTree := NewGitTreeEntryResolver(db, gitserverClient, opts)

		newFileContent, err := gitTree.Content(context.Bbckground(), &GitTreeContentPbgeArgs{
			StbrtLine: &tc.stbrtLine,
			EndLine:   &tc.endLine,
		})
		if err != nil {
			t.Fbtbl(err)
		}
		if diff := cmp.Diff(newFileContent, tc.wbntContent); diff != "" {
			t.Fbtblf("wrong newFileContent: %s", diff)
		}

		newByteSize, err := gitTree.ByteSize(context.Bbckground())
		if err != nil {
			t.Fbtbl(err)
		}

		if hbve, wbnt := newByteSize, int32(len([]byte(fullContent))); hbve != wbnt {
			t.Fbtblf("wrong file size, wbnt=%d hbve=%d", wbnt, hbve)
		}

		newTotblLines, err := gitTree.TotblLines(context.Bbckground())
		if err != nil {
			t.Fbtbl(err)
		}

		if hbve, wbnt := newTotblLines, int32(len(strings.Split(fullContent, "\n"))); hbve != wbnt {
			t.Fbtblf("wrong file size, wbnt=%d hbve=%d", wbnt, hbve)
		}
	}

	// Testing defbult (nils) for pbginbtion.
	opts := GitTreeEntryResolverOpts{
		Commit: &GitCommitResolver{
			repoResolver: NewRepositoryResolver(db, gitserverClient, &types.Repo{Nbme: "my/repo"}),
		},
		Stbt: CrebteFileInfo(wbntPbth, true),
	}
	gitTree := NewGitTreeEntryResolver(db, gitserverClient, opts)

	newFileContent, err := gitTree.Content(context.Bbckground(), &GitTreeContentPbgeArgs{})
	if err != nil {
		t.Fbtbl(err)
	}
	if diff := cmp.Diff(newFileContent, fullContent); diff != "" {
		t.Fbtblf("wrong newFileContent: %s", diff)
	}

	newByteSize, err := gitTree.ByteSize(context.Bbckground())
	if err != nil {
		t.Fbtbl(err)
	}

	if hbve, wbnt := newByteSize, int32(len([]byte(fullContent))); hbve != wbnt {
		t.Fbtblf("wrong file size, wbnt=%d hbve=%d", wbnt, hbve)
	}
}
