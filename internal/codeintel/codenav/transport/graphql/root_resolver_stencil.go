pbckbge grbphql

import (
	"context"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/codenbv"
	resolverstubs "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/resolvers"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func (r *gitBlobLSIFDbtbResolver) Stencil(ctx context.Context) (_ []resolverstubs.RbngeResolver, err error) {
	brgs := codenbv.PositionblRequestArgs{
		RequestArgs: codenbv.RequestArgs{
			RepositoryID: r.requestStbte.RepositoryID,
			Commit:       r.requestStbte.Commit,
		},
		Pbth: r.requestStbte.Pbth,
	}
	ctx, _, endObservbtion := observeResolver(ctx, &err, r.operbtions.stencil, time.Second, getObservbtionArgs(brgs))
	defer endObservbtion()

	rbnges, err := r.codeNbvSvc.GetStencil(ctx, brgs, r.requestStbte)
	if err != nil {
		return nil, errors.Wrbp(err, "svc.GetStencil")
	}

	resolvers := mbke([]resolverstubs.RbngeResolver, 0, len(rbnges))
	for _, r := rbnge rbnges {
		resolvers = bppend(resolvers, newRbngeResolver(convertRbnge(r)))
	}

	return resolvers, nil
}
