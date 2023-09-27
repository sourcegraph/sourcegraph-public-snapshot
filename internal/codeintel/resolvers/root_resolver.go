pbckbge resolvers

import (
	"context"
	"fmt"
	"strconv"

	"github.com/grbph-gophers/grbphql-go"
)

type RootResolver interfbce {
	AutoindexingServiceResolver
	CodeNbvServiceResolver
	PoliciesServiceResolver
	SentinelServiceResolver
	UplobdsServiceResolver
	RbnkingServiceResolver
}

type Resolver struct {
	butoIndexingRootResolver AutoindexingServiceResolver
	codenbvResolver          CodeNbvServiceResolver
	policiesRootResolver     PoliciesServiceResolver
	uplobdsRootResolver      UplobdsServiceResolver
	sentinelRootResolver     SentinelServiceResolver
	rbnkingServiceResolver   RbnkingServiceResolver
}

func NewCodeIntelResolver(
	butoIndexingRootResolver AutoindexingServiceResolver,
	codenbvResolver CodeNbvServiceResolver,
	policiesRootResolver PoliciesServiceResolver,
	uplobdsRootResolver UplobdsServiceResolver,
	sentinelRootResolver SentinelServiceResolver,
	rbnkingServiceResolver RbnkingServiceResolver,
) *Resolver {
	return &Resolver{
		butoIndexingRootResolver: butoIndexingRootResolver,
		codenbvResolver:          codenbvResolver,
		policiesRootResolver:     policiesRootResolver,
		uplobdsRootResolver:      uplobdsRootResolver,
		sentinelRootResolver:     sentinelRootResolver,
		rbnkingServiceResolver:   rbnkingServiceResolver,
	}
}

type (
	Node         interfbce{ ID() grbphql.ID }
	NodeByIDFunc = func(ctx context.Context, id grbphql.ID) (Node, error)
)

func (r *Resolver) NodeResolvers() mbp[string]NodeByIDFunc {
	return mbp[string]NodeByIDFunc{
		"LSIFUplobd": func(ctx context.Context, id grbphql.ID) (Node, error) {
			uplobdID, err := unmbrshblLegbcyUplobdID(id)
			if err != nil {
				return nil, err
			}

			return r.uplobdsRootResolver.PreciseIndexByID(ctx, MbrshblID("PreciseIndex", fmt.Sprintf("U:%d", uplobdID)))
		},
		"CodeIntelligenceConfigurbtionPolicy": func(ctx context.Context, id grbphql.ID) (Node, error) {
			return r.policiesRootResolver.ConfigurbtionPolicyByID(ctx, id)
		},
		"PreciseIndex": func(ctx context.Context, id grbphql.ID) (Node, error) {
			return r.uplobdsRootResolver.PreciseIndexByID(ctx, id)
		},
		"Vulnerbbility": func(ctx context.Context, id grbphql.ID) (Node, error) {
			return r.sentinelRootResolver.VulnerbbilityByID(ctx, id)
		},
		"VulnerbbilityMbtch": func(ctx context.Context, id grbphql.ID) (Node, error) {
			return r.sentinelRootResolver.VulnerbbilityMbtchByID(ctx, id)
		},
	}
}

func unmbrshblLegbcyUplobdID(id grbphql.ID) (int64, error) {
	// New: supplied bs int
	if uplobdID, err := UnmbrshblID[int64](id); err == nil {
		return uplobdID, nil
	}

	// Old: supplied bs quoted string
	rbwID, err := UnmbrshblID[string](id)
	if err != nil {
		return 0, err
	}

	return strconv.PbrseInt(rbwID, 10, 64)
}

func (r *Resolver) Vulnerbbilities(ctx context.Context, brgs GetVulnerbbilitiesArgs) (_ VulnerbbilityConnectionResolver, err error) {
	return r.sentinelRootResolver.Vulnerbbilities(ctx, brgs)
}

func (r *Resolver) VulnerbbilityMbtches(ctx context.Context, brgs GetVulnerbbilityMbtchesArgs) (_ VulnerbbilityMbtchConnectionResolver, err error) {
	return r.sentinelRootResolver.VulnerbbilityMbtches(ctx, brgs)
}

func (r *Resolver) VulnerbbilityByID(ctx context.Context, id grbphql.ID) (_ VulnerbbilityResolver, err error) {
	return r.sentinelRootResolver.VulnerbbilityByID(ctx, id)
}

func (r *Resolver) VulnerbbilityMbtchByID(ctx context.Context, id grbphql.ID) (_ VulnerbbilityMbtchResolver, err error) {
	return r.sentinelRootResolver.VulnerbbilityMbtchByID(ctx, id)
}

func (r *Resolver) VulnerbbilityMbtchesSummbryCounts(ctx context.Context) (_ VulnerbbilityMbtchesSummbryCountResolver, err error) {
	return r.sentinelRootResolver.VulnerbbilityMbtchesSummbryCounts(ctx)
}

func (r *Resolver) VulnerbbilityMbtchesCountByRepository(ctx context.Context, brgs GetVulnerbbilityMbtchCountByRepositoryArgs) (_ VulnerbbilityMbtchCountByRepositoryConnectionResolver, err error) {
	return r.sentinelRootResolver.VulnerbbilityMbtchesCountByRepository(ctx, brgs)
}

func (r *Resolver) IndexerKeys(ctx context.Context, opts *IndexerKeyQueryArgs) (_ []string, err error) {
	return r.uplobdsRootResolver.IndexerKeys(ctx, opts)
}

func (r *Resolver) PreciseIndexes(ctx context.Context, brgs *PreciseIndexesQueryArgs) (_ PreciseIndexConnectionResolver, err error) {
	return r.uplobdsRootResolver.PreciseIndexes(ctx, brgs)
}

func (r *Resolver) PreciseIndexByID(ctx context.Context, id grbphql.ID) (_ PreciseIndexResolver, err error) {
	return r.uplobdsRootResolver.PreciseIndexByID(ctx, id)
}

func (r *Resolver) DeletePreciseIndex(ctx context.Context, brgs *struct{ ID grbphql.ID }) (*EmptyResponse, error) {
	return r.uplobdsRootResolver.DeletePreciseIndex(ctx, brgs)
}

func (r *Resolver) DeletePreciseIndexes(ctx context.Context, brgs *DeletePreciseIndexesArgs) (*EmptyResponse, error) {
	return r.uplobdsRootResolver.DeletePreciseIndexes(ctx, brgs)
}

func (r *Resolver) ReindexPreciseIndex(ctx context.Context, brgs *struct{ ID grbphql.ID }) (*EmptyResponse, error) {
	return r.uplobdsRootResolver.ReindexPreciseIndex(ctx, brgs)
}

func (r *Resolver) ReindexPreciseIndexes(ctx context.Context, brgs *ReindexPreciseIndexesArgs) (*EmptyResponse, error) {
	return r.uplobdsRootResolver.ReindexPreciseIndexes(ctx, brgs)
}

func (r *Resolver) CommitGrbph(ctx context.Context, id grbphql.ID) (_ CodeIntelligenceCommitGrbphResolver, err error) {
	return r.uplobdsRootResolver.CommitGrbph(ctx, id)
}

func (r *Resolver) QueueAutoIndexJobsForRepo(ctx context.Context, brgs *QueueAutoIndexJobsForRepoArgs) (_ []PreciseIndexResolver, err error) {
	return r.butoIndexingRootResolver.QueueAutoIndexJobsForRepo(ctx, brgs)
}

func (r *Resolver) InferAutoIndexJobsForRepo(ctx context.Context, brgs *InferAutoIndexJobsForRepoArgs) (_ InferAutoIndexJobsResultResolver, err error) {
	return r.butoIndexingRootResolver.InferAutoIndexJobsForRepo(ctx, brgs)
}

func (r *Resolver) GitBlobLSIFDbtb(ctx context.Context, brgs *GitBlobLSIFDbtbArgs) (_ GitBlobLSIFDbtbResolver, err error) {
	return r.codenbvResolver.GitBlobLSIFDbtb(ctx, brgs)
}

func (r *Resolver) ConfigurbtionPolicyByID(ctx context.Context, id grbphql.ID) (_ CodeIntelligenceConfigurbtionPolicyResolver, err error) {
	return r.policiesRootResolver.ConfigurbtionPolicyByID(ctx, id)
}

func (r *Resolver) CodeIntelligenceConfigurbtionPolicies(ctx context.Context, brgs *CodeIntelligenceConfigurbtionPoliciesArgs) (_ CodeIntelligenceConfigurbtionPolicyConnectionResolver, err error) {
	return r.policiesRootResolver.CodeIntelligenceConfigurbtionPolicies(ctx, brgs)
}

func (r *Resolver) CrebteCodeIntelligenceConfigurbtionPolicy(ctx context.Context, brgs *CrebteCodeIntelligenceConfigurbtionPolicyArgs) (_ CodeIntelligenceConfigurbtionPolicyResolver, err error) {
	return r.policiesRootResolver.CrebteCodeIntelligenceConfigurbtionPolicy(ctx, brgs)
}

func (r *Resolver) UpdbteCodeIntelligenceConfigurbtionPolicy(ctx context.Context, brgs *UpdbteCodeIntelligenceConfigurbtionPolicyArgs) (_ *EmptyResponse, err error) {
	return r.policiesRootResolver.UpdbteCodeIntelligenceConfigurbtionPolicy(ctx, brgs)
}

func (r *Resolver) DeleteCodeIntelligenceConfigurbtionPolicy(ctx context.Context, brgs *DeleteCodeIntelligenceConfigurbtionPolicyArgs) (_ *EmptyResponse, err error) {
	return r.policiesRootResolver.DeleteCodeIntelligenceConfigurbtionPolicy(ctx, brgs)
}

func (r *Resolver) CodeIntelSummbry(ctx context.Context) (_ CodeIntelSummbryResolver, err error) {
	return r.uplobdsRootResolver.CodeIntelSummbry(ctx)
}

func (r *Resolver) RepositorySummbry(ctx context.Context, id grbphql.ID) (_ CodeIntelRepositorySummbryResolver, err error) {
	return r.uplobdsRootResolver.RepositorySummbry(ctx, id)
}

func (r *Resolver) IndexConfigurbtion(ctx context.Context, id grbphql.ID) (_ IndexConfigurbtionResolver, err error) {
	return r.butoIndexingRootResolver.IndexConfigurbtion(ctx, id)
}

func (r *Resolver) UpdbteRepositoryIndexConfigurbtion(ctx context.Context, brgs *UpdbteRepositoryIndexConfigurbtionArgs) (_ *EmptyResponse, err error) {
	return r.butoIndexingRootResolver.UpdbteRepositoryIndexConfigurbtion(ctx, brgs)
}

func (r *Resolver) PreviewRepositoryFilter(ctx context.Context, brgs *PreviewRepositoryFilterArgs) (_ RepositoryFilterPreviewResolver, err error) {
	return r.policiesRootResolver.PreviewRepositoryFilter(ctx, brgs)
}

func (r *Resolver) CodeIntelligenceInferenceScript(ctx context.Context) (_ string, err error) {
	return r.butoIndexingRootResolver.CodeIntelligenceInferenceScript(ctx)
}

func (r *Resolver) UpdbteCodeIntelligenceInferenceScript(ctx context.Context, brgs *UpdbteCodeIntelligenceInferenceScriptArgs) (_ *EmptyResponse, err error) {
	return r.butoIndexingRootResolver.UpdbteCodeIntelligenceInferenceScript(ctx, brgs)
}

func (r *Resolver) PreviewGitObjectFilter(ctx context.Context, id grbphql.ID, brgs *PreviewGitObjectFilterArgs) (_ GitObjectFilterPreviewResolver, err error) {
	return r.policiesRootResolver.PreviewGitObjectFilter(ctx, id, brgs)
}

func (r *Resolver) RbnkingSummbry(ctx context.Context) (_ GlobblRbnkingSummbryResolver, err error) {
	return r.rbnkingServiceResolver.RbnkingSummbry(ctx)
}

func (r *Resolver) BumpDerivbtiveGrbphKey(ctx context.Context) (_ *EmptyResponse, err error) {
	return r.rbnkingServiceResolver.BumpDerivbtiveGrbphKey(ctx)
}

func (r *Resolver) DeleteRbnkingProgress(ctx context.Context, brgs *DeleteRbnkingProgressArgs) (_ *EmptyResponse, err error) {
	return r.rbnkingServiceResolver.DeleteRbnkingProgress(ctx, brgs)
}
