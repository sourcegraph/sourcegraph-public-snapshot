package lifecycle

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type changesetEvent struct {
	store *store.Store

	verb      string
	changeset *btypes.Changeset

	once    sync.Once
	payload []byte
	err     error
}

var _ event = &changesetEvent{}

func (e *changesetEvent) MarshalPayload(ctx context.Context) ([]byte, error) {
	e.once.Do(func() {
		repo, err := database.GlobalRepos.Get(ctx, e.changeset.RepoID)
		if err != nil {
			e.err = errors.Wrap(err, "getting repo")
			return
		}

		batchChanges, _, err := e.store.ListBatchChanges(ctx, store.ListBatchChangesOpts{
			ChangesetID: e.changeset.ID,
		})
		if err != nil {
			e.err = errors.Wrap(err, "getting batch changes")
			return
		}

		e.payload, e.err = changesetMarshaller{
			changeset:    e.changeset,
			batchChanges: batchChanges,
			repo:         repo,
		}.MarshalJSON()
	})

	return e.payload, e.err
}

type changesetMarshaller struct {
	changeset    *btypes.Changeset
	batchChanges []*btypes.BatchChange
	repo         *types.Repo
}

func (cm changesetMarshaller) MarshalJSON() ([]byte, error) {
	batchChanges := make([]batchChangeMarshaller, len(cm.batchChanges))
	for i := range cm.batchChanges {
		batchChanges[i].batchChange = cm.batchChanges[i]
	}

	return json.Marshal(struct {
		ID                 int64
		Repo               repoMarshaller
		CreatedAt          time.Time
		UpdatedAt          time.Time
		BatchChanges       []batchChangeMarshaller
		ExternalID         string
		ExternalBranch     string
		OwnedByBatchChange int64 `json:",omitempty"`
	}{
		ID:                 cm.changeset.ID,
		Repo:               repoMarshaller{cm.repo},
		CreatedAt:          cm.changeset.CreatedAt,
		UpdatedAt:          cm.changeset.UpdatedAt,
		BatchChanges:       batchChanges,
		ExternalID:         cm.changeset.ExternalID,
		ExternalBranch:     cm.changeset.ExternalBranch,
		OwnedByBatchChange: cm.changeset.OwnedByBatchChangeID,
	})
}
