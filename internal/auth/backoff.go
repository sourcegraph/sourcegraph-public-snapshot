package auth

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/conf"
)

type Backoff int

const (
	// so the zero value means "from conf"
	ConfBackoff Backoff = 0
	ZeroBackoff Backoff = 1
)

func (b Backoff) SyncUserBackoff() time.Duration {
	if b == ZeroBackoff {
		return time.Duration(0)
	}

	seconds := conf.Get().PermissionsSyncUsersBackoffSeconds
	if seconds <= 0 {
		return 60 * time.Second
	}
	return time.Duration(seconds) * time.Second
}

func (b Backoff) SyncRepoBackoff() time.Duration {
	if b == ZeroBackoff {
		return time.Duration(0)
	}

	seconds := conf.Get().PermissionsSyncReposBackoffSeconds
	if seconds <= 0 {
		return 60 * time.Second
	}
	return time.Duration(seconds) * time.Second
}
