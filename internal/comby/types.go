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
