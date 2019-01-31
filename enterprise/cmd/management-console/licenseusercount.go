package main

import (
	"context"
	"strconv"

	"github.com/garyburd/redigo/redis"

	"github.com/sourcegraph/sourcegraph/cmd/management-console/shared"
	"github.com/sourcegraph/sourcegraph/pkg/redispool"
)

var noLicenseMaximumAllowedUserCount int32 = 200
var (
	pool      = redispool.Store
	keyPrefix = "license_user_count:"

	started bool
)

func init() {
	// Make the Site.productSubscription.actualUserCount and Site.productSubscription.actualUserCountDate
	// GraphQL fields return the proper max user count and timestamp on the current license.
	shared.ActualUserCount = actualUserCount
	shared.ActualUserCountDate = actualUserCountDate
}

// GetMaxUsers gets the max users associated with a license key.
func GetMaxUsers(signature string) (int, string, error) {
	c := pool.Get()
	defer c.Close()

	if signature == "" {
		// No license key is in use.
		return 0, "", nil
	}

	return getMaxUsers(c, signature)
}

func getMaxUsers(c redis.Conn, key string) (int, string, error) {
	lastMax, err := redis.String(c.Do("HGET", maxUsersKey(), key))
	if err != nil && err != redis.ErrNil {
		return 0, "", err
	}
	lastMaxInt := 0
	if lastMax != "" {
		lastMaxInt, err = strconv.Atoi(lastMax)
		if err != nil {
			return 0, "", err
		}
	}
	lastMaxDate, err := redis.String(c.Do("HGET", maxUsersTimeKey(), key))
	if err != nil && err != redis.ErrNil {
		return 0, "", err
	}
	return lastMaxInt, lastMaxDate, nil
}

func maxUsersKey() string {
	return keyPrefix + "max"
}

func maxUsersTimeKey() string {
	return keyPrefix + "max_time"
}

// actualUserCount returns the actual max number of users that have had accounts on the
// Sourcegraph instance, under the current license.
func actualUserCount(ctx context.Context) (int32, error) {
	_, signature, err := GetConfiguredProductLicenseInfoWithSignature()
	if err != nil || signature == "" {
		return 0, err
	}

	count, _, err := GetMaxUsers(signature)
	return int32(count), err
}

// actualUserCountDate returns the timestamp when the actual max number of users that have
// had accounts on the Sourcegraph instance, under the current license, was reached.
func actualUserCountDate(ctx context.Context) (string, error) {
	_, signature, err := GetConfiguredProductLicenseInfoWithSignature()
	if err != nil || signature == "" {
		return "", err
	}

	_, date, err := GetMaxUsers(signature)
	return date, err
}
