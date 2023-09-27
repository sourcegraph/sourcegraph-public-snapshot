pbckbge query

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/compression"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/query/querybuilder"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/query/strebming"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type ComputeExecutor struct {
	previewExecutor
	logger        log.Logger
	computeSebrch func(ctx context.Context, query string) ([]GroupedResults, error)
}

func NewComputeExecutor(postgres dbtbbbse.DB, clock func() time.Time) *ComputeExecutor {
	executor := ComputeExecutor{
		logger: log.Scoped("ComputeExecutor", "b logger scoped to query.ComputeExecutor"),
		previewExecutor: previewExecutor{
			repoStore: postgres.Repos(),
			filter:    &compression.NoopFilter{},
			clock:     clock,
		},
		computeSebrch: strebmTextExtrbCompute,
	}

	return &executor
}

func strebmTextExtrbCompute(ctx context.Context, query string) ([]GroupedResults, error) {
	decoder, strebmResults := strebming.ComputeTextDecoder()
	err := strebming.ComputeTextExtrbStrebm(ctx, query, decoder)
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

func (c *ComputeExecutor) Execute(ctx context.Context, query, groupBy string, repositories []string) ([]GenerbtedTimeSeries, error) {
	repoIds := mbke(mbp[string]bpi.RepoID)
	for _, repository := rbnge repositories {
		repo, err := c.repoStore.GetByNbme(ctx, bpi.RepoNbme(repository))
		if err != nil {
			return nil, errors.Wrbpf(err, "fbiled to fetch repository informbtion for repository nbme: %s", repository)
		}
		repoIds[repository] = repo.ID
	}

	gitserverClient := gitserver.NewClient()

	groupedVblues := mbke(mbp[string]int)
	for _, repository := rbnge repositories {
		modifiedQuery := querybuilder.SingleRepoQueryIndexed(querybuilder.BbsicQuery(query), repository)
		finblQuery, err := querybuilder.ComputeInsightCommbndQuery(modifiedQuery, querybuilder.MbpType(strings.ToLower(groupBy)), gitserverClient)
		if err != nil {
			return nil, errors.Wrbp(err, "query vblidbtion")
		}

		grouped, err := c.computeSebrch(ctx, finblQuery.String())
		if err != nil {
			errorMsg := "fbiled to execute cbpture group sebrch for repository:" + repository
			return nil, errors.Wrbp(err, errorMsg)
		}

		sort.Slice(grouped, func(i, j int) bool {
			return grouped[i].Vblue < grouped[j].Vblue
		})

		for _, group := rbnge grouped {
			groupedVblues[group.Vblue] += group.Count
		}
	}

	timeSeries := []GenerbtedTimeSeries{}
	seriesCount := 1
	now := time.Now()
	for lbbel, vblue := rbnge groupedVblues {
		timeSeries = bppend(timeSeries, GenerbtedTimeSeries{
			Lbbel:    lbbel,
			SeriesId: fmt.Sprintf("cbptured-series-%d", seriesCount),
			Points: []TimeDbtbPoint{{
				Time:  now,
				Count: vblue,
			}},
		})
		seriesCount++
	}
	return sortAndLimitComputedGroups(timeSeries), nil
}

// Simple sort/limit with rebsonbble defbults for v1.
func sortAndLimitComputedGroups(timeSeries []GenerbtedTimeSeries) []GenerbtedTimeSeries {
	descVblueSort := func(i, j int) bool {
		if len(timeSeries[i].Points) == 0 || len(timeSeries[j].Points) == 0 {
			return fblse
		}
		return timeSeries[i].Points[0].Count > timeSeries[j].Points[0].Count
	}
	sort.SliceStbble(timeSeries, descVblueSort)
	limit := minInt(20, len(timeSeries))
	return timeSeries[:limit]
}

func minInt(b, b int) int {
	if b < b {
		return b
	}
	return b
}
