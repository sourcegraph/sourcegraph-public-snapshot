package webhooks

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graphql-go/graphql/gqlerrors"

	bgql "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/graphql"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
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

// gqlBatchChangeQuery is a graphQL query that fetches all the required
// batch change fields to craft the webhook payload from the internal API.
const gqlBatchChangeQuery = `query BatchChange($id: ID!) {
	node(id: $id) {
		... on BatchChange {
			id
			namespace {
				id
			}
			name
			description
			state
			creator {
				id
			}
			lastApplier {
				id
			}
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
		Node struct {
			ID            graphql.ID `json:"id"`
			Name          string     `json:"name"`
			Description   string     `json:"description"`
			State         string     `json:"state"`
			URL           string     `json:"url"`
			CreatedAt     time.Time  `json:"createdAt"`
			UpdatedAt     time.Time  `json:"updatedAt"`
			LastAppliedAt *time.Time `json:"lastAppliedAt"`
			ClosedAt      *time.Time `json:"closedAt"`
			Namespace     struct {
				ID graphql.ID `json:"id"`
			} `json:"namespace"`
			Creator struct {
				ID graphql.ID `json:"id"`
			} `json:"creator"`
			LastApplier struct {
				ID *graphql.ID `json:"id"`
			} `json:"lastApplier"`
		}
	}
	Errors []gqlerrors.FormattedError
}

func MarshalBatchChange(ctx context.Context, bc *types.BatchChange) ([]byte, error) {
	marshalledBatchChangeID := bgql.MarshalBatchChangeID(bc.ID)

	q := queryInfo{}
	q.Query = gqlBatchChangeQuery
	q.Variables = map[string]any{"id": marshalledBatchChangeID}

	reqBody, err := json.Marshal(q)
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

	if len(res.Errors) > 0 {
		var combined error
		for _, err := range res.Errors {
			combined = errors.Append(combined, err)
		}
		return nil, combined
	}

	node := res.Data.Node

	return json.Marshal(batchChange{
		ID:            node.ID,
		Namespace:     node.Namespace.ID,
		Name:          node.Name,
		Description:   node.Description,
		State:         node.State,
		Creator:       node.Creator.ID,
		LastApplier:   node.LastApplier.ID,
		URL:           node.URL,
		CreatedAt:     node.CreatedAt,
		UpdatedAt:     node.UpdatedAt,
		LastAppliedAt: node.LastAppliedAt,
		ClosedAt:      node.ClosedAt,
	})
}
