package graphqlbackend

import (
	"context"
	"testing"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestAccessRequestsQuery_All(t *testing.T) {
	mockAccessRequests := []*types.AccessRequest{
		{ID: 1, Email: "a1@example.com", Name: "a1", CreatedAt: time.Now(), AdditionalInfo: "af1"},
		{ID: 2, Email: "a2@example.com", Name: "a2", CreatedAt: time.Now()},
	}
	newMockAccessRequestStore := func(t *testing.T, list []*types.AccessRequest) database.AccessRequestStore {
		mockStore := database.NewMockAccessRequestStore()
		mockStore.ListFunc.SetDefaultReturn(list, nil)
		mockStore.CountFunc.SetDefaultReturn(len(list), nil)
		return mockStore
	}

	t.Parallel()

	t.Run("non-admin user", func(t *testing.T) {
		// setup
		users := database.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1, SiteAdmin: false}, nil)

		db := database.NewMockDB()
		db.UsersFunc.SetDefaultReturn(users)

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
		rootResolver := newSchemaResolver(db, gitserver.NewClient())

		// test
		_, err := rootResolver.AccessRequests(ctx, &database.AccessRequestsFilterOptions{})
		require.Error(t, err)
		require.Equal(t, err, auth.ErrMustBeSiteAdmin)
	})

	t.Run("admin user", func(t *testing.T) {
		// setup
		accessRequestStore := newMockAccessRequestStore(t, mockAccessRequests)
		users := database.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)

		db := database.NewMockDB()
		db.AccessRequestsFunc.SetDefaultReturn(accessRequestStore)
		db.UsersFunc.SetDefaultReturn(users)

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
		rootResolver := newSchemaResolver(db, gitserver.NewClient())

		// test - admin user should be able to see all access requests
		accessRequestsResolver, err := rootResolver.AccessRequests(ctx, &database.AccessRequestsFilterOptions{})
		require.NoError(t, err)

		// test - count should be correct
		count, err := accessRequestsResolver.TotalCount(ctx)
		require.NoError(t, err)
		require.Equal(t, int32(2), count)

		// test - nodes should be correct
		accessRequestResolvers, err := accessRequestsResolver.Nodes(ctx, nil)
		require.NoError(t, err)
		var accessRequests = make([]*types.AccessRequest, len(mockAccessRequests))
		for index, accessRequestResolver := range accessRequestResolvers {
			id, err := UnmarshalAccessRequestID(accessRequestResolver.ID())
			require.NoError(t, err)
			accessRequests[index] = &types.AccessRequest{
				ID:        id,
				Name:      accessRequestResolver.Name(ctx),
				Email:     accessRequestResolver.Email(ctx),
				CreatedAt: accessRequestResolver.CreatedAt(ctx).Time,
			}
			additionalInfo := accessRequestResolver.AdditionalInfo(ctx)
			if additionalInfo != nil {
				accessRequests[index].AdditionalInfo = *additionalInfo
			}
		}
		require.Equal(t, accessRequests, mockAccessRequests)
	})
}

func TestAccessRequestsMutation_SetAccessRequestStatus(t *testing.T) {
	newMockAccessRequestStore := func(t *testing.T, existing *types.AccessRequest, want *types.AccessRequest) database.AccessRequestStore {
		mockStore := database.NewMockAccessRequestStore()
		mockStore.TransactFunc.SetDefaultReturn(mockStore, nil)
		mockStore.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int32) (*types.AccessRequest, error) {
			if id == existing.ID {
				return existing, nil
			}
			return nil, errors.Newf("access request with id %d not found", id)
		})
		mockStore.UpdateFunc.SetDefaultHook(func(ctx context.Context, accessRequest *types.AccessRequest) (*types.AccessRequest, error) {
			require.Equal(t, &want, &accessRequest)
			return existing, nil
		})
		return mockStore
	}

	t.Parallel()

	t.Run("non-admin user", func(t *testing.T) {
		// setup
		users := database.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1, SiteAdmin: false}, nil)

		db := database.NewMockDB()
		db.UsersFunc.SetDefaultReturn(users)

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
		rootResolver := newSchemaResolver(db, gitserver.NewClient())

		// test
		graphqlArgs := struct {
			ID     graphql.ID
			Status types.AccessRequestStatus
		}{
			ID:     MarshalAccessRequestID(1),
			Status: types.AccessRequestStatusApproved,
		}
		_, err := rootResolver.SetAccessRequestStatus(ctx, &graphqlArgs)
		require.Error(t, err)
		require.Equal(t, err, auth.ErrMustBeSiteAdmin)
	})

	t.Run("existing access request", func(t *testing.T) {
		// setup
		existingAccessRequest := &types.AccessRequest{
			ID:             1,
			Email:          "a1@example.com",
			Name:           "a1",
			CreatedAt:      time.Now(),
			AdditionalInfo: "af1",
			Status:         types.AccessRequestStatusPending,
		}
		accessRequestStore := newMockAccessRequestStore(t, existingAccessRequest, &types.AccessRequest{
			ID:     existingAccessRequest.ID,
			Status: types.AccessRequestStatusApproved,
		})
		users := database.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)

		db := database.NewMockDB()
		db.AccessRequestsFunc.SetDefaultReturn(accessRequestStore)
		db.UsersFunc.SetDefaultReturn(users)

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
		rootResolver := newSchemaResolver(db, gitserver.NewClient())

		// test
		graphqlArgs := struct {
			ID     graphql.ID
			Status types.AccessRequestStatus
		}{
			ID:     MarshalAccessRequestID(existingAccessRequest.ID),
			Status: types.AccessRequestStatusApproved,
		}
		_, err := rootResolver.SetAccessRequestStatus(ctx, &graphqlArgs)
		require.NoError(t, err)
	})

	t.Run("non-existing access request", func(t *testing.T) {
		// setup
		existingAccessRequest := &types.AccessRequest{
			ID:             1,
			Email:          "a1@example.com",
			Name:           "a1",
			CreatedAt:      time.Now(),
			AdditionalInfo: "af1",
			Status:         types.AccessRequestStatusPending,
		}
		accessRequestStore := newMockAccessRequestStore(t, existingAccessRequest, &types.AccessRequest{
			ID:     existingAccessRequest.ID,
			Status: types.AccessRequestStatusApproved,
		})
		users := database.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)

		db := database.NewMockDB()
		db.AccessRequestsFunc.SetDefaultReturn(accessRequestStore)
		db.UsersFunc.SetDefaultReturn(users)

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
		rootResolver := newSchemaResolver(db, gitserver.NewClient())

		// test
		graphqlArgs := struct {
			ID     graphql.ID
			Status types.AccessRequestStatus
		}{
			ID:     MarshalAccessRequestID(123),
			Status: types.AccessRequestStatusApproved,
		}
		_, err := rootResolver.SetAccessRequestStatus(ctx, &graphqlArgs)
		require.Error(t, err)
		require.Equal(t, err.Error(), "access request with id 123 not found")
	})
}
