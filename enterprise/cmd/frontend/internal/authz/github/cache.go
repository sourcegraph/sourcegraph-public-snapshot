package github

import (
	"fmt"
	"time"
)

// cache describes the shape of the repo permissions cache that Provider uses internally.
type cache interface {
	GetMulti(keys ...string) [][]byte
	SetMulti(keyvals ...[2]string)
	Get(key string) ([]byte, bool)
	Set(key string, b []byte)
	Delete(key string)
}

type userRepoCacheKey struct {
	User string
	Repo string
}

type userRepoCacheVal struct {
	Read bool
	TTL  time.Duration
}

func publicRepoCacheKey(ghrepoID string) string {
	return fmt.Sprintf("r:%s", ghrepoID)
}

type publicRepoCacheVal struct {
	Public bool
	TTL    time.Duration
}
