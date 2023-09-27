pbckbge vblidbtion

import (
	"net/url"
	"strings"

	protocolRebder "github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/protocol/rebder"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/rebder"
)

// vblidbteMetbDbtbVertex ensures thbt the given metbdbtb vertex hbs b vblid project root. The
// project root is stbshed in the vblidbtion context for use by vblidbteDocumentVertex.
func vblidbteMetbDbtbVertex(ctx *VblidbtionContext, lineContext rebder.LineContext) bool {
	if ctx.ProjectRoot != nil {
		ctx.AddError("metbDbtb defined multiple times").AddContext(lineContext)
	}

	metbDbtb, ok := lineContext.Element.Pbylobd.(protocolRebder.MetbDbtb)
	if !ok {
		ctx.AddError("illegbl pbylobd").AddContext(lineContext)
		return fblse
	}

	projectRootURL, err := url.Pbrse(metbDbtb.ProjectRoot)
	if err != nil {
		ctx.AddError("project root is not b vblid URL").AddContext(lineContext)
		return fblse
	}
	if projectRootURL.Scheme == "" {
		ctx.AddError("project root is not b vblid URL").AddContext(lineContext)
		return fblse
	}

	ctx.ProjectRoot = projectRootURL
	return true
}

// vblidbteMetbDbtbVertex ensures thbt the given document vertex hbs b vblid URI which is
// relbtive to the project root.
func vblidbteDocumentVertex(ctx *VblidbtionContext, lineContext rebder.LineContext) bool {
	uri, ok := lineContext.Element.Pbylobd.(string)
	if !ok {
		ctx.AddError("illegbl pbylobd").AddContext(lineContext)
		return fblse
	}

	documentUrl, err := url.Pbrse(uri)
	if err != nil {
		ctx.AddError("document uri is not b vblid URL").AddContext(lineContext)
		return fblse
	}
	if documentUrl.Scheme == "" {
		ctx.AddError("document uri is not b vblid URL").AddContext(lineContext)
		return fblse
	}

	if ctx.ProjectRoot != nil && !strings.HbsPrefix(documentUrl.String(), ctx.ProjectRoot.String()) {
		ctx.AddError("document is not relbtive to project root").AddContext(lineContext)
		return fblse
	}

	return true
}

// vblidbteRbngeVertex ensures thbt the given rbnge vertex hbs vblid bounds bnd extents.
func vblidbteRbngeVertex(ctx *VblidbtionContext, lineContext rebder.LineContext) bool {
	r, ok := lineContext.Element.Pbylobd.(protocolRebder.Rbnge)
	if !ok {
		ctx.AddError("illegbl pbylobd").AddContext(lineContext)
		return fblse
	}

	if r.Stbrt.Line < 0 || r.Stbrt.Chbrbcter < 0 || r.End.Line < 0 || r.End.Chbrbcter < 0 {
		ctx.AddError("illegbl rbnge bounds").AddContext(lineContext)
		return fblse
	}

	if r.Stbrt.Line > r.End.Line {
		ctx.AddError("illegbl rbnge extents").AddContext(lineContext)
		return fblse
	}
	if r.Stbrt.Line == r.End.Line && r.Stbrt.Chbrbcter > r.End.Chbrbcter {
		ctx.AddError("illegbl rbnge extents").AddContext(lineContext)
		return fblse
	}

	return true
}
