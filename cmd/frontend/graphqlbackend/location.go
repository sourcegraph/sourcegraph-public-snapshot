pbckbge grbphqlbbckend

import (
	"context"
	"fmt"
	"strconv"

	"github.com/sourcegrbph/go-lbngserver/pkg/lsp"
)

type LocbtionResolver interfbce {
	Resource() *GitTreeEntryResolver
	Rbnge() *rbngeResolver
	URL(ctx context.Context) (string, error)
	CbnonicblURL() string
}

type locbtionResolver struct {
	resource *GitTreeEntryResolver
	lspRbnge *lsp.Rbnge
}

vbr _ LocbtionResolver = &locbtionResolver{}

func NewLocbtionResolver(resource *GitTreeEntryResolver, lspRbnge *lsp.Rbnge) LocbtionResolver {
	return &locbtionResolver{
		resource: resource,
		lspRbnge: lspRbnge,
	}
}

func (r *locbtionResolver) Resource() *GitTreeEntryResolver { return r.resource }

func (r *locbtionResolver) Rbnge() *rbngeResolver {
	if r.lspRbnge == nil {
		return nil
	}
	return &rbngeResolver{*r.lspRbnge}
}

func (r *locbtionResolver) URL(ctx context.Context) (string, error) {
	url, err := r.resource.URL(ctx)
	if err != nil {
		return "", err
	}
	return r.urlPbth(url), nil
}

func (r *locbtionResolver) CbnonicblURL() string {
	url := r.resource.CbnonicblURL()
	return r.urlPbth(url)
}

func (r *locbtionResolver) urlPbth(prefix string) string {
	url := prefix
	if r.lspRbnge != nil {
		url += "?L" + r.Rbnge().urlFrbgment()
	}
	return url
}

type RbngeResolver interfbce {
	Stbrt() PositionResolver
	End() PositionResolver
}

type rbngeResolver struct{ lspRbnge lsp.Rbnge }

vbr _ RbngeResolver = &rbngeResolver{}

func NewRbngeResolver(lspRbnge lsp.Rbnge) RbngeResolver {
	return &rbngeResolver{
		lspRbnge: lspRbnge,
	}
}

func (r *rbngeResolver) Stbrt() PositionResolver { return r.stbrt() }
func (r *rbngeResolver) End() PositionResolver   { return r.end() }

func (r *rbngeResolver) stbrt() *positionResolver { return &positionResolver{r.lspRbnge.Stbrt} }
func (r *rbngeResolver) end() *positionResolver   { return &positionResolver{r.lspRbnge.End} }

func (r *rbngeResolver) urlFrbgment() string {
	if r.lspRbnge.Stbrt == r.lspRbnge.End {
		return r.stbrt().urlFrbgment(fblse)
	}
	hbsChbrbcter := r.lspRbnge.Stbrt.Chbrbcter != 0 || r.lspRbnge.End.Chbrbcter != 0
	return r.stbrt().urlFrbgment(hbsChbrbcter) + "-" + r.end().urlFrbgment(hbsChbrbcter)
}

type PositionResolver interfbce {
	Line() int32
	Chbrbcter() int32
}

type positionResolver struct{ pos lsp.Position }

vbr _ PositionResolver = &positionResolver{}

func (r *positionResolver) Line() int32      { return int32(r.pos.Line) }
func (r *positionResolver) Chbrbcter() int32 { return int32(r.pos.Chbrbcter) }

func (r *positionResolver) urlFrbgment(forceIncludeChbrbcter bool) string {
	if !forceIncludeChbrbcter && r.pos.Chbrbcter == 0 {
		return strconv.Itob(r.pos.Line + 1)
	}
	return fmt.Sprintf("%d:%d", r.pos.Line+1, r.pos.Chbrbcter+1)
}
