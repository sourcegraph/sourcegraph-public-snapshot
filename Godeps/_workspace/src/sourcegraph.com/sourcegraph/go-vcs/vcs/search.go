package vcs

type Searcher interface {
	// Search searches the text of a repository at the given commit
	// ID.
	Search(CommitID, SearchOptions) ([]*SearchResult, error)
}

const (
	// FixedQuery is a value for SearchOptions.QueryType that
	// indicates the query is a fixed string, not a regex.
	FixedQuery = "fixed"

	// TODO(sqs): allow regexp searches, extended regexp searches, etc.
)
