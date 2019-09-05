package graphqlbackend

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
)

func TestCampaigns(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := dbtesting.TestContext(t)
	ctx = backend.WithAuthzBypass(ctx)

	s := db.NewCampaignsStore(dbconn.Global)

	sr := schemaResolver{CampaignsStore: s}

	user, err := db.Users.Create(ctx, db.NewUser{
		Email:           "alice@example.com",
		EmailIsVerified: true,
		Username:        "alice",
		DisplayName:     "alice",
		Password:        "test",
	})
	if err != nil {
		t.Fatalf("user creation failed: %s", err)
	}

	now := time.Now().UTC().Truncate(time.Microsecond)
	campaigns := make([]*types.Campaign, 3)
	for i := range campaigns {
		campaigns[i] = &types.Campaign{
			Name:            fmt.Sprintf("Upgrade ES-Lint %d", i),
			Description:     "All the Javascripts are belong to us",
			AuthorID:        user.ID,
			NamespaceUserID: user.ID,
			CreatedAt:       now,
			UpdatedAt:       now,
		}

		err := s.CreateCampaign(ctx, campaigns[i])
		if err != nil {
			t.Fatal(err)
		}
	}

	query := `
		query Campaigns {
			campaigns(first: 3) {
				nodes {
					id
					name
					description
					createdAt
					updatedAt
					author {
						id
						email
						username
					}
					namespace {
						... on User {
							id
							email
							username
						}
					}
				}
				totalCount
				pageInfo { hasNextPage }
			}
		}
	`

	type pageInfo struct {
		HasNextPage bool `json:"hasNextPage"`
	}

	type node struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		CreatedAt   string `json:"createdAt"`
		UpdatedAt   string `json:"updatedAt"`
	}

	type campaignsQuery struct {
		Nodes      []node   `json:"nodes"`
		TotalCount int      `json:"totalCount"`
		PageInfo   pageInfo `json:"pageInfo"`
	}

	type queryResult struct {
		Campaigns campaignsQuery `json:"campaigns"`
	}

	want := queryResult{
		Campaigns: campaignsQuery{
			Nodes:      []node{},
			TotalCount: len(campaigns),
			PageInfo:   pageInfo{HasNextPage: false},
		},
	}
	for _, c := range campaigns {
		want.Campaigns.Nodes = append(want.Campaigns.Nodes, node{
			ID:          string(marshalCampaignID(c.ID)),
			Name:        c.Name,
			Description: c.Description,
			CreatedAt:   c.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   c.UpdatedAt.Format(time.RFC3339),
		})
	}

	schema, err := graphql.ParseSchema(Schema, &sr)
	if err != nil {
		t.Fatal(err)
	}

	result := schema.Exec(ctx, query, "", nil)
	if len(result.Errors) != 0 {
		t.Fatalf("exec produced errors: %+v", result.Errors)
	}

	var have queryResult
	if err := json.Unmarshal(result.Data, &have); err != nil {
		t.Fatalf("json unmarshal failed: %s", err)
	}
	if !reflect.DeepEqual(have, want) {
		t.Fatalf("wrong result: %s", cmp.Diff(have, want))
	}
}
