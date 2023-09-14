package webhooks

import (
	"context"
	"encoding/json"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graphql-go/graphql/gqlerrors"

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

func marshalBatchChange(ctx context.Context, client httpcli.Doer, id graphql.ID) ([]byte, error) {
	q := queryInfo{
		Name:      "BatchChange",
		Query:     gqlBatchChangeQuery,
		Variables: map[string]any{"id": id},
	}

	var res gqlBatchChangeResponse
	if err := makeRequest(ctx, q, client, &res); err != nil {
		return nil, err
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
