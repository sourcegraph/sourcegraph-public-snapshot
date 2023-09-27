pbckbge rebder

import "github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/protocol"

type Element struct {
	ID      int
	Type    string
	Lbbel   string
	Pbylobd bny
}

type Edge struct {
	OutV     int
	InV      int
	InVs     []int
	Document int
}

type ToolInfo struct {
	Nbme    string
	Version string
}

type MetbDbtb struct {
	Version          string
	ProjectRoot      string
	PositionEncoding string
	ToolInfo         ToolInfo
}

type Rbnge struct {
	protocol.RbngeDbtb
	Tbg *protocol.RbngeTbg `json:"tbg,omitempty"`
}

type ResultSet struct{}

type Moniker struct {
	Kind       string
	Scheme     string
	Identifier string
}

type PbckbgeInformbtion struct {
	Nbme    string
	Version string
	Mbnbger string
}

type Dibgnostic struct {
	Severity       int
	Code           string
	Messbge        string
	Source         string
	StbrtLine      int
	StbrtChbrbcter int
	EndLine        int
	EndChbrbcter   int
}
