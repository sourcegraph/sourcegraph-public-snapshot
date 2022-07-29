package resolvers

import (
	"context"
	"sort"
	"strings"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query/querybuilder"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query/streaming"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var _ graphqlbackend.RelatedInsightsInlineResolver = &relatedInsightsInlineResolver{}
var _ graphqlbackend.RelatedInsightsResolver = &relatedInsightsResolver{}

func (r *Resolver) RelatedInsightsInline(ctx context.Context, args graphqlbackend.RelatedInsightsArgs) ([]graphqlbackend.RelatedInsightsInlineResolver, error) {
	validator := PermissionsValidatorFromBase(&r.baseInsightResolver)
	validator.loadUserContext(ctx)

	allSeries, err := r.insightStore.GetAll(ctx, store.InsightQueryArgs{
		Repo:   args.Input.Repo,
		UserID: validator.userIds,
		OrgID:  validator.orgIds,
	})
	if err != nil {
		return nil, errors.Wrap(err, "GetAll")
	}
	allSeries = limitSeries(allSeries)

	seriesMatches := map[string]*relatedInsightInlineMetadata{}
	for _, series := range allSeries {
		decoder, metadataResult := streaming.MetadataDecoder()
		modifiedQuery, err := querybuilder.SingleFileQuery(querybuilder.BasicQuery(series.Query), args.Input.Repo, args.Input.File, args.Input.Revision, querybuilder.CodeInsightsQueryDefaults(false))
		if err != nil {
			return nil, errors.Wrap(err, "SingleFileQuery")
		}
		err = streaming.Search(ctx, modifiedQuery.String(), decoder)
		if err != nil {
			return nil, errors.Wrap(err, "streaming.Search")
		}
		mr := *metadataResult
		if len(mr.Errors) > 0 {
			r.logger.Warn("related insights errors", log.Strings("errors", mr.Errors))
		}
		if len(mr.Alerts) > 0 {
			r.logger.Warn("related insights alerts", log.Strings("alerts", mr.Alerts))
		}
		if len(mr.SkippedReasons) > 0 {
			r.logger.Warn("related insights skipped", log.Strings("reasons", mr.SkippedReasons))
		}

		for _, match := range mr.Matches {
			for _, lineMatch := range match.LineMatches {
				if seriesMatches[series.UniqueID] == nil {
					seriesMatches[series.UniqueID] = &relatedInsightInlineMetadata{title: series.Title, lineNumbers: []int32{lineMatch.LineNumber}}
				} else if !containsInt(seriesMatches[series.UniqueID].lineNumbers, lineMatch.LineNumber) {
					seriesMatches[series.UniqueID].lineNumbers = append(seriesMatches[series.UniqueID].lineNumbers, lineMatch.LineNumber)
				}
			}
		}
	}

	var resolvers []graphqlbackend.RelatedInsightsInlineResolver
	for insightId, metadata := range seriesMatches {
		sort.SliceStable(metadata.lineNumbers, func(i, j int) bool {
			return metadata.lineNumbers[i] < metadata.lineNumbers[j]
		})
		resolvers = append(resolvers, &relatedInsightsInlineResolver{viewID: insightId, title: metadata.title, lineNumbers: metadata.lineNumbers})
	}
	return resolvers, nil
}

type relatedInsightInlineMetadata struct {
	title       string
	lineNumbers []int32
}

type relatedInsightsInlineResolver struct {
	viewID      string
	title       string
	lineNumbers []int32

	baseInsightResolver
}

func (r *relatedInsightsInlineResolver) ViewID() string {
	return r.viewID
}

func (r *relatedInsightsInlineResolver) Title() string {
	return r.title
}

func (r *relatedInsightsInlineResolver) LineNumbers() []int32 {
	return r.lineNumbers
}

func (r *Resolver) RelatedInsightsForFile(ctx context.Context, args graphqlbackend.RelatedInsightsArgs) ([]graphqlbackend.RelatedInsightsResolver, error) {
	validator := PermissionsValidatorFromBase(&r.baseInsightResolver)
	validator.loadUserContext(ctx)

	allSeries, err := r.insightStore.GetAll(ctx, store.InsightQueryArgs{
		Repo:                     args.Input.Repo,
		ContainingQuerySubstring: "select:file",
		UserID:                   validator.userIds,
		OrgID:                    validator.orgIds,
	})
	if err != nil {
		return nil, errors.Wrap(err, "GetAll")
	}
	allSeries = limitSeries(allSeries)

	var resolvers []graphqlbackend.RelatedInsightsResolver
	matchedInsightViews := map[string]bool{}
	for _, series := range allSeries {
		// We stop processing if we have matched on this insight view before.
		if _, ok := matchedInsightViews[series.UniqueID]; ok {
			continue
		}

		decoder, metadataResult := streaming.MetadataDecoder()
		modifiedQuery, err := querybuilder.SingleFileQuery(querybuilder.BasicQuery(series.Query), args.Input.Repo, args.Input.File, args.Input.Revision, querybuilder.CodeInsightsQueryDefaults(false))
		if err != nil {
			return nil, errors.Wrap(err, "SingleFileQuery")
		}
		err = streaming.Search(ctx, modifiedQuery.String(), decoder)
		if err != nil {
			return nil, errors.Wrap(err, "streaming.Search")
		}
		mr := *metadataResult
		if len(mr.Errors) > 0 {
			r.logger.Warn("file related insights errors", log.Strings("errors", mr.Errors))
		}
		if len(mr.Alerts) > 0 {
			r.logger.Warn("file related insights alerts", log.Strings("alerts", mr.Alerts))
		}
		if len(mr.SkippedReasons) > 0 {
			r.logger.Warn("file related insights skipped", log.Strings("reasons", mr.SkippedReasons))
		}

		for _, match := range mr.Matches {
			if len(match.Path) > 0 && strings.EqualFold(match.Path, args.Input.File) {
				matchedInsightViews[series.UniqueID] = true
				resolvers = append(resolvers, &relatedInsightsResolver{viewID: series.UniqueID, title: series.Title})
			}
		}
	}

	return resolvers, nil
}

func (r *Resolver) RelatedInsightsForRepo(ctx context.Context, args graphqlbackend.RelatedInsightsRepoArgs) ([]graphqlbackend.RelatedInsightsResolver, error) {
	validator := PermissionsValidatorFromBase(&r.baseInsightResolver)
	validator.loadUserContext(ctx)

	allSeries, err := r.insightStore.GetAll(ctx, store.InsightQueryArgs{
		Repo:                     args.Input.Repo,
		ContainingQuerySubstring: "select:repo",
		UserID:                   validator.userIds,
		OrgID:                    validator.orgIds,
	})
	if err != nil {
		return nil, errors.Wrap(err, "GetAll")
	}
	allSeries = limitSeries(allSeries)

	var resolvers []graphqlbackend.RelatedInsightsResolver
	matchedInsightViews := map[string]bool{}
	for _, series := range allSeries {
		// We stop processing if we have matched on this insight view before.
		if _, ok := matchedInsightViews[series.UniqueID]; ok {
			continue
		}

		decoder, metadataResult := streaming.MetadataDecoder()
		modifiedQuery, err := querybuilder.SingleRepoQuery(querybuilder.BasicQuery(series.Query), args.Input.Repo, args.Input.Revision, querybuilder.CodeInsightsQueryDefaults(false))
		if err != nil {
			return nil, errors.Wrap(err, "SingleRepoQuery")
		}
		err = streaming.Search(ctx, modifiedQuery.String(), decoder)
		if err != nil {
			return nil, errors.Wrap(err, "streaming.Search")
		}
		mr := *metadataResult
		if len(mr.Errors) > 0 {
			r.logger.Warn("file related insights errors", log.Strings("errors", mr.Errors))
		}
		if len(mr.Alerts) > 0 {
			r.logger.Warn("file related insights alerts", log.Strings("alerts", mr.Alerts))
		}
		if len(mr.SkippedReasons) > 0 {
			r.logger.Warn("file related insights skipped", log.Strings("reasons", mr.SkippedReasons))
		}

		for _, match := range mr.Matches {
			if len(match.RepositoryName) > 0 && strings.EqualFold(match.RepositoryName, args.Input.Repo) {
				matchedInsightViews[series.UniqueID] = true
				resolvers = append(resolvers, &relatedInsightsResolver{viewID: series.UniqueID, title: series.Title})
			}
		}
	}

	return resolvers, nil
}

type relatedInsightsResolver struct {
	viewID string
	title  string

	baseInsightResolver
}

func (r *relatedInsightsResolver) ViewID() string {
	return r.viewID
}

func (r *relatedInsightsResolver) Title() string {
	return r.title
}

func containsInt(array []int32, findElement int32) bool {
	for _, currentElement := range array {
		if findElement == currentElement {
			return true
		}
	}
	return false
}

// Limiting the number of series/queries to 50 will have no impact on the vast majority of customers.
// However, our own test environments have hundreds of insights and we need to limit these queries in some way
// so that the endpoints do not time out.
// This returns the 50 most recent series.
func limitSeries(series []types.InsightViewSeries) []types.InsightViewSeries {
	sort.SliceStable(series, func(i, j int) bool {
		return series[i].CreatedAt.After(series[j].CreatedAt)
	})
	return series[:minInt(50, int32(len(series)))]
}
