package licensing

import (
	"context"
	"strconv"

	"github.com/gomodule/redigo/redis"

	"github.com/sourcegraph/sourcegraph/internal/redispool"
)

func getMaxUsers(kv redispool.KeyValue, signature string) (int, string, error) {
	if signature == "" {
		// No license key is in use.
		return 0, "", nil
	}

	key := signature

	lastMax, err := kv.HGet(MaxUsersKey(), key).String()
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
	lastMaxDate, err := kv.HGet(MaxUsersTimeKey(), key).String()
	if err != nil && err != redis.ErrNil {
		return 0, "", err
	}
	return lastMaxInt, lastMaxDate, nil
}

// ActualUserCount returns the actual max number of users that have had accounts on the
// Sourcegraph instance, under the current license.
func ActualUserCount(ctx context.Context, kv redispool.KeyValue) (int32, error) {
	_, signature, err := GetConfiguredProductLicenseInfoWithSignature()
	if err != nil || signature == "" {
		return 0, err
	}

	count, _, err := getMaxUsers(kv, signature)
	return int32(count), err
}

// ActualUserCountDate returns the timestamp when the actual max number of users that have
// had accounts on the Sourcegraph instance, under the current license, was reached.
func ActualUserCountDate(ctx context.Context, kv redispool.KeyValue) (string, error) {
	_, signature, err := GetConfiguredProductLicenseInfoWithSignature()
	if err != nil || signature == "" {
		return "", err
	}

	_, date, err := getMaxUsers(kv, signature)
	return date, err
}

// NoLicenseMaximumAllowedUserCount is the maximum number of user accounts that may exist when
// running without a license. Exceeding this number of user accounts requires a license.
const NoLicenseMaximumAllowedUserCount int32 = 10

// NoLicenseWarningUserCount is the number of user accounts when all users are shown a warning (when running
// without a license).
const NoLicenseWarningUserCount int32 = 10

var (
	keyPrefix = "license_user_count:"
)

func MaxUsersKey() string {
	return keyPrefix + "max"
}

func MaxUsersTimeKey() string {
	return keyPrefix + "max_time"
}
