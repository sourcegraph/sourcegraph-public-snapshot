package comby

type Input interface {
	Value()
}

type ZipPath string
type DirPath string

func (z ZipPath) Value() {}
func (d DirPath) Value() {}

type Args struct {
	// An Input to process (either a path to a directory or zip file)
	Input

	// A template pattern that expresses what to match
	MatchTemplate string

	// A rule that places constraints on matching or rewriting
	Rule string

	// A template pattern that expresses how matches should be rewritten
	RewriteTemplate string

	// Matcher is a file extension (e.g., '.go') which denotes which language parser to use
	Matcher string

	// If MatchOnly is set to true, then comby will only find matches and not perform replacement
	MatchOnly bool

	// FilePatterns is a list of file patterns (suffixes) to filter and process
	FilePatterns []string

	// NumWorkers is the number of worker processes to fork in parallel
	NumWorkers int
}

// Location is the location in a file
type Location struct {
	Offset int `json:"offset"`
	Line   int `json:"line"`
	Column int `json:"column"`
}

// Range is a range of start location to end location
type Range struct {
	Start Location `json:"start"`
	End   Location `json:"end"`
}

// Match represents a range of matched characters and the matched content
type Match struct {
	Range   Range  `json:"range"`
	Matched string `json:"matched"`
}

// FileMatch represents all the matches in a single file
type FileMatch struct {
	URI     string  `json:"uri"`
	Matches []Match `json:"matches"`
}

// FileDiff represents a diff for a file
type FileDiff struct {
	URI  string `json:"uri"`
	Diff string `json:"diff"`
}
