package graphql

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/shared"
	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type rootResolver struct {
	rankingSvc       RankingService
	siteAdminChecker sharedresolvers.SiteAdminChecker
	operations       *operations
}

func NewRootResolver(
	observationCtx *observation.Context,
	rankingSvc *ranking.Service,
	siteAdminChecker sharedresolvers.SiteAdminChecker,
) resolverstubs.RankingServiceResolver {
	return &rootResolver{
		rankingSvc:       rankingSvc,
		siteAdminChecker: siteAdminChecker,
		operations:       newOperations(observationCtx),
	}
}

// 🚨 SECURITY: Only site admins may view ranking job summaries.
func (r *rootResolver) RankingSummary(ctx context.Context) (_ []resolverstubs.RankingSummaryResolver, err error) {
	ctx, _, endObservation := r.operations.rankingSummary.With(ctx, &err, observation.Args{})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	if err := r.siteAdminChecker.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	summaries, err := r.rankingSvc.Summaries(ctx)
	if err != nil {
		return nil, err
	}

	resolvers := make([]resolverstubs.RankingSummaryResolver, 0, len(summaries))
	for _, summary := range summaries {
		resolvers = append(resolvers, &rankingSummaryResolver{
			summary: summary,
		})
	}

	return resolvers, nil
}

type rankingSummaryResolver struct {
	summary shared.Summary
}

func (r *rankingSummaryResolver) GraphKey() string {
	return r.summary.GraphKey
}

func (r *rankingSummaryResolver) PathMapperProgress() resolverstubs.RankingSummaryProgressResolver {
	return &progressResolver{progress: r.summary.PathMapperProgress}
}

func (r *rankingSummaryResolver) ReferenceMapperProgress() resolverstubs.RankingSummaryProgressResolver {
	return &progressResolver{progress: r.summary.ReferenceMapperProgress}
}

func (r *rankingSummaryResolver) ReducerProgress() resolverstubs.RankingSummaryProgressResolver {
	if r.summary.ReducerProgress == nil {
		return nil
	}

	return &progressResolver{progress: *r.summary.ReducerProgress}
}

type progressResolver struct {
	progress shared.Progress
}

func (r *progressResolver) StartedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.progress.StartedAt}
}

func (r *progressResolver) CompletedAt() *gqlutil.DateTime {
	return gqlutil.DateTimeOrNil(r.progress.CompletedAt)
}

func (r *progressResolver) Processed() int32 {
	return int32(r.progress.Processed)
}

func (r *progressResolver) Total() int32 {
	return int32(r.progress.Total)
}
