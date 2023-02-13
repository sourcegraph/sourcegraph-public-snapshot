package webhooks

import (
	"context"
	"encoding/json"
	"time"

	"github.com/graph-gophers/graphql-go"

	bgql "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/graphql"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// batchChange represents a batch change in a webhook payload.
type batchChange struct {
	ID            graphql.ID  `json:"id"`
	Namespace     graphql.ID  `json:"namespace_id"`
	Name          string      `json:"name"`
	Description   string      `json:"description"`
	State         string      `json:"state"`
	Creator       graphql.ID  `json:"creator_user_id"`
	LastApplier   *graphql.ID `json:"last_applier_user_id"`
	URL           string      `json:"url"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
	LastAppliedAt *time.Time  `json:"last_applied_at"`
	ClosedAt      *time.Time  `json:"closed_at"`
}

func MarshalBatchChange(ctx context.Context, db basestore.ShareableStore, bc *types.BatchChange) ([]byte, error) {
	namespaceID, err := bgql.MarshalNamespaceID(bc.NamespaceUserID, bc.NamespaceOrgID)
	if err != nil {
		return nil, errors.Wrap(err, "marshalling namespace")
	}

	namespace, err := database.NamespacesWith(db).GetByID(ctx, bc.NamespaceOrgID, bc.NamespaceUserID)
	if err != nil {
		return nil, errors.Wrap(err, "querying namespace")
	}

	url, err := bc.URL(ctx, namespace.Name)
	if err != nil {
		return nil, errors.Wrap(err, "building URL")
	}

	payload := batchChange{
		ID:            bgql.MarshalBatchChangeID(bc.ID),
		Namespace:     namespaceID,
		Name:          bc.Name,
		Description:   bc.Description,
		State:         bc.State().ToGraphQL(),
		Creator:       bgql.MarshalUserID(bc.CreatorID),
		LastApplier:   nullableMap(bc.LastApplierID, bgql.MarshalUserID),
		URL:           url,
		CreatedAt:     bc.CreatedAt,
		UpdatedAt:     bc.UpdatedAt,
		LastAppliedAt: nullable(bc.LastAppliedAt),
		ClosedAt:      nullable(bc.ClosedAt),
	}

	return json.Marshal(&payload)
}
