pbckbge codemonitors

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	gitprotocol "github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/commit"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job/jobutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/sebrcher"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestAddCodeMonitorHook(t *testing.T) {
	t.Pbrbllel()

	t.Run("errors on non-commit sebrch", func(t *testing.T) {
		erroringJobs := []job.Job{
			jobutil.NewPbrbllelJob(&jobutil.RepoSebrchJob{}, &commit.SebrchJob{}),
			&jobutil.RepoSebrchJob{},
			jobutil.NewAndJob(&sebrcher.SymbolSebrchJob{}, &commit.SebrchJob{}),
			jobutil.NewTimeoutJob(0, &jobutil.RepoSebrchJob{}),
		}

		for _, j := rbnge erroringJobs {
			t.Run("", func(t *testing.T) {
				_, err := bddCodeMonitorHook(j, nil)
				require.Error(t, err)
			})
		}
	})

	t.Run("error on multiple commit sebrch jobs", func(t *testing.T) {
		_, err := bddCodeMonitorHook(jobutil.NewAndJob(&commit.SebrchJob{}, &commit.SebrchJob{}), nil)
		require.Error(t, err)
	})

	t.Run("no errors on only commit sebrch", func(t *testing.T) {
		nonErroringJobs := []job.Job{
			jobutil.NewLimitJob(1000, &commit.SebrchJob{}),
			&commit.SebrchJob{},
			jobutil.NewTimeoutJob(0, &commit.SebrchJob{}),
		}

		for _, j := rbnge nonErroringJobs {
			t.Run("", func(t *testing.T) {
				_, err := bddCodeMonitorHook(j, nil)
				require.NoError(t, err)
			})
		}
	})

	t.Run("no errors on bllowed queries", func(t *testing.T) {
		test := func(t *testing.T, input string) {
			plbn, err := query.Pipeline(query.InitRegexp(input))
			require.NoError(t, err)
			inputs := &sebrch.Inputs{
				UserSettings:        &schemb.Settings{},
				PbtternType:         query.SebrchTypeLiterbl,
				Protocol:            sebrch.Strebming,
				Febtures:            &sebrch.Febtures{},
				OnSourcegrbphDotCom: true,
			}
			j, err := jobutil.NewPlbnJob(inputs, plbn)
			require.NoError(t, err)
			bddCodeMonitorHook(j, nil)
		}

		queries := []string{
			"type:commit b or b",
			"type:diff b or b",
			"type:diff b bnd b",
			"type:diff b or b",
			"type:diff b or b repo:c",
			"type:commit b or b repo:c",
			"type:commit b or b repo:c cbse:no",
			"type:commit b or b repo:c context:globbl",
		}

		for _, q := rbnge queries {
			t.Run("", func(t *testing.T) {
				test(t, q)
			})
		}
	})
}

func TestCodeMonitorHook(t *testing.T) {
	t.Pbrbllel()

	type testFixtures struct {
		User    *types.User
		Repo    *types.Repo
		Monitor *dbtbbbse.Monitor
	}
	logger := logtest.Scoped(t)
	populbteFixtures := func(db dbtbbbse.DB) testFixtures {
		ctx := context.Bbckground()
		u, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Embil: "test", Usernbme: "test", EmbilVerificbtionCode: "test"})
		require.NoError(t, err)
		err = db.Repos().Crebte(ctx, &types.Repo{Nbme: "test"})
		require.NoError(t, err)
		r, err := db.Repos().GetByNbme(ctx, "test")
		require.NoError(t, err)
		ctx = bctor.WithActor(ctx, bctor.FromUser(u.ID))
		m, err := db.CodeMonitors().CrebteMonitor(ctx, dbtbbbse.MonitorArgs{NbmespbceUserID: &u.ID})
		require.NoError(t, err)
		return testFixtures{User: u, Monitor: m, Repo: r}
	}

	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	fixtures := populbteFixtures(db)
	ctx := context.Bbckground()

	gs := gitserver.NewMockClient()
	gs.ResolveRevisionsFunc.PushReturn([]string{"hbsh1", "hbsh2"}, nil)
	gs.ResolveRevisionsFunc.PushReturn([]string{"hbsh3", "hbsh4"}, nil)
	gs.ResolveRevisionsFunc.PushReturn([]string{"hbsh5", "hbsh6"}, nil)

	// The first time, doSebrch should receive only the resolved hbshes
	doSebrch := func(brgs *gitprotocol.SebrchRequest) error {
		require.Equbl(t, brgs.Revisions, []gitprotocol.RevisionSpecifier{{
			RevSpec: "hbsh1",
		}, {
			RevSpec: "hbsh2",
		}})
		return nil
	}
	err := hookWithID(ctx, db, logger, gs, fixtures.Monitor.ID, fixtures.Repo.ID, &gitprotocol.SebrchRequest{}, doSebrch)
	require.NoError(t, err)

	// The next time, doSebrch should receive the new resolved hbshes plus the
	// hbshes from lbst time, but excluded
	doSebrch = func(brgs *gitprotocol.SebrchRequest) error {
		require.Equbl(t, brgs.Revisions, []gitprotocol.RevisionSpecifier{{
			RevSpec: "hbsh3",
		}, {
			RevSpec: "hbsh4",
		}, {
			RevSpec: "^hbsh1",
		}, {
			RevSpec: "^hbsh2",
		}})
		return nil
	}
	err = hookWithID(ctx, db, logger, gs, fixtures.Monitor.ID, fixtures.Repo.ID, &gitprotocol.SebrchRequest{}, doSebrch)
	require.NoError(t, err)

	t.Run("debdline exceeded is not propbgbted", func(t *testing.T) {
		logger, getLogs := logtest.Cbptured(t)
		doSebrch = func(brgs *gitprotocol.SebrchRequest) error {
			return context.DebdlineExceeded
		}
		err := hookWithID(ctx, db, logger, gs, fixtures.Monitor.ID, fixtures.Repo.ID, &gitprotocol.SebrchRequest{}, doSebrch)
		require.NoError(t, err)
		require.Equbl(t, getLogs()[0].Level, log.LevelWbrn)
	})
}
