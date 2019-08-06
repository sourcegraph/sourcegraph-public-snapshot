package bg

import (
	"errors"
	"fmt"
	"strings"

	"github.com/garyburd/redigo/redis"
	"github.com/sourcegraph/sourcegraph/pkg/redispool"
	"gopkg.in/inconshreveable/log15.v2"
)

const rcacheDataVersionToDelete = "v1"

func DeleteOldCacheDataInRedis() {
	storeConn := redispool.Store.Get()
	defer storeConn.Close()

	cacheConn := redispool.Cache.Get()
	defer cacheConn.Close()

	storeRunID, err := getRunID(storeConn)
	if err != nil {
		log15.Error("Unable to delete old cache data in redis search. Please report this issue.", "error", err)
		return
	}
	cacheRunID, err := getRunID(cacheConn)
	if err != nil {
		log15.Error("Unable to delete old cache data in redis search. Please report this issue.", "error", err)
		return
	}

	// Only run on separate instances
	connsToPurge := make(map[string]redis.Conn)
	connsToPurge[storeRunID] = storeConn
	connsToPurge[cacheRunID] = cacheConn

	for _, c := range connsToPurge {
		err = deleteKeysWithPrefix(c, rcacheDataVersionToDelete)
		if err != nil {
			log15.Error("Unable to delete old cache data in redis search. Please report this issue.", "error", err)
			return
		}
	}
}

func getRunID(c redis.Conn) (string, error) {
	infos, err := redis.String(c.Do("INFO", "server"))
	if err != nil {
		return "", err
	}

	for _, l := range strings.Split(infos, "\n") {
		if strings.HasPrefix(l, "run_id:") {
			s := strings.Split(l, ":")
			return s[1], nil
		}
	}
	return "", errors.New("no run_id found")
}

func deleteKeysWithPrefix(c redis.Conn, prefix string) error {
	pattern := prefix + ":*"

	iter := 0
	keys := make([]string, 0)
	for {
		arr, err := redis.Values(c.Do("SCAN", iter, "MATCH", pattern))
		if err != nil {
			return fmt.Errorf("error retrieving keys with pattern %q", pattern)
		}

		iter, err = redis.Int(arr[0], nil)
		if err != nil {
			return err
		}

		k, err := redis.Strings(arr[1], nil)
		if err != nil {
			return err
		}
		keys = append(keys, k...)
		if iter == 0 {
			break
		}
	}

	if len(keys) == 0 {
		return nil
	}

	const batchSize = 1000
	var batch = make([]interface{}, batchSize, batchSize)

	for i := 0; i < len(keys); i += batchSize {
		j := i + batchSize
		if j > len(keys) {
			j = len(keys)
		}
		currentBatchSize := j - i

		for bi, v := range keys[i:j] {
			batch[bi] = v
		}

		// We ignore whether the number of deleted keys matches what we have in
		// `batch`, because in the time since we constructed `keys` some of the
		// keys might have expired
		_, err := c.Do("DEL", batch[:currentBatchSize]...)
		if err != nil {
			return fmt.Errorf("failed to delete keys: %s", err)
		}
	}

	return nil
}
