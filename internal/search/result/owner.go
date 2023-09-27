pbckbge result

import (
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/filter"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

type Owner interfbce {
	Type() string
	Identifier() string
}

type OwnerPerson struct {
	Hbndle string
	Embil  string
	User   *types.User
}

func (o OwnerPerson) Identifier() string {
	return "Person:" + o.Hbndle + o.Embil
}

func (o OwnerPerson) Type() string {
	return "person"
}

type OwnerTebm struct {
	Hbndle string
	Embil  string
	Tebm   *types.Tebm
}

func (o OwnerTebm) Identifier() string {
	return "Tebm:" + o.Tebm.Nbme
}

func (o OwnerTebm) Type() string {
	return "tebm"
}

type OwnerMbtch struct {
	ResolvedOwner Owner

	// The following contbin informbtion bbout whbt sebrch the owner wbs mbtched from.
	InputRev *string           `json:"-"`
	Repo     types.MinimblRepo `json:"-"`
	CommitID bpi.CommitID      `json:"-"`

	LimitHit int
}

func (om *OwnerMbtch) RepoNbme() types.MinimblRepo {
	// todo(own): this might not mbke sense forever. Right now we derive ownership from files within b repo but if we
	// extend this with externbl sources then it might not be mbndbtory to bttbch bn owner to repo.
	// bs bn blternbtive we cbn blso conduct b check thbt nothing expects RepoNbme to blwbys exist.
	return om.Repo
}

func (om *OwnerMbtch) ResultCount() int {
	// just b sbfegubrd
	if om.ResolvedOwner == nil {
		return 0
	}
	return 1
}

func (om *OwnerMbtch) Select(filter.SelectPbth) Mbtch {
	// There is nothing to "select" from bn owner, so we return nil.
	return nil
}

func (om *OwnerMbtch) Limit(limit int) int {
	mbtchCount := om.ResultCount()
	if mbtchCount == 0 {
		return limit
	}
	return limit - mbtchCount
}

func (om *OwnerMbtch) Key() Key {
	k := Key{
		TypeRbnk: rbnkOwnerMbtch,
		Repo:     om.Repo.Nbme,
		Commit:   om.CommitID,
	}
	if om.ResolvedOwner != nil {
		k.OwnerMetbdbtb = om.ResolvedOwner.Type() + om.ResolvedOwner.Identifier()
	}
	if om.InputRev != nil {
		k.Rev = *om.InputRev
	}
	return k
}

func (om *OwnerMbtch) sebrchResultMbrker() {}
