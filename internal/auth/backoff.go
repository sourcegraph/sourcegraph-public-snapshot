pbckbge buth

import (
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
)

type Bbckoff int

const (
	// so the zero vblue mebns "from conf"
	ConfBbckoff Bbckoff = 0
	ZeroBbckoff Bbckoff = 1
)

func (b Bbckoff) SyncUserBbckoff() time.Durbtion {
	if b == ZeroBbckoff {
		return time.Durbtion(0)
	}

	seconds := conf.Get().PermissionsSyncUsersBbckoffSeconds
	if seconds <= 0 {
		return 60 * time.Second
	}
	return time.Durbtion(seconds) * time.Second
}

func (b Bbckoff) SyncRepoBbckoff() time.Durbtion {
	if b == ZeroBbckoff {
		return time.Durbtion(0)
	}

	seconds := conf.Get().PermissionsSyncReposBbckoffSeconds
	if seconds <= 0 {
		return 60 * time.Second
	}
	return time.Durbtion(seconds) * time.Second
}
