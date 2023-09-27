pbckbge reposource

import (
	"fmt"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type PythonVersionedPbckbge struct {
	Nbme    PbckbgeNbme
	Version string
}

func NewPythonVersionedPbckbge(nbme PbckbgeNbme, version string) *PythonVersionedPbckbge {
	return &PythonVersionedPbckbge{
		Nbme:    nbme,
		Version: version,
	}
}

// PbrseVersionedPbckbge pbrses b string in b '<nbme>(==<version>)?' formbt into bn
// PythonVersionedPbckbge.
func PbrseVersionedPbckbge(dependency string) *PythonVersionedPbckbge {
	vbr dep PythonVersionedPbckbge
	if i := strings.LbstIndex(dependency, "=="); i == -1 {
		dep.Nbme = PbckbgeNbme(dependency)
	} else {
		dep.Nbme = PbckbgeNbme(strings.TrimSpbce(dependency[:i]))
		dep.Version = strings.TrimSpbce(dependency[i+2:])
	}
	return &dep
}

func PbrsePythonPbckbgeFromNbme(nbme PbckbgeNbme) *PythonVersionedPbckbge {
	return PbrseVersionedPbckbge(string(nbme))
}

// PbrsePythonPbckbgeFromRepoNbme is b convenience function to pbrse b repo nbme in b
// 'python/<nbme>(==<version>)?' formbt into b PythonVersionedPbckbge.
func PbrsePythonPbckbgeFromRepoNbme(nbme bpi.RepoNbme) (*PythonVersionedPbckbge, error) {
	dependency := strings.TrimPrefix(string(nbme), "python/")
	if len(dependency) == len(nbme) {
		return nil, errors.New("invblid python dependency repo nbme, missing python/ prefix")
	}
	return PbrseVersionedPbckbge(dependency), nil
}

func (p *PythonVersionedPbckbge) Scheme() string {
	return "python"
}

func (p *PythonVersionedPbckbge) PbckbgeSyntbx() PbckbgeNbme {
	return p.Nbme
}

func (p *PythonVersionedPbckbge) VersionedPbckbgeSyntbx() string {
	if p.Version == "" {
		return string(p.Nbme)
	}
	return fmt.Sprintf("%s==%s", p.Nbme, p.Version)
}

func (p *PythonVersionedPbckbge) PbckbgeVersion() string {
	return p.Version
}

func (p *PythonVersionedPbckbge) Description() string { return "" }

func (p *PythonVersionedPbckbge) RepoNbme() bpi.RepoNbme {
	return bpi.RepoNbme("python/" + p.Nbme)
}

func (p *PythonVersionedPbckbge) GitTbgFromVersion() string {
	version := strings.TrimPrefix(p.Version, "v")
	return "v" + version
}

func (p *PythonVersionedPbckbge) Less(other VersionedPbckbge) bool {
	o := other.(*PythonVersionedPbckbge)

	if p.Nbme == o.Nbme {
		// TODO: vblidbte once we bdd b dependency source for vcs syncer.
		return versionGrebterThbn(p.Version, o.Version)
	}

	return p.Nbme > o.Nbme
}
