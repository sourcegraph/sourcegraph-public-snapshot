package resolvers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"

	"github.com/google/uuid"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	sgerrors "github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestListWebhooks(t *testing.T) {
	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)
	// There is only a user with ID = 1. User with ID = 2 doesn't exist.
	users.GetByIDFunc.SetDefaultHook(func(_ context.Context, id int32) (*types.User, error) {
		if id == 1 {
			return &types.User{Username: "alice"}, nil
		}
		return nil, database.NewUserNotFoundError(id)
	})

	ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
	webhookStore := dbmocks.NewMockWebhookStore()
	webhooks := []*types.Webhook{
		{
			ID:              1,
			CodeHostKind:    extsvc.KindGitHub,
			CreatedByUserID: 1,
		},
		{
			ID:              2,
			CodeHostKind:    extsvc.KindGitLab,
			CreatedByUserID: 1,
		},
		{
			ID:              3,
			CodeHostKind:    extsvc.KindGitHub,
			CreatedByUserID: 1,
			UpdatedByUserID: 1,
		},
		{
			ID:              4,
			CodeHostKind:    extsvc.KindGitHub,
			CreatedByUserID: 1,
			UpdatedByUserID: 2,
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

	webhookLogsStore := dbmocks.NewMockWebhookLogStore()
	webhookLogs := []*types.WebhookLog{
		{
			ID:         1,
			WebhookID:  &webhooks[0].ID,
			StatusCode: 404,
		},
		{
			ID:         2,
			WebhookID:  &webhooks[0].ID,
			StatusCode: 200,
		},
		{
			ID:         3,
			WebhookID:  &webhooks[1].ID,
			StatusCode: 404,
		},
	}
	webhookLogsStore.ListFunc.SetDefaultHook(func(ctx2 context.Context, opts database.WebhookLogListOpts) ([]*types.WebhookLog, int64, error) {
		if opts.WebhookID == nil {
			return webhookLogs, 0, nil
		}
		filteredWebhooks := []*types.WebhookLog{}
		for _, webhookLog := range webhookLogs {
			if *webhookLog.WebhookID == *opts.WebhookID {
				if opts.OnlyErrors {
					if webhookLog.StatusCode >= 400 {
						filteredWebhooks = append(filteredWebhooks, webhookLog)
					}
				} else {
					filteredWebhooks = append(filteredWebhooks, webhookLog)
				}
			}
		}
		return filteredWebhooks, 0, nil
	})
	webhookLogsStore.CountFunc.SetDefaultHook(func(ctx2 context.Context, opts database.WebhookLogListOpts) (int64, error) {
		whs, _, err := webhookLogsStore.List(ctx2, opts)
		return int64(len(whs)), err
	})

	db := dbmocks.NewMockDB()
	db.WebhooksFunc.SetDefaultReturn(webhookStore)
	db.UsersFunc.SetDefaultReturn(users)
	db.WebhookLogsFunc.SetDefaultReturn(webhookLogsStore)
	gqlSchema := createGqlSchema(t, db)
	graphqlbackend.RunTests(t, []*graphqlbackend.Test{
		{
			Label:   "basic",
			Context: ctx,
			Schema:  gqlSchema,
			Query: `
				{
					webhooks {
						nodes {
							id
							updatedBy {
								username
							}
							createdBy {
								username
							}
						}
						totalCount
						pageInfo { hasNextPage }
					}
				}
			`,
			ExpectedResult: `{"webhooks":
				{
					"nodes":[
						{
							"id":"V2ViaG9vazox",
							"createdBy": {
								"username": "alice"
							},
							"updatedBy": null
						},
						{
							"id":"V2ViaG9vazoy",
							"createdBy": {
								"username": "alice"
							},
							"updatedBy": null
						},
						{
							"id":"V2ViaG9vazoz",
							"createdBy": {
								"username": "alice"
							},
							"updatedBy": {
								"username": "alice"
							}
						},
						{
							"id":"V2ViaG9vazo0",
							"createdBy": {
								"username": "alice"
							},
							"updatedBy": null
						}
					],
					"totalCount":4,
					"pageInfo":{"hasNextPage":false}
				}}`,
		},
		{
			Label:   "specify first",
			Context: ctx,
			Schema:  gqlSchema,
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
			Schema:  gqlSchema,
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
			Schema:  gqlSchema,
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
		{
			Label:   "with logs",
			Context: ctx,
			Schema:  gqlSchema,
			Query: `
				{
					webhooks {
						nodes {
							id
							webhookLogs {
								totalCount
 							}
						}
						totalCount
						pageInfo { hasNextPage }
					}
				}
			`,
			ExpectedResult: `{"webhooks":
				{
					"nodes":[
						{ "id":"V2ViaG9vazox", "webhookLogs": { "totalCount": 2 } },
						{ "id":"V2ViaG9vazoy", "webhookLogs": { "totalCount": 1 } },
						{ "id":"V2ViaG9vazoz", "webhookLogs": { "totalCount": 0 } },
						{ "id":"V2ViaG9vazo0", "webhookLogs": { "totalCount": 0 } }
					],
					"totalCount":4,
					"pageInfo":{"hasNextPage":false}
				}}`,
		},
		{
			Label:   "with logs only errors",
			Context: ctx,
			Schema:  gqlSchema,
			Query: `
				{
					webhooks {
						nodes {
							id
							webhookLogs(onlyErrors: true) {
								totalCount
							}
						}
						totalCount
						pageInfo { hasNextPage }
					}
				}
			`,
			ExpectedResult: `{"webhooks":
				{
					"nodes":[
						{ "id":"V2ViaG9vazox", "webhookLogs": { "totalCount": 1 } },
						{ "id":"V2ViaG9vazoy", "webhookLogs": { "totalCount": 1 } },
						{ "id":"V2ViaG9vazoz", "webhookLogs": { "totalCount": 0 } },
						{ "id":"V2ViaG9vazo0", "webhookLogs": { "totalCount": 0 } }
					],
					"totalCount":4,
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

	store := dbmocks.NewMockWebhookStore()
	db := dbmocks.NewMockDB()
	db.WebhooksFunc.SetDefaultReturn(store)

	users := dbmocks.NewMockUserStore()
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
		graphqlbackend.RunTests(t, []*graphqlbackend.Test{
			{
				Schema: createGqlSchema(t, db),
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

		graphqlbackend.RunTest(t, &graphqlbackend.Test{
			Schema: createGqlSchema(t, db),
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

		graphqlbackend.RunTest(t, &graphqlbackend.Test{
			Schema: createGqlSchema(t, db),
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

		graphqlbackend.RunTest(t, &graphqlbackend.Test{
			Schema: createGqlSchema(t, db),
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

		graphqlbackend.RunTest(t, &graphqlbackend.Test{
			Schema:         createGqlSchema(t, db),
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
	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

	ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
	webhookStore := dbmocks.NewMockWebhookStore()
	whUUID, err := uuid.NewUUID()
	assert.NoError(t, err)
	expectedWebhook := types.Webhook{
		ID: 1, UUID: whUUID, Name: "webhookName",
	}
	webhookStore.CreateFunc.SetDefaultReturn(&expectedWebhook, nil)

	db := dbmocks.NewMockDB()
	db.WebhooksFunc.SetDefaultReturn(webhookStore)
	db.UsersFunc.SetDefaultReturn(users)
	queryStr := `mutation CreateWebhook($name: String!, $codeHostKind: String!, $codeHostURN: String!, $secret: String) {
				createWebhook(name: $name, codeHostKind: $codeHostKind, codeHostURN: $codeHostURN, secret: $secret) {
					name
					id
					uuid
				}
			}`
	gqlSchema := createGqlSchema(t, db)

	graphqlbackend.RunTests(t, []*graphqlbackend.Test{
		{
			Label:   "basic",
			Context: ctx,
			Schema:  gqlSchema,
			Query:   queryStr,
			ExpectedResult: fmt.Sprintf(`
				{
					"createWebhook": {
						"name": "webhookName",
						"id": "V2ViaG9vazox",
						"uuid": "%s"
					}
				}
			`, whUUID),
			Variables: map[string]any{
				"name":         "webhookName",
				"codeHostKind": "GITHUB",
				"codeHostURN":  "https://github.com",
			},
		},
		{
			Label:          "invalid code host",
			Context:        ctx,
			Schema:         gqlSchema,
			Query:          queryStr,
			ExpectedResult: "null",
			ExpectedErrors: []*errors.QueryError{
				{
					Message: "webhooks are not supported for code host kind InvalidKind",
					Path:    []any{"createWebhook"},
				},
			},
			Variables: map[string]any{
				"name":         "webhookName",
				"codeHostKind": "InvalidKind",
				"codeHostURN":  "https://github.com",
			},
		},
		{
			Label:          "secrets not supported for code host",
			Context:        ctx,
			Schema:         gqlSchema,
			Query:          queryStr,
			ExpectedResult: "null",
			ExpectedErrors: []*errors.QueryError{
				{
					Message: "webhooks do not support secrets for code host kind BITBUCKETCLOUD",
					Path:    []any{"createWebhook"},
				},
			},
			Variables: map[string]any{
				"name":         "webhookName",
				"codeHostKind": "BITBUCKETCLOUD",
				"codeHostURN":  "https://bitbucket.com",
				"secret":       "mysupersecret",
			},
		},
	})

	// validate error if not site admin
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: false}, nil)
	graphqlbackend.RunTest(t, &graphqlbackend.Test{
		Label:          "only site admin can create webhook",
		Context:        ctx,
		Schema:         gqlSchema,
		Query:          queryStr,
		ExpectedResult: "null",
		ExpectedErrors: []*errors.QueryError{
			{
				Message: "must be site admin",
				Path:    []any{"createWebhook"},
			},
		},
		Variables: map[string]any{
			"name":         "webhookName",
			"codeHostKind": "GITHUB",
			"codeHostURN":  "https://github.com",
			"secret":       "mysupersecret",
		},
	})
}

func TestGetWebhookWithURL(t *testing.T) {
	users := dbmocks.NewMockUserStore()
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
	webhookStore := dbmocks.NewMockWebhookStore()
	whUUID, err := uuid.NewUUID()
	assert.NoError(t, err)
	expectedWebhook := types.Webhook{
		ID: webhookID, UUID: whUUID,
	}
	webhookStore.GetByIDFunc.SetDefaultReturn(&expectedWebhook, nil)

	db := dbmocks.NewMockDB()
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
	gqlSchema := createGqlSchema(t, db)

	graphqlbackend.RunTest(t, &graphqlbackend.Test{
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

	graphqlbackend.RunTest(t, &graphqlbackend.Test{
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
	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: false}, nil)

	ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
	webhookStore := dbmocks.NewMockWebhookStore()
	webhookStore.DeleteFunc.SetDefaultReturn(sgerrors.New("oops"))

	db := dbmocks.NewMockDB()
	db.WebhooksFunc.SetDefaultReturn(webhookStore)
	db.UsersFunc.SetDefaultReturn(users)
	id := marshalWebhookID(42)
	queryStr := `mutation DeleteWebhook($id: ID!) {
				deleteWebhook(id: $id) {
					alwaysNil
				}
			}`
	gqlSchema := createGqlSchema(t, db)

	// validate error if not site admin
	graphqlbackend.RunTest(t, &graphqlbackend.Test{
		Label:          "only site admin can delete webhook",
		Context:        ctx,
		Schema:         gqlSchema,
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

	graphqlbackend.RunTest(t, &graphqlbackend.Test{
		Label:          "database error",
		Context:        ctx,
		Schema:         gqlSchema,
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

	graphqlbackend.RunTest(t, &graphqlbackend.Test{
		Label:   "webhook successfully deleted",
		Context: ctx,
		Schema:  gqlSchema,
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

func TestUpdateWebhook(t *testing.T) {
	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: false}, nil)

	ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
	webhookStore := dbmocks.NewMockWebhookStore()
	webhookStore.UpdateFunc.SetDefaultHook(func(ctx context.Context, webhook *types.Webhook) (*types.Webhook, error) {
		return nil, sgerrors.New("oops")
	})
	whUUID := uuid.New()
	ghURN, err := extsvc.NewCodeHostBaseURL("https://github.com")
	require.NoError(t, err)
	webhookStore.GetByIDFunc.SetDefaultReturn(&types.Webhook{Name: "old name", ID: 1, UUID: whUUID, CodeHostURN: ghURN}, nil)

	db := dbmocks.NewMockDB()
	db.WebhooksFunc.SetDefaultReturn(webhookStore)
	db.UsersFunc.SetDefaultReturn(users)
	id := marshalWebhookID(42)
	mutateStr := `mutation UpdateWebhook($id: ID!, $name: String, $codeHostKind: String, $codeHostURN: String, $secret: String) {
                updateWebhook(id: $id, name: $name, codeHostKind: $codeHostKind, codeHostURN: $codeHostURN, secret: $secret) {
                    name
                    id
                    uuid
                    codeHostURN
				}
			}`
	gqlSchema := createGqlSchema(t, db)

	// validate error if not site admin
	graphqlbackend.RunTest(t, &graphqlbackend.Test{
		Label:          "only site admin can update webhook",
		Context:        ctx,
		Schema:         gqlSchema,
		Query:          mutateStr,
		ExpectedResult: "null",
		ExpectedErrors: []*errors.QueryError{
			{
				Message: "must be site admin",
				Path:    []any{"updateWebhook"},
			},
		},
		Variables: map[string]any{
			"id":           string(id),
			"name":         "new name",
			"codeHostKind": nil,
			"codeHostURN":  nil,
			"secret":       nil,
		},
	})

	// User is site admin
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

	graphqlbackend.RunTest(t, &graphqlbackend.Test{
		Label:          "database error",
		Context:        ctx,
		Schema:         gqlSchema,
		Query:          mutateStr,
		ExpectedResult: "null",
		ExpectedErrors: []*errors.QueryError{
			{
				Message: "update webhook: oops",
				Path:    []any{"updateWebhook"},
			},
		},
		Variables: map[string]any{
			"id":           string(id),
			"name":         "new name",
			"codeHostKind": nil,
			"codeHostURN":  nil,
			"secret":       nil,
		},
	})

	// database layer behaves
	webhookStore.UpdateFunc.SetDefaultHook(func(ctx context.Context, webhook *types.Webhook) (*types.Webhook, error) {
		return webhook, nil
	})

	graphqlbackend.RunTest(t, &graphqlbackend.Test{
		Label:   "webhook successfully updated 1 field",
		Context: ctx,
		Schema:  gqlSchema,
		Query:   mutateStr,
		ExpectedResult: fmt.Sprintf(`
				{
					"updateWebhook": {
						"name": "new name",
						"id": "V2ViaG9vazox",
						"uuid": "%s",
                        "codeHostURN": "https://github.com/"
					}
				}
			`, whUUID),
		Variables: map[string]any{
			"id":           string(id),
			"name":         "new name",
			"codeHostKind": nil,
			"codeHostURN":  nil,
			"secret":       nil,
		},
	})

	graphqlbackend.RunTest(t, &graphqlbackend.Test{
		Label:   "webhook successfully updated multiple fields",
		Context: ctx,
		Schema:  gqlSchema,
		Query:   mutateStr,
		ExpectedResult: fmt.Sprintf(`
				{
					"updateWebhook": {
						"name": "new name",
						"id": "V2ViaG9vazox",
						"uuid": "%s",
                        "codeHostURN": "https://example.github.com/"
					}
				}
			`, whUUID),
		Variables: map[string]any{
			"id":           string(id),
			"name":         "new name",
			"codeHostKind": nil,
			"codeHostURN":  "https://example.github.com",
			"secret":       nil,
		},
	})

	graphqlbackend.RunTest(t, &graphqlbackend.Test{
		Label:   "BitBucket Cloud webhook successfully updated without a secret",
		Context: ctx,
		Schema:  gqlSchema,
		Query:   mutateStr,
		ExpectedResult: fmt.Sprintf(`
				{
					"updateWebhook": {
						"name": "new name",
						"id": "V2ViaG9vazox",
						"uuid": "%s",
                        "codeHostURN": "https://sg.bitbucket.org/"
					}
				}
			`, whUUID),
		Variables: map[string]any{
			"id":           string(id),
			"name":         "new name",
			"codeHostKind": extsvc.KindBitbucketCloud,
			"codeHostURN":  "https://sg.bitbucket.org",
			"secret":       nil,
		},
	})

	webhookStore.GetByIDFunc.SetDefaultReturn(nil, &database.WebhookNotFoundError{ID: 2})

	graphqlbackend.RunTest(t, &graphqlbackend.Test{
		Label:          "error for non-existent webhook",
		Context:        ctx,
		Schema:         gqlSchema,
		Query:          mutateStr,
		ExpectedResult: "null",
		ExpectedErrors: []*errors.QueryError{
			{
				Message: "update webhook: webhook with ID 2 not found",
				Path:    []any{"updateWebhook"},
			},
		},
		Variables: map[string]any{
			"id":           string(id),
			"name":         "new name",
			"codeHostKind": nil,
			"codeHostURN":  "https://example.github.com",
			"secret":       nil,
		},
	})
}

func createGqlSchema(t *testing.T, db database.DB) *graphql.Schema {
	t.Helper()
	gqlSchema, err := graphqlbackend.NewSchemaWithWebhooksResolver(db, NewWebhooksResolver(db))
	if err != nil {
		t.Fatal(err)
	}
	return gqlSchema
}
