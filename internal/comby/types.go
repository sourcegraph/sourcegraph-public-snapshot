package comby

import "archive/tar"

type Input interface {
	input()
}

type Tar struct {
	TarInputEventC chan TarInputEvent
}

type TarInputEvent struct {
	Header  tar.Header
	Content []byte
}

type ZipPath string
type DirPath string
type FileContent []byte

func (ZipPath) input()     {}
func (DirPath) input()     {}
func (FileContent) input() {}
func (Tar) input()         {}

type resultKind int

const (
	// MatchOnly means comby returns matches satisfying a pattern (no replacement)
	MatchOnly resultKind = iota
	// Replacement means comby returns the result of performing an in-place operation on file contents
	Replacement
	// Diff means comby returns a diff after performing an in-place operation on file contents
	Diff
	// NewlineSeparatedOutput means output the result of substituting the rewrite
	// template, newline-separated for each result.
	NewlineSeparatedOutput
)

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

	ResultKind resultKind

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

type ChunkMatch struct {
	Content string   `json:"content"`
	Start   Location `json:"start"`
	Ranges  []Range  `json:"ranges"`
}

type Result interface {
	result()
}

var (
	_ Result = (*FileMatchWithChunks)(nil)
	_ Result = (*FileMatch)(nil)
	_ Result = (*FileDiff)(nil)
	_ Result = (*FileReplacement)(nil)
	_ Result = (*Output)(nil)
)

func (*FileMatchWithChunks) result() {}
func (*FileMatch) result()           {}
func (*FileDiff) result()            {}
func (*FileReplacement) result()     {}
func (*Output) result()              {}

// FileMatchWithChunks represents all the chunk matches in a single file.
type FileMatchWithChunks struct {
	URI          string       `json:"uri"`
	ChunkMatches []ChunkMatch `json:"matches"`
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

// FileReplacement represents a file content been modified by a rewrite operation.
type FileReplacement struct {
	URI     string `json:"uri"`
	Content string `json:"rewritten_source"`
}

// Output represents content output by substituting variables in a rewrite template.
type Output struct {
	Value []byte // corresponds to stdout of a comby invocation.
}
