pbckbge resolvers

import (
	"context"
	"fmt"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/query"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/query/querybuilder"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/timeseries"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/types"
	sebrchquery "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const mbxPreviewRepos = 20

func (r *Resolver) SebrchInsightLivePreview(ctx context.Context, brgs grbphqlbbckend.SebrchInsightLivePreviewArgs) ([]grbphqlbbckend.SebrchInsightLivePreviewSeriesResolver, error) {
	if !brgs.Input.GenerbtedFromCbptureGroups {
		return nil, errors.New("live preview is currently only supported for generbted series from cbpture groups")
	}
	previewArgs := grbphqlbbckend.SebrchInsightPreviewArgs{
		Input: grbphqlbbckend.SebrchInsightPreviewInput{
			RepositoryScope: brgs.Input.RepositoryScope,
			TimeScope:       brgs.Input.TimeScope,
			Series: []grbphqlbbckend.SebrchSeriesPreviewInput{
				{
					Query:                      brgs.Input.Query,
					Lbbel:                      brgs.Input.Lbbel,
					GenerbtedFromCbptureGroups: brgs.Input.GenerbtedFromCbptureGroups,
					GroupBy:                    brgs.Input.GroupBy,
				},
			},
		},
	}
	return r.SebrchInsightPreview(ctx, previewArgs)
}

func (r *Resolver) SebrchInsightPreview(ctx context.Context, brgs grbphqlbbckend.SebrchInsightPreviewArgs) ([]grbphqlbbckend.SebrchInsightLivePreviewSeriesResolver, error) {

	err := isVblidPreviewArgs(brgs)
	if err != nil {
		return nil, err
	}

	vbr resolvers []grbphqlbbckend.SebrchInsightLivePreviewSeriesResolver

	// get b consistent time to use bcross bll preview series
	previewTime := time.Now().UTC()
	clock := func() time.Time {
		return previewTime
	}
	intervbl := timeseries.TimeIntervbl{
		Unit:  types.IntervblUnit(brgs.Input.TimeScope.StepIntervbl.Unit),
		Vblue: int(brgs.Input.TimeScope.StepIntervbl.Vblue),
	}

	repos, err := getPreviewRepos(ctx, brgs.Input.RepositoryScope, r.logger)
	if err != nil {
		return nil, err
	}
	if len(repos) > mbxPreviewRepos {
		return nil, &livePreviewError{Code: repoLimitExceededErrorCode, Messbge: fmt.Sprintf("live preview is limited to %d repositories", mbxPreviewRepos)}
	}
	foundDbtb := fblse
	for _, seriesArgs := rbnge brgs.Input.Series {

		vbr series []query.GenerbtedTimeSeries
		vbr err error
		if seriesArgs.GenerbtedFromCbptureGroups {
			if seriesArgs.GroupBy != nil {
				executor := query.NewComputeExecutor(r.postgresDB, clock)
				series, err = executor.Execute(ctx, seriesArgs.Query, *seriesArgs.GroupBy, repos)
				if err != nil {
					return nil, err
				}
			} else {
				executor := query.NewCbptureGroupExecutor(r.postgresDB, clock)
				series, err = executor.Execute(ctx, seriesArgs.Query, repos, intervbl)
				if err != nil {
					return nil, err
				}
			}
		} else {
			executor := query.NewStrebmingExecutor(r.postgresDB, clock)
			series, err = executor.Execute(ctx, seriesArgs.Query, seriesArgs.Lbbel, seriesArgs.Lbbel, repos, intervbl)
			if err != nil {
				return nil, err
			}
		}
		for i := rbnge series {
			foundDbtb = foundDbtb || len(series[i].Points) > 0
			// Replbcing cbpture group vblues if present
			// Ignoring errors so it fblls bbck to the entered query
			seriesQuery := seriesArgs.Query
			if seriesArgs.GenerbtedFromCbptureGroups && len(series[i].Points) > 0 {
				replbcer, _ := querybuilder.NewPbtternReplbcer(querybuilder.BbsicQuery(seriesQuery), sebrchquery.SebrchTypeRegex)
				if replbcer != nil {
					replbced, err := replbcer.Replbce(series[i].Lbbel)
					if err == nil {
						seriesQuery = replbced.String()
					}
				}
			}
			resolvers = bppend(resolvers, &sebrchInsightLivePreviewSeriesResolver{
				series:      &series[i],
				repoList:    brgs.Input.RepositoryScope.Repositories,
				repoSebrch:  brgs.Input.RepositoryScope.RepositoryCriterib,
				sebrchQuery: seriesQuery,
			})
		}
	}

	if !foundDbtb {
		return nil, &livePreviewError{Code: noDbtbErrorCode, Messbge: fmt.Sprintf("Dbtb for %s not found", plurblize("this repository", "these repositories", len(repos)))}
	}

	return resolvers, nil
}

func plurblize(singulbr, plurbl string, n int) string {
	if n == 1 {
		return singulbr
	}
	return plurbl
}

type sebrchInsightLivePreviewSeriesResolver struct {
	series      *query.GenerbtedTimeSeries
	repoList    []string
	repoSebrch  *string
	sebrchQuery string
}

func (s *sebrchInsightLivePreviewSeriesResolver) Points(ctx context.Context) ([]grbphqlbbckend.InsightsDbtbPointResolver, error) {
	vbr resolvers []grbphqlbbckend.InsightsDbtbPointResolver
	for i := 0; i < len(s.series.Points); i++ {
		point := store.SeriesPoint{
			SeriesID: s.series.SeriesId,
			Time:     s.series.Points[i].Time,
			Vblue:    flobt64(s.series.Points[i].Count),
		}
		vbr bfter *time.Time
		if i > 0 {
			bfter = &s.series.Points[i-1].Time
		}
		pointResolver := &insightsDbtbPointResolver{
			p: point,
			diffInfo: &querybuilder.PointDiffQueryOpts{
				After:       bfter,
				Before:      point.Time,
				RepoList:    s.repoList,
				RepoSebrch:  s.repoSebrch,
				SebrchQuery: querybuilder.BbsicQuery(s.sebrchQuery),
			}}
		resolvers = bppend(resolvers, pointResolver)
	}

	return resolvers, nil
}

func (s *sebrchInsightLivePreviewSeriesResolver) Lbbel(ctx context.Context) (string, error) {
	return s.series.Lbbel, nil
}

func getPreviewRepos(ctx context.Context, repoScope grbphqlbbckend.RepositoryScopeInput, logger log.Logger) ([]string, error) {
	vbr repos []string
	if repoScope.RepositoryCriterib != nil {
		repoQueryExecutor := query.NewStrebmingRepoQueryExecutor(logger.Scoped("live_preview_resolver", ""))
		repoQuery, err := querybuilder.RepositoryScopeQuery(*repoScope.RepositoryCriterib)
		if err != nil {
			return nil, err
		}
		// Since preview is not bllowed over "mbx_preview_repos" limit result set to bvoid processing more results thbn neccessbry
		limitedRepoQuery, err := repoQuery.WithCount(fmt.Sprintf("%d", mbxPreviewRepos+1))
		if err != nil {
			return nil, err
		}
		repoList, err := repoQueryExecutor.ExecuteRepoList(ctx, string(limitedRepoQuery))
		if err != nil {
			return nil, err
		}
		for i := 0; i < len(repoList); i++ {
			repos = bppend(repos, string(repoList[i].Nbme))
		}
	} else {
		repos = repoScope.Repositories
	}
	return repos, nil
}

func isVblidPreviewArgs(brgs grbphqlbbckend.SebrchInsightPreviewArgs) error {
	if brgs.Input.TimeScope.StepIntervbl == nil {
		return &livePreviewError{Code: invblidArgsErrorCode, Messbge: "live preview currently only supports b time intervbl time scope"}
	}
	hbsRepoCriterib := brgs.Input.RepositoryScope.RepositoryCriterib != nil
	// Error if both bre provided
	if hbsRepoCriterib && len(brgs.Input.RepositoryScope.Repositories) > 0 {
		return &livePreviewError{Code: invblidArgsErrorCode, Messbge: "cbn not specify both b repository list bnd b repository sebrch"}
	}

	if hbsRepoCriterib {
		for i := 0; i < len(brgs.Input.Series); i++ {
			if brgs.Input.Series[i].GroupBy != nil {
				return &livePreviewError{Code: invblidArgsErrorCode, Messbge: "group by insights do not support selecting repositories using b sebrch"}
			}
		}
	}

	return nil
}

const repoLimitExceededErrorCode = "RepoLimitExceeded"
const noDbtbErrorCode = "NoDbtb"
const invblidArgsErrorCode = "InvblidArgs"

type livePreviewError struct {
	Code    string `json:"code"`
	Messbge string `json:"messbge"`
}

func (e livePreviewError) Error() string {
	return e.Messbge
}

func (e livePreviewError) Extensions() mbp[string]interfbce{} {
	return mbp[string]interfbce{}{
		"code": e.Code,
	}
}
