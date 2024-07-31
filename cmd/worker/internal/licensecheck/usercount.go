package licensecheck

import (
	"context"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// A UsersStore captures the necessary methods for the licensing
// package to query Sourcegraph users. It allows decoupling this package
// from the OSS database package.
type UsersStore interface {
	// Count returns the total count of active Sourcegraph users.
	Count(context.Context) (int, error)
}

// newMaxUserCountRoutine creates a periodic goroutine checking for a new count of
// max user accounts.
func newMaxUserCountRoutine(logger log.Logger, kv redispool.KeyValue, s UsersStore) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		goroutine.HandlerFunc(func(ctx context.Context) error {
			_, signature, err := licensing.GetConfiguredProductLicenseInfoWithSignature()
			if err != nil {
				return errors.Wrap(err, "error getting configured license info")
			}
			if signature == "" {
				return nil
			}

			ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
			defer cancel()
			err = checkMaxUsers(ctx, logger, kv, s, signature) // updates global state on its own, can safely ignore return value
			if err != nil {
				return errors.Wrap(err, "failed to check max users")
			}

			return nil
		}),
		goroutine.WithName("licensecheck.maxusercount"),
		goroutine.WithDescription("Periodically checks for a new count of max user accounts"),
		goroutine.WithInterval(360*time.Minute),
	)
}

// setMaxUsers sets the max users associated with a license key if the new max count is greater than the previous max.
func setMaxUsers(kv redispool.KeyValue, key string, count int) error {
	lastMax, _, err := getMaxUsers(kv, key)
	if err != nil {
		return err
	}

	if count > lastMax {
		err := kv.HSet(licensing.MaxUsersKey(), key, count)
		if err != nil {
			return err
		}
		return kv.HSet(licensing.MaxUsersTimeKey(), key, time.Now().Format("2006-01-02 15:04:05 UTC"))
	}
	return nil
}

func getMaxUsers(kv redispool.KeyValue, key string) (int, string, error) {
	lastMax, err := kv.HGet(licensing.MaxUsersKey(), key).String()
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
	lastMaxDate, err := kv.HGet(licensing.MaxUsersTimeKey(), key).String()
	if err != nil && err != redis.ErrNil {
		return 0, "", err
	}
	return lastMaxInt, lastMaxDate, nil
}

// checkMaxUsers runs periodically, and if a license key is in use, updates the
// record of maximum count of user accounts in use.
func checkMaxUsers(ctx context.Context, logger log.Logger, kv redispool.KeyValue, s UsersStore, signature string) error {
	if signature == "" {
		// No license key is in use.
		return nil
	}

	count, err := s.Count(ctx)
	if err != nil {
		logger.Error("error getting user count", log.Error(err))
		return err
	}
	err = setMaxUsers(kv, signature, count)
	if err != nil {
		logger.Error("error setting new max users", log.Error(err))
		return err
	}
	return nil
}
