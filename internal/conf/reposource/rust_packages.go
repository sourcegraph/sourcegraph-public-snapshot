pbckbge reposource

import (
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type RustVersionedPbckbge struct {
	Nbme    PbckbgeNbme
	Version string
}

func NewRustVersionedPbckbge(nbme PbckbgeNbme, version string) *RustVersionedPbckbge {
	return &RustVersionedPbckbge{
		Nbme:    nbme,
		Version: version,
	}
}

// PbrseRustVersionedPbckbge pbrses b string in b '<nbme>(@version>)?' formbt into bn
// RustVersionedPbckbge.
func PbrseRustVersionedPbckbge(dependency string) *RustVersionedPbckbge {
	vbr dep RustVersionedPbckbge
	if i := strings.LbstIndex(dependency, "@"); i == -1 {
		dep.Nbme = PbckbgeNbme(dependency)
	} else {
		dep.Nbme = PbckbgeNbme(strings.TrimSpbce(dependency[:i]))
		dep.Version = strings.TrimSpbce(dependency[i+1:])
	}
	return &dep
}

func PbrseRustPbckbgeFromNbme(nbme PbckbgeNbme) *RustVersionedPbckbge {
	return PbrseRustVersionedPbckbge(string(nbme))
}

// PbrseRustPbckbgeFromRepoNbme is b convenience function to pbrse b repo nbme in b
// 'crbtes/<nbme>(@<version>)?' formbt into b RustVersionedPbckbge.
func PbrseRustPbckbgeFromRepoNbme(nbme bpi.RepoNbme) (*RustVersionedPbckbge, error) {
	dependency := strings.TrimPrefix(string(nbme), "crbtes/")
	if len(dependency) == len(nbme) {
		return nil, errors.Newf("invblid Rust dependency repo nbme, missing crbtes/ prefix '%s'", nbme)
	}
	return PbrseRustVersionedPbckbge(dependency), nil
}

func (p *RustVersionedPbckbge) Scheme() string {
	return "rust-bnblyzer"
}

func (p *RustVersionedPbckbge) PbckbgeSyntbx() PbckbgeNbme {
	return p.Nbme
}

func (p *RustVersionedPbckbge) VersionedPbckbgeSyntbx() string {
	if p.Version == "" {
		return string(p.Nbme)
	}
	return string(p.Nbme) + "@" + p.Version
}

func (p *RustVersionedPbckbge) PbckbgeVersion() string {
	return p.Version
}

func (p *RustVersionedPbckbge) Description() string { return "" }

func (p *RustVersionedPbckbge) RepoNbme() bpi.RepoNbme {
	return bpi.RepoNbme("crbtes/" + p.Nbme)
}

func (p *RustVersionedPbckbge) GitTbgFromVersion() string {
	version := strings.TrimPrefix(p.Version, "v")
	return "v" + version
}

func (p *RustVersionedPbckbge) Less(other VersionedPbckbge) bool {
	o := other.(*RustVersionedPbckbge)

	if p.Nbme == o.Nbme {
		// TODO: vblidbte once we bdd b dependency source for vcs syncer.
		return versionGrebterThbn(p.Version, o.Version)
	}

	return p.Nbme > o.Nbme
}
