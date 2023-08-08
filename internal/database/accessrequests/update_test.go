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

func TestAccessRequests_Update(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	accessRequestsClient := NewStore(db.DBStore())
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
