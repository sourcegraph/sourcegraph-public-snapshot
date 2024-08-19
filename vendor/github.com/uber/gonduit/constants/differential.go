package constants

// DifferentialQueryCommitHashType are the commit types.
type DifferentialQueryCommitHashType string

const (
	// DifferentialQueryGitCommit is a git commit.
	DifferentialQueryGitCommit DifferentialQueryCommitHashType = "gtcm"
	// DifferentialQueryGitTr is something.
	DifferentialQueryGitTr DifferentialQueryCommitHashType = "gttr"
	// DifferentialQueryHGCommit is a mercurial commit.
	DifferentialQueryHGCommit DifferentialQueryCommitHashType = "hgcm"
)

// DifferentialStatusLegacy is the status of a differential revision.
type DifferentialStatusLegacy int

const (
	// DifferentialStatusLegacyNeedsReview is needs-review status.
	DifferentialStatusLegacyNeedsReview DifferentialStatusLegacy = 0
	// DifferentialStatusLegacyNeedsRevision is needs-revision status.
	DifferentialStatusLegacyNeedsRevision DifferentialStatusLegacy = 1
	// DifferentialStatusLegacyAccepted is accepted status.
	DifferentialStatusLegacyAccepted DifferentialStatusLegacy = 2
	// DifferentialStatusLegacyPublished is published (aka "closed") status.
	DifferentialStatusLegacyPublished DifferentialStatusLegacy = 3
	// DifferentialStatusLegacyAbandoned is abandoned status.
	DifferentialStatusLegacyAbandoned DifferentialStatusLegacy = 4
	// DifferentialStatusLegacyChangesPlanned is changes-planned status.
	DifferentialStatusLegacyChangesPlanned DifferentialStatusLegacy = 5
	// DifferentialStatusLegacyDraft is draft status. Value is the same as
	// for "needs-review" status because there were no legacy value for this
	// type.
	DifferentialStatusLegacyDraft DifferentialStatusLegacy = 0
)

// DifferentialQueryOrder is the order in which query results cna be ordered.
type DifferentialQueryOrder string

const (
	// DifferentialQueryOrderModified orders results by date modified.
	DifferentialQueryOrderModified DifferentialQueryOrder = "order-modified"
	// DifferentialQueryOrderCreated orders results by date created.
	DifferentialQueryOrderCreated DifferentialQueryOrder = "order-created"
)

// DifferentialGetCommitMessageEditType is value of edit type field.
type DifferentialGetCommitMessageEditType string

const (
	// DifferentialGetCommitMessageEdit mode hides read-only fields.
	DifferentialGetCommitMessageEdit DifferentialGetCommitMessageEditType = "edit"

	// DifferentialGetCommitMessageCreate mode hides read-only fields. "Field:"
	// templates are shown for some fields even if they are empty.
	DifferentialGetCommitMessageCreate DifferentialGetCommitMessageEditType = "create"

	// DifferentialGetCommitMessageRead shows all fields including read-only.
	// Value is empty string on purpose.
	DifferentialGetCommitMessageRead DifferentialGetCommitMessageEditType = ""
)
