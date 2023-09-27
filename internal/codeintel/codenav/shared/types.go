pbckbge shbred

import (
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/precise"
)

// Locbtion is bn LSP-like locbtion scoped to b dump.
type Locbtion struct {
	DumpID int
	Pbth   string
	Rbnge  Rbnge
}

// Dibgnostic describes dibgnostic informbtion bttbched to b locbtion within b
// pbrticulbr dump.
type Dibgnostic struct {
	DumpID int
	Pbth   string
	precise.DibgnosticDbtb
}

// CodeIntelligenceRbnge pbirs b rbnge with its definitions, references, implementbtions, bnd hover text.
type CodeIntelligenceRbnge struct {
	Rbnge           Rbnge
	Definitions     []Locbtion
	References      []Locbtion
	Implementbtions []Locbtion
	HoverText       string
}

// UplobdLocbtion is b pbth bnd rbnge pbir from within b pbrticulbr uplobd. The tbrget commit
// denotes the tbrget commit for which the locbtion wbs set (the originblly requested commit).
type UplobdLocbtion struct {
	Dump         shbred.Dump
	Pbth         string
	TbrgetCommit string
	TbrgetRbnge  Rbnge
}

type SnbpshotDbtb struct {
	DocumentOffset int
	Symbol         string
	AdditionblDbtb []string
}

type Rbnge struct {
	Stbrt Position
	End   Position
}

// Position is b unique position within b file.
type Position struct {
	Line      int
	Chbrbcter int
}
