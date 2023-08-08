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

func TestAccessRequests_GetByID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	client := NewStore(db.DBStore())

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
