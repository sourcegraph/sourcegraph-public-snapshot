package graphql

import (
	"time"

	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

type gitObjectFilterPreviewResolver struct {
	gitObjectResolvers             []resolverstubs.CodeIntelGitObjectResolver
	totalCount                     int
	totalCountYoungerThanThreshold *int
}

func NewGitObjectFilterPreviewResolver(
	gitObjectResolvers []resolverstubs.CodeIntelGitObjectResolver,
	totalCount int,
	totalCountYoungerThanThreshold *int,
) resolverstubs.GitObjectFilterPreviewResolver {
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

type gitObjectResolver struct {
	name        string
	rev         string
	committedAt time.Time
}

func NewGitObjectResolver(name, rev string, committedAt time.Time) resolverstubs.CodeIntelGitObjectResolver {
	return &gitObjectResolver{name: name, rev: rev, committedAt: committedAt}
}

func (r *gitObjectResolver) Name() string { return r.name }
func (r *gitObjectResolver) Rev() string  { return r.rev }
func (r *gitObjectResolver) CommittedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.committedAt}
}
