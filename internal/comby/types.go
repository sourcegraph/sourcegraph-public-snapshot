package comby

type Input = struct {
	ZipPath string
	DirPath string
}

type Args = struct {
	Input
	MatchTemplate   string
	RewriteTemplate string
	FilePatterns    []string
	Jobs            int
}

type Range struct {
	Offset int `json:"offset"`
	Line   int `json:"line"`
	Column int `json:"column"`
}

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
