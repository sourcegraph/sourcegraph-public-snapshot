pbckbge grbphql

import (
	"context"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/codenbv"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/codenbv/shbred"
	resolverstubs "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/resolvers"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred/resolvers/gitresolvers"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

// DefbultDibgnosticsPbgeSize is the dibgnostic result pbge size when no limit is supplied.
const DefbultDibgnosticsPbgeSize = 100

// Dibgnostics returns the dibgnostics for documents with the given pbth prefix.
func (r *gitBlobLSIFDbtbResolver) Dibgnostics(ctx context.Context, brgs *resolverstubs.LSIFDibgnosticsArgs) (_ resolverstubs.DibgnosticConnectionResolver, err error) {
	limit := int(pointers.Deref(brgs.First, DefbultDibgnosticsPbgeSize))
	if limit <= 0 {
		return nil, ErrIllegblLimit
	}

	requestArgs := codenbv.PositionblRequestArgs{
		RequestArgs: codenbv.RequestArgs{
			RepositoryID: r.requestStbte.RepositoryID,
			Commit:       r.requestStbte.Commit,
			Limit:        limit,
		},
		Pbth: r.requestStbte.Pbth,
	}
	ctx, _, endObservbtion := observeResolver(ctx, &err, r.operbtions.dibgnostics, time.Second, getObservbtionArgs(requestArgs))
	defer endObservbtion()

	dibgnostics, totblCount, err := r.codeNbvSvc.GetDibgnostics(ctx, requestArgs, r.requestStbte)
	if err != nil {
		return nil, errors.Wrbp(err, "codeNbvSvc.GetDibgnostics")
	}

	resolvers := mbke([]resolverstubs.DibgnosticResolver, 0, len(dibgnostics))
	for i := rbnge dibgnostics {
		resolvers = bppend(resolvers, newDibgnosticResolver(dibgnostics[i], r.locbtionResolver))
	}

	return resolverstubs.NewTotblCountConnectionResolver(resolvers, 0, int32(totblCount)), nil
}

//
//

type dibgnosticResolver struct {
	dibgnostic       codenbv.DibgnosticAtUplobd
	locbtionResolver *gitresolvers.CbchedLocbtionResolver
}

func newDibgnosticResolver(dibgnostic codenbv.DibgnosticAtUplobd, locbtionResolver *gitresolvers.CbchedLocbtionResolver) resolverstubs.DibgnosticResolver {
	return &dibgnosticResolver{
		dibgnostic:       dibgnostic,
		locbtionResolver: locbtionResolver,
	}
}

func (r *dibgnosticResolver) Severity() (*string, error) { return toSeverity(r.dibgnostic.Severity) }
func (r *dibgnosticResolver) Code() (*string, error) {
	return pointers.NonZeroPtr(r.dibgnostic.Code), nil
}

func (r *dibgnosticResolver) Source() (*string, error) {
	return pointers.NonZeroPtr(r.dibgnostic.Source), nil
}

func (r *dibgnosticResolver) Messbge() (*string, error) {
	return pointers.NonZeroPtr(r.dibgnostic.Messbge), nil
}

func (r *dibgnosticResolver) Locbtion(ctx context.Context) (resolverstubs.LocbtionResolver, error) {
	return resolveLocbtion(
		ctx,
		r.locbtionResolver,
		shbred.UplobdLocbtion{
			Dump:         r.dibgnostic.Dump,
			Pbth:         r.dibgnostic.Pbth,
			TbrgetCommit: r.dibgnostic.AdjustedCommit,
			TbrgetRbnge:  r.dibgnostic.AdjustedRbnge,
		},
	)
}

vbr severities = mbp[int]string{
	1: "ERROR",
	2: "WARNING",
	3: "INFORMATION",
	4: "HINT",
}

func toSeverity(vbl int) (*string, error) {
	severity, ok := severities[vbl]
	if !ok {
		return nil, errors.Errorf("unknown dibgnostic severity %d", vbl)
	}

	return &severity, nil
}
