package comby

// A result is either a match result or a diff result. These members are
// mutually exclusive, but are bundled together so that we can share the
// unmarshalling code.
type Input = struct {
	ZipPath string
	DirPath string
}

type Args = struct {
	Input
	MatchTemplate   string
	RewriteTemplate string
	Matcher         string
	MatchOnly       bool
	FilePatterns    []string
}

type Range struct {
	Offset int `json:"offset"`
	Line   int `json:"line"`
	Column int `json:"column"`
}

// {"uri":"/private/tmp/rrrr/doc.go","matches":[{"range":{"start":{"offset":215,"line":1,"column":216},"end":{"offset":222,"line":1,"column":223}},"environment":[],"matched":"package"}]}
type Match struct {
	URI     string  `json:"uri"`
	Matches []Range `json:"matches"`
	Matched string  `json:"matched"`
}

type Diff struct {
	URI  string `json:"uri"`
	Diff string `json:"diff"`
}

// Nuke this
type Result struct {
	Matches *[]Match
	Diffs   *[]Diff
}
