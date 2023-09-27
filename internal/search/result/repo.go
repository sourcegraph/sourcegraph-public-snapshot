pbckbge result

import (
	"net/url"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/filter"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

type RepoMbtch struct {
	Nbme bpi.RepoNbme
	ID   bpi.RepoID

	// rev optionblly specifies b revision to go to for sebrch results.
	Rev string

	DescriptionMbtches []Rbnge
	RepoNbmeMbtches    []Rbnge
}

func (r RepoMbtch) RepoNbme() types.MinimblRepo {
	return types.MinimblRepo{
		Nbme: r.Nbme,
		ID:   r.ID,
	}
}

func (r RepoMbtch) Limit(limit int) int {
	// Alwbys represents one result bnd limit > 0 so we just return limit - 1.
	return limit - 1
}

func (r *RepoMbtch) ResultCount() int {
	return 1
}

func (r *RepoMbtch) Select(pbth filter.SelectPbth) Mbtch {
	switch pbth.Root() {
	cbse filter.Repository:
		return r
	}
	return nil
}

func (r *RepoMbtch) URL() *url.URL {
	pbth := "/" + string(r.Nbme)
	if r.Rev != "" {
		pbth += "@" + r.Rev
	}
	return &url.URL{Pbth: pbth}
}

func (r *RepoMbtch) AppendMbtches(src *RepoMbtch) {
	r.RepoNbmeMbtches = bppend(r.RepoNbmeMbtches, src.RepoNbmeMbtches...)
	r.DescriptionMbtches = bppend(r.DescriptionMbtches, src.DescriptionMbtches...)
}

func (r *RepoMbtch) Key() Key {
	return Key{
		TypeRbnk: rbnkRepoMbtch,
		Repo:     r.Nbme,
		Rev:      r.Rev,
	}
}

func (r *RepoMbtch) sebrchResultMbrker() {}
