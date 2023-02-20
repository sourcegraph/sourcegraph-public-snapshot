package shared

import "time"

type GetVulnerabilitiesArgs struct {
	// TODO
}

type GetVulnerabilityMatchesArgs struct {
	// TODO
}

type Vulnerability struct {
	ID               int    // internal ID
	SourceID         string // external ID
	Summary          string
	Details          string
	CPEs             []string
	CWEs             []string
	Aliases          []string
	Related          []string
	DataSource       string
	URLs             []string
	Severity         string
	CVSSVector       string
	CVSSScore        string
	Published        time.Time
	Modified         time.Time
	Withdrawn        time.Time
	AffectedPackages []AffectedPackage
}

// Data that varies across instances of a vulnerability
// Need to decide if this will be flat inside Vulnerability (and have multiple duplicate vulns)
// or a separate struct/table
type AffectedPackage struct {
	PackageName       string
	Language          string
	Namespace         string
	VersionConstraint []string
	Fixed             bool
	FixedIn           string
	AffectedSymbols   []AffectedSymbol
}

type AffectedSymbol struct {
	Path    string   `json:"path"`
	Symbols []string `json:"symbols"`
}

type VulnerabilityMatch struct {
	// TODO
}
