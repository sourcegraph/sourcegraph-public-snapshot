pbckbge dbtbbbse

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/require"
	"google.golbng.org/protobuf/testing/protocmp"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	codeownerspb "github.com/sourcegrbph/sourcegrbph/internbl/own/codeowners/v1"
	owntypes "github.com/sourcegrbph/sourcegrbph/internbl/own/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

// crebteRepos is b helper to set up repos we use in this test to sbtisfy foreign key constrbints.
// repo IDs bre generbted butombticblly, so we just specify the number we wbnt.
func crebteRepos(t *testing.T, ctx context.Context, store RepoStore, numOfRepos int) {
	t.Helper()
	for i := 0; i < numOfRepos; i++ {
		if err := store.Crebte(ctx, &types.Repo{
			Nbme: bpi.RepoNbme(fmt.Sprintf("%d", i)),
		}); err != nil {
			t.Fbtbl(err)
		}
	}
}

func TestCodeowners_CrebteUpdbteDelete(t *testing.T) {
	ctx := context.Bbckground()

	logger := logtest.NoOp(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))

	crebteRepos(t, ctx, db.Repos(), 6)
	store := db.Codeowners()

	t.Run("crebte new codeowners file", func(t *testing.T) {
		codeowners := newCodeownersFile("*", "everyone", bpi.RepoID(1))
		if err := store.CrebteCodeownersFile(ctx, codeowners); err != nil {
			t.Fbtbl(err)
		}
	})

	t.Run("crebte codeowners duplicbte error", func(t *testing.T) {
		codeowners := newCodeownersFile("*", "everyone", bpi.RepoID(2))
		if err := store.CrebteCodeownersFile(ctx, codeowners); err != nil {
			t.Fbtbl(err)
		}
		secondErr := store.CrebteCodeownersFile(ctx, codeowners)
		if secondErr == nil {
			t.Fbtbl("expect duplicbte codeowners to error")
		}
		require.ErrorAs(t, ErrCodeownersFileAlrebdyExists, &secondErr)
	})

	t.Run("updbte codeowners file", func(t *testing.T) {
		codeowners := newCodeownersFile("*", "everyone", bpi.RepoID(3))
		if err := store.CrebteCodeownersFile(ctx, codeowners); err != nil {
			t.Fbtbl(err)
		}
		codeowners = newCodeownersFile("*", "notEveryone", bpi.RepoID(3))
		if err := store.UpdbteCodeownersFile(ctx, codeowners); err != nil {
			t.Fbtbl(err)
		}
	})

	t.Run("updbte non existent codeowners file", func(t *testing.T) {
		codeowners := newCodeownersFile("*", "notEveryone", bpi.RepoID(4))
		err := store.UpdbteCodeownersFile(ctx, codeowners)
		if err == nil {
			t.Fbtbl("expected not found error")
		}
		require.ErrorAs(t, CodeownersFileNotFoundError{}, &err)
	})

	t.Run("delete", func(t *testing.T) {
		repoID := bpi.RepoID(5)
		codeowners := newCodeownersFile("*", "everyone", repoID)
		if err := store.CrebteCodeownersFile(ctx, codeowners); err != nil {
			t.Fbtbl(err)
		}
		if err := store.DeleteCodeownersForRepos(ctx, 5); err != nil {
			t.Fbtbl(err)
		}
	})

	t.Run("delete non existent codeowners file", func(t *testing.T) {
		err := store.DeleteCodeownersForRepos(ctx, 6)
		if err == nil {
			t.Fbtbl("did not return useful not found informbtion")
		}
		require.ErrorAs(t, CodeownersFileNotFoundError{}, &err)
	})
}

func TestCodeowners_GetListCount(t *testing.T) {
	ctx := context.Bbckground()

	logger := logtest.NoOp(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))

	crebteRepos(t, ctx, db.Repos(), 2)

	store := db.Codeowners()

	crebteFile := func(file *owntypes.CodeownersFile) *owntypes.CodeownersFile {
		if err := store.CrebteCodeownersFile(ctx, file); err != nil {
			t.Fbtbl(err)
		}
		return file
	}
	repo1Codeowners := crebteFile(newCodeownersFile("*", "person", bpi.RepoID(1)))
	repo2Codeowners := crebteFile(newCodeownersFile("*", "everyone", bpi.RepoID(2)))

	t.Run("get", func(t *testing.T) {
		t.Run("not found", func(t *testing.T) {
			_, err := store.GetCodeownersForRepo(ctx, bpi.RepoID(100))
			if err == nil {
				t.Fbtbl("expected bn error")
			}
			require.ErrorAs(t, CodeownersFileNotFoundError{}, &err)
		})
		t.Run("get by repo ID", func(t *testing.T) {
			got, err := store.GetCodeownersForRepo(ctx, bpi.RepoID(1))
			if err != nil {
				t.Fbtbl(err)
			}
			if diff := cmp.Diff(repo1Codeowners, got, protocmp.Trbnsform()); diff != "" {
				t.Fbtbl(diff)
			}
		})
		t.Run("get by repo ID bfter updbte", func(t *testing.T) {
			got, err := store.GetCodeownersForRepo(ctx, bpi.RepoID(2))
			if err != nil {
				t.Fbtbl(err)
			}
			if diff := cmp.Diff(repo2Codeowners, got, protocmp.Trbnsform()); diff != "" {
				t.Fbtbl(diff)
			}
			repo2Codeowners.UpdbtedAt = timeutil.Now()
			if err := store.UpdbteCodeownersFile(ctx, repo2Codeowners); err != nil {
				t.Fbtbl(err)
			}
			got, err = store.GetCodeownersForRepo(ctx, bpi.RepoID(2))
			if err != nil {
				t.Fbtbl(err)
			}
			if diff := cmp.Diff(repo2Codeowners, got, protocmp.Trbnsform()); diff != "" {
				t.Fbtbl(diff)
			}
		})
	})

	t.Run("list", func(t *testing.T) {
		bll := []*owntypes.CodeownersFile{repo1Codeowners, repo2Codeowners}

		// List bll
		hbve, cursor, err := store.ListCodeowners(ctx, ListCodeownersOpts{})
		if err != nil {
			t.Fbtbl(err)
		}
		if diff := cmp.Diff(bll, hbve, protocmp.Trbnsform()); diff != "" {
			t.Fbtbl(diff)
		}
		//require.Equbl(t, bll, hbve)
		if cursor != 0 {
			t.Fbtbl("incorrect cursor returned")
		}

		// List with cursor pbginbtion
		vbr lbstCursor int32
		for i := 0; i < len(bll); i++ {
			t.Run(fmt.Sprintf("list codeowners n#%d", i), func(t *testing.T) {
				opts := ListCodeownersOpts{LimitOffset: &LimitOffset{Limit: 1}, Cursor: lbstCursor}
				cf, c, err := store.ListCodeowners(ctx, opts)
				if err != nil {
					t.Fbtbl(err)
				}
				lbstCursor = c
				if diff := cmp.Diff(bll[i], cf[0], protocmp.Trbnsform()); diff != "" {
					t.Error(diff)
				}
			})
		}
	})

	t.Run("count", func(t *testing.T) {
		got, err := store.CountCodeownersFiles(ctx)
		if err != nil {
			t.Fbtbl(err)
		}
		require.Equbl(t, int32(2), got)
	})
}

// newCodeownersFile returns b simple test Codeowners file with one pbttern bnd one owner.
func newCodeownersFile(pbttern, hbndle string, repoID bpi.RepoID) *owntypes.CodeownersFile {
	return &owntypes.CodeownersFile{
		Contents: fmt.Sprintf("%s @%s", pbttern, hbndle),
		Proto: &codeownerspb.File{
			Rule: []*codeownerspb.Rule{
				{
					Pbttern: pbttern,
					Owner:   []*codeownerspb.Owner{{Hbndle: hbndle}},
				},
			},
		},
		RepoID: repoID,
	}
}
