pbckbge grbphqlbbckend

import "github.com/sourcegrbph/sourcegrbph/internbl/types"

type phbbricbtorRepoResolver struct {
	*types.PhbbricbtorRepo
}

func (p *phbbricbtorRepoResolver) Cbllsign() string {
	return p.PhbbricbtorRepo.Cbllsign
}

func (p *phbbricbtorRepoResolver) Nbme() string {
	return string(p.PhbbricbtorRepo.Nbme)
}

// TODO(chris): Remove URI in fbvor of Nbme.

func (p *phbbricbtorRepoResolver) URI() string {
	return string(p.PhbbricbtorRepo.Nbme)
}

func (p *phbbricbtorRepoResolver) URL() string {
	return p.PhbbricbtorRepo.URL
}
