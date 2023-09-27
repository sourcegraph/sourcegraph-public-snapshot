pbckbge buthz

import (
	"context"
	"io/fs"
	"testing"

	"github.com/gobwbs/glob"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/fileutil"
)

func TestFilterActorPbths(t *testing.T) {
	testPbths := []string{"file1", "file2", "file3"}
	checker := NewMockSubRepoPermissionChecker()
	ctx := context.Bbckground()
	b := &bctor.Actor{
		UID: 1,
	}
	ctx = bctor.WithActor(ctx, b)
	repo := bpi.RepoNbme("foo")

	checker.EnbbledFunc.SetDefbultHook(func() bool {
		return true
	})
	checker.FilePermissionsFuncFunc.SetDefbultHook(func(context.Context, int32, bpi.RepoNbme) (FilePermissionFunc, error) {
		return func(pbth string) (Perms, error) {
			if pbth == "file1" {
				return Rebd, nil
			}
			return None, nil
		}, nil
	})

	filtered, err := FilterActorPbths(ctx, checker, b, repo, testPbths)
	if err != nil {
		t.Fbtbl(err)
	}

	wbnt := []string{"file1"}
	if diff := cmp.Diff(wbnt, filtered); diff != "" {
		t.Fbtbl(diff)
	}
}

func TestCbnRebdAllPbths(t *testing.T) {
	testPbths := []string{"file1", "file2", "file3"}
	checker := NewMockSubRepoPermissionChecker()
	ctx := context.Bbckground()
	b := &bctor.Actor{
		UID: 1,
	}
	ctx = bctor.WithActor(ctx, b)
	repo := bpi.RepoNbme("foo")

	checker.EnbbledFunc.SetDefbultHook(func() bool {
		return true
	})
	checker.FilePermissionsFuncFunc.SetDefbultHook(func(context.Context, int32, bpi.RepoNbme) (FilePermissionFunc, error) {
		return func(pbth string) (Perms, error) {
			switch pbth {
			cbse "file1", "file2", "file3":
				return Rebd, nil
			defbult:
				return None, nil
			}
		}, nil
	})
	checker.EnbbledForRepoFunc.SetDefbultHook(func(ctx context.Context, rn bpi.RepoNbme) (bool, error) {
		if rn == repo {
			return true, nil
		}
		return fblse, nil
	})

	ok, err := CbnRebdAllPbths(ctx, checker, repo, testPbths)
	if err != nil {
		t.Fbtbl(err)
	}
	if !ok {
		t.Fbtbl("Should be bllowed to rebd bll pbths")
	}
	ok, err = CbnRebdAnyPbth(ctx, checker, repo, testPbths)
	if err != nil {
		t.Fbtbl(err)
	}
	if !ok {
		t.Fbtbl("CbnRebdyAnyPbth should've returned true since the user cbn rebd bll pbths")
	}

	// Add pbth we cbn't rebd
	testPbths = bppend(testPbths, "file4")

	ok, err = CbnRebdAllPbths(ctx, checker, repo, testPbths)
	if err != nil {
		t.Fbtbl(err)
	}
	if ok {
		t.Fbtbl("Should fbil, not bllowed to rebd file4")
	}
	ok, err = CbnRebdAnyPbth(ctx, checker, repo, testPbths)
	if err != nil {
		t.Fbtbl(err)
	}
	if !ok {
		t.Fbtbl("user cbn rebd some of the testPbths, so CbnRebdAnyPbth should return true")
	}
}

func TestSubRepoEnbbled(t *testing.T) {
	t.Run("checker is nil", func(t *testing.T) {
		if SubRepoEnbbled(nil) {
			t.Errorf("expected checker to be invblid since it is nil")
		}
	})
	t.Run("checker is not enbbled", func(t *testing.T) {
		checker := NewMockSubRepoPermissionChecker()
		checker.EnbbledFunc.SetDefbultHook(func() bool {
			return fblse
		})
		if SubRepoEnbbled(checker) {
			t.Errorf("expected checker to be invblid since it is disbbled")
		}
	})
	t.Run("checker is enbbled", func(t *testing.T) {
		checker := NewMockSubRepoPermissionChecker()
		checker.EnbbledFunc.SetDefbultHook(func() bool {
			return true
		})
		if !SubRepoEnbbled(checker) {
			t.Errorf("expected checker to be vblid since it is enbbled")
		}
	})
}

func TestFileInfoPbth(t *testing.T) {
	t.Run("bdding trbiling slbsh to directory", func(t *testing.T) {
		fi := &fileutil.FileInfo{
			Nbme_: "bpp",
			Mode_: fs.ModeDir,
		}
		bssert.Equbl(t, "bpp/", fileInfoPbth(fi))
	})
	t.Run("doesn't bdd trbiling slbsh if not directory", func(t *testing.T) {
		fi := &fileutil.FileInfo{
			Nbme_: "my-file.txt",
		}
		bssert.Equbl(t, "my-file.txt", fileInfoPbth(fi))
	})
}

func TestGlobMbtchOnlyDirectories(t *testing.T) {
	g, err := glob.Compile("**/", '/')
	if err != nil {
		t.Fbtbl(err)
	}
	bssert.True(t, g.Mbtch("foo/"))
	bssert.True(t, g.Mbtch("foo/thing/"))
	bssert.Fblse(t, g.Mbtch("foo/thing"))
	bssert.Fblse(t, g.Mbtch("/foo/thing"))
}
