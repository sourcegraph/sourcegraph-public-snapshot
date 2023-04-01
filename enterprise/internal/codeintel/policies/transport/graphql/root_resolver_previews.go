package graphql

import (
	"context"
	"strings"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/opentracing/opentracing-go/log"

	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

const (
	DefaultRepositoryFilterPreviewPageSize = 15 // TEMP: 50
	DefaultGitObjectFilterPreviewPageSize  = 15 // TEMP: 100
)

func (r *rootResolver) PreviewRepositoryFilter(ctx context.Context, args *resolverstubs.PreviewRepositoryFilterArgs) (_ resolverstubs.RepositoryFilterPreviewResolver, err error) {
	ctx, _, endObservation := r.operations.previewRepoFilter.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int32("first", resolverstubs.Deref(args.First, 0)),
		log.String("patterns", strings.Join(args.Patterns, ", ")),
	}})
	defer endObservation(1, observation.Args{})

	pageSize := DefaultRepositoryFilterPreviewPageSize
	if args.First != nil {
		pageSize = int(*args.First)
	}

	ids, totalMatches, matchesAll, repositoryMatchLimit, err := r.policySvc.GetPreviewRepositoryFilter(ctx, args.Patterns, pageSize)
	if err != nil {
		return nil, err
	}

	resv := make([]resolverstubs.RepositoryResolver, 0, len(ids))
	for _, id := range ids {
		res, err := sharedresolvers.NewRepositoryFromID(ctx, r.repoStore, id)
		if err != nil {
			return nil, err
		}

		resv = append(resv, res)
	}

	limitedCount := totalMatches
	if repositoryMatchLimit != nil && *repositoryMatchLimit < limitedCount {
		limitedCount = *repositoryMatchLimit
	}

	return newRepositoryFilterPreviewResolver(resv, limitedCount, totalMatches, matchesAll, repositoryMatchLimit), nil
}

func (r *rootResolver) PreviewGitObjectFilter(ctx context.Context, id graphql.ID, args *resolverstubs.PreviewGitObjectFilterArgs) (_ resolverstubs.GitObjectFilterPreviewResolver, err error) {
	ctx, _, endObservation := r.operations.previewGitObjectFilter.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int32("first", resolverstubs.Deref(args.First, 0)),
		log.String("type", string(args.Type)),
		log.String("pattern", args.Pattern),
	}})
	defer endObservation(1, observation.Args{})

	repositoryID, err := resolverstubs.UnmarshalID[int](id)
	if err != nil {
		return nil, err
	}

	gitObjects, totalCount, totalCountYoungerThanThreshold, err := r.policySvc.GetPreviewGitObjectFilter(
		ctx,
		repositoryID,
		types.GitObjectType(args.Type),
		args.Pattern,
		int(args.Limit(DefaultGitObjectFilterPreviewPageSize)),
		args.CountObjectsYoungerThanHours,
	)
	if err != nil {
		return nil, err
	}

	var gitObjectResolvers []resolverstubs.CodeIntelGitObjectResolver
	for _, gitObject := range gitObjects {
		gitObjectResolvers = append(gitObjectResolvers, newGitObjectResolver(gitObject.Name, gitObject.Rev, gitObject.CommittedAt))
	}

	return newGitObjectFilterPreviewResolver(gitObjectResolvers, totalCount, totalCountYoungerThanThreshold), nil
}

//
//

type repositoryFilterPreviewResolver struct {
	repositoryResolvers []resolverstubs.RepositoryResolver
	totalCount          int
	totalMatches        int
	matchesAllRepos     bool
	limit               *int
}

func newRepositoryFilterPreviewResolver(repositoryResolvers []resolverstubs.RepositoryResolver, totalCount, totalMatches int, matchesAllRepos bool, limit *int) resolverstubs.RepositoryFilterPreviewResolver {
	return &repositoryFilterPreviewResolver{
		repositoryResolvers: repositoryResolvers,
		totalCount:          totalCount,
		totalMatches:        totalMatches,
		matchesAllRepos:     matchesAllRepos,
		limit:               limit,
	}
}

func (r *repositoryFilterPreviewResolver) Nodes() []resolverstubs.RepositoryResolver {
	return r.repositoryResolvers
}

func (r *repositoryFilterPreviewResolver) TotalCount() int32 {
	return int32(r.totalCount)
}

func (r *repositoryFilterPreviewResolver) TotalMatches() int32 {
	return int32(r.totalMatches)
}

func (r *repositoryFilterPreviewResolver) MatchesAllRepos() bool {
	return r.matchesAllRepos
}

func (r *repositoryFilterPreviewResolver) Limit() *int32 {
	if r.limit == nil {
		return nil
	}

	v := int32(*r.limit)
	return &v
}

//
//

type gitObjectFilterPreviewResolver struct {
	gitObjectResolvers             []resolverstubs.CodeIntelGitObjectResolver
	totalCount                     int
	totalCountYoungerThanThreshold *int
}

func newGitObjectFilterPreviewResolver(gitObjectResolvers []resolverstubs.CodeIntelGitObjectResolver, totalCount int, totalCountYoungerThanThreshold *int) resolverstubs.GitObjectFilterPreviewResolver {
	return &gitObjectFilterPreviewResolver{
		gitObjectResolvers:             gitObjectResolvers,
		totalCount:                     totalCount,
		totalCountYoungerThanThreshold: totalCountYoungerThanThreshold,
	}
}

func (r *gitObjectFilterPreviewResolver) Nodes() []resolverstubs.CodeIntelGitObjectResolver {
	return r.gitObjectResolvers
}

func (r *gitObjectFilterPreviewResolver) TotalCount() int32 {
	return int32(r.totalCount)
}

func (r *gitObjectFilterPreviewResolver) TotalCountYoungerThanThreshold() *int32 {
	return toInt32(r.totalCountYoungerThanThreshold)
}

//
//

type gitObjectResolver struct {
	name        string
	rev         string
	committedAt time.Time
}

func newGitObjectResolver(name, rev string, committedAt time.Time) resolverstubs.CodeIntelGitObjectResolver {
	return &gitObjectResolver{name: name, rev: rev, committedAt: committedAt}
}

func (r *gitObjectResolver) Name() string { return r.name }
func (r *gitObjectResolver) Rev() string  { return r.rev }
func (r *gitObjectResolver) CommittedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.committedAt}
}

//
//

func toInt32(val *int) *int32 {
	if val == nil {
		return nil
	}

	v := int32(*val)
	return &v
}
