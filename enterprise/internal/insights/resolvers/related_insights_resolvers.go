package resolvers

import (
	"context"
	"fmt"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query/querybuilder"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query/streaming"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var _ graphqlbackend.RelatedInsightsInlineResolver = &relatedInsightsInlineResolver{}
var _ graphqlbackend.RelatedInsightsInlineMetadataResolver = &relatedInsightsInlineMetadataResolver{}

func (r *Resolver) Insights(ctx context.Context, args graphqlbackend.RelatedInsightsInlineArgs) ([]graphqlbackend.RelatedInsightsInlineResolver, error) {
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
	fmt.Printf("found %d total series\n", len(allSeries))

	resolvers := []*relatedInsightsInlineMetadataResolver{}

	for _, series := range allSeries {
		decoder, metadataResult := streaming.MetadataDecoder()
		modifiedQuery, err := querybuilder.SingleFileQuery(querybuilder.BasicQuery(series.Query), args.Input.Repo, args.Input.File, args.Input.Revision, querybuilder.CodeInsightsQueryDefaults(false))
		if err != nil {
			return nil, errors.Wrap(err, "SingleFileQuery")
		}
		fmt.Printf("query: %s\n", series.Query)
		fmt.Printf("modified query: %s\n", modifiedQuery.String())

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
				fmt.Println("Found a match!")
				fmt.Println(lineMatch.Line)
				fmt.Println(lineMatch.LineNumber)
				fmt.Println(lineMatch.OffsetAndLengths)

				// maybe we want to store multiple line numbers per insight instead?
				// append to resolvers here?
			}
		}
	}

	// TODO format the results and return them. Also possible all of this should go below in the Insights resolver.

	return nil, nil
}

type relatedInsightsInlineResolver struct {
	series string
	baseInsightResolver
}

func (r *relatedInsightsInlineResolver) Insights(ctx context.Context) ([]relatedInsightsInlineMetadataResolver, error) {
	return nil, nil
}

type relatedInsightsInlineMetadataResolver struct {
	viewID      string
	title       string
	lineNumbers []int32

	baseInsightResolver
}

func (r *relatedInsightsInlineMetadataResolver) ViewID() string {
	return r.viewID
}

func (r *relatedInsightsInlineMetadataResolver) Title() string {
	return r.title
}

func (r *relatedInsightsInlineMetadataResolver) LineNumbers() []int32 {
	return r.lineNumbers
}
