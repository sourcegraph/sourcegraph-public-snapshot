package webhooks

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/graph-gophers/graphql-go"

	bgql "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/graphql"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
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

const gqlBatchChangeQuery = `query BatchChange($id: ID!) {
	node(id: $id) {
		... on BatchChange {
			id
			namespace
			name
			description
			state
			creator
			lastApplier
			url
			createdAt
			updatedAt
			lastAppliedAt
			closedAt
		}
	}
}`

type gqlBatchChangeResponse struct {
	Data struct {
		Node batchChange
	}
}

func MarshalBatchChange(ctx context.Context, db basestore.ShareableStore, bc *types.BatchChange) ([]byte, error) {
	marshalledBatchChangeID := bgql.MarshalBatchChangeID(bc.ID)
	input := map[string]any{"id": marshalledBatchChangeID}

	reqBody, err := json.Marshal(map[string]any{"query": gqlBatchChangeQuery})
	if err != nil {
		return nil, errors.Wrap(err, "marshal request body")
	}

	url, err := gqlURL("BatchChange")
	if err != nil {
		return nil, errors.Wrap(err, "construct frontend URL")
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, errors.Wrap(err, "construct request")
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpcli.InternalDoer.Do(req.WithContext(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "do request")
	}
	defer resp.Body.Close()

	var res gqlBatchChangeResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, errors.Wrap(err, "decode response")
	}

	return json.Marshal(res.Data.Node)
}
