package resolvers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/batches/resolvers/apitest"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	bt "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

func TestExternalServicesWithoutWebhooksResolver(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	db := dbtest.NewDB(t, "")

	now := timeutil.Now()
	clock := func() time.Time { return now }
	bstore := store.NewWithClock(db, &observation.TestContext, nil, clock)

	fixture := bt.NewExternalServiceWebhookFixture(t, actor.WithInternalActor(ctx), bstore)
	adminCtx := actor.WithActor(ctx, actor.FromUser(fixture.AdminUser.ID))
	userCtx := actor.WithActor(ctx, actor.FromUser(fixture.RegularUser.ID))

	t.Run("empty", func(t *testing.T) {
		bc := &batchChangeResolver{
			store:       bstore,
			batchChange: fixture.EmptyBatchChange,
		}

		conn, err := bc.ExternalServicesWithoutWebhooks(adminCtx, &graphqlbackend.ListExternalServicesArgs{})
		assert.Nil(t, err)

		assertExternalServiceConnection(
			t, adminCtx,
			apitest.ExternalServiceConnection{
				TotalCount: 0,
			},
			conn,
		)
	})

	t.Run("other", func(t *testing.T) {
		bc := &batchChangeResolver{
			store:       bstore,
			batchChange: fixture.OtherBatchChange,
		}

		conn, err := bc.ExternalServicesWithoutWebhooks(
			adminCtx,
			&graphqlbackend.ListExternalServicesArgs{
				First: 50,
			},
		)
		assert.Nil(t, err)

		assertExternalServiceConnection(
			t, adminCtx,
			apitest.ExternalServiceConnection{
				Nodes: []apitest.ExternalService{
					{
						Kind:        fixture.OtherSvc.Kind,
						DisplayName: fixture.OtherSvc.DisplayName,
					},
				},
				TotalCount: 1,
			},
			conn,
		)
	})

	t.Run("primary", func(t *testing.T) {
		bc := &batchChangeResolver{
			store:       bstore,
			batchChange: fixture.PrimaryBatchChange,
		}

		t.Run("as user", func(t *testing.T) {
			// This should never succeed, since regular users cannot access
			// external services under any circumstances.
			conn, err := bc.ExternalServicesWithoutWebhooks(
				userCtx,
				&graphqlbackend.ListExternalServicesArgs{
					First: 2,
				},
			)
			assert.NotNil(t, err)
			assert.True(t, errors.Is(err, backend.ErrMustBeSiteAdmin))
			assert.Nil(t, conn)
		})

		t.Run("as admin", func(t *testing.T) {
			// We'll set up the pagination to return only two services at a
			// time. As the underlying store methods enforce ASC ordering on the
			// external service ID, this means that the first page will be
			// GitHub and GitLab, then the second page will be Bitbucket Server.
			// (This is specifically due to the ordering of when the external
			// services are created in the fixture.)
			conn, err := bc.ExternalServicesWithoutWebhooks(
				adminCtx,
				&graphqlbackend.ListExternalServicesArgs{
					First: 2,
				},
			)
			assert.Nil(t, err)

			// The cursor is the ID of the first element of the _next_ page,
			// which is the Bitbucket Server external service.
			wantCursor := fmt.Sprint(fixture.BitbucketServerSvc.ID)
			assertExternalServiceConnection(
				t, adminCtx,
				apitest.ExternalServiceConnection{
					Nodes: []apitest.ExternalService{
						{
							Kind:        fixture.GitHubSvc.Kind,
							DisplayName: fixture.GitHubSvc.DisplayName,
						},
						{
							Kind:        fixture.GitLabSvc.Kind,
							DisplayName: fixture.GitLabSvc.DisplayName,
						},
					},
					PageInfo: apitest.PageInfo{
						EndCursor:   &wantCursor,
						HasNextPage: true,
					},
					TotalCount: 3,
				},
				conn,
			)

			// Grab the second page.
			conn, err = bc.ExternalServicesWithoutWebhooks(
				adminCtx,
				&graphqlbackend.ListExternalServicesArgs{
					First: 2,
					After: &wantCursor,
				},
			)
			assert.Nil(t, err)

			assertExternalServiceConnection(
				t, adminCtx,
				apitest.ExternalServiceConnection{
					Nodes: []apitest.ExternalService{
						{
							Kind:        fixture.BitbucketServerSvc.Kind,
							DisplayName: fixture.BitbucketServerSvc.DisplayName,
						},
					},
					PageInfo: apitest.PageInfo{
						EndCursor:   nil,
						HasNextPage: false,
					},
					TotalCount: 3,
				},
				conn,
			)
		})
	})
}

func assertExternalServiceConnection(
	t *testing.T,
	ctx context.Context,
	want apitest.ExternalServiceConnection,
	conn graphqlbackend.ExternalServiceConnectionResolver,
) {
	t.Helper()

	// Fully testing every field in each node is unnecessarily duplicative,
	// since the actual node resolver is graphqlbackend.ExternalServiceResolver
	// and it's tested there. Instead, we'll only look at the external service
	// kind and display name.
	nodes, err := conn.Nodes(ctx)
	assert.Nil(t, err, "Nodes error")
	have := make([]apitest.ExternalService, 0, len(nodes))
	for _, node := range nodes {
		have = append(have, apitest.ExternalService{
			Kind:        node.Kind(),
			DisplayName: node.DisplayName(),
		})
	}
	for _, es := range want.Nodes {
		assert.Contains(t, have, es, "Nodes contains value")
	}

	pageInfo, err := conn.PageInfo(ctx)
	assert.Nil(t, err, "PageInfo error")
	assert.Equal(t, want.PageInfo.EndCursor, pageInfo.EndCursor(), "PageInfo.EndCursor")
	assert.Equal(t, want.PageInfo.HasNextPage, pageInfo.HasNextPage(), "PageInfo.HasNextPage")

	count, err := conn.TotalCount(ctx)
	assert.Nil(t, err, "TotalCount error")
	assert.EqualValues(t, want.TotalCount, count, "TotalCount value")
}
