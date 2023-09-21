package graphqlbackend

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestSpongeLogNode(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	t.Run("log without an interpreter", func(t *testing.T) {
		log := database.SpongeLog{
			ID:   uuid.New(),
			Text: "example log text",
		}
		require.NoError(t, db.SpongeLogs().Save(ctx, log))
		RunTest(t, &Test{
			Schema:  mustParseGraphQLSchema(t, db),
			Context: ctx,
			Query: `
			query SpongeLogByID($id: ID!){
				node(id: $id) {
					__typename
					... on SpongeLog {
						log
						interpreter
					}
				}
			}`,
			ExpectedResult: `
			{
				"node": {
					"__typename": "SpongeLog",
					"log": "example log text",
					"interpreter": null
				}
			}`,
			Variables: map[string]any{
				"id": string(relay.MarshalID(spongeLogIDKind, log.ID.String())),
			},
		})
	})

	t.Run("log with an interpreter", func(t *testing.T) {
		log := database.SpongeLog{
			ID:          uuid.New(),
			Text:        "example log text",
			Interpreter: "example",
		}
		require.NoError(t, db.SpongeLogs().Save(ctx, log))

		RunTest(t, &Test{
			Schema:  mustParseGraphQLSchema(t, db),
			Context: ctx,
			Query: `
			query SpongeLogByID($id: ID!){
				node(id: $id) {
					__typename
					... on SpongeLog {
						log
						interpreter
					}
				}
			}`,
			ExpectedResult: `
			{
				"node": {
					"__typename": "SpongeLog",
					"log": "example log text",
					"interpreter": "example"
				}
			}`,
			Variables: map[string]any{
				"id": string(relay.MarshalID(spongeLogIDKind, log.ID.String())),
			},
		})
	})
}
