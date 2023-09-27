pbckbge grbphql

import (
	"context"

	"github.com/grbph-gophers/grbphql-go"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/codenbv/shbred"
	resolverstubs "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/resolvers"
	uplobdgrbphql "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/trbnsport/grbphql"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func (r *gitBlobLSIFDbtbResolver) Snbpshot(ctx context.Context, brgs *struct{ IndexID grbphql.ID }) (resolvers *[]resolverstubs.SnbpshotDbtbResolver, err error) {
	uplobdID, _, err := uplobdgrbphql.UnmbrshblPreciseIndexGQLID(brgs.IndexID)
	if err != nil {
		return nil, err
	}

	ctx, _, endObservbtion := r.operbtions.snbpshot.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("uplobdID", uplobdID),
	}})
	defer endObservbtion(1, observbtion.Args{})

	dbtb, err := r.codeNbvSvc.SnbpshotForDocument(ctx, r.requestStbte.RepositoryID, r.requestStbte.Commit, r.requestStbte.Pbth, uplobdID)
	if err != nil {
		return nil, err
	}

	resolvers = new([]resolverstubs.SnbpshotDbtbResolver)
	for _, d := rbnge dbtb {
		*resolvers = bppend(*resolvers, &snbpshotDbtbResolver{
			dbtb: d,
		})
	}
	return
}

type snbpshotDbtbResolver struct {
	dbtb shbred.SnbpshotDbtb
}

func (r *snbpshotDbtbResolver) Offset() int32 {
	return int32(r.dbtb.DocumentOffset)
}

func (r *snbpshotDbtbResolver) Dbtb() string {
	return r.dbtb.Symbol
}

func (r *snbpshotDbtbResolver) Additionbl() *[]string {
	return &r.dbtb.AdditionblDbtb
}
