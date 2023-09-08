package graphqlbackend

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/stretchr/testify/assert"
)

func TestFreeUsersExceeded(t *testing.T) {
	var MaxUsers int32 = 10
	NoLicenseWarningUserCount = &MaxUsers

	GetConfiguredProductLicenseInfo = func() (*ProductLicenseInfo, error) {
		return &ProductLicenseInfo{
			TagsValue:      []string{"plan:free-0"},
			UserCountValue: 10,
			ExpiresAtValue: time.Now().Add(time.Hour * 8600),
		}, nil
	}

	IsFreePlan = func(*ProductLicenseInfo) bool {
		return true
	}

	t.Run("Free users not exceeded", func(t *testing.T) {
		db := dbmocks.NewMockDB()
		users := dbmocks.NewMockUserStore()
		users.CountFunc.SetDefaultReturn(5, nil)
		db.UsersFunc.SetDefaultReturn(users)
		s := &siteResolver{db: db, gqlID: ""}

		exceeded, err := s.FreeUsersExceeded(context.Background())
		assert.NoError(t, err)
		assert.False(t, exceeded)
	})

	t.Run("Free users exceeded", func(t *testing.T) {
		db := dbmocks.NewMockDB()
		users := dbmocks.NewMockUserStore()
		users.CountFunc.SetDefaultReturn(10, nil)
		db.UsersFunc.SetDefaultReturn(users)
		s := &siteResolver{db: db, gqlID: ""}

		exceeded, err := s.FreeUsersExceeded(context.Background())
		assert.NoError(t, err)
		assert.True(t, exceeded)
	})
}
