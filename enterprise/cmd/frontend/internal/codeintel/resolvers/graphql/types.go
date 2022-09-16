package graphql

import (
	"time"

	autoindexingShared "github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/shared"
	store "github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/lsifstore"
	uploadsShared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
)

// AdjustedLocation is a path and range pair from within a particular upload. The adjusted commit
// denotes the target commit for which the location was adjusted (the originally requested commit).
type AdjustedLocation struct {
	Dump           store.Dump
	Path           string
	AdjustedCommit string
	AdjustedRange  lsifstore.Range
}

// AdjustedDiagnostic is a diagnostic from within a particular upload. The adjusted commit denotes
// the target commit for which the location was adjusted (the originally requested commit).
type AdjustedDiagnostic struct {
	lsifstore.Diagnostic
	Dump           store.Dump
	AdjustedCommit string
	AdjustedRange  lsifstore.Range
}

// AdjustedCodeIntelligenceRange stores definition, reference, and hover information for all ranges
// within a block of lines. The definition and reference locations have been adjusted to fit the
// target (originally requested) commit.
type AdjustedCodeIntelligenceRange struct {
	Range           lsifstore.Range
	Definitions     []AdjustedLocation
	References      []AdjustedLocation
	Implementations []AdjustedLocation
	HoverText       string
}

type RepositorySummary struct {
	RecentUploads           []uploadsShared.UploadsWithRepositoryNamespace
	RecentIndexes           []autoindexingShared.IndexesWithRepositoryNamespace
	LastUploadRetentionScan *time.Time
	LastIndexScan           *time.Time
}

type RetentionPolicyMatchCandidate struct {
	*store.ConfigurationPolicy
	Matched           bool
	ProtectingCommits []string
}
