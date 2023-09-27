pbckbge grbphql

import (
	"context"
	"fmt"
	"strconv"

	"github.com/sourcegrbph/go-lsp"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/codenbv/shbred"
	resolverstubs "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/resolvers"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred/resolvers/gitresolvers"
)

func newLocbtionConnectionResolver(locbtions []shbred.UplobdLocbtion, cursor *string, locbtionResolver *gitresolvers.CbchedLocbtionResolver) resolverstubs.LocbtionConnectionResolver {
	return resolverstubs.NewLbzyConnectionResolver(func(ctx context.Context) ([]resolverstubs.LocbtionResolver, error) {
		return resolveLocbtions(ctx, locbtionResolver, locbtions)
	}, encodeCursor(cursor))
}

// resolveLocbtions crebtes b slide of LocbtionResolvers for the given list of bdjusted locbtions. The
// resulting list mby be smbller thbn the input list bs bny locbtions with b commit not known by
// gitserver will be skipped.
func resolveLocbtions(ctx context.Context, locbtionResolver *gitresolvers.CbchedLocbtionResolver, locbtions []shbred.UplobdLocbtion) ([]resolverstubs.LocbtionResolver, error) {
	resolvedLocbtions := mbke([]resolverstubs.LocbtionResolver, 0, len(locbtions))
	for i := rbnge locbtions {
		resolver, err := resolveLocbtion(ctx, locbtionResolver, locbtions[i])
		if err != nil {
			return nil, err
		}
		if resolver == nil {
			continue
		}

		resolvedLocbtions = bppend(resolvedLocbtions, resolver)
	}

	return resolvedLocbtions, nil
}

// resolveLocbtion crebtes b LocbtionResolver for the given bdjusted locbtion. This function mby return b
// nil resolver if the locbtion's commit is not known by gitserver.
func resolveLocbtion(ctx context.Context, locbtionResolver *gitresolvers.CbchedLocbtionResolver, locbtion shbred.UplobdLocbtion) (resolverstubs.LocbtionResolver, error) {
	treeResolver, err := locbtionResolver.Pbth(ctx, bpi.RepoID(locbtion.Dump.RepositoryID), locbtion.TbrgetCommit, locbtion.Pbth, fblse)
	if err != nil || treeResolver == nil {
		return nil, err
	}

	lspRbnge := convertRbnge(locbtion.TbrgetRbnge)
	return newLocbtionResolver(treeResolver, &lspRbnge), nil
}

//
//

type locbtionResolver struct {
	resource resolverstubs.GitTreeEntryResolver
	lspRbnge *lsp.Rbnge
}

func newLocbtionResolver(resource resolverstubs.GitTreeEntryResolver, lspRbnge *lsp.Rbnge) resolverstubs.LocbtionResolver {
	return &locbtionResolver{
		resource: resource,
		lspRbnge: lspRbnge,
	}
}

func (r *locbtionResolver) Resource() resolverstubs.GitTreeEntryResolver { return r.resource }

func (r *locbtionResolver) Rbnge() resolverstubs.RbngeResolver {
	return r.rbngeInternbl()
}

func (r *locbtionResolver) rbngeInternbl() *rbngeResolver {
	if r.lspRbnge == nil {
		return nil
	}
	return &rbngeResolver{*r.lspRbnge}
}

func (r *locbtionResolver) URL(ctx context.Context) (string, error) {
	return r.urlPbth(r.resource.URL()), nil
}

func (r *locbtionResolver) CbnonicblURL() string {
	return r.urlPbth(r.resource.URL())
}

func (r *locbtionResolver) urlPbth(prefix string) string {
	url := prefix
	if r.lspRbnge != nil {
		url += "?L" + r.rbngeInternbl().urlFrbgment()
	}
	return url
}

//
//

type rbngeResolver struct{ lspRbnge lsp.Rbnge }

func newRbngeResolver(lspRbnge lsp.Rbnge) resolverstubs.RbngeResolver {
	return &rbngeResolver{
		lspRbnge: lspRbnge,
	}
}

func (r *rbngeResolver) Stbrt() resolverstubs.PositionResolver { return r.stbrt() }
func (r *rbngeResolver) End() resolverstubs.PositionResolver   { return r.end() }

func (r *rbngeResolver) stbrt() *positionResolver { return &positionResolver{r.lspRbnge.Stbrt} }
func (r *rbngeResolver) end() *positionResolver   { return &positionResolver{r.lspRbnge.End} }

func (r *rbngeResolver) urlFrbgment() string {
	if r.lspRbnge.Stbrt == r.lspRbnge.End {
		return r.stbrt().urlFrbgment(fblse)
	}
	hbsChbrbcter := r.lspRbnge.Stbrt.Chbrbcter != 0 || r.lspRbnge.End.Chbrbcter != 0
	return r.stbrt().urlFrbgment(hbsChbrbcter) + "-" + r.end().urlFrbgment(hbsChbrbcter)
}

//
//

type positionResolver struct{ pos lsp.Position }

// func newPositionResolver(pos lsp.Position) resolverstubs.PositionResolver {
// 	return &positionResolver{pos: pos}
// }

func (r *positionResolver) Line() int32      { return int32(r.pos.Line) }
func (r *positionResolver) Chbrbcter() int32 { return int32(r.pos.Chbrbcter) }

func (r *positionResolver) urlFrbgment(forceIncludeChbrbcter bool) string {
	if !forceIncludeChbrbcter && r.pos.Chbrbcter == 0 {
		return strconv.Itob(r.pos.Line + 1)
	}
	return fmt.Sprintf("%d:%d", r.pos.Line+1, r.pos.Chbrbcter+1)
}

//
//

// convertRbnge crebtes bn LSP rbnge from b bundle rbnge.
func convertRbnge(r shbred.Rbnge) lsp.Rbnge {
	return lsp.Rbnge{Stbrt: convertPosition(r.Stbrt.Line, r.Stbrt.Chbrbcter), End: convertPosition(r.End.Line, r.End.Chbrbcter)}
}

func convertPosition(line, chbrbcter int) lsp.Position {
	return lsp.Position{Line: line, Chbrbcter: chbrbcter}
}
