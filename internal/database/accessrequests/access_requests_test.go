package accessrequests

import (
	"context"
	"strconv"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/tj/assert"
)

func TestAccessRequests_Create(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	t.Run("valid input", func(t *testing.T) {
		accessRequest, err := NewARClient(db.Client()).Create(ctx, &types.AccessRequest{
			Email:          "a1@example.com",
			Name:           "a1",
			AdditionalInfo: "info1",
		})
		assert.NoError(t, err)
		assert.Equal(t, "a1", accessRequest.Name)
		assert.Equal(t, "info1", accessRequest.AdditionalInfo)
		assert.Equal(t, "a1@example.com", accessRequest.Email)
		assert.Equal(t, types.AccessRequestStatusPending, accessRequest.Status)
	})

	t.Run("existing access request email", func(t *testing.T) {
		_, err := NewARClient(db.Client()).Create(ctx, &types.AccessRequest{
			Email:          "a2@example.com",
			Name:           "a1",
			AdditionalInfo: "info1",
		})
		assert.NoError(t, err)

		_, err = NewARClient(db.Client()).Create(ctx, &types.AccessRequest{
			Email:          "a2@example.com",
			Name:           "a2",
			AdditionalInfo: "info2",
		})
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "cannot create user: err_access_request_with_such_email_exists")
	})

	t.Run("existing verified user email", func(t *testing.T) {
		_, err := db.Users().Create(ctx, database.NewUser{
			Username:        "u",
			Email:           "u@example.com",
			EmailIsVerified: true,
		})
		if err != nil {
			t.Fatal(err)
		}

		_, err = NewARClient(db.Client()).Create(ctx, &types.AccessRequest{
			Email:          "u@example.com",
			Name:           "a3",
			AdditionalInfo: "info3",
		})
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "cannot create user: err_user_with_such_email_exists")
	})
}

func TestAccessRequests_Update(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	accessRequestsClient := NewARClient(db.Client())
	usersStore := db.Users()
	user, _ := usersStore.Create(ctx, database.NewUser{Username: "u1", Email: "u1@email", EmailIsVerified: true})

	t.Run("non-existent access request", func(t *testing.T) {
		nonExistentAccessRequestID := int32(1234)
		updated, err := accessRequestsClient.Update(ctx, &types.AccessRequest{ID: nonExistentAccessRequestID, Status: types.AccessRequestStatusApproved, DecisionByUserID: &user.ID})
		assert.Error(t, err)
		assert.Nil(t, updated)
		assert.Equal(t, err, &ErrNotFound{ID: nonExistentAccessRequestID})
	})

	t.Run("existing access request", func(t *testing.T) {
		accessRequest, err := accessRequestsClient.Create(ctx, &types.AccessRequest{
			Email:          "a1@example.com",
			Name:           "a1",
			AdditionalInfo: "info1",
		})
		assert.NoError(t, err)
		assert.Equal(t, accessRequest.Status, types.AccessRequestStatusPending)
		updated, err := accessRequestsClient.Update(ctx, &types.AccessRequest{ID: accessRequest.ID, Status: types.AccessRequestStatusApproved, DecisionByUserID: &user.ID})
		assert.NotNil(t, updated)
		assert.NoError(t, err)
		assert.Equal(t, updated.Status, types.AccessRequestStatusApproved)
		assert.Equal(t, updated.DecisionByUserID, &user.ID)
	})
}

func TestAccessRequests_GetByID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	client := NewARClient(db.Client())

	t.Run("non-existing access request", func(t *testing.T) {
		nonExistentAccessRequestID := int32(1234)
		accessRequest, err := client.GetByID(ctx, nonExistentAccessRequestID)
		assert.Error(t, err)
		assert.Nil(t, accessRequest)
		assert.Equal(t, err, &ErrNotFound{ID: nonExistentAccessRequestID})
	})
	t.Run("existing access request", func(t *testing.T) {
		createdAccessRequest, err := client.Create(ctx, &types.AccessRequest{Email: "a1@example.com", Name: "a1", AdditionalInfo: "info1"})
		assert.NoError(t, err)
		accessRequest, err := client.GetByID(ctx, createdAccessRequest.ID)
		assert.NoError(t, err)
		assert.Equal(t, accessRequest, createdAccessRequest)
	})
}

func TestAccessRequests_GetByEmail(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := db.AccessRequests()
	client := NewARClient(db.Client())

	t.Run("non-existing access request", func(t *testing.T) {
		nonExistingAccessRequestEmail := "non-existing@example"
		accessRequest, err := store.GetByEmail(ctx, nonExistingAccessRequestEmail)
		assert.Error(t, err)
		assert.Nil(t, accessRequest)
		assert.Equal(t, err, &ErrNotFound{Email: nonExistingAccessRequestEmail})
	})
	t.Run("existing access request", func(t *testing.T) {
		createdAccessRequest, err := client.Create(ctx, &types.AccessRequest{Email: "a1@example.com", Name: "a1", AdditionalInfo: "info1"})
		assert.NoError(t, err)
		accessRequest, err := store.GetByEmail(ctx, createdAccessRequest.Email)
		assert.NoError(t, err)
		assert.Equal(t, accessRequest, createdAccessRequest)
	})
}

func TestAccessRequests_Count(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	accessRequestStore := db.AccessRequests()
	accessRequestClient := NewARClient(db.Client())
	client := NewARClient(db.Client())

	usersStore := db.Users()
	user, _ := usersStore.Create(ctx, database.NewUser{Username: "u1", Email: "u1@email", EmailIsVerified: true})

	ar1, err := client.Create(ctx, &types.AccessRequest{Email: "a1@example.com", Name: "a1", AdditionalInfo: "info1"})
	assert.NoError(t, err)
	ar2, err := client.Create(ctx, &types.AccessRequest{Email: "a2@example.com", Name: "a2", AdditionalInfo: "info2"})
	assert.NoError(t, err)
	_, err = client.Create(ctx, &types.AccessRequest{Email: "a3@example.com", Name: "a3", AdditionalInfo: "info3"})
	assert.NoError(t, err)

	t.Run("all", func(t *testing.T) {
		count, err := accessRequestStore.Count(ctx, &database.AccessRequestsFilterArgs{})
		assert.NoError(t, err)
		assert.Equal(t, count, 3)
	})

	t.Run("by status", func(t *testing.T) {
		accessRequestClient.Update(ctx, &types.AccessRequest{ID: ar1.ID, Status: types.AccessRequestStatusApproved, DecisionByUserID: &user.ID})
		accessRequestClient.Update(ctx, &types.AccessRequest{ID: ar2.ID, Status: types.AccessRequestStatusRejected, DecisionByUserID: &user.ID})

		pending := types.AccessRequestStatusPending
		count, err := accessRequestStore.Count(ctx, &database.AccessRequestsFilterArgs{Status: &pending})
		assert.NoError(t, err)
		assert.Equal(t, 1, count)

		rejected := types.AccessRequestStatusRejected
		count, err = accessRequestStore.Count(ctx, &database.AccessRequestsFilterArgs{Status: &rejected})
		assert.NoError(t, err)
		assert.Equal(t, 1, count)

		approved := types.AccessRequestStatusApproved
		count, err = accessRequestStore.Count(ctx, &database.AccessRequestsFilterArgs{Status: &approved})
		assert.NoError(t, err)
		assert.Equal(t, count, 1)
	})
}

func TestAccessRequests_List(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	accessRequestStore := db.AccessRequests()
	accessRequestClient := NewARClient(db.Client())

	usersStore := db.Users()
	user, _ := usersStore.Create(ctx, database.NewUser{Username: "u1", Email: "u1@email", EmailIsVerified: true})

	ar1, err := accessRequestClient.Create(ctx, &types.AccessRequest{Email: "a1@example.com", Name: "a1", AdditionalInfo: "info1"})
	assert.NoError(t, err)
	ar2, err := accessRequestClient.Create(ctx, &types.AccessRequest{Email: "a2@example.com", Name: "a2", AdditionalInfo: "info2"})
	assert.NoError(t, err)
	_, err = accessRequestClient.Create(ctx, &types.AccessRequest{Email: "a3@example.com", Name: "a3", AdditionalInfo: "info3"})
	assert.NoError(t, err)

	t.Run("all", func(t *testing.T) {
		accessRequests, err := accessRequestStore.List(ctx, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, len(accessRequests), 3)

		// map to names
		names := make([]string, len(accessRequests))
		for i, ar := range accessRequests {
			names[i] = ar.Name
		}

		assert.Equal(t, []string{"a3", "a2", "a1"}, names)
	})

	t.Run("order", func(t *testing.T) {
		accessRequests, err := accessRequestStore.List(ctx, nil, &database.PaginationArgs{OrderBy: database.OrderBy{{Field: "name"}}, Ascending: true})
		assert.NoError(t, err)
		assert.Equal(t, len(accessRequests), 3)

		// map to names
		names := make([]string, len(accessRequests))
		for i, ar := range accessRequests {
			names[i] = ar.Name
		}

		assert.Equal(t, names, []string{"a1", "a2", "a3"})
	})

	t.Run("limit & pagination", func(t *testing.T) {
		one := 1
		accessRequests, err := accessRequestStore.List(ctx, nil, &database.PaginationArgs{First: &one})
		assert.NoError(t, err)
		assert.Equal(t, len(accessRequests), 1)

		// map to names
		names := make([]string, len(accessRequests))
		for i, ar := range accessRequests {
			names[i] = ar.Name
		}

		assert.Equal(t, names, []string{"a3"})

		after := strconv.Itoa(int(accessRequests[0].ID))
		two := int(2)
		accessRequests, err = accessRequestStore.List(ctx, nil, &database.PaginationArgs{First: &two, After: &after, OrderBy: database.OrderBy{{Field: string(ListID)}}})
		assert.NoError(t, err)
		assert.Equal(t, 2, len(accessRequests))

		// map to names
		names = make([]string, len(accessRequests))
		for i, ar := range accessRequests {
			names[i] = ar.Name
		}

		assert.Equal(t, names, []string{"a2", "a1"})
	})

	t.Run("by status", func(t *testing.T) {
		accessRequestClient.Update(ctx, &types.AccessRequest{ID: ar1.ID, Status: types.AccessRequestStatusApproved, DecisionByUserID: &user.ID})
		accessRequestClient.Update(ctx, &types.AccessRequest{ID: ar2.ID, Status: types.AccessRequestStatusRejected, DecisionByUserID: &user.ID})

		// list all pending
		pending := types.AccessRequestStatusPending
		accessRequests, err := accessRequestStore.List(ctx, &database.AccessRequestsFilterArgs{Status: &pending}, nil)
		assert.NoError(t, err)
		assert.Equal(t, len(accessRequests), 1)

		// list all rejected
		rejected := types.AccessRequestStatusRejected
		accessRequests, err = accessRequestStore.List(ctx, &database.AccessRequestsFilterArgs{Status: &rejected}, nil)
		assert.NoError(t, err)
		assert.Equal(t, len(accessRequests), 1)

		// list all approved
		approved := types.AccessRequestStatusApproved
		accessRequests, err = accessRequestStore.List(ctx, &database.AccessRequestsFilterArgs{Status: &approved}, nil)
		assert.NoError(t, err)
		assert.Equal(t, len(accessRequests), 1)
	})
}
