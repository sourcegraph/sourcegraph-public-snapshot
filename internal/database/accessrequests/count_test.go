package accessrequests

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestAccessRequests_Count(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	accessRequestClient := NewStore(db.DBStore())
	client := NewStore(db.DBStore())

	usersStore := db.Users()
	user, _ := usersStore.Create(ctx, database.NewUser{Username: "u1", Email: "u1@email", EmailIsVerified: true})

	ar1, err := client.Create(ctx, &types.AccessRequest{Email: "a1@example.com", Name: "a1", AdditionalInfo: "info1"})
	assert.NoError(t, err)
	ar2, err := client.Create(ctx, &types.AccessRequest{Email: "a2@example.com", Name: "a2", AdditionalInfo: "info2"})
	assert.NoError(t, err)
	_, err = client.Create(ctx, &types.AccessRequest{Email: "a3@example.com", Name: "a3", AdditionalInfo: "info3"})
	assert.NoError(t, err)

	t.Run("all", func(t *testing.T) {
		count, err := accessRequestClient.Count(ctx, &FilterArgs{})
		assert.NoError(t, err)
		assert.Equal(t, count, 3)
	})

	t.Run("by status", func(t *testing.T) {
		accessRequestClient.Update(ctx, &types.AccessRequest{ID: ar1.ID, Status: types.AccessRequestStatusApproved, DecisionByUserID: &user.ID})
		accessRequestClient.Update(ctx, &types.AccessRequest{ID: ar2.ID, Status: types.AccessRequestStatusRejected, DecisionByUserID: &user.ID})

		pending := types.AccessRequestStatusPending
		count, err := accessRequestClient.Count(ctx, &FilterArgs{Status: &pending})
		assert.NoError(t, err)
		assert.Equal(t, 1, count)

		rejected := types.AccessRequestStatusRejected
		count, err = accessRequestClient.Count(ctx, &FilterArgs{Status: &rejected})
		assert.NoError(t, err)
		assert.Equal(t, 1, count)

		approved := types.AccessRequestStatusApproved
		count, err = accessRequestClient.Count(ctx, &FilterArgs{Status: &approved})
		assert.NoError(t, err)
		assert.Equal(t, count, 1)
	})
}
