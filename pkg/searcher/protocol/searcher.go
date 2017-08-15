// Package protocol contains structures used by the searcher API.
package protocol

// Request represents a request to searcher
type Request struct {
	// Repo is which repository to search. eg "github.com/gorilla/mux"
	Repo string

	// Commit is which commit to search. It is required to be resolved,
	// not a ref like HEAD or master. eg
	// "599cba5e7b6137d46ddf58fb1765f5d928e69604"
	Commit string

	PatternInfo
}

// PatternInfo describes a search request on a repo. Most of the fields
// are based on PatternInfo used in vscode.
type PatternInfo struct {
	// Pattern is the search query. It is a regular expression if IsRegExp
	// is true, otherwise a fixed string. eg "route variable"
	Pattern string

	// IsRegExp if true will treat the Pattern as a regular expression.
	IsRegExp bool

	// IsWordMatch if true will only match the pattern at word boundaries.
	IsWordMatch bool

	// IsCaseSensitive if false will ignore the case of text and pattern
	// when finding matches.
	IsCaseSensitive bool

	// ExcludePattern is a glob pattern that should not match the returned files.
	// eg '**/node_modules'
	ExcludePattern string

	// Include pattern is a glob pattern that should match the returned files.
	// eg '**/*.go' to search go files.
	IncludePattern string

	// FileMatchLimit limits the number of files with matches that are returned.
	FileMatchLimit int
}

// Response represents the response from a Search request.
type Response struct {
	Matches []FileMatch

	// LimitHit is true if Matches may not include all FileMatches.
	LimitHit bool
}

// FileMatch is the struct used by vscode to receive search results
type FileMatch struct {
	Path        string
	LineMatches []LineMatch

	// LimitHit is true if LineMatches may not include all LineMatches.
	LimitHit bool
}

// LineMatch is the struct used by vscode to receive search results for a line.
type LineMatch struct {
	// Preview is the matched line.
	Preview string

	// LineNumber is the 0-based line number. Note: Our editors present
	// 1-based line numbers, but internally vscode uses 0-based.
	LineNumber int

	// OffsetAndLengths is a slice of 2-tuples (Offset, Length)
	// representing each match on a line.
	OffsetAndLengths [][]int

	// LimitHit is true if OffsetAndLengths may not include all OffsetAndLengths.
	LimitHit bool
}
