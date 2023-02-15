package versions

import (
	"encoding/json"

	"github.com/gomodule/redigo/redis"

	"github.com/sourcegraph/sourcegraph/internal/redispool"
)

var (
	store       = redispool.Store
	versionsKey = "extsvcversions"
)

type Version struct {
	ExternalServiceKind string `json:"external_service_kind"`
	Version             string `json:"version"`
	Key                 string `json:"-"`
}

func storeVersions(versions []*Version) error {
	payload, err := json.Marshal(versions)
	if err != nil {
		return err
	}

	return store.Set(versionsKey, payload)
}

func GetVersions() ([]*Version, error) {
	if MockGetVersions != nil {
		return MockGetVersions()
	}
	var versions []*Version

	raw, err := store.Get(versionsKey).Bytes()
	if err != nil && err != redis.ErrNil {
		return versions, err
	}

	if err := json.Unmarshal(raw, &versions); err != nil {
		return versions, err
	}

	return versions, nil
}
