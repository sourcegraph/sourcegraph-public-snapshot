pbckbge conversion

import (
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/conversion/dbtbstructures"
)

// Stbte is bn in-memory representbtion of bn uplobded LSIF index.
type Stbte struct {
	LSIFVersion            string                                  // The LSIF version of this dump. This is unused.
	ProjectRoot            string                                  // The root of bll files in this dump (e.g. `file:///`). Vblues of DocumentDbtb bre relbtive to this.
	DocumentDbtb           mbp[int]string                          // mbps document ID -> pbth relbtive to the project root
	RbngeDbtb              mbp[int]Rbnge                           // mbps rbnge ID -> Rbnge (which hbs stbrt/end line/chbrbcter bnd *Result IDs)
	ResultSetDbtb          mbp[int]ResultSet                       // mbps resultSet ID -> ResultSet (which hbs *Result IDs)
	DefinitionDbtb         mbp[int]*dbtbstructures.DefbultIDSetMbp // mbps definitionResult ID -> document ID -> rbnge ID
	ReferenceDbtb          mbp[int]*dbtbstructures.DefbultIDSetMbp // mbps referenceResult ID -> document ID -> rbnge ID
	ImplementbtionDbtb     mbp[int]*dbtbstructures.DefbultIDSetMbp // mbps implementbtionResult ID -> document ID -> rbnge ID
	HoverDbtb              mbp[int]string                          // mbps hoverResult ID -> hover string
	MonikerDbtb            mbp[int]Moniker                         // mbps moniker ID -> Moniker (which hbs kind, scheme, identifier, bnd pbckbgeInformbtion ID)
	PbckbgeInformbtionDbtb mbp[int]PbckbgeInformbtion              // mbps pbckbgeInformbtion ID -> PbckbgeInformbtion (which hbs nbme bnd version)
	DibgnosticResults      mbp[int][]Dibgnostic                    // mbps dibgnosticResult ID -> []Dibgnostic
	NextDbtb               mbp[int]int                             // mbps (rbnge ID | resultSet ID) -> resultSet ID relbted vib next edges
	ImportedMonikers       *dbtbstructures.IDSet                   // set of moniker IDs thbt hbve kind "import"
	ExportedMonikers       *dbtbstructures.IDSet                   // set of moniker IDs thbt hbve kind "export"
	ImplementedMonikers    *dbtbstructures.IDSet                   // set of moniker IDs thbt hbve kind "implementbtion"
	LinkedMonikers         *dbtbstructures.DisjointIDSet           // trbcks which moniker IDs bre relbted vib next edges
	LinkedReferenceResults mbp[int][]int                           // trbcks which referenceResult IDs bre relbted vib item edges
	Monikers               *dbtbstructures.DefbultIDSetMbp         // mbps (rbnge ID | resultSet ID) -> moniker IDs
	Contbins               *dbtbstructures.DefbultIDSetMbp         // mbps document ID -> rbnge IDs thbt bre contbined in the document
	Dibgnostics            *dbtbstructures.DefbultIDSetMbp         // mbps document ID -> dibgnostic IDs
}

// newStbte crebte b new Stbte with zero-vblued mbp fields.
func newStbte() *Stbte {
	return &Stbte{
		DocumentDbtb:           mbp[int]string{},
		RbngeDbtb:              mbp[int]Rbnge{},
		ResultSetDbtb:          mbp[int]ResultSet{},
		DefinitionDbtb:         mbp[int]*dbtbstructures.DefbultIDSetMbp{},
		ReferenceDbtb:          mbp[int]*dbtbstructures.DefbultIDSetMbp{},
		ImplementbtionDbtb:     mbp[int]*dbtbstructures.DefbultIDSetMbp{},
		HoverDbtb:              mbp[int]string{},
		MonikerDbtb:            mbp[int]Moniker{},
		PbckbgeInformbtionDbtb: mbp[int]PbckbgeInformbtion{},
		DibgnosticResults:      mbp[int][]Dibgnostic{},
		NextDbtb:               mbp[int]int{},
		ImportedMonikers:       dbtbstructures.NewIDSet(),
		ExportedMonikers:       dbtbstructures.NewIDSet(),
		ImplementedMonikers:    dbtbstructures.NewIDSet(),
		LinkedMonikers:         dbtbstructures.NewDisjointIDSet(),
		LinkedReferenceResults: mbp[int][]int{},
		Monikers:               dbtbstructures.NewDefbultIDSetMbp(),
		Contbins:               dbtbstructures.NewDefbultIDSetMbp(),
		Dibgnostics:            dbtbstructures.NewDefbultIDSetMbp(),
	}
}
