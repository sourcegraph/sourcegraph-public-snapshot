package reader

import "github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol"

type Element struct {
	ID      int
	Type    string
	Label   string
	Payload any
}

type Edge struct {
	OutV     int
	InV      int
	InVs     []int
	Document int
}

type ToolInfo struct {
	Name    string
	Version string
}

type MetaData struct {
	Version          string
	ProjectRoot      string
	PositionEncoding string
	ToolInfo         ToolInfo
}

type Range struct {
	protocol.RangeData
	Tag *protocol.RangeTag `json:"tag,omitempty"`
}

type ResultSet struct{}

type Moniker struct {
	Kind       string
	Scheme     string
	Identifier string
}

type PackageInformation struct {
	Name    string
	Version string
	Manager string
}

type Diagnostic struct {
	Severity       int
	Code           string
	Message        string
	Source         string
	StartLine      int
	StartCharacter int
	EndLine        int
	EndCharacter   int
}
