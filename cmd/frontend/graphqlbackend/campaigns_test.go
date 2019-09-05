package graphqlbackend

import (
	"testing"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
)

func TestCampaigns(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := dbtesting.TestContext(t)
	s := db.NewCampaignsStore(dbconn.Global)
	sr := schemaResolver{CampaignsStore: s}

	query := `
		query Campaigns {
			campaigns(first: 3) {
				id
				name
				description
				createdAt
				updatedAt
			}
		}
	`

	schema, err := graphql.ParseSchema(Schema, &sr, graphql.UseFieldResolvers())
	if err != nil {
		t.Fatal(err)
	}

	result := schema.Exec(ctx, query, "", nil)

	t.Logf("%+v", result)
}
