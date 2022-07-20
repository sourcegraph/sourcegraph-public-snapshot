package types

type BatchChangeChangesetArchivedState string

const (
	BatchChangeChangesetUnarchived     BatchChangeChangesetArchivedState = ""
	BatchChangeChangesetPendingArchive                                   = "pending"
	BatchChangeChangesetArchived                                         = "archived"
)

type BatchChangeChangesetAssociation struct {
	BatchChangeID int64
	ChangesetID   int64
	Detach        bool
	Archived      BatchChangeChangesetArchivedState
}
