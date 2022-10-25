package graphqlbackend

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/google/uuid"
	"github.com/graph-gophers/graphql-go/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	sgerrors "github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestListWebhooks(t *testing.T) {
	users := database.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

	ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
	webhookStore := database.NewMockWebhookStore()
	webhooks := []*types.Webhook{
		{
			ID:           1,
			CodeHostKind: extsvc.KindGitHub,
		},
		{
			ID:           2,
			CodeHostKind: extsvc.KindGitLab,
		},
		{
			ID:           3,
			CodeHostKind: extsvc.KindGitHub,
		},
		{
			ID:           4,
			CodeHostKind: extsvc.KindGitHub,
		},
	}
	webhookStore.ListFunc.SetDefaultHook(func(ctx2 context.Context, options database.WebhookListOptions) ([]*types.Webhook, error) {
		if options.Kind == extsvc.KindGitHub {
			return append([]*types.Webhook{webhooks[0]}, webhooks[2:4]...), nil
		}
		if options.Cursor != nil && options.Cursor.Value != "" {
			cursorVal, err := strconv.Atoi(options.Cursor.Value)
			assert.NoError(t, err)
			return webhooks[cursorVal-1:], nil
		}
		if options.LimitOffset != nil {
			return webhooks[options.Offset:options.Limit], nil
		}
		return webhooks, nil
	})
	webhookStore.CountFunc.SetDefaultHook(func(ctx context.Context, opts database.WebhookListOptions) (int, error) {
		whs, err := webhookStore.List(ctx, opts)
		return len(whs), err
	})

	db := database.NewMockDB()
	db.WebhooksFunc.SetDefaultReturn(webhookStore)
	db.UsersFunc.SetDefaultReturn(users)
	schema := mustParseGraphQLSchema(t, db)
	RunTests(t, []*Test{
		{
			Label:   "basic",
			Context: ctx,
			Schema:  schema,
			Query: `
				{
					webhooks {
						nodes { id }
						totalCount
						pageInfo { hasNextPage }
					}
				}
			`,
			ExpectedResult: `{"webhooks":
				{
					"nodes":[
						{"id":"V2ViaG9vazox"},
						{"id":"V2ViaG9vazoy"},
						{"id":"V2ViaG9vazoz"},
						{"id":"V2ViaG9vazo0"}
					],
					"totalCount":4,
					"pageInfo":{"hasNextPage":false}
				}}`,
		},
		{
			Label:   "specify first",
			Context: ctx,
			Schema:  schema,
			Query: `query Webhooks($first: Int!) {
						webhooks(first: $first) {
							nodes { id }
							totalCount
							pageInfo { hasNextPage endCursor }
						}
					}
			`,
			Variables: map[string]any{
				"first": 2,
			},
			ExpectedResult: `{"webhooks":
				{
					"nodes":[
						{"id":"V2ViaG9vazox"},
						{"id":"V2ViaG9vazoy"}
					],
					"totalCount":2,
					"pageInfo":{"hasNextPage":true, "endCursor": "V2ViaG9va0N1cnNvcjp7IkNvbHVtbiI6ImlkIiwiVmFsdWUiOiIzIiwiRGlyZWN0aW9uIjoibmV4dCJ9"}
				}}`,
		},
		{
			Label:   "specify kind",
			Context: ctx,
			Schema:  schema,
			Query: `query Webhooks($kind: ExternalServiceKind) {
						webhooks(kind: $kind) {
							nodes { id }
							totalCount
							pageInfo { hasNextPage }
						}
					}
			`,
			Variables: map[string]any{
				"kind": extsvc.KindGitHub,
			},
			ExpectedResult: `{"webhooks":
				{
					"nodes":[
						{"id":"V2ViaG9vazox"},
						{"id":"V2ViaG9vazoz"},
						{"id":"V2ViaG9vazo0"}
					],
					"totalCount":3,
					"pageInfo":{"hasNextPage":false}
				}}`,
		},
		{
			Label:   "specify cursor",
			Context: ctx,
			Schema:  schema,
			Query: `query Webhooks($first: Int!, $after: String!) {
						webhooks(first: $first, after: $after) {
							nodes { id }
							totalCount
							pageInfo { hasNextPage }
						}
					}
			`,
			Variables: map[string]any{
				"first": 2,
				"after": "V2ViaG9va0N1cnNvcjp7IkNvbHVtbiI6ImlkIiwiVmFsdWUiOiIzIiwiRGlyZWN0aW9uIjoibmV4dCJ9",
			},
			ExpectedResult: `{"webhooks":
				{
					"nodes":[
						{"id":"V2ViaG9vazoz"},
						{"id":"V2ViaG9vazo0"}
					],
					"totalCount":2,
					"pageInfo":{"hasNextPage":false}
				}}`,
		},
	})
}

func TestWebhooks_CursorPagination(t *testing.T) {
	gitlabURN, err := extsvc.NewCodeHostBaseURL("gitlab.com")
	require.NoError(t, err)
	githubURN, err := extsvc.NewCodeHostBaseURL("github.com")
	require.NoError(t, err)
	bbURN, err := extsvc.NewCodeHostBaseURL("bb.com")
	require.NoError(t, err)

	mockWebhooks := []*types.Webhook{
		{ID: 0, CodeHostURN: gitlabURN},
		{ID: 1, CodeHostURN: githubURN},
		{ID: 2, CodeHostURN: bbURN},
	}

	store := database.NewMockWebhookStore()
	db := database.NewMockDB()
	db.WebhooksFunc.SetDefaultReturn(store)

	users := database.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)
	db.UsersFunc.SetDefaultReturn(users)

	buildQuery := func(first int, after string) string {
		var args []string
		if first != 0 {
			args = append(args, fmt.Sprintf("first: %d", first))
		}
		if after != "" {
			args = append(args, fmt.Sprintf("after: %q", after))
		}

		return fmt.Sprintf(`{ webhooks(%s) { nodes { id } pageInfo { endCursor } } }`, strings.Join(args, ", "))
	}

	t.Run("Initial page without a cursor present", func(t *testing.T) {
		store.ListFunc.SetDefaultReturn(mockWebhooks[0:2], nil)
		RunTests(t, []*Test{
			{
				Schema: mustParseGraphQLSchema(t, db),
				Query:  buildQuery(1, ""),
				ExpectedResult: `
				{
					"webhooks": {
						"nodes": [{
							"id": "V2ViaG9vazow"
						}],
						"pageInfo": {
						  "endCursor": "V2ViaG9va0N1cnNvcjp7IkNvbHVtbiI6ImlkIiwiVmFsdWUiOiIxIiwiRGlyZWN0aW9uIjoibmV4dCJ9"
						}
					}
				}
			`,
			},
		})
	})

	t.Run("Second page", func(t *testing.T) {
		store.ListFunc.SetDefaultReturn(mockWebhooks[1:], nil)

		RunTest(t, &Test{
			Schema: mustParseGraphQLSchema(t, db),
			Query:  buildQuery(1, "V2ViaG9va0N1cnNvcjp7IkNvbHVtbiI6ImlkIiwiVmFsdWUiOiIxIiwiRGlyZWN0aW9uIjoibmV4dCJ9"),
			ExpectedResult: `
				{
					"webhooks": {
						"nodes": [{
							"id": "V2ViaG9vazox"
						}],
						"pageInfo": {
						  "endCursor": "V2ViaG9va0N1cnNvcjp7IkNvbHVtbiI6ImlkIiwiVmFsdWUiOiIyIiwiRGlyZWN0aW9uIjoibmV4dCJ9"
						}
					}
				}
			`,
		})
	})

	t.Run("Initial page with no further rows to fetch", func(t *testing.T) {
		store.ListFunc.SetDefaultReturn(mockWebhooks, nil)

		RunTest(t, &Test{
			Schema: mustParseGraphQLSchema(t, db),
			Query:  buildQuery(3, ""),
			ExpectedResult: `
				{
					"webhooks": {
						"nodes": [{
							"id": "V2ViaG9vazow"
						}, {
							"id": "V2ViaG9vazox"
						}, {
							"id": "V2ViaG9vazoy"
						}],
						"pageInfo": {
						  "endCursor": null
						}
					}
				}
			`,
		})
	})

	t.Run("With no webhooks present", func(t *testing.T) {
		store.ListFunc.SetDefaultReturn(nil, nil)

		RunTest(t, &Test{
			Schema: mustParseGraphQLSchema(t, db),
			Query:  buildQuery(1, ""),
			ExpectedResult: `
				{
					"webhooks": {
						"nodes": [],
						"pageInfo": {
						  "endCursor": null
						}
					}
				}
			`,
		})
	})

	t.Run("With an invalid cursor provided", func(t *testing.T) {
		store.ListFunc.SetDefaultReturn(nil, nil)

		RunTest(t, &Test{
			Schema:         mustParseGraphQLSchema(t, db),
			Query:          buildQuery(1, "invalid-cursor-value"),
			ExpectedResult: "null",
			ExpectedErrors: []*errors.QueryError{
				{
					Path:          []any{"webhooks"},
					Message:       `cannot unmarshal webhook cursor type: ""`,
					ResolverError: errors.Errorf(`cannot unmarshal webhook cursor type: ""`),
				},
			},
		})
	})
}

func TestCreateWebhook(t *testing.T) {
	users := database.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

	ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
	webhookStore := database.NewMockWebhookStore()
	whUUID, err := uuid.NewUUID()
	assert.NoError(t, err)
	expectedWebhook := types.Webhook{
		ID: 1, UUID: whUUID,
	}
	webhookStore.CreateFunc.SetDefaultReturn(&expectedWebhook, nil)

	db := database.NewMockDB()
	db.WebhooksFunc.SetDefaultReturn(webhookStore)
	db.UsersFunc.SetDefaultReturn(users)
	queryStr := `mutation CreateWebhook($codeHostKind: String!, $codeHostURN: String!, $secret: String) {
				createWebhook(codeHostKind: $codeHostKind, codeHostURN: $codeHostURN, secret: $secret) {
					id
					uuid
				}
			}`
	schema := mustParseGraphQLSchema(t, db)

	RunTests(t, []*Test{
		{
			Label:   "basic",
			Context: ctx,
			Schema:  schema,
			Query:   queryStr,
			ExpectedResult: fmt.Sprintf(`
				{
					"createWebhook": {
						"id": "V2ViaG9vazox",
						"uuid": "%s"
					}
				}
			`, whUUID),
			Variables: map[string]any{
				"codeHostKind": "GITHUB",
				"codeHostURN":  "https://github.com",
			},
		},
		{
			Label:          "invalid code host",
			Context:        ctx,
			Schema:         schema,
			Query:          queryStr,
			ExpectedResult: "null",
			ExpectedErrors: []*errors.QueryError{
				{
					Message: "webhooks are not supported for code host kind InvalidKind",
					Path:    []any{"createWebhook"},
				},
			},
			Variables: map[string]any{
				"codeHostKind": "InvalidKind",
				"codeHostURN":  "https://github.com",
			},
		},
		{
			Label:          "secrets not supported for code host",
			Context:        ctx,
			Schema:         schema,
			Query:          queryStr,
			ExpectedResult: "null",
			ExpectedErrors: []*errors.QueryError{
				{
					Message: "webhooks do not support secrets for code host kind BITBUCKETCLOUD",
					Path:    []any{"createWebhook"},
				},
			},
			Variables: map[string]any{
				"codeHostKind": "BITBUCKETCLOUD",
				"codeHostURN":  "https://bitbucket.com",
				"secret":       "mysupersecret",
			},
		},
	})

	// validate error if not site admin
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: false}, nil)
	RunTest(t, &Test{
		Label:          "only site admin can create webhook",
		Context:        ctx,
		Schema:         schema,
		Query:          queryStr,
		ExpectedResult: "null",
		ExpectedErrors: []*errors.QueryError{
			{
				Message: "must be site admin",
				Path:    []any{"createWebhook"},
			},
		},
		Variables: map[string]any{
			"codeHostKind": "GITHUB",
			"codeHostURN":  "https://github.com",
			"secret":       "mysupersecret",
		},
	})
}

func TestGetWebhookWithURL(t *testing.T) {
	users := database.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

	testURL := "https://testurl.com"
	invalidURL := "https://invalid.com/%+o"
	webhookID := int32(1)
	webhookIDMarshaled := marshalWebhookID(webhookID)
	conf.Mock(
		&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				ExternalURL: testURL,
			},
		},
	)

	ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
	webhookStore := database.NewMockWebhookStore()
	whUUID, err := uuid.NewUUID()
	assert.NoError(t, err)
	expectedWebhook := types.Webhook{
		ID: webhookID, UUID: whUUID,
	}
	webhookStore.GetByIDFunc.SetDefaultReturn(&expectedWebhook, nil)

	db := database.NewMockDB()
	db.WebhooksFunc.SetDefaultReturn(webhookStore)
	db.UsersFunc.SetDefaultReturn(users)
	queryStr := `query GetWebhook($id: ID!) {
                node(id: $id) {
                    ... on Webhook {
                        id
                        uuid
                        url
                    }
                }
			}`
	gqlSchema := mustParseGraphQLSchema(t, db)

	RunTest(t, &Test{
		Label:   "basic",
		Context: ctx,
		Schema:  gqlSchema,
		Query:   queryStr,
		ExpectedResult: fmt.Sprintf(`
				{
					"node": {
						"id": %q,
						"uuid": %q,
                        "url": "%s/.api/webhooks/%s"
					}
				}
			`, webhookIDMarshaled, whUUID.String(), testURL, whUUID.String()),
		Variables: map[string]any{
			"id": "V2ViaG9vazox",
		},
	})

	conf.Mock(
		&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				ExternalURL: invalidURL,
			},
		},
	)
	RunTest(t, &Test{
		Label:          "error if external URL invalid",
		Context:        ctx,
		Schema:         gqlSchema,
		Query:          queryStr,
		ExpectedResult: `{"node": null}`,
		ExpectedErrors: []*errors.QueryError{
			{
				Message: strings.Join([]string{
					"could not parse site config external URL:",
					` parse "https://invalid.com/%+o": invalid URL escape "%+o"`,
				}, ""),
				Path: []any{"node", "url"},
			},
		},
		Variables: map[string]any{
			"id": "V2ViaG9vazox",
		},
	})
}

func TestWebhookCursor(t *testing.T) {
	var (
		webhookCursor       = types.Cursor{Column: "id", Value: "4", Direction: "next"}
		opaqueWebhookCursor = "V2ViaG9va0N1cnNvcjp7IkNvbHVtbiI6ImlkIiwiVmFsdWUiOiI0IiwiRGlyZWN0aW9uIjoibmV4dCJ9"
	)
	t.Run("Marshal", func(t *testing.T) {
		if got, want := MarshalWebhookCursor(&webhookCursor), opaqueWebhookCursor; got != want {
			t.Errorf("got opaque cursor %q, want %q", got, want)
		}
	})
	t.Run("Unmarshal", func(t *testing.T) {
		cursor, err := UnmarshalWebhookCursor(&opaqueWebhookCursor)
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(cursor, &webhookCursor); diff != "" {
			t.Fatal(diff)
		}
	})
}

func TestDeleteWebhook(t *testing.T) {
	users := database.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: false}, nil)

	ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
	webhookStore := database.NewMockWebhookStore()
	webhookStore.DeleteFunc.SetDefaultReturn(sgerrors.New("oops"))

	db := database.NewMockDB()
	db.WebhooksFunc.SetDefaultReturn(webhookStore)
	db.UsersFunc.SetDefaultReturn(users)
	id := marshalWebhookID(42)
	queryStr := `mutation DeleteWebhook($id: ID!) {
				deleteWebhook(id: $id) {
					alwaysNil
				}
			}`
	schema := mustParseGraphQLSchema(t, db)

	// validate error if not site admin
	RunTest(t, &Test{
		Label:          "only site admin can delete webhook",
		Context:        ctx,
		Schema:         schema,
		Query:          queryStr,
		ExpectedResult: "null",
		ExpectedErrors: []*errors.QueryError{
			{
				Message: "must be site admin",
				Path:    []any{"deleteWebhook"},
			},
		},
		Variables: map[string]any{
			"id": string(id),
		},
	})

	// User is site admin
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

	RunTest(t, &Test{
		Label:          "database error",
		Context:        ctx,
		Schema:         schema,
		Query:          queryStr,
		ExpectedResult: "null",
		ExpectedErrors: []*errors.QueryError{
			{
				Message: "delete webhook: oops",
				Path:    []any{"deleteWebhook"},
			},
		},
		Variables: map[string]any{
			"id": string(id),
		},
	})

	// database layer behaves
	webhookStore.DeleteFunc.SetDefaultReturn(nil)

	RunTest(t, &Test{
		Label:   "webhook successfully deleted",
		Context: ctx,
		Schema:  schema,
		Query:   queryStr,
		ExpectedResult: `
				{
					"deleteWebhook": {
						"alwaysNil": null
					}
				}
			`,
		Variables: map[string]any{
			"id": string(id),
		},
	})
}
