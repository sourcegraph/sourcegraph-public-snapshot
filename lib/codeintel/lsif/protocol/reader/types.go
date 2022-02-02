package reader

import "github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol"

type Element struct {
	ID      int
	Type    string
	Label   string
	Payload interface{}
}

type Edge struct {
	OutV     int   `json:"outV"`
	InV      int   `json:"inV"`
	InVs     []int `json:"inVs"`
	Document int   `json:"document"`
}

type ToolInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type MetaData struct {
	Version          string   `json:"version"`
	ProjectRoot      string   `json:"projectRoot"`
	PositionEncoding string   `json:"positionEncoding"`
	ToolInfo         ToolInfo `json:"toolInfo"`
}

type Range struct {
	protocol.RangeData
	Tag *protocol.RangeTag `json:"tag,omitempty"`
}

type ResultSet struct{}

type Moniker struct {
	Kind       string `json:"kind"`
	Scheme     string `json:"scheme"`
	Identifier string `json:"identifier"`
}

type PackageInformation struct {
	Name    string `json:"name"`
	Version string `json:"version"`
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
