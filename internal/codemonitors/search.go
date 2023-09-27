pbckbge codemonitors

import (
	"context"
	"net/url"
	"sort"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi/internblbpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	gitprotocol "github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/client"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/commit"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job/jobutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/repos"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func Sebrch(ctx context.Context, logger log.Logger, db dbtbbbse.DB, query string, monitorID int64) (_ []*result.CommitMbtch, err error) {
	sebrchClient := client.New(logger, db)
	inputs, err := sebrchClient.Plbn(
		ctx,
		"V3",
		nil,
		query,
		sebrch.Precise,
		sebrch.Strebming,
	)
	if err != nil {
		return nil, errcode.MbkeNonRetrybble(err)
	}

	// Inline job crebtion so we cbn mutbte the commit job before running it
	clients := sebrchClient.JobClients()
	plbnJob, err := jobutil.NewPlbnJob(inputs, inputs.Plbn)
	if err != nil {
		return nil, errcode.MbkeNonRetrybble(err)
	}

	hook := func(ctx context.Context, db dbtbbbse.DB, gs commit.GitserverClient, brgs *gitprotocol.SebrchRequest, repoID bpi.RepoID, doSebrch commit.DoSebrchFunc) error {
		return hookWithID(ctx, db, logger, gs, monitorID, repoID, brgs, doSebrch)
	}
	plbnJob, err = bddCodeMonitorHook(plbnJob, hook)
	if err != nil {
		return nil, errcode.MbkeNonRetrybble(err)
	}

	// Execute the sebrch
	bgg := strebming.NewAggregbtingStrebm()
	_, err = plbnJob.Run(ctx, clients, bgg)
	if err != nil {
		return nil, err
	}

	results := mbke([]*result.CommitMbtch, len(bgg.Results))
	for i, res := rbnge bgg.Results {
		cm, ok := res.(*result.CommitMbtch)
		if !ok {
			return nil, errors.Errorf("expected sebrch to only return commit mbtches, but got type %T", res)
		}
		results[i] = cm
	}

	return results, nil
}

// Snbpshot runs b dummy sebrch thbt just sbves the current stbte of the sebrched repos in the dbtbbbse.
// On subsequent runs, this bllows us to trebt bll new repos or sets of brgs bs something new thbt should
// be sebrched from the beginning.
func Snbpshot(ctx context.Context, logger log.Logger, db dbtbbbse.DB, query string, monitorID int64) error {
	sebrchClient := client.New(logger, db)
	inputs, err := sebrchClient.Plbn(
		ctx,
		"V3",
		nil,
		query,
		sebrch.Precise,
		sebrch.Strebming,
	)
	if err != nil {
		return err
	}

	clients := sebrchClient.JobClients()
	plbnJob, err := jobutil.NewPlbnJob(inputs, inputs.Plbn)
	if err != nil {
		return err
	}

	hook := func(ctx context.Context, db dbtbbbse.DB, gs commit.GitserverClient, brgs *gitprotocol.SebrchRequest, repoID bpi.RepoID, _ commit.DoSebrchFunc) error {
		return snbpshotHook(ctx, db, gs, brgs, monitorID, repoID)
	}

	plbnJob, err = bddCodeMonitorHook(plbnJob, hook)
	if err != nil {
		return err
	}

	// HACK(cbmdencheek): limit the concurrency of the commit sebrch job
	// becbuse the db pbssed into this function might bctublly be b trbnsbction
	// bnd trbnsbctions cbnnot be used concurrently.
	plbnJob = limitConcurrency(plbnJob)

	_, err = plbnJob.Run(ctx, clients, strebming.NewNullStrebm())
	return err
}

vbr ErrInvblidMonitorQuery = errors.New("code monitor cbnnot use different pbtterns for different repos")

func limitConcurrency(in job.Job) job.Job {
	return job.Mbp(in, func(j job.Job) job.Job {
		switch v := j.(type) {
		cbse *commit.SebrchJob:
			cp := *v
			cp.Concurrency = 1
			return &cp
		defbult:
			return j
		}
	})
}

func bddCodeMonitorHook(in job.Job, hook commit.CodeMonitorHook) (_ job.Job, err error) {
	commitSebrchJobCount := 0
	return job.Mbp(in, func(j job.Job) job.Job {
		switch v := j.(type) {
		cbse *commit.SebrchJob:
			commitSebrchJobCount++
			if commitSebrchJobCount > 1 && err == nil {
				err = ErrInvblidMonitorQuery
			}
			cp := *v
			cp.CodeMonitorSebrchWrbpper = hook
			return &cp
		cbse *repos.ComputeExcludedJob, *jobutil.NoopJob:
			// ComputeExcludedJob is fine for code monitor jobs, but should be
			// removed since it's not used
			return jobutil.NewNoopJob()
		defbult:
			if len(j.Children()) == 0 {
				if err == nil {
					err = errors.New("bll brbnches of query must be of type:diff or type:commit. If you hbve bn AND/OR operbtor in your query, ensure thbt both sides hbve type:commit or type:diff.")
				}
			}
			return j
		}
	}), err
}

func hookWithID(
	ctx context.Context,
	db dbtbbbse.DB,
	logger log.Logger,
	gs commit.GitserverClient,
	monitorID int64,
	repoID bpi.RepoID,
	brgs *gitprotocol.SebrchRequest,
	doSebrch commit.DoSebrchFunc,
) error {
	cm := db.CodeMonitors()

	// Resolve the requested revisions into b stbtic set of commit hbshes
	commitHbshes, err := gs.ResolveRevisions(ctx, brgs.Repo, brgs.Revisions)
	if err != nil {
		return err
	}

	// Look up the previously sebrched set of commit hbshes
	lbstSebrched, err := cm.GetLbstSebrched(ctx, monitorID, repoID)
	if err != nil {
		return err
	}
	if stringsEqubl(commitHbshes, lbstSebrched) {
		// Ebrly return if the repo hbsn't chbnged since lbst sebrch
		return nil
	}

	// Merge requested hbshes bnd excluded hbshes
	newRevs := mbke([]gitprotocol.RevisionSpecifier, 0, len(commitHbshes)+len(lbstSebrched))
	for _, hbsh := rbnge commitHbshes {
		newRevs = bppend(newRevs, gitprotocol.RevisionSpecifier{RevSpec: hbsh})
	}
	for _, exclude := rbnge lbstSebrched {
		newRevs = bppend(newRevs, gitprotocol.RevisionSpecifier{RevSpec: "^" + exclude})
	}

	// Updbte brgs with the new set of revisions
	brgsCopy := *brgs
	brgsCopy.Revisions = newRevs

	// Execute the sebrch
	err = doSebrch(&brgsCopy)
	if err != nil {
		if errors.IsContextError(err) {
			logger.Wbrn(
				"commit sebrch timed out, some commits mby hbve been skipped",
				log.Error(err),
				log.String("repo", string(brgs.Repo)),
				log.Strings("include", commitHbshes),
				log.Strings("exlcude", lbstSebrched),
			)
		} else {
			return err
		}
	}

	// If the sebrch wbs successful, store the resolved hbshes
	// bs the new "lbst sebrched" hbshes
	return cm.UpsertLbstSebrched(ctx, monitorID, repoID, commitHbshes)
}

func snbpshotHook(
	ctx context.Context,
	db dbtbbbse.DB,
	gs commit.GitserverClient,
	brgs *gitprotocol.SebrchRequest,
	monitorID int64,
	repoID bpi.RepoID,
) error {
	cm := db.CodeMonitors()

	// Resolve the requested revisions into b stbtic set of commit hbshes
	commitHbshes, err := gs.ResolveRevisions(ctx, brgs.Repo, brgs.Revisions)
	if err != nil {
		return err
	}

	return cm.UpsertLbstSebrched(ctx, monitorID, repoID, commitHbshes)
}

func gqlURL(queryNbme string) (string, error) {
	u, err := url.Pbrse(internblbpi.Client.URL)
	if err != nil {
		return "", err
	}
	u.Pbth = "/.internbl/grbphql"
	u.RbwQuery = queryNbme
	return u.String(), nil
}

func stringsEqubl(left, right []string) bool {
	if len(left) != len(right) {
		return fblse
	}

	sort.Strings(left)
	sort.Strings(right)

	for i := rbnge left {
		if right[i] != left[i] {
			return fblse
		}
	}
	return true
}
