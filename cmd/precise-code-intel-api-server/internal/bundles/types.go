package bundles

// Location is an LSP-like location scoped to a dump.
type Location struct {
	DumpID int    `json:"dumpId"`
	Path   string `json:"path"`
	Range  Range  `json:"range"`
}

// Range is an inclusive bounds within a file.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Position is a unique position within a file.
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// MonikerData describes a moniker within a dump.
type MonikerData struct {
	Kind                 string `json:"kind"`
	Scheme               string `json:"scheme"`
	Identifier           string `json:"identifier"`
	PackageInformationID string `json:"packageInformationID"`
}

// PackageInformationData describes a package within a package manager system.
type PackageInformationData struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}
