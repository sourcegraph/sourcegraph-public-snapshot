package webhooks

import (
	"context"
	"encoding/json"
	"time"

	"github.com/graph-gophers/graphql-go"

	bgql "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/graphql"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// changeset represents a changeset in a webhook payload.
type changeset struct {
	ID                  graphql.ID   `json:"id"`
	ExternalID          string       `json:"external_id"`
	BatchChangeIDs      []graphql.ID `json:"batch_change_ids"`
	OwningBatchChangeID *graphql.ID  `json:"owning_batch_change_id"`
	RepositoryID        graphql.ID   `json:"repository_id"`
	CreatedAt           time.Time    `json:"created_at"`
	UpdatedAt           time.Time    `json:"updated_at"`
	Title               *string      `json:"title"`
	Body                *string      `json:"body"`
	AuthorName          *string      `json:"author_name"`
	State               string       `json:"state"`
	Labels              []string     `json:"labels"`
	ExternalURL         *string      `json:"external_url"`
	ForkNamespace       *string      `json:"fork_namespace"`
	ReviewState         *string      `json:"review_state"`
	CheckState          *string      `json:"check_state"`
	Error               *string      `json:"error"`
	SyncerError         *string      `json:"syncer_error"`
}

func MarshalChangeset(ctx context.Context, db basestore.ShareableStore, cs *types.Changeset) ([]byte, error) {
	batchChangeIDs := make([]graphql.ID, 0, len(cs.BatchChanges))
	for _, assoc := range cs.BatchChanges {
		batchChangeIDs = append(batchChangeIDs, bgql.MarshalBatchChangeID(assoc.BatchChangeID))
	}

	labels := cs.Labels()
	labelNames := make([]string, 0, len(cs.Labels()))
	for _, label := range labels {
		labelNames = append(labelNames, label.Name)
	}

	var (
		title       *string
		body        *string
		authorName  *string
		externalURL *string
	)

	if cs.Published() && !cs.IsImporting() {
		for name, field := range map[string]struct {
			method func() (string, error)
			out    **string
		}{
			"title":        {cs.Title, &title},
			"body":         {cs.Body, &body},
			"author name":  {cs.AuthorName, &authorName},
			"external URL": {cs.URL, &externalURL},
		} {
			value, err := field.method()
			if err != nil {
				return nil, errors.Wrapf(err, "getting %s", name)
			}
			*field.out = &value
		}
	}

	payload := changeset{
		ID:                  bgql.MarshalChangesetID(cs.ID),
		ExternalID:          cs.ExternalID,
		BatchChangeIDs:      batchChangeIDs,
		OwningBatchChangeID: nullableMap(cs.OwnedByBatchChangeID, bgql.MarshalBatchChangeID),
		RepositoryID:        bgql.MarshalRepoID(cs.RepoID),
		CreatedAt:           cs.CreatedAt,
		UpdatedAt:           cs.UpdatedAt,
		Title:               title,
		Body:                body,
		AuthorName:          authorName,
		State:               string(cs.State),
		Labels:              labelNames,
		ExternalURL:         externalURL,
		ForkNamespace:       nullable(cs.ExternalForkNamespace),
		ReviewState:         nullable(string(cs.ExternalReviewState)),
		CheckState:          nullable(string(cs.ExternalCheckState)),
		Error:               cs.FailureMessage,
		SyncerError:         cs.SyncErrorMessage,
	}

	return json.Marshal(&payload)
}
