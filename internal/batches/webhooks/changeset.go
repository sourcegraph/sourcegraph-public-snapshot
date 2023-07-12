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

// changeset represents a changeset in a webhook payload.
type changeset struct {
	ID                 graphql.ID   `json:"id"`
	ExternalID         string       `json:"external_id"`
	BatchChangeIDs     []graphql.ID `json:"batch_change_ids"`
	RepositoryID       graphql.ID   `json:"repository_id"`
	CreatedAt          time.Time    `json:"created_at"`
	UpdatedAt          time.Time    `json:"updated_at"`
	Title              *string      `json:"title"`
	Body               *string      `json:"body"`
	AuthorName         *string      `json:"author_name"`
	AuthorEmail        *string      `json:"author_email"`
	State              string       `json:"state"`
	Labels             []string     `json:"labels"`
	ExternalURL        *string      `json:"external_url"`
	ForkNamespace      *string      `json:"fork_namespace"`
	ReviewState        *string      `json:"review_state"`
	CheckState         *string      `json:"check_state"`
	Error              *string      `json:"error"`
	SyncerError        *string      `json:"syncer_error"`
	ForkName           *string      `json:"fork_name"`
	OwnedByBatchChange *graphql.ID  `json:"owning_batch_change_id"`
}

const gqlChangesetQuery = `query Changeset($id: ID!) {
	node(id: $id) {
		... on ExternalChangeset {
			id
			externalID
			batchChanges {
				nodes {
					id
				}
			}
			repository {
				id
			}
			createdAt
			updatedAt
			title
			body
			author {
				name
				email
			}
			state
			labels {
				text
			}
			externalURL {
				url
			}
			forkNamespace
			reviewState
			checkState
			error
			syncerError
			forkName
			ownedByBatchChange
		}
	}
}`

type gqlChangesetResponse struct {
	Data struct {
		Node struct {
			ID           graphql.ID `json:"id"`
			ExternalID   string     `json:"externalId"`
			BatchChanges struct {
				Nodes []struct {
					ID graphql.ID `json:"id"`
				} `json:"nodes"`
			} `json:"batchChanges"`
			Repository struct {
				ID graphql.ID `json:"id"`
			} `json:"repository"`
			CreatedAt time.Time `json:"createdAt"`
			UpdatedAt time.Time `json:"updatedAt"`
			Title     *string   `json:"title"`
			Body      *string   `json:"body"`
			Author    *struct {
				Name  string `json:"name"`
				Email string `json:"email"`
			} `json:"author"`
			State  string `json:"state"`
			Labels []struct {
				Text string `json:"text"`
			} `json:"labels"`
			ExternalURL *struct {
				URL string `json:"url"`
			} `json:"externalURL"`
			ForkNamespace      *string     `json:"forkNamespace"`
			ForkName           *string     `json:"forkName"`
			ReviewState        *string     `json:"reviewState"`
			CheckState         *string     `json:"checkState"`
			Error              *string     `json:"error"`
			SyncerError        *string     `json:"syncerError"`
			OwnedByBatchChange *graphql.ID `json:"ownedByBatchChange"`
		}
	}
	Errors []gqlerrors.FormattedError
}

func marshalChangeset(ctx context.Context, client httpcli.Doer, id graphql.ID) ([]byte, error) {
	q := queryInfo{
		Name:      "Changeset",
		Query:     gqlChangesetQuery,
		Variables: map[string]any{"id": id},
	}

	var res gqlChangesetResponse
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
	var batchChangeIDs []graphql.ID
	for _, bc := range node.BatchChanges.Nodes {
		batchChangeIDs = append(batchChangeIDs, bc.ID)
	}

	var labels []string
	for _, label := range node.Labels {
		labels = append(labels, label.Text)
	}

	var authorName *string
	if node.Author != nil {
		authorName = &node.Author.Name
	}

	var authorEmail *string
	if node.Author != nil {
		authorEmail = &node.Author.Email
	}

	var externalURL *string
	if node.ExternalURL != nil {
		externalURL = &node.ExternalURL.URL
	}

	return json.Marshal(changeset{
		ID:                 node.ID,
		ExternalID:         node.ExternalID,
		BatchChangeIDs:     batchChangeIDs,
		RepositoryID:       node.Repository.ID,
		CreatedAt:          node.CreatedAt,
		UpdatedAt:          node.UpdatedAt,
		Title:              node.Title,
		Body:               node.Body,
		AuthorName:         authorName,
		AuthorEmail:        authorEmail,
		State:              node.State,
		Labels:             labels,
		ExternalURL:        externalURL,
		ForkNamespace:      node.ForkNamespace,
		ForkName:           node.ForkName,
		ReviewState:        node.ReviewState,
		CheckState:         node.CheckState,
		Error:              node.Error,
		SyncerError:        node.SyncerError,
		OwnedByBatchChange: node.OwnedByBatchChange,
	})
}
