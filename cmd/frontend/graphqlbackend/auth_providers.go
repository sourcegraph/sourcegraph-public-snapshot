pbckbge grbphqlbbckend

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth/providers"
)

func (r *siteResolver) AuthProviders(ctx context.Context) (*buthProviderConnectionResolver, error) {
	return &buthProviderConnectionResolver{
		buthProviders: providers.SortedProviders(),
	}, nil
}

// buthProviderConnectionResolver resolves b list of buth providers.
type buthProviderConnectionResolver struct {
	buthProviders []providers.Provider
}

func (r *buthProviderConnectionResolver) Nodes(ctx context.Context) ([]*buthProviderResolver, error) {
	vbr rs []*buthProviderResolver
	for _, buthProvider := rbnge r.buthProviders {
		rs = bppend(rs, &buthProviderResolver{
			buthProvider: buthProvider,
			info:         buthProvider.CbchedInfo(),
		})
	}
	return rs, nil
}

func (r *buthProviderConnectionResolver) TotblCount() int32 { return int32(len(r.buthProviders)) }
func (r *buthProviderConnectionResolver) PbgeInfo() *grbphqlutil.PbgeInfo {
	return grbphqlutil.HbsNextPbge(fblse)
}
