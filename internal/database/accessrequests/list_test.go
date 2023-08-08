package accessrequests

import (
	"context"
	"strconv"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestAccessRequests_List(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	accessRequestClient := NewStore(db.DBStore())

	usersStore := db.Users()
	user, _ := usersStore.Create(ctx, database.NewUser{Username: "u1", Email: "u1@email", EmailIsVerified: true})

	ar1, err := accessRequestClient.Create(ctx, &types.AccessRequest{Email: "a1@example.com", Name: "a1", AdditionalInfo: "info1"})
	assert.NoError(t, err)
	ar2, err := accessRequestClient.Create(ctx, &types.AccessRequest{Email: "a2@example.com", Name: "a2", AdditionalInfo: "info2"})
	assert.NoError(t, err)
	_, err = accessRequestClient.Create(ctx, &types.AccessRequest{Email: "a3@example.com", Name: "a3", AdditionalInfo: "info3"})
	assert.NoError(t, err)

	t.Run("all", func(t *testing.T) {
		accessRequests, err := accessRequestClient.List(ctx, nil, nil)
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
		accessRequests, err := accessRequestClient.List(ctx, nil, &database.PaginationArgs{OrderBy: database.OrderBy{{Field: "name"}}, Ascending: true})
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
		accessRequests, err := accessRequestClient.List(ctx, nil, &database.PaginationArgs{First: &one})
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
		accessRequests, err = accessRequestClient.List(ctx, nil, &database.PaginationArgs{First: &two, After: &after, OrderBy: database.OrderBy{{Field: string(ListID)}}})
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
		accessRequests, err := accessRequestClient.List(ctx, &FilterArgs{Status: &pending}, nil)
		assert.NoError(t, err)
		assert.Equal(t, len(accessRequests), 1)

		// list all rejected
		rejected := types.AccessRequestStatusRejected
		accessRequests, err = accessRequestClient.List(ctx, &FilterArgs{Status: &rejected}, nil)
		assert.NoError(t, err)
		assert.Equal(t, len(accessRequests), 1)

		// list all approved
		approved := types.AccessRequestStatusApproved
		accessRequests, err = accessRequestClient.List(ctx, &FilterArgs{Status: &approved}, nil)
		assert.NoError(t, err)
		assert.Equal(t, len(accessRequests), 1)
	})
}
