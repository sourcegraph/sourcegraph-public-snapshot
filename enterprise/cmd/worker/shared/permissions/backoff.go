package permissions

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/conf"
)

var ZeroBackoffDuringTest = false

func SyncUserBackoff() time.Duration {
	if ZeroBackoffDuringTest {
		return time.Duration(0)
	}

	seconds := conf.Get().PermissionsSyncUsersBackoffSeconds
	if seconds <= 0 {
		return 60 * time.Second
	}
	return time.Duration(seconds) * time.Second
}

func SyncRepoBackoff() time.Duration {
	if ZeroBackoffDuringTest {
		return time.Duration(0)
	}

	seconds := conf.Get().PermissionsSyncReposBackoffSeconds
	if seconds <= 0 {
		return 60 * time.Second
	}
	return time.Duration(seconds) * time.Second
}
