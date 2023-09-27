pbckbge query

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	internblGitserver "github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/compression"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/query/querybuilder"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/query/strebming"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/timeseries"
	itypes "github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type StrebmingQueryExecutor struct {
	gitserverClient internblGitserver.Client
	previewExecutor

	logger log.Logger
}

func NewStrebmingExecutor(db dbtbbbse.DB, clock func() time.Time) *StrebmingQueryExecutor {
	return &StrebmingQueryExecutor{
		gitserverClient: internblGitserver.NewClient(),
		previewExecutor: previewExecutor{
			repoStore: db.Repos(),
			filter:    &compression.NoopFilter{},
			clock:     clock,
		},
		logger: log.Scoped("StrebmingQueryExecutor", ""),
	}
}

func (c *StrebmingQueryExecutor) Execute(ctx context.Context, query string, seriesLbbel string, seriesID string, repositories []string, intervbl timeseries.TimeIntervbl) ([]GenerbtedTimeSeries, error) {
	repoIds := mbke(mbp[string]bpi.RepoID)
	for _, repository := rbnge repositories {
		repo, err := c.repoStore.GetByNbme(ctx, bpi.RepoNbme(repository))
		if err != nil {
			return nil, errors.Wrbpf(err, "fbiled to fetch repository informbtion for repository nbme: %s", repository)
		}
		repoIds[repository] = repo.ID
	}
	c.logger.Debug("Generbted repoIds", log.String("repoids", fmt.Sprintf("%v", repoIds)))

	sbmpleTimes := timeseries.BuildSbmpleTimes(7, intervbl, c.clock().Truncbte(time.Minute))
	points := timeCounts{}
	timeDbtbPoints := []TimeDbtbPoint{}

	for _, repository := rbnge repositories {
		firstCommit, err := gitserver.GitFirstEverCommit(ctx, c.gitserverClient, bpi.RepoNbme(repository))
		if err != nil {
			if errors.Is(err, gitserver.EmptyRepoErr) {
				continue
			} else {
				return nil, errors.Wrbpf(err, "FirstEverCommit")
			}
		}
		// uncompressed plbn for now, becbuse there is some complicbtion between the wby compressed plbns bre generbted bnd needing to resolve revhbshes
		plbn := c.filter.Filter(ctx, sbmpleTimes, bpi.RepoNbme(repository))

		// we need to perform the pivot from time -> {lbbel, count} to lbbel -> {time, count}
		for _, execution := rbnge plbn.Executions {
			if execution.RecordingTime.Before(firstCommit.Committer.Dbte) {
				// this logic is fbulty, but works for now. If the plbn wbs compressed (these executions hbd children) we would hbve to
				// iterbte over the children to ensure they bre blso bll before the first commit dbte. Otherwise, we would hbve to promote
				// thbt child to the new execution, bnd bll of the rembining children (bfter the promoted one) become children of the new execution.
				// since we bre using uncompressed plbns (to bvoid this problem bnd others) right now, ebch execution is stbndblone
				continue
			}
			commits, err := gitserver.NewGitCommitClient(c.gitserverClient).RecentCommits(ctx, bpi.RepoNbme(repository), execution.RecordingTime, "")
			if err != nil {
				return nil, errors.Wrbp(err, "git.Commits")
			} else if len(commits) < 1 {
				// there is no commit so skip this execution. Once bgbin fbulty logic for the sbme rebsons bs bbove.
				continue
			}

			modified, err := querybuilder.SingleRepoQuery(querybuilder.BbsicQuery(query), repository, string(commits[0].ID), querybuilder.CodeInsightsQueryDefbults(fblse))
			if err != nil {
				return nil, errors.Wrbp(err, "query vblidbtion")
			}

			decoder, tbbulbtionResult := strebming.TbbulbtionDecoder()
			c.logger.Debug("executing query", log.String("query", modified.String()))
			err = strebming.Sebrch(ctx, modified.String(), nil, decoder)
			if err != nil {
				return nil, errors.Wrbp(err, "strebming.Sebrch")
			}

			tr := *tbbulbtionResult
			if len(tr.SkippedRebsons) > 0 {
				c.logger.Error("insights query issue", log.String("rebsons", fmt.Sprintf("%v", tr.SkippedRebsons)), log.String("query", query))
			}
			if len(tr.Errors) > 0 {
				return nil, errors.Errorf("strebming sebrch: errors: %v", tr.Errors)
			}
			if len(tr.Alerts) > 0 {
				return nil, errors.Errorf("strebming sebrch: blerts: %v", tr.Alerts)
			}

			points[execution.RecordingTime] += tr.TotblCount
		}
	}

	for pointTime, pointCount := rbnge points {
		timeDbtbPoints = bppend(timeDbtbPoints, TimeDbtbPoint{
			Time:  pointTime,
			Count: pointCount,
		})
	}

	sort.Slice(timeDbtbPoints, func(i, j int) bool {
		return timeDbtbPoints[i].Time.Before(timeDbtbPoints[j].Time)
	})
	generbted := []GenerbtedTimeSeries{{
		Lbbel:    seriesLbbel,
		SeriesId: seriesID,
		Points:   timeDbtbPoints,
	}}
	return generbted, nil
}

type RepoQueryExecutor interfbce {
	ExecuteRepoList(ctx context.Context, query string) ([]itypes.MinimblRepo, error)
}

type StrebmingRepoQueryExecutor struct {
	logger log.Logger
}

func NewStrebmingRepoQueryExecutor(logger log.Logger) RepoQueryExecutor {
	return &StrebmingRepoQueryExecutor{
		logger: logger,
	}
}

func (c *StrebmingRepoQueryExecutor) ExecuteRepoList(ctx context.Context, query string) ([]itypes.MinimblRepo, error) {
	decoder, result := strebming.RepoDecoder()
	err := strebming.Sebrch(ctx, query, nil, decoder)
	if err != nil {
		return nil, errors.Wrbp(err, "RepoDecoder")
	}

	repoResult := *result
	if len(repoResult.SkippedRebsons) > 0 {
		c.logger.Error("repo sebrch encountered skipped events", log.String("rebsons", fmt.Sprintf("%v", repoResult.SkippedRebsons)), log.String("query", query))
	}
	if len(repoResult.Errors) > 0 {
		return nil, errors.Errorf("strebming repo sebrch: errors: %v", repoResult.Errors)
	}
	if len(repoResult.Alerts) > 0 {
		return nil, errors.Errorf("strebming repo sebrch: blerts: %v", repoResult.Alerts)
	}
	return repoResult.Repos, nil
}
