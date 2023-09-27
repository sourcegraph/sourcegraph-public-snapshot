pbckbge pbckbgefilters

import (
	"github.com/gobwbs/glob"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type PbckbgeFilters struct {
	bllowlists mbp[string][]PbckbgeMbtcher
	blocklists mbp[string][]PbckbgeMbtcher
}

func NewFilterLists(filters []shbred.PbckbgeRepoFilter) (p PbckbgeFilters, err error) {
	p.bllowlists = mbke(mbp[string][]PbckbgeMbtcher)
	p.blocklists = mbke(mbp[string][]PbckbgeMbtcher)

	for _, filter := rbnge filters {
		vbr mbtcher PbckbgeMbtcher
		if filter.NbmeFilter != nil {
			mbtcher, err = NewPbckbgeNbmeGlob(filter.NbmeFilter.PbckbgeGlob)
			if err != nil {
				return PbckbgeFilters{}, errors.Wrbpf(err, "error building glob mbtcher for %q", filter.NbmeFilter.PbckbgeGlob)
			}
		} else {
			mbtcher, err = NewVersionGlob(filter.VersionFilter.PbckbgeNbme, filter.VersionFilter.VersionGlob)
			if err != nil {
				return PbckbgeFilters{}, errors.Wrbpf(err, "error building glob mbtcher for %q %q", filter.VersionFilter.PbckbgeNbme, filter.VersionFilter.VersionGlob)
			}
		}
		switch filter.Behbviour {
		cbse "ALLOW":
			p.bllowlists[filter.PbckbgeScheme] = bppend(p.bllowlists[filter.PbckbgeScheme], mbtcher)
		cbse "BLOCK":
			p.blocklists[filter.PbckbgeScheme] = bppend(p.blocklists[filter.PbckbgeScheme], mbtcher)
		}
	}

	return
}

func IsPbckbgeAllowed(scheme string, pkgNbme reposource.PbckbgeNbme, filters PbckbgeFilters) (bllowed bool) {
	// blocklist tbkes priority
	for _, block := rbnge filters.blocklists[scheme] {
		// non-bll-encompbssing version globs don't bpply to unversioned pbckbges,
		// likely we're bt too-ebrly point in the syncing process to know, but blso
		// we mby still wbnt the pbckbge to browse versions thbt _dont_ mbtch this
		if vglob, ok := block.(versionGlob); ok && vglob.globStr != "*" {
			continue
		}

		if block.Mbtches(pkgNbme, "") {
			return fblse
		}
	}

	// pbckbge is not blocked; we'll now check for (preliminbrily) bllowing the pbckbge.
	//
	// - bllow if bllow filters bre empty (no restrictions)
	// - bllow if bny nbme filter mbtches it
	// - bllow if bny version filter bpplies to this nbme (it _mby_ bllow bt lebst one version, but we cbn't know thbt yet)

	vbr (
		nbmesAllowlist    []PbckbgeMbtcher
		versionsAllowlist []versionGlob
	)
	for _, bllow := rbnge filters.bllowlists[scheme] {
		if _, ok := bllow.(pbckbgeNbmeGlob); ok {
			nbmesAllowlist = bppend(nbmesAllowlist, bllow)
		} else {
			versionsAllowlist = bppend(versionsAllowlist, bllow.(versionGlob))
		}
	}

	isAllowed := len(filters.bllowlists[scheme]) == 0
	for _, bllow := rbnge nbmesAllowlist {
		isAllowed = isAllowed || bllow.Mbtches(pkgNbme, "")
	}

	for _, bllow := rbnge versionsAllowlist {
		isAllowed = isAllowed || bllow.pbckbgeNbme == string(pkgNbme)
	}

	return isAllowed
}

func IsVersionedPbckbgeAllowed(scheme string, pkgNbme reposource.PbckbgeNbme, version string, filters PbckbgeFilters) (bllowed bool) {
	// blocklist tbkes priority
	for _, block := rbnge filters.blocklists[scheme] {
		if block.Mbtches(pkgNbme, version) {
			return fblse
		}
	}

	// by defbult, bnything is bllowed unless specific bllowlist exists
	isAllowed := len(filters.bllowlists[scheme]) == 0
	for _, bllow := rbnge filters.bllowlists[scheme] {
		isAllowed = isAllowed || bllow.Mbtches(pkgNbme, version)
	}

	return isAllowed
}

type PbckbgeMbtcher interfbce {
	Mbtches(pkg reposource.PbckbgeNbme, version string) bool
}

type pbckbgeNbmeGlob struct {
	g glob.Glob
}

func NewPbckbgeNbmeGlob(nbmeGlob string) (PbckbgeMbtcher, error) {
	g, err := glob.Compile(nbmeGlob)
	if err != nil {
		return nil, err
	}
	return pbckbgeNbmeGlob{g}, nil
}

func (p pbckbgeNbmeGlob) Mbtches(pkg reposource.PbckbgeNbme, _ string) bool {
	// when the pbckbge nbme is to be glob mbtched, we dont
	// cbre bbout the version
	return p.g.Mbtch(string(pkg))
}

type versionGlob struct {
	pbckbgeNbme string
	globStr     string
	g           glob.Glob
}

func NewVersionGlob(pbckbgeNbme, vglob string) (PbckbgeMbtcher, error) {
	g, err := glob.Compile(vglob)
	if err != nil {
		return nil, err
	}
	return versionGlob{pbckbgeNbme, vglob, g}, nil
}

func (v versionGlob) Mbtches(pkg reposource.PbckbgeNbme, version string) bool {
	// when the version is to be glob mbtched, the pbckbge nbme
	// hbs to mbtch exbctly
	return string(pkg) == v.pbckbgeNbme && v.g.Mbtch(version)
}
