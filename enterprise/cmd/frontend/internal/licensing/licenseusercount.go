package licensing

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/hooks"
	"github.com/sourcegraph/sourcegraph/pkg/redispool"

	log15 "gopkg.in/inconshreveable/log15.v2"
)

var (
	fetchOnce sync.Once
	pool      = redispool.Store
	keyPrefix = "license_user_count:"

	started bool
)

func init() {
	// Start counting max users on the instance on launch.
	hooks.AfterDBInit = func() {
		go StartMaxUserCount()
	}
	// Make the Site.productSubscription.actualUserCount and Site.productSubscription.actualUserCountDate
	// GraphQL fields return the proper max user count and timestamp on the current license.
	graphqlbackend.ActualUserCount = actualUserCount
	graphqlbackend.ActualUserCountDate = actualUserCountDate
}

// setMaxUsers sets the max users associated with a license key if the new max count is greater than the previous max.
func setMaxUsers(key string, count int) error {
	c := pool.Get()
	defer c.Close()

	lastMax, _, err := getMaxUsers(c, key)
	if err != nil {
		return err
	}

	if count > lastMax {
		err = c.Send("HSET", maxUsersKey(), key, count)
		if err != nil {
			return err
		}
		return c.Send("HSET", maxUsersTimeKey(), key, time.Now().Format("2006-01-02 15:04:05 UTC"))
	}
	return nil
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

// checkMaxUsers runs periodically, and if a license key is in use, updates the
// record of maximum count of user accounts in use.
func checkMaxUsers(ctx context.Context, signature string) error {
	if signature == "" {
		// No license key is in use.
		return nil
	}

	count, err := db.Users.Count(ctx, nil)
	if err != nil {
		log15.Error("licensing.checkMaxUsers: error getting user count", "error", err)
		return err
	}
	err = setMaxUsers(signature, int(count))
	if err != nil {
		log15.Error("licensing.checkMaxUsers: error setting new max users", "error", err)
		return err
	}
	return nil
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

// StartMaxUserCount starts checking for a new count of max user accounts periodically.
func StartMaxUserCount() {
	if started {
		panic("already started")
	}
	started = true

	ctx := context.Background()
	const delay = 360 * time.Minute
	for {
		_, signature, err := GetConfiguredProductLicenseInfoWithSignature()
		if err != nil {
			log15.Error("licensing.StartMaxUserCount: error getting configured license info")
		} else if signature != "" {
			ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
			_ = checkMaxUsers(ctx, signature) // updates global state on its own, can safely ignore return value
			cancel()
		}
		time.Sleep(delay)
	}
}
