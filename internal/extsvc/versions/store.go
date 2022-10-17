package versions

import (
	"encoding/json"

	"github.com/gomodule/redigo/redis"

	"github.com/sourcegraph/sourcegraph/internal/redispool"
)

var (
	pool        = redispool.Store
	versionsKey = "extsvcversions"
)

type Version struct {
	ExternalServiceKind string `json:"external_service_kind"`
	Version             string `json:"version"`
	Key                 string `json:"-"`
}

func storeVersions(versions []*Version) error {
	c := pool.Get()
	defer c.Close()

	payload, err := json.Marshal(versions)
	if err != nil {
		return err
	}

	return c.Send("SET", versionsKey, payload)
}

func GetVersions() ([]*Version, error) {
	if MockGetVersions != nil {
		return MockGetVersions()
	}
	c := pool.Get()
	defer c.Close()

	var versions []*Version

	raw, err := redis.Bytes(c.Do("GET", versionsKey))
	if err != nil && err != redis.ErrNil {
		return versions, err
	}

	if err := json.Unmarshal(raw, &versions); err != nil {
		return versions, err
	}

	return versions, nil
}
