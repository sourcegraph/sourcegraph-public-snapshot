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

func TestAccessRequests_Create(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	t.Run("valid input", func(t *testing.T) {
		accessRequest, err := NewStore(db.DBStore()).Create(ctx, &types.AccessRequest{
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
		_, err := NewStore(db.DBStore()).Create(ctx, &types.AccessRequest{
			Email:          "a2@example.com",
			Name:           "a1",
			AdditionalInfo: "info1",
		})
		assert.NoError(t, err)

		_, err = NewStore(db.DBStore()).Create(ctx, &types.AccessRequest{
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

		_, err = NewStore(db.DBStore()).Create(ctx, &types.AccessRequest{
			Email:          "u@example.com",
			Name:           "a3",
			AdditionalInfo: "info3",
		})
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "cannot create user: err_user_with_such_email_exists")
	})
}
