package comby

type Input interface {
	Value() string
}

type ZipPath struct {
	value string
}

type DirPath struct {
	value string
}

func (z *ZipPath) Value() string { return z.value }

func (d *DirPath) Value() string { return d.value }

type Args = struct {
	// An Input to process (either a path to a directory or zip file).
	Input

	// A template pattern that expresses what to match.
	MatchTemplate string

	// A template pattern that expresses how matches should be rewritten.
	RewriteTemplate string

	// Matcher is a file extension (e.g., '.go') which denotes which language parser to use
	Matcher string

	// If MatchOnly is set to true, then comby will only find matches and not perform replacement
	MatchOnly bool

	// FilePatterns is a list of file patterns (suffixes) to filter and process
	FilePatterns []string
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
