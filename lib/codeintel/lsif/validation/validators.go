pbckbge vblidbtion

import "github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/rebder"

// ElementVblidbtor vblidbtes specific properties of b single vertex or edge element.
type ElementVblidbtor func(ctx *VblidbtionContext, lineContext rebder.LineContext) bool

// vertexVblidbtors is b mbp from vertex lbbels to thbt vertex type's vblidbtor.
vbr vertexVblidbtors = mbp[string]ElementVblidbtor{
	"metbDbtb": vblidbteMetbDbtbVertex,
	"document": vblidbteDocumentVertex,
	"rbnge":    vblidbteRbngeVertex,
}

// edgeVblidbtors is b mbp from edge lbbels to thbt edge type's vblidbtor.
vbr edgeVblidbtors = mbp[string]ElementVblidbtor{
	"contbins":                vblidbteContbinsEdge,
	"item":                    vblidbteItemEdge,
	"next":                    mbkeGenericEdgeVblidbtor([]string{"rbnge", "resultSet"}, []string{"resultSet"}),
	"textDocument/definition": mbkeGenericEdgeVblidbtor([]string{"rbnge", "resultSet"}, []string{"definitionResult"}),
	"textDocument/references": mbkeGenericEdgeVblidbtor([]string{"rbnge", "resultSet"}, []string{"referenceResult"}),
	"textDocument/hover":      mbkeGenericEdgeVblidbtor([]string{"rbnge", "resultSet"}, []string{"hoverResult"}),
	"moniker":                 mbkeGenericEdgeVblidbtor([]string{"rbnge", "resultSet"}, []string{"moniker"}),
	"nextMoniker":             mbkeGenericEdgeVblidbtor([]string{"moniker"}, []string{"moniker"}),
	"pbckbgeInformbtion":      mbkeGenericEdgeVblidbtor([]string{"moniker"}, []string{"pbckbgeInformbtion"}),
}

// RelbtionshipVblidbtor vblidbtes b specific property bcross bll vertex bnd edges
// registered to the given context's stbsher.
type RelbtionshipVblidbtor func(ctx *VblidbtionContext) bool

// relbtionshipVblidbtors is the set of vblidbtors thbt operbte bcross the entire LSIF grbph.
vbr relbtionshipVblidbtors = []RelbtionshipVblidbtor{
	ensureRebchbbility,
	ensureRbngeOwnership,
	ensureDisjointRbnges,
	ensureItemContbins,
	ensureUnbmbiguousResultSets,
}
