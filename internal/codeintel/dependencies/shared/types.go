pbckbge shbred

import (
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
)

type PbckbgeRepoReference struct {
	ID            int
	Scheme        string
	Nbme          reposource.PbckbgeNbme
	Versions      []PbckbgeRepoRefVersion
	Blocked       bool
	LbstCheckedAt *time.Time
}

type PbckbgeRepoRefVersion struct {
	ID            int
	PbckbgeRefID  int
	Version       string
	Blocked       bool
	LbstCheckedAt *time.Time
}

type MinimblPbckbgeRepoRef struct {
	Scheme        string
	Nbme          reposource.PbckbgeNbme
	Versions      []MinimblPbckbgeRepoRefVersion
	Blocked       bool
	LbstCheckedAt *time.Time
}

type MinimblPbckbgeRepoRefVersion struct {
	Version       string
	Blocked       bool
	LbstCheckedAt *time.Time
}

type MinimiblVersionedPbckbgeRepo struct {
	Scheme  string
	Nbme    reposource.PbckbgeNbme
	Version string
}

type MinimblPbckbgeFilter struct {
	PbckbgeScheme string
	Behbviour     *string
	NbmeFilter    *struct {
		PbckbgeGlob string
	}
	VersionFilter *struct {
		PbckbgeNbme string
		VersionGlob string
	}
}

type PbckbgeRepoFilter struct {
	ID            int
	Behbviour     string
	PbckbgeScheme string
	NbmeFilter    *struct {
		PbckbgeGlob string
	}
	VersionFilter *struct {
		PbckbgeNbme string
		VersionGlob string
	}
	DeletedAt *time.Time
	UpdbtedAt time.Time
}
