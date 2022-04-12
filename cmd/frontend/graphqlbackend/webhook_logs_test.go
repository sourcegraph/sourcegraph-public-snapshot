package graphqlbackend

import (
	"context"
	"testing"
	"time"

	mockassert "github.com/derision-test/go-mockgen/testutil/assert"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestWebhookLogsArgs(t *testing.T) {
	// Create two times for easier reuse in test cases.
	var (
		now   = time.Date(2021, 11, 1, 18, 25, 10, 0, time.UTC)
		later = now.Add(1 * time.Hour)
	)

	t.Run("success", func(t *testing.T) {
		for name, tc := range map[string]struct {
			id    webhookLogsExternalServiceID
			input webhookLogsArgs
			want  database.WebhookLogListOpts
		}{
			"no arguments": {
				id:    webhookLogsAllExternalServices,
				input: webhookLogsArgs{},
				want: database.WebhookLogListOpts{
					Limit: 50,
				},
			},
			"OnlyErrors false": {
				id: webhookLogsUnmatchedExternalService,
				input: webhookLogsArgs{
					OnlyErrors: boolPtr(false),
				},
				want: database.WebhookLogListOpts{
					Limit:             50,
					ExternalServiceID: int64Ptr(0),
					OnlyErrors:        false,
				},
			},
			"all arguments": {
				id: webhookLogsExternalServiceID(1),
				input: webhookLogsArgs{
					ConnectionArgs: graphqlutil.ConnectionArgs{
						First: int32Ptr(25),
					},
					After:      stringPtr("40"),
					OnlyErrors: boolPtr(true),
					Since:      timePtr(now),
					Until:      timePtr(later),
				},
				want: database.WebhookLogListOpts{
					Limit:             25,
					Cursor:            40,
					ExternalServiceID: int64Ptr(1),
					OnlyErrors:        true,
					Since:             timePtr(now),
					Until:             timePtr(later),
				},
			},
		} {
			t.Run(name, func(t *testing.T) {
				have, err := tc.input.toListOpts(tc.id)
				assert.Nil(t, err)
				assert.Equal(t, tc.want, have)
			})
		}
	})

	t.Run("errors", func(t *testing.T) {
		for _, input := range []string{
			"",
			"-",
			"0.0",
			"foo",
		} {
			t.Run(input, func(t *testing.T) {
				_, err := (&webhookLogsArgs{After: &input}).toListOpts(webhookLogsUnmatchedExternalService)
				assert.NotNil(t, err)
			})
		}
	})
}

func TestNewWebhookLogConnectionResolver(t *testing.T) {
	// We'll test everything else below, but let's just make sure the admin
	// check occurs.
	t.Run("unauthenticated user", func(t *testing.T) {
		users := database.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(nil, nil)

		db := database.NewMockDB()
		db.UsersFunc.SetDefaultReturn(users)

		_, err := newWebhookLogConnectionResolver(context.Background(), db, nil, webhookLogsUnmatchedExternalService)
		assert.ErrorIs(t, err, backend.ErrNotAuthenticated)
	})

	t.Run("regular user", func(t *testing.T) {
		users := database.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{}, nil)

		db := database.NewMockDB()
		db.UsersFunc.SetDefaultReturn(users)

		_, err := newWebhookLogConnectionResolver(context.Background(), db, nil, webhookLogsUnmatchedExternalService)
		assert.ErrorIs(t, err, backend.ErrMustBeSiteAdmin)
	})

	t.Run("admin user", func(t *testing.T) {
		users := database.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

		db := database.NewMockDB()
		db.UsersFunc.SetDefaultReturn(users)

		_, err := newWebhookLogConnectionResolver(context.Background(), db, nil, webhookLogsUnmatchedExternalService)
		assert.Nil(t, err)
	})
}

func TestWebhookLogConnectionResolver(t *testing.T) {
	ctx := context.Background()

	// We'll set up a fake page of 20 logs.
	logs := []*types.WebhookLog{}
	for i := 0; i < 20; i++ {
		logs = append(logs, &types.WebhookLog{})
	}

	// We also need a fake TransactableHandle to be able to construct
	// webhookLogResolvers.
	db := &basestore.TransactableHandle{}

	createMockStore := func(logs []*types.WebhookLog, next int64, err error) *database.MockWebhookLogStore {
		store := database.NewMockWebhookLogStore()
		store.ListFunc.SetDefaultReturn(logs, next, err)
		store.HandleFunc.SetDefaultReturn(db)

		return store
	}

	t.Run("empty and has no further pages", func(t *testing.T) {
		store := createMockStore([]*types.WebhookLog{}, 0, nil)

		r := &webhookLogConnectionResolver{
			args: &webhookLogsArgs{
				ConnectionArgs: graphqlutil.ConnectionArgs{
					First: int32Ptr(20),
				},
			},
			externalServiceID: webhookLogsExternalServiceID(1),
			store:             store,
		}

		nodes, err := r.Nodes(ctx)
		assert.Len(t, nodes, 0)
		assert.Nil(t, err)

		page, err := r.PageInfo(context.Background())
		assert.False(t, page.HasNextPage())
		assert.Nil(t, err)

		mockassert.CalledOnceWith(
			t, store.ListFunc,
			mockassert.Values(
				mockassert.Skip,
				database.WebhookLogListOpts{
					ExternalServiceID: int64Ptr(1),
					Limit:             20,
				},
			),
		)
	})

	t.Run("full and has next page", func(t *testing.T) {
		store := createMockStore(logs, 20, nil)

		r := &webhookLogConnectionResolver{
			args: &webhookLogsArgs{
				ConnectionArgs: graphqlutil.ConnectionArgs{
					First: int32Ptr(20),
				},
			},
			externalServiceID: webhookLogsExternalServiceID(1),
			store:             store,
		}

		nodes, err := r.Nodes(ctx)
		for i, node := range nodes {
			assert.Equal(t, logs[i], node.log)
		}
		assert.Nil(t, err)

		page, err := r.PageInfo(context.Background())
		assert.True(t, page.HasNextPage())
		assert.Nil(t, err)

		mockassert.CalledOnceWith(
			t, store.ListFunc,
			mockassert.Values(
				mockassert.Skip,
				database.WebhookLogListOpts{
					ExternalServiceID: int64Ptr(1),
					Limit:             20,
				},
			),
		)
	})

	t.Run("errors", func(t *testing.T) {
		want := errors.New("error")
		store := createMockStore(nil, 0, want)

		r := &webhookLogConnectionResolver{
			args: &webhookLogsArgs{
				ConnectionArgs: graphqlutil.ConnectionArgs{
					First: int32Ptr(20),
				},
			},
			externalServiceID: webhookLogsExternalServiceID(1),
			store:             store,
		}

		_, err := r.PageInfo(context.Background())
		assert.ErrorIs(t, err, want)
	})
}

func TestWebhookLogConnectionResolver_TotalCount(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		store := database.NewMockWebhookLogStore()
		store.CountFunc.SetDefaultReturn(40, nil)

		r := &webhookLogConnectionResolver{
			args: &webhookLogsArgs{
				OnlyErrors: boolPtr(true),
			},
			externalServiceID: webhookLogsExternalServiceID(1),
			store:             store,
		}

		have, err := r.TotalCount(context.Background())
		assert.EqualValues(t, 40, have)
		assert.Nil(t, err)

		mockassert.CalledOnceWith(
			t, store.CountFunc,
			mockassert.Values(
				mockassert.Skip,
				database.WebhookLogListOpts{
					ExternalServiceID: int64Ptr(1),
					Limit:             50,
					OnlyErrors:        true,
				},
			),
		)
	})

	t.Run("errors", func(t *testing.T) {
		want := errors.New("error")
		store := database.NewMockWebhookLogStore()
		store.CountFunc.SetDefaultReturn(0, want)

		r := &webhookLogConnectionResolver{
			args: &webhookLogsArgs{
				OnlyErrors: boolPtr(true),
			},
			externalServiceID: webhookLogsExternalServiceID(1),
			store:             store,
		}

		_, err := r.TotalCount(context.Background())
		assert.ErrorIs(t, err, want)
	})
}

func boolPtr(v bool) *bool           { return &v }
func int32Ptr(v int32) *int32        { return &v }
func int64Ptr(v int64) *int64        { return &v }
func stringPtr(v string) *string     { return &v }
func timePtr(v time.Time) *time.Time { return &v }
