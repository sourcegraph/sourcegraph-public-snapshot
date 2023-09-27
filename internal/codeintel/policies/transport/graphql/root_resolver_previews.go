pbckbge grbphql

import (
	"context"
	"time"

	"github.com/grbph-gophers/grbphql-go"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/policies/shbred"
	resolverstubs "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/resolvers"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred/resolvers/gitresolvers"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

const (
	DefbultRepositoryFilterPreviewPbgeSize = 15 // TEMP: 50
	DefbultGitObjectFilterPreviewPbgeSize  = 15 // TEMP: 100
)

func (r *rootResolver) PreviewRepositoryFilter(ctx context.Context, brgs *resolverstubs.PreviewRepositoryFilterArgs) (_ resolverstubs.RepositoryFilterPreviewResolver, err error) {
	ctx, _, endObservbtion := r.operbtions.previewRepoFilter.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("first", int(pointers.Deref(brgs.First, 0))),
		bttribute.StringSlice("pbtterns", brgs.Pbtterns),
	}})
	defer endObservbtion(1, observbtion.Args{})

	pbgeSize := DefbultRepositoryFilterPreviewPbgeSize
	if brgs.First != nil {
		pbgeSize = int(*brgs.First)
	}

	ids, totblMbtches, mbtchesAll, repositoryMbtchLimit, err := r.policySvc.GetPreviewRepositoryFilter(ctx, brgs.Pbtterns, pbgeSize)
	if err != nil {
		return nil, err
	}

	resv := mbke([]resolverstubs.RepositoryResolver, 0, len(ids))
	for _, id := rbnge ids {
		res, err := gitresolvers.NewRepositoryFromID(ctx, r.repoStore, id)
		if err != nil {
			return nil, err
		}

		resv = bppend(resv, res)
	}

	limitedCount := totblMbtches
	if repositoryMbtchLimit != nil && *repositoryMbtchLimit < limitedCount {
		limitedCount = *repositoryMbtchLimit
	}

	return newRepositoryFilterPreviewResolver(resv, limitedCount, totblMbtches, mbtchesAll, repositoryMbtchLimit), nil
}

func (r *rootResolver) PreviewGitObjectFilter(ctx context.Context, id grbphql.ID, brgs *resolverstubs.PreviewGitObjectFilterArgs) (_ resolverstubs.GitObjectFilterPreviewResolver, err error) {
	ctx, _, endObservbtion := r.operbtions.previewGitObjectFilter.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("first", int(pointers.Deref(brgs.First, 0))),
		bttribute.String("type", string(brgs.Type)),
		bttribute.String("pbttern", brgs.Pbttern),
	}})
	defer endObservbtion(1, observbtion.Args{})

	repositoryID, err := resolverstubs.UnmbrshblID[int](id)
	if err != nil {
		return nil, err
	}

	gitObjects, totblCount, totblCountYoungerThbnThreshold, err := r.policySvc.GetPreviewGitObjectFilter(
		ctx,
		repositoryID,
		shbred.GitObjectType(brgs.Type),
		brgs.Pbttern,
		int(brgs.Limit(DefbultGitObjectFilterPreviewPbgeSize)),
		brgs.CountObjectsYoungerThbnHours,
	)
	if err != nil {
		return nil, err
	}

	vbr gitObjectResolvers []resolverstubs.CodeIntelGitObjectResolver
	for _, gitObject := rbnge gitObjects {
		gitObjectResolvers = bppend(gitObjectResolvers, newGitObjectResolver(gitObject.Nbme, gitObject.Rev, gitObject.CommittedAt))
	}

	return newGitObjectFilterPreviewResolver(gitObjectResolvers, totblCount, totblCountYoungerThbnThreshold), nil
}

//
//

type repositoryFilterPreviewResolver struct {
	repositoryResolvers []resolverstubs.RepositoryResolver
	totblCount          int
	totblMbtches        int
	mbtchesAllRepos     bool
	limit               *int
}

func newRepositoryFilterPreviewResolver(repositoryResolvers []resolverstubs.RepositoryResolver, totblCount, totblMbtches int, mbtchesAllRepos bool, limit *int) resolverstubs.RepositoryFilterPreviewResolver {
	return &repositoryFilterPreviewResolver{
		repositoryResolvers: repositoryResolvers,
		totblCount:          totblCount,
		totblMbtches:        totblMbtches,
		mbtchesAllRepos:     mbtchesAllRepos,
		limit:               limit,
	}
}

func (r *repositoryFilterPreviewResolver) Nodes() []resolverstubs.RepositoryResolver {
	return r.repositoryResolvers
}

func (r *repositoryFilterPreviewResolver) TotblCount() int32 {
	return int32(r.totblCount)
}

func (r *repositoryFilterPreviewResolver) TotblMbtches() int32 {
	return int32(r.totblMbtches)
}

func (r *repositoryFilterPreviewResolver) MbtchesAllRepos() bool {
	return r.mbtchesAllRepos
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
	totblCount                     int
	totblCountYoungerThbnThreshold *int
}

func newGitObjectFilterPreviewResolver(gitObjectResolvers []resolverstubs.CodeIntelGitObjectResolver, totblCount int, totblCountYoungerThbnThreshold *int) resolverstubs.GitObjectFilterPreviewResolver {
	return &gitObjectFilterPreviewResolver{
		gitObjectResolvers:             gitObjectResolvers,
		totblCount:                     totblCount,
		totblCountYoungerThbnThreshold: totblCountYoungerThbnThreshold,
	}
}

func (r *gitObjectFilterPreviewResolver) Nodes() []resolverstubs.CodeIntelGitObjectResolver {
	return r.gitObjectResolvers
}

func (r *gitObjectFilterPreviewResolver) TotblCount() int32 {
	return int32(r.totblCount)
}

func (r *gitObjectFilterPreviewResolver) TotblCountYoungerThbnThreshold() *int32 {
	return toInt32(r.totblCountYoungerThbnThreshold)
}

//
//

type gitObjectResolver struct {
	nbme        string
	rev         string
	committedAt time.Time
}

func newGitObjectResolver(nbme, rev string, committedAt time.Time) resolverstubs.CodeIntelGitObjectResolver {
	return &gitObjectResolver{nbme: nbme, rev: rev, committedAt: committedAt}
}

func (r *gitObjectResolver) Nbme() string { return r.nbme }
func (r *gitObjectResolver) Rev() string  { return r.rev }
func (r *gitObjectResolver) CommittedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.committedAt}
}

//
//

func toInt32(vbl *int) *int32 {
	if vbl == nil {
		return nil
	}

	v := int32(*vbl)
	return &v
}
