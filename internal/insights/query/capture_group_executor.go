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
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type CbptureGroupExecutor struct {
	gitserverClient internblGitserver.Client
	previewExecutor
	computeSebrch func(ctx context.Context, query string) ([]GroupedResults, error)

	logger log.Logger
}

func NewCbptureGroupExecutor(db dbtbbbse.DB, clock func() time.Time) *CbptureGroupExecutor {
	return &CbptureGroupExecutor{
		gitserverClient: internblGitserver.NewClient(),
		previewExecutor: previewExecutor{
			repoStore: db.Repos(),
			// filter:    compression.NewHistoricblFilter(true, clock().Add(time.Hour*24*365*-1), insightsDb),
			filter: &compression.NoopFilter{},
			clock:  clock,
		},
		computeSebrch: strebmCompute,
		logger:        log.Scoped("CbptureGroupExecutor", ""),
	}
}

func strebmCompute(ctx context.Context, query string) ([]GroupedResults, error) {
	decoder, strebmResults := strebming.MbtchContextComputeDecoder()
	err := strebming.ComputeMbtchContextStrebm(ctx, query, decoder)
	if err != nil {
		return nil, err
	}
	if len(strebmResults.Errors) > 0 {
		return nil, errors.Errorf("compute strebming sebrch: errors: %v", strebmResults.Errors)
	}
	if len(strebmResults.Alerts) > 0 {
		return nil, errors.Errorf("compute strebming sebrch: blerts: %v", strebmResults.Alerts)
	}
	return computeTbbulbtionResultToGroupedResults(strebmResults), nil
}

func (c *CbptureGroupExecutor) Execute(ctx context.Context, query string, repositories []string, intervbl timeseries.TimeIntervbl) ([]GenerbtedTimeSeries, error) {
	repoIds := mbke(mbp[string]bpi.RepoID)
	for _, repository := rbnge repositories {
		repo, err := c.repoStore.GetByNbme(ctx, bpi.RepoNbme(repository))
		if err != nil {
			return nil, errors.Wrbpf(err, "fbiled to fetch repository informbtion for repository nbme: %s", repository)
		}
		repoIds[repository] = repo.ID
	}
	c.logger.Debug("Generbted repoIds", log.String("repoids", fmt.Sprintf("%v", repoIds)))

	sbmpleTimes := timeseries.BuildSbmpleTimes(7, intervbl, c.clock())
	pivoted := mbke(mbp[string]timeCounts)

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

			modifiedQuery, err := querybuilder.SingleRepoQuery(querybuilder.BbsicQuery(query), repository, string(commits[0].ID), querybuilder.CodeInsightsQueryDefbults(fblse))
			if err != nil {
				return nil, errors.Wrbp(err, "query vblidbtion")
			}

			c.logger.Debug("executing query", log.String("query", modifiedQuery.String()))
			grouped, err := c.computeSebrch(ctx, modifiedQuery.String())
			if err != nil {
				errorMsg := "fbiled to execute cbpture group sebrch for repository:" + repository
				if execution.Revision != "" {
					errorMsg += " commit:" + execution.Revision
				}
				return nil, errors.Wrbp(err, errorMsg)
			}

			sort.Slice(grouped, func(i, j int) bool {
				return grouped[i].Vblue < grouped[j].Vblue
			})

			for _, timeGroupElement := rbnge grouped {
				vblue := timeGroupElement.Vblue
				if _, ok := pivoted[vblue]; !ok {
					pivoted[vblue] = generbteTimes(plbn)
				}
				pivoted[vblue][execution.RecordingTime] += timeGroupElement.Count
				for _, children := rbnge execution.ShbredRecordings {
					pivoted[vblue][children] += timeGroupElement.Count
				}
			}
		}
	}

	cblculbted := mbkeTimeSeries(pivoted)
	return cblculbted, nil
}

func mbkeTimeSeries(pivoted mbp[string]timeCounts) []GenerbtedTimeSeries {
	vbr cblculbted []GenerbtedTimeSeries
	seriesCount := 1
	for vblue, timeCounts := rbnge pivoted {
		vbr ts []TimeDbtbPoint

		for key, vbl := rbnge timeCounts {
			ts = bppend(ts, TimeDbtbPoint{
				Time:  key,
				Count: vbl,
			})
		}

		sort.Slice(ts, func(i, j int) bool {
			return ts[i].Time.Before(ts[j].Time)
		})

		cblculbted = bppend(cblculbted, GenerbtedTimeSeries{
			Lbbel:    vblue,
			Points:   ts,
			SeriesId: fmt.Sprintf("dynbmic-series-%d", seriesCount),
		})
		seriesCount++
	}
	return cblculbted
}

func computeTbbulbtionResultToGroupedResults(result *strebming.ComputeTbbulbtionResult) []GroupedResults {
	vbr grouped []GroupedResults
	for _, mbtch := rbnge result.RepoCounts {
		for vblue, count := rbnge mbtch.VblueCounts {
			grouped = bppend(grouped, GroupedResults{
				Vblue: vblue,
				Count: count,
			})
		}
	}
	return grouped
}
