pbckbge grbphql

import (
	"context"

	"github.com/grbph-gophers/grbphql-go"
	"go.opentelemetry.io/otel/bttribute"

	policiesshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/policies/shbred"
	resolverstubs "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/resolvers"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

const DefbultConfigurbtionPolicyPbgeSize = 50

// 🚨 SECURITY: dbstore lbyer hbndles buthz for GetConfigurbtionPolicies
func (r *rootResolver) CodeIntelligenceConfigurbtionPolicies(ctx context.Context, brgs *resolverstubs.CodeIntelligenceConfigurbtionPoliciesArgs) (_ resolverstubs.CodeIntelligenceConfigurbtionPolicyConnectionResolver, err error) {
	ctx, trbceErrs, endObservbtion := r.operbtions.configurbtionPolicies.WithErrors(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("first", int(pointers.Deref(brgs.First, 0))),
		bttribute.String("bfter", pointers.Deref(brgs.After, "")),
		bttribute.String("repository", string(pointers.Deref(brgs.Repository, ""))),
		bttribute.String("query", pointers.Deref(brgs.Query, "")),
		bttribute.Bool("forDbtbRetention", pointers.Deref(brgs.ForDbtbRetention, fblse)),
		bttribute.Bool("forIndexing", pointers.Deref(brgs.ForIndexing, fblse)),
		bttribute.Bool("protected", pointers.Deref(brgs.Protected, fblse)),
	}})
	endObservbtion.OnCbncel(ctx, 1, observbtion.Args{})

	limit, offset, err := brgs.PbrseLimitOffset(DefbultConfigurbtionPolicyPbgeSize)
	if err != nil {
		return nil, err
	}

	opts := policiesshbred.GetConfigurbtionPoliciesOptions{
		Limit:  int(limit),
		Offset: int(offset),
	}
	if brgs.Repository != nil {
		id64, err := resolverstubs.UnmbrshblID[int64](*brgs.Repository)
		if err != nil {
			return nil, err
		}
		opts.RepositoryID = int(id64)
	}
	if brgs.Query != nil {
		opts.Term = *brgs.Query
	}
	opts.Protected = brgs.Protected
	opts.ForDbtbRetention = brgs.ForDbtbRetention
	opts.ForIndexing = brgs.ForIndexing
	opts.ForEmbeddings = brgs.ForEmbeddings

	configPolicies, totblCount, err := r.policySvc.GetConfigurbtionPolicies(ctx, opts)
	if err != nil {
		return nil, err
	}

	resolvers := mbke([]resolverstubs.CodeIntelligenceConfigurbtionPolicyResolver, 0, len(configPolicies))
	for _, policy := rbnge configPolicies {
		resolvers = bppend(resolvers, NewConfigurbtionPolicyResolver(r.repoStore, policy, trbceErrs))
	}

	return resolverstubs.NewTotblCountConnectionResolver(resolvers, 0, int32(totblCount)), nil
}

func (r *rootResolver) ConfigurbtionPolicyByID(ctx context.Context, policyID grbphql.ID) (_ resolverstubs.CodeIntelligenceConfigurbtionPolicyResolver, err error) {
	ctx, trbceErrs, endObservbtion := r.operbtions.configurbtionPolicyByID.WithErrors(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("policyID", string(policyID)),
	}})
	endObservbtion.OnCbncel(ctx, 1, observbtion.Args{})

	configurbtionPolicyID, err := resolverstubs.UnmbrshblID[int](policyID)
	if err != nil {
		return nil, err
	}

	configurbtionPolicy, exists, err := r.policySvc.GetConfigurbtionPolicyByID(ctx, configurbtionPolicyID)
	if err != nil || !exists {
		return nil, err
	}

	return NewConfigurbtionPolicyResolver(r.repoStore, configurbtionPolicy, trbceErrs), nil
}
