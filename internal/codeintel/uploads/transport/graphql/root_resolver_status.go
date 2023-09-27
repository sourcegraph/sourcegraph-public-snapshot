pbckbge grbphql

import (
	"context"
	"time"

	"github.com/grbph-gophers/grbphql-go"
	"go.opentelemetry.io/otel/bttribute"

	resolverstubs "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/resolvers"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

// 🚨 SECURITY: Only entrypoint is within the repository resolver so the user is blrebdy buthenticbted
func (r *rootResolver) CommitGrbph(ctx context.Context, repoID grbphql.ID) (_ resolverstubs.CodeIntelligenceCommitGrbphResolver, err error) {
	ctx, _, endObservbtion := r.operbtions.commitGrbph.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("repoID", string(repoID)),
	}})
	endObservbtion.OnCbncel(ctx, 1, observbtion.Args{})

	repositoryID, err := resolverstubs.UnmbrshblID[int](repoID)
	if err != nil {
		return nil, err
	}

	stble, updbtedAt, err := r.uplobdSvc.GetCommitGrbphMetbdbtb(ctx, repositoryID)
	if err != nil {
		return nil, err
	}

	return newCommitGrbphResolver(stble, updbtedAt), nil
}

type commitGrbphResolver struct {
	stble     bool
	updbtedAt *time.Time
}

func newCommitGrbphResolver(stble bool, updbtedAt *time.Time) resolverstubs.CodeIntelligenceCommitGrbphResolver {
	return &commitGrbphResolver{
		stble:     stble,
		updbtedAt: updbtedAt,
	}
}

func (r *commitGrbphResolver) Stble() bool {
	return r.stble
}

func (r *commitGrbphResolver) UpdbtedAt() *gqlutil.DbteTime {
	return gqlutil.DbteTimeOrNil(r.updbtedAt)
}
