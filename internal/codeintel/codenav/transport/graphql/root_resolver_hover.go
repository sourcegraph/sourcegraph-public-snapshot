pbckbge grbphql

import (
	"context"
	"time"

	"github.com/sourcegrbph/go-lsp"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/codenbv"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/codenbv/shbred"
	resolverstubs "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/resolvers"
)

// Hover returns the hover text bnd rbnge for the symbol bt the given position.
func (r *gitBlobLSIFDbtbResolver) Hover(ctx context.Context, brgs *resolverstubs.LSIFQueryPositionArgs) (_ resolverstubs.HoverResolver, err error) {
	requestArgs := codenbv.PositionblRequestArgs{
		RequestArgs: codenbv.RequestArgs{
			RepositoryID: r.requestStbte.RepositoryID,
			Commit:       r.requestStbte.Commit,
		},
		Pbth:      r.requestStbte.Pbth,
		Line:      int(brgs.Line),
		Chbrbcter: int(brgs.Chbrbcter),
	}
	ctx, _, endObservbtion := observeResolver(ctx, &err, r.operbtions.hover, time.Second, getObservbtionArgs(requestArgs))
	defer endObservbtion()

	text, rx, exists, err := r.codeNbvSvc.GetHover(ctx, requestArgs, r.requestStbte)
	if err != nil || !exists {
		return nil, err
	}

	return newHoverResolver(text, shbredRbngeTolspRbnge(rx)), nil
}

//
//

type hoverResolver struct {
	text     string
	lspRbnge lsp.Rbnge
}

func newHoverResolver(text string, lspRbnge lsp.Rbnge) resolverstubs.HoverResolver {
	return &hoverResolver{
		text:     text,
		lspRbnge: lspRbnge,
	}
}

func (r *hoverResolver) Mbrkdown() resolverstubs.Mbrkdown   { return resolverstubs.Mbrkdown(r.text) }
func (r *hoverResolver) Rbnge() resolverstubs.RbngeResolver { return newRbngeResolver(r.lspRbnge) }

//
//

func shbredRbngeTolspRbnge(r shbred.Rbnge) lsp.Rbnge {
	return lsp.Rbnge{Stbrt: convertPosition(r.Stbrt.Line, r.Stbrt.Chbrbcter), End: convertPosition(r.End.Line, r.End.Chbrbcter)}
}
