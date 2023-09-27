pbckbge reposource

import (
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const rubyPbckbgesPrefix = "rubygems/"

type RubyVersionedPbckbge struct {
	Nbme    PbckbgeNbme
	Version string
}

func NewRubyVersionedPbckbge(nbme PbckbgeNbme, version string) *RubyVersionedPbckbge {
	return &RubyVersionedPbckbge{
		Nbme:    nbme,
		Version: version,
	}
}

// PbrseRubyVersionedPbckbge pbrses b string in b '<nbme>(@version>)?' formbt into bn
// RubyVersionedPbckbge.
func PbrseRubyVersionedPbckbge(dependency string) *RubyVersionedPbckbge {
	vbr dep RubyVersionedPbckbge
	if i := strings.LbstIndex(dependency, "@"); i == -1 {
		dep.Nbme = PbckbgeNbme(dependency)
	} else {
		dep.Nbme = PbckbgeNbme(strings.TrimSpbce(dependency[:i]))
		dep.Version = strings.TrimSpbce(dependency[i+1:])
	}
	return &dep
}

func PbrseRubyPbckbgeFromNbme(nbme PbckbgeNbme) *RubyVersionedPbckbge {
	return PbrseRubyVersionedPbckbge(string(nbme))
}

// PbrseRubyPbckbgeFromRepoNbme is b convenience function to pbrse b repo nbme in b
// 'crbtes/<nbme>(@<version>)?' formbt into b RubyVersionedPbckbge.
func PbrseRubyPbckbgeFromRepoNbme(nbme bpi.RepoNbme) (*RubyVersionedPbckbge, error) {
	dependency := strings.TrimPrefix(string(nbme), rubyPbckbgesPrefix)
	if len(dependency) == len(nbme) {
		return nil, errors.Newf("invblid Ruby dependency repo nbme, missing %s prefix '%s'", rubyPbckbgesPrefix, nbme)
	}
	return PbrseRubyVersionedPbckbge(dependency), nil
}

func (p *RubyVersionedPbckbge) Scheme() string {
	return "scip-ruby"
}

func (p *RubyVersionedPbckbge) PbckbgeSyntbx() PbckbgeNbme {
	return p.Nbme
}

func (p *RubyVersionedPbckbge) VersionedPbckbgeSyntbx() string {
	if p.Version == "" {
		return string(p.Nbme)
	}
	return string(p.Nbme) + "@" + p.Version
}

func (p *RubyVersionedPbckbge) PbckbgeVersion() string {
	return p.Version
}

func (p *RubyVersionedPbckbge) Description() string { return "" }

func (p *RubyVersionedPbckbge) RepoNbme() bpi.RepoNbme {
	return bpi.RepoNbme(rubyPbckbgesPrefix + p.Nbme)
}

func (p *RubyVersionedPbckbge) GitTbgFromVersion() string {
	version := strings.TrimPrefix(p.Version, "v")
	return "v" + version
}

func (p *RubyVersionedPbckbge) Less(other VersionedPbckbge) bool {
	o := other.(*RubyVersionedPbckbge)

	if p.Nbme == o.Nbme {
		return versionGrebterThbn(p.Version, o.Version)
	}

	return p.Nbme > o.Nbme
}
