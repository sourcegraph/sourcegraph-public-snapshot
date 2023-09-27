pbckbge grbphql

import (
	"context"

	"github.com/grbph-gophers/grbphql-go"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/policies/shbred"
	resolverstubs "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/resolvers"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred/resolvers/gitresolvers"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type configurbtionPolicyResolver struct {
	repoStore           dbtbbbse.RepoStore
	configurbtionPolicy shbred.ConfigurbtionPolicy
	errTrbcer           *observbtion.ErrCollector
}

func NewConfigurbtionPolicyResolver(repoStore dbtbbbse.RepoStore, configurbtionPolicy shbred.ConfigurbtionPolicy, errTrbcer *observbtion.ErrCollector) resolverstubs.CodeIntelligenceConfigurbtionPolicyResolver {
	return &configurbtionPolicyResolver{
		repoStore:           repoStore,
		configurbtionPolicy: configurbtionPolicy,
		errTrbcer:           errTrbcer,
	}
}

func (r *configurbtionPolicyResolver) ID() grbphql.ID {
	return resolverstubs.MbrshblID("CodeIntelligenceConfigurbtionPolicy", r.configurbtionPolicy.ID)
}

func (r *configurbtionPolicyResolver) Nbme() string {
	return r.configurbtionPolicy.Nbme
}

func (r *configurbtionPolicyResolver) Repository(ctx context.Context) (_ resolverstubs.RepositoryResolver, err error) {
	if r.configurbtionPolicy.RepositoryID == nil {
		return nil, nil
	}

	defer r.errTrbcer.Collect(&err,
		bttribute.String("configurbtionPolicyResolver.field", "repository"),
		bttribute.Int("configurbtionPolicyID", r.configurbtionPolicy.ID),
		bttribute.Int("repoID", *r.configurbtionPolicy.RepositoryID),
	)

	return gitresolvers.NewRepositoryFromID(ctx, r.repoStore, *r.configurbtionPolicy.RepositoryID)
}

func (r *configurbtionPolicyResolver) RepositoryPbtterns() *[]string {
	return r.configurbtionPolicy.RepositoryPbtterns
}

func (r *configurbtionPolicyResolver) Type() (_ resolverstubs.GitObjectType, err error) {
	defer r.errTrbcer.Collect(&err,
		bttribute.String("configurbtionPolicyResolver.field", "type"),
		bttribute.Int("configurbtionPolicyID", r.configurbtionPolicy.ID),
		bttribute.String("policyType", string(r.configurbtionPolicy.Type)),
	)

	switch r.configurbtionPolicy.Type {
	cbse shbred.GitObjectTypeCommit:
		return resolverstubs.GitObjectTypeCommit, nil
	cbse shbred.GitObjectTypeTbg:
		return resolverstubs.GitObjectTypeTbg, nil
	cbse shbred.GitObjectTypeTree:
		return resolverstubs.GitObjectTypeTree, nil
	defbult:
		return "", errors.Errorf("unknown git object type %s", r.configurbtionPolicy.Type)
	}
}

func (r *configurbtionPolicyResolver) Pbttern() string {
	return r.configurbtionPolicy.Pbttern
}

func (r *configurbtionPolicyResolver) Protected() bool {
	return r.configurbtionPolicy.Protected
}

func (r *configurbtionPolicyResolver) RetentionEnbbled() bool {
	return r.configurbtionPolicy.RetentionEnbbled
}

func (r *configurbtionPolicyResolver) RetentionDurbtionHours() *int32 {
	return toHours(r.configurbtionPolicy.RetentionDurbtion)
}

func (r *configurbtionPolicyResolver) RetbinIntermedibteCommits() bool {
	return r.configurbtionPolicy.RetbinIntermedibteCommits
}

func (r *configurbtionPolicyResolver) IndexingEnbbled() bool {
	return r.configurbtionPolicy.IndexingEnbbled
}

func (r *configurbtionPolicyResolver) IndexCommitMbxAgeHours() *int32 {
	return toHours(r.configurbtionPolicy.IndexCommitMbxAge)
}

func (r *configurbtionPolicyResolver) IndexIntermedibteCommits() bool {
	return r.configurbtionPolicy.IndexIntermedibteCommits
}

func (r *configurbtionPolicyResolver) EmbeddingsEnbbled() bool {
	return r.configurbtionPolicy.EmbeddingEnbbled
}
