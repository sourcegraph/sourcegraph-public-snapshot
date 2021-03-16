package api

// Progress is an aggregate type representing a progress update.
type Progress struct {
	// Done is true if this is a final progress event.
	Done bool `json:"done"`

	// RepositoriesCount is the number of repositories being searched. It is
	// non-nil once the set of repositories has been resolved.
	RepositoriesCount *int `json:"repositoriesCount,omitempty"`

	// MatchCount is number of non-overlapping matches. If skipped is
	// non-empty, then this is a lower bound.
	MatchCount int `json:"matchCount"`

	// DurationMs is the wall clock time in milliseconds for this search.
	DurationMs int `json:"durationMs"`

	// Skipped is a description of shards or documents that were skipped. This
	// has a deterministic ordering. More important reasons will be listed
	// first. If a search is repeated, the final skipped list will be the
	// same.  However, within a search stream when a new skipped reason is
	// found, it may appear anywhere in the list.
	Skipped []Skipped `json:"skipped"`

	// Trace is the URL of an associated trace if the query is logging one.
	Trace string `json:"trace,omitempty"`
}

// Skipped is a description of shards or documents that were skipped.
type Skipped struct {
	// Reason is why a document/shard/repository was skipped. We group counts
	// by reason. eg ShardTimeout
	Reason SkippedReason `json:"reason"`
	// Title is a short message. eg "1,200 timed out".
	Title string `json:"title"`
	// Message is a message to show the user. Usually includes information
	// explaining the reason, count as well as a sample of the missing items.
	Message  string          `json:"message"`
	Severity SkippedSeverity `json:"severity"`
	// Suggested is a query expression to remedy the skip. eg "archived:yes".
	Suggested *SkippedSuggested `json:"suggested,omitempty"`
}

// SkippedSuggested is a query to suggest to the user to resolve the reason
// for skipping.
type SkippedSuggested struct {
	Title           string `json:"title"`
	QueryExpression string `json:"queryExpression"`
}

// SkippedReason is an enum for Skipped.Reason.
type SkippedReason string

const (
	// DocumentMatchLimit is when we found too many matches in a document, so
	// we stopped searching it.
	DocumentMatchLimit SkippedReason = "document-match-limit"
	// ShardMatchLimit is when we found too many matches in a
	// shard/repository, so we stopped searching it.
	ShardMatchLimit = "shard-match-limit"
	// RepositoryLimit is when we did not search a repository because the set
	// of repositories to search was too large.
	RepositoryLimit = "repository-limit"
	// ShardTimeout is when we ran out of time before searching a
	// shard/repository.
	ShardTimeout = "shard-timeout"
	// RepositoryCloning is when we could not search a repository because it
	// is not cloned.
	RepositoryCloning = "repository-cloning"
	// RepositoryMissing is when we could not search a repository because it
	// is not cloned and we failed to find it on the remote code host.
	RepositoryMissing = "repository-missing"
	// ExcludedFork is when we did not search a repository because it is a
	// fork.
	ExcludedFork = "repository-fork"
	// ExcludedArchive is when we did not search a repository because it is
	// archived.
	ExcludedArchive = "excluded-archive"
)

// SkippedSeverity is an enum for Skipped.Severity.
type SkippedSeverity string

const (
	SeverityInfo SkippedSeverity = "info"
	SeverityWarn                 = "warn"
)
