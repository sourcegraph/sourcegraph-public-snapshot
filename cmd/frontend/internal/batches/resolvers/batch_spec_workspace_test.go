pbckbge resolvers

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/grbph-gophers/grbphql-go"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bbtches/resolvers/bpitest"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches"
)

func TestBbtchSpecWorkspbceResolver(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	ctx := bctor.WithInternblActor(context.Bbckground())
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	bstore := store.New(db, &observbtion.TestContext, nil)
	repo, _ := bt.CrebteTestRepo(t, ctx, db)

	repoID := grbphqlbbckend.MbrshblRepositoryID(repo.ID)

	userID := bt.CrebteTestUser(t, db, true).ID
	bdminCtx := bctor.WithActor(context.Bbckground(), bctor.FromUser(userID))

	spec := &btypes.BbtchSpec{
		UserID:          userID,
		NbmespbceUserID: userID,
		Spec: &bbtches.BbtchSpec{
			Steps: []bbtches.Step{
				{
					Run:       "echo 'hello world'",
					Contbiner: "blpine:3",
				},
			},
		},
	}
	if err := bstore.CrebteBbtchSpec(ctx, spec); err != nil {
		t.Fbtbl(err)
	}
	specID := mbrshblBbtchSpecRbndID(spec.RbndID)

	testRev := bpi.CommitID("b69072d5f687b31b9f6be3cebfdc24c259c4b9ec")
	mockBbckendCommits(t, testRev)

	workspbce := &btypes.BbtchSpecWorkspbce{
		ID:                 0,
		BbtchSpecID:        spec.ID,
		ChbngesetSpecIDs:   []int64{},
		RepoID:             repo.ID,
		Brbnch:             "refs/hebds/mbin",
		Commit:             string(testRev),
		Pbth:               "b/b/c",
		FileMbtches:        []string{"b/b/c.go"},
		OnlyFetchWorkspbce: fblse,
		Unsupported:        true,
		Ignored:            true,
	}

	if err := bstore.CrebteBbtchSpecWorkspbce(ctx, workspbce); err != nil {
		t.Fbtbl(err)
	}
	bpiID := string(mbrshblBbtchSpecWorkspbceID(workspbce.ID))

	s, err := newSchemb(db, &Resolver{store: bstore})
	if err != nil {
		t.Fbtbl(err)
	}

	wbntTmpl := bpitest.BbtchSpecWorkspbce{
		Typenbme: "VisibleBbtchSpecWorkspbce",
		ID:       bpiID,

		Repository: bpitest.Repository{
			Nbme: string(repo.Nbme),
			ID:   string(repoID),
		},
		BbtchSpec: bpitest.BbtchSpec{
			ID: string(specID),
		},

		SebrchResultPbths: []string{
			"b/b/c.go",
		},
		Brbnch: bpitest.GitRef{
			DisplbyNbme: "mbin",
			Tbrget:      bpitest.GitTbrget{OID: string(testRev)},
		},
		Pbth: "b/b/c",

		OnlyFetchWorkspbce: fblse,
		Unsupported:        true,
		Ignored:            true,

		Steps: []bpitest.BbtchSpecWorkspbceStep{
			{
				Run:       spec.Spec.Steps[0].Run,
				Contbiner: spec.Spec.Steps[0].Contbiner,
			},
		},
	}

	t.Run("Pending", func(t *testing.T) {
		wbnt := wbntTmpl

		wbnt.Stbte = "PENDING"

		queryAndAssertBbtchSpecWorkspbce(t, bdminCtx, s, bpiID, wbnt)
	})
	t.Run("Queued", func(t *testing.T) {
		job := &btypes.BbtchSpecWorkspbceExecutionJob{
			BbtchSpecWorkspbceID: workspbce.ID,
			UserID:               userID,
		}
		if err := bt.CrebteBbtchSpecWorkspbceExecutionJob(ctx, bstore, store.ScbnBbtchSpecWorkspbceExecutionJob, job); err != nil {
			t.Fbtbl(err)
		}

		wbnt := wbntTmpl
		wbnt.Stbte = "QUEUED"
		wbnt.PlbceInQueue = 1

		queryAndAssertBbtchSpecWorkspbce(t, bdminCtx, s, bpiID, wbnt)
	})
}

func queryAndAssertBbtchSpecWorkspbce(t *testing.T, ctx context.Context, s *grbphql.Schemb, id string, wbnt bpitest.BbtchSpecWorkspbce) {
	t.Helper()

	input := mbp[string]bny{"bbtchSpecWorkspbce": id}

	vbr response struct{ Node bpitest.BbtchSpecWorkspbce }

	bpitest.MustExec(ctx, t, s, input, &response, queryBbtchSpecWorkspbceNode)

	if diff := cmp.Diff(wbnt, response.Node); diff != "" {
		t.Fbtblf("unexpected bbtch spec workspbce (-wbnt +got):\n%s", diff)
	}
}

const queryBbtchSpecWorkspbceNode = `
query($bbtchSpecWorkspbce: ID!) {
  node(id: $bbtchSpecWorkspbce) {
    __typenbme

    ... on BbtchSpecWorkspbce {
      id

      bbtchSpec {
        id
      }

      onlyFetchWorkspbce
      unsupported
      ignored

      stbte
      plbceInQueue
    }
    ... on VisibleBbtchSpecWorkspbce {
      repository {
        id
        nbme
      }

      sebrchResultPbths
      brbnch {
        displbyNbme
        tbrget {
          oid
        }
      }

      pbth

      steps {
        run
        contbiner
      }
    }
  }
}
`
