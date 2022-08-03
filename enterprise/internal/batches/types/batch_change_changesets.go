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

// AssociatedChangeset extends the changeset type to include the batch changes
// it is associated with. This does not directly represent a database table.
//
// TODO: store method to access this; store method to persist this.
type AssociatedChangeset struct {
	*Changeset
	associations map[int64]*BatchChangeChangesetAssociation
}

func NewAssociatedChangeset(cs *Changeset) *AssociatedChangeset {
	return &AssociatedChangeset{
		Changeset:    cs,
		associations: make(map[int64]*BatchChangeChangesetAssociation),
	}
}

func (ac *AssociatedChangeset) AddAssociation(assoc *BatchChangeChangesetAssociation) {
	ac.associations[assoc.BatchChangeID] = assoc
}

// RemoveBatchChangeID removes a batch change from the association by its ID.
// If the ID is not in the association, this method does nothing successfully.
func (ac *AssociatedChangeset) RemoveBatchChangeID(id int64) {
	delete(ac.associations, id)
}

// AttachedTo returns true if the changeset is currently attached to the batch
// change with the given batchChangeID.
func (ac *AssociatedChangeset) AttachedTo(batchChangeID int64) bool {
	_, ok := ac.associations[batchChangeID]
	return ok
}

// Attach attaches the batch change with the given ID to the changeset.
// If the batch change is already attached, this is a noop.
// If the batch change is still attached but is marked as to be detached,
// the detach flag is removed.
func (ac *AssociatedChangeset) Attach(batchChangeID int64) *BatchChangeChangesetAssociation {
	assoc := &BatchChangeChangesetAssociation{
		BatchChangeID: batchChangeID,
		ChangesetID:   ac.Changeset.ID,
		Detach:        false,
		Archived:      BatchChangeChangesetUnarchived,
	}
	ac.associations[batchChangeID] = assoc
	return assoc
}

// Detach marks the given batch change as to-be-detached. Returns true, if the
// batch change currently is attached to the batch change. This function is a noop,
// if the given batch change was not attached to the changeset.
func (ac *AssociatedChangeset) Detach(batchChangeID int64) bool {
	if assoc := ac.associations[batchChangeID]; assoc != nil {
		assoc.Detach = true
		return true
	}
	return false
}

// Archive marks the given changeset as to-be-archived. Returns true, if the
// changeset currently is attached to the batch change and *not* archived. This
// function is a noop, if the given changeset was already archived.
func (ac *AssociatedChangeset) Archive(batchChangeID int64) bool {
	if assoc := ac.associations[batchChangeID]; assoc != nil {
		assoc.Archived = BatchChangeChangesetArchived
		return true
	}
	return false
}

// ArchivedIn checks whether the changeset is archived in the given batch change.
func (ac *AssociatedChangeset) ArchivedIn(batchChangeID int64) bool {
	if assoc := ac.associations[batchChangeID]; assoc != nil {
		return assoc.Archived == BatchChangeChangesetArchived
	}
	return false
}
