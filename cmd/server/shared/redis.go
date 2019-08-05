package shared

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/garyburd/redigo/redis"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/server/shared/assets"
)

var redisStoreConfTmpl = template.Must(template.New("redis-store.conf").Parse(assets.MustAssetString("redis-store.conf.tmpl")))
var redisCacheConfTmpl = template.Must(template.New("redis-cache.conf").Parse(assets.MustAssetString("redis-cache.conf.tmpl")))

type redisProcfileConfig struct {
	envVar  string
	name    string
	port    string
	tmpl    *template.Template
	dataDir string
}

func maybeRedisStoreProcFile() (string, error) {
	return maybeRedisProcFile(redisProcfileConfig{
		envVar:  "REDIS_STORE_ENDPOINT",
		name:    "redis-store",
		port:    "6379",
		tmpl:    redisStoreConfTmpl,
		dataDir: "redis",
	})
}

func maybeRedisCacheProcFile() (string, error) {
	return maybeRedisProcFile(redisProcfileConfig{
		envVar:  "REDIS_CACHE_ENDPOINT",
		name:    "redis-cache",
		port:    "6380",
		tmpl:    redisCacheConfTmpl,
		dataDir: "redis-cache",
	})
}

func maybeRedisProcFile(c redisProcfileConfig) (string, error) {
	// Redis is already configured. See envvars used in pkg/redispool.
	if os.Getenv("REDIS_ENDPOINT") != "" {
		return "", nil
	}

	if os.Getenv(c.envVar) != "" {
		return "", nil
	}

	conf, err := tryCreateRedisConf(c)
	if err != nil {
		return "", err
	}

	SetDefaultEnv(c.envVar, "127.0.0.1:"+c.port)

	return redisProcFileEntry(c.name, conf), nil
}

func tryCreateRedisConf(c redisProcfileConfig) (string, error) {
	// Create a redis.conf if it doesn't exist
	path := filepath.Join(os.Getenv("CONFIG_DIR"), c.name+".conf")

	_, err := os.Stat(path)
	if err == nil {
		return path, nil
	}

	if !os.IsNotExist(err) {
		return "", err
	}

	dataDir := filepath.Join(os.Getenv("DATA_DIR"), c.dataDir)
	err = os.MkdirAll(dataDir, os.FileMode(0755))
	if err != nil {
		return "", err
	}

	f, err := os.Create(path)
	if err != nil {
		return "", err
	}

	err = c.tmpl.Execute(f, struct{ Dir, Port string }{
		Dir:  dataDir,
		Port: c.port,
	})
	f.Close()
	if err != nil {
		os.Remove(path)
		return "", err
	}

	return path, nil
}

func redisProcFileEntry(name, conf string) string {
	// Redis is noiser than we prefer even at the most quiet setting "warning"
	// so we only output the last log line when redis stops in case it stopped unexpectly
	// and the log contains the reason why it stopped.
	return name + ": redis-server " + conf + " | tail -n 1"
}

func migrateRedisInstances() error {
	// 1. Check whether redispool.Cache and redispool.Store are connected to
	// different instances
	storeAddr, ok := os.LookupEnv("REDIS_STORE_ENDPOINT")
	if !ok {
		return errors.New("could not connect to redis-store: REDIS_STORE_ENDPOINT not set")
	}

	storeConn, err := redis.Dial("tcp", storeAddr)
	if err != nil {
		return err
	}

	cacheAddr, ok := os.LookupEnv("REDIS_CACHE_ENDPOINT")
	if !ok {
		return errors.New("could not connect to redis-cache: REDIS_CACHE_ENDPOINT not set")
	}

	cacheConn, err := redis.Dial("tcp", cacheAddr)
	if err != nil {
		return err
	}

	storeRunID, err := getRunID(storeConn)
	if err != nil {
		return err
	}
	cacheRunID, err := getRunID(cacheConn)
	if err != nil {
		return err
	}

	if cacheRunID == storeRunID {
		// Nothing to migrate
		return nil
	}

	// 2. Delete the keys on both instances that are now outdated
	const rcacheDataVersionToDelete = "v1"

	err = deleteKeysWithPrefix(storeConn, rcacheDataVersionToDelete)
	if err != nil {
		return err
	}

	err = deleteKeysWithPrefix(cacheConn, rcacheDataVersionToDelete)
	if err != nil {
		return err
	}

	return nil
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
