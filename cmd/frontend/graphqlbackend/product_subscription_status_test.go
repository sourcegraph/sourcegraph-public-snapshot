package graphqlbackend

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMaximumAllowedUserCount(t *testing.T) {
	var MaxFreeUsers int32 = 10
	NoLicenseMaximumAllowedUserCount = &MaxFreeUsers

	subscriptionStatus := productSubscriptionStatus{}

	t.Run("free license", func(t *testing.T) {
		var expectedUsers uint = 10
		GetConfiguredProductLicenseInfo = func() (*ProductLicenseInfo, error) {
			return &ProductLicenseInfo{
				TagsValue:      []string{"plan:free-0"},
				UserCountValue: expectedUsers,
				ExpiresAtValue: time.Now().Add(time.Hour * 8600),
			}, nil
		}

		IsFreePlan = func(*ProductLicenseInfo) bool {
			return true
		}
		users, err := subscriptionStatus.MaximumAllowedUserCount(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, MaxFreeUsers, *users)
	})

	t.Run("Free users exceeded", func(t *testing.T) {
		var expectedUsers uint = 20
		GetConfiguredProductLicenseInfo = func() (*ProductLicenseInfo, error) {
			return &ProductLicenseInfo{
				TagsValue:      []string{"plan:business-0"},
				UserCountValue: expectedUsers,
				ExpiresAtValue: time.Now().Add(time.Hour * 8600),
			}, nil
		}
		IsFreePlan = func(*ProductLicenseInfo) bool {
			return false
		}
		users, err := subscriptionStatus.MaximumAllowedUserCount(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, int32(expectedUsers), *users)
	})
}
