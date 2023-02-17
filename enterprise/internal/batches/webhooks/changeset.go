package webhooks

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/graph-gophers/graphql-go"

	bgql "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/graphql"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
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
	ForkName            *string      `json:"fork_name"`
	ForkNamespace       *string      `json:"fork_namespace"`
	ReviewState         *string      `json:"review_state"`
	CheckState          *string      `json:"check_state"`
	Error               *string      `json:"error"`
	SyncerError         *string      `json:"syncer_error"`
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
		}
	}
}`

type gqlChangesetResponse struct {
	Data struct {
		Node struct {
			ID           graphql.ID `json:"id"`
			ExternalID   string     `json:"externalId"`
			BatchChanges struct {
				Nodes struct {
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
				Name string `json:"name"`
			} `json:"author"`
			State  string `json:"state"`
			Labels []struct {
				Text string `json:"text"`
			} `json:"labels"`
			ExternalURL *struct {
				URL string `json:"url"`
			} `json:"externalURL"`
			ForkNamespace *string `json:"forkNamespace"`
			ReviewState   *string `json:"reviewState"`
			CheckState    *string `json:"checkState"`
			Error         *string `json:"error"`
			SyncerError   *string `json:"syncerError"`
		}
	}
}

func marshalChangeset(ctx context.Context, id graphql.ID) ([]byte, error) {
	q := queryInfo{}
	q.Query = gqlChangesetQuery
	q.Variables = map[string]any{"id": id}

	reqBody, err := json.Marshal(q)
	if err != nil {
		return nil, errors.Wrap(err, "marshal request body")
	}

	url, err := gqlURL("Changeset")
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

	return json.Marshal(changeset{
		ID:                  node.ID,
		ExternalID:          node.E,
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
		ForkName:            nullable(cs.ExternalForkName),
		ReviewState:         nullable(string(cs.ExternalReviewState)),
		CheckState:          nullable(string(cs.ExternalCheckState)),
		Error:               cs.FailureMessage,
		SyncerError:         cs.SyncErrorMessage,
	})
}
