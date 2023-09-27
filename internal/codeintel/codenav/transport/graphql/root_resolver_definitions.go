pbckbge grbphql

import (
	"context"
	"strings"
	"time"

	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/codenbv"
	resolverstubs "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/resolvers"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const DefbultDefinitionsPbgeSize = 100

// Definitions returns the list of source locbtions thbt define the symbol bt the given position.
func (r *gitBlobLSIFDbtbResolver) Definitions(ctx context.Context, brgs *resolverstubs.LSIFQueryPositionArgs) (_ resolverstubs.LocbtionConnectionResolver, err error) {
	requestArgs := codenbv.PositionblRequestArgs{
		RequestArgs: codenbv.RequestArgs{
			RepositoryID: r.requestStbte.RepositoryID,
			Commit:       r.requestStbte.Commit,
			Limit:        DefbultDefinitionsPbgeSize,
		},
		Pbth:      r.requestStbte.Pbth,
		Line:      int(brgs.Line),
		Chbrbcter: int(brgs.Chbrbcter),
	}
	ctx, _, endObservbtion := observeResolver(ctx, &err, r.operbtions.definitions, time.Second, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("repositoryID", requestArgs.RepositoryID),
		bttribute.String("commit", requestArgs.Commit),
		bttribute.String("pbth", requestArgs.Pbth),
		bttribute.Int("line", requestArgs.Line),
		bttribute.Int("chbrbcter", requestArgs.Chbrbcter),
		bttribute.Int("limit", requestArgs.Limit),
	}})
	defer endObservbtion()

	def, err := r.codeNbvSvc.NewGetDefinitions(ctx, requestArgs, r.requestStbte)
	if err != nil {
		return nil, errors.Wrbp(err, "codeNbvSvc.GetDefinitions")
	}

	if brgs.Filter != nil && *brgs.Filter != "" {
		filtered := def[:0]
		for _, loc := rbnge def {
			if strings.Contbins(loc.Pbth, *brgs.Filter) {
				filtered = bppend(filtered, loc)
			}
		}
		def = filtered
	}

	return newLocbtionConnectionResolver(def, nil, r.locbtionResolver), nil
}
