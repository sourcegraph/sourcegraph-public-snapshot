package resolvers

import (
	"context"
	"sort"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query/querybuilder"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query/streaming"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var _ graphqlbackend.RelatedInsightsInlineResolver = &relatedInsightsInlineResolver{}
var _ graphqlbackend.RelatedInsightsForFileResolver = &relatedInsightsForFileResolver{}

func (r *Resolver) RelatedInsightsInline(ctx context.Context, args graphqlbackend.RelatedInsightsArgs) ([]graphqlbackend.RelatedInsightsInlineResolver, error) {
	validator := PermissionsValidatorFromBase(&r.baseInsightResolver)
	validator.loadUserContext(ctx)

	allSeries, err := r.insightStore.GetAll(ctx, store.InsightQueryArgs{
		Repo:   &args.Input.Repo,
		UserID: validator.userIds,
		OrgID:  validator.orgIds,
	})
	if err != nil {
		return nil, errors.Wrap(err, "GetAll")
	}

	seriesMatches := map[string]*relatedInsightMetadata{}
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
			log15.Warn("related insights errors", mr.Errors)
		}
		if len(mr.Alerts) > 0 {
			log15.Warn("related insights alerts", mr.Alerts)
		}
		if len(mr.SkippedReasons) > 0 {
			log15.Warn("related insights skipped", mr.SkippedReasons)
		}

		for _, match := range mr.Matches {
			for _, lineMatch := range match.LineMatches {
				if seriesMatches[series.UniqueID] == nil {
					seriesMatches[series.UniqueID] = &relatedInsightMetadata{title: series.Title, lineNumbers: []int32{lineMatch.LineNumber}}
				} else {
					// Since insights can have multiple series, we might get duplicate matches.
					if !containsInt(seriesMatches[series.UniqueID].lineNumbers, lineMatch.LineNumber) {
						seriesMatches[series.UniqueID].lineNumbers = append(seriesMatches[series.UniqueID].lineNumbers, lineMatch.LineNumber)
					}
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

type relatedInsightMetadata struct {
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

func (r *Resolver) RelatedInsightsForFile(ctx context.Context, args graphqlbackend.RelatedInsightsArgs) ([]graphqlbackend.RelatedInsightsForFileResolver, error) {
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

	seriesMatches := map[string]*relatedInsightMetadata{}
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
			log15.Warn("related insights errors", mr.Errors)
		}
		if len(mr.Alerts) > 0 {
			log15.Warn("related insights alerts", mr.Alerts)
		}
		if len(mr.SkippedReasons) > 0 {
			log15.Warn("related insights skipped", mr.SkippedReasons)
		}

		for _, match := range mr.Matches {
			for _, lineMatch := range match.LineMatches {
				if seriesMatches[series.UniqueID] == nil {
					seriesMatches[series.UniqueID] = &relatedInsightMetadata{title: series.Title, lineNumbers: []int32{lineMatch.LineNumber}}
				} else {
					// Since insights can have multiple series, we might get duplicate matches.
					if !containsInt(seriesMatches[series.UniqueID].lineNumbers, lineMatch.LineNumber) {
						seriesMatches[series.UniqueID].lineNumbers = append(seriesMatches[series.UniqueID].lineNumbers, lineMatch.LineNumber)
					}
				}
			}
		}
	}

	var resolvers []graphqlbackend.RelatedInsightsForFileResolver
	for insightId, metadata := range seriesMatches {
		sort.SliceStable(metadata.lineNumbers, func(i, j int) bool {
			return metadata.lineNumbers[i] < metadata.lineNumbers[j]
		})
		resolvers = append(resolvers, &relatedInsightsForFileResolver{viewID: insightId, title: metadata.title, file: args.Input.File})
	}
	return resolvers, nil
}

type relatedInsightsForFileResolver struct {
	viewID string
	title  string
	file   string

	baseInsightResolver
}

func (r *relatedInsightsForFileResolver) ViewID() string {
	return r.viewID
}

func (r *relatedInsightsForFileResolver) Title() string {
	return r.title
}

func (r *relatedInsightsForFileResolver) File() string {
	return r.file
}

func containsInt(array []int32, findElement int32) bool {
	for _, currentElement := range array {
		if findElement == currentElement {
			return true
		}
	}
	return false
}
