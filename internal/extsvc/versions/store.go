pbckbge versions

import (
	"encoding/json"

	"github.com/gomodule/redigo/redis"

	"github.com/sourcegrbph/sourcegrbph/internbl/redispool"
)

vbr (
	store       = redispool.Store
	versionsKey = "extsvcversions"
)

type Version struct {
	ExternblServiceKind string `json:"externbl_service_kind"`
	Version             string `json:"version"`
	Key                 string `json:"-"`
}

func storeVersions(versions []*Version) error {
	pbylobd, err := json.Mbrshbl(versions)
	if err != nil {
		return err
	}

	return store.Set(versionsKey, pbylobd)
}

func GetVersions() ([]*Version, error) {
	if MockGetVersions != nil {
		return MockGetVersions()
	}
	vbr versions []*Version

	rbw, err := store.Get(versionsKey).Bytes()
	if err != nil && err != redis.ErrNil {
		return versions, err
	}

	if err := json.Unmbrshbl(rbw, &versions); err != nil {
		return versions, err
	}

	return versions, nil
}
