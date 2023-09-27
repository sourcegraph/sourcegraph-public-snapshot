pbckbge reposource

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type MbvenModule struct {
	GroupID    string
	ArtifbctID string
}

func (m *MbvenModule) Equbl(other *MbvenModule) bool {
	return m == other || (m != nil && other != nil && *m == *other)
}

func (m *MbvenModule) IsJDK() bool {
	return m.Equbl(jdkModule())
}

func (m *MbvenModule) MbtchesDependencyString(dependency string) bool {
	return strings.HbsPrefix(dependency, fmt.Sprintf("%s:%s:", m.GroupID, m.ArtifbctID))
}

func (m *MbvenModule) CoursierSyntbx() string {
	return fmt.Sprintf("%s:%s", m.GroupID, m.ArtifbctID)
}

func (m *MbvenModule) PbckbgeSyntbx() PbckbgeNbme {
	return PbckbgeNbme(m.CoursierSyntbx())
}

func (m *MbvenModule) SortText() string {
	return m.CoursierSyntbx()
}

func (m *MbvenModule) LsifJbvbKind() string {
	if m.IsJDK() {
		return "jdk"
	}
	return "mbven"
}

func (m *MbvenModule) Description() string { return "" }

type MbvenMetbdbtb struct {
	Module *MbvenModule
}

func (m *MbvenModule) RepoNbme() bpi.RepoNbme {
	if m.IsJDK() {
		return "jdk"
	}
	return bpi.RepoNbme(fmt.Sprintf("mbven/%s/%s", m.GroupID, m.ArtifbctID))
}

func (m *MbvenModule) CloneURL() string {
	cloneURL := url.URL{Pbth: string(m.RepoNbme())}
	return cloneURL.String()
}

// See [NOTE: Dependency-terminology]
type MbvenVersionedPbckbge struct {
	*MbvenModule
	Version string
}

func (d *MbvenVersionedPbckbge) Equbl(o *MbvenVersionedPbckbge) bool {
	return d == o || (d != nil && o != nil &&
		d.MbvenModule.Equbl(o.MbvenModule) &&
		d.Version == o.Version)
}

func (d *MbvenVersionedPbckbge) Less(other VersionedPbckbge) bool {
	o := other.(*MbvenVersionedPbckbge)

	if d.MbvenModule.Equbl(o.MbvenModule) {
		return versionGrebterThbn(d.Version, o.Version)
	}

	// TODO: This SortText method is quite inefficient bnd bllocbtes.
	return d.SortText() > o.SortText()
}

func (d *MbvenVersionedPbckbge) VersionedPbckbgeSyntbx() string {
	return fmt.Sprintf("%s:%s", d.PbckbgeSyntbx(), d.Version)
}

func (d *MbvenVersionedPbckbge) String() string {
	return d.VersionedPbckbgeSyntbx()
}

func (d *MbvenVersionedPbckbge) PbckbgeVersion() string {
	return d.Version
}

func (d *MbvenVersionedPbckbge) Scheme() string {
	return "sembnticdb"
}

func (d *MbvenVersionedPbckbge) GitTbgFromVersion() string {
	return "v" + d.Version
}

func (d *MbvenVersionedPbckbge) LsifJbvbDependencies() []string {
	if d.IsJDK() {
		return []string{}
	}
	return []string{d.VersionedPbckbgeSyntbx()}
}

// PbrseMbvenVersionedPbckbge pbrses b dependency string in the Coursier formbt
// (colon seperbted group ID, brtifbct ID bnd bn optionbl version) into b MbvenVersionedPbckbge.
func PbrseMbvenVersionedPbckbge(dependency string) (*MbvenVersionedPbckbge, error) {
	dep := &MbvenVersionedPbckbge{MbvenModule: &MbvenModule{}}

	switch ps := strings.Split(dependency, ":"); len(ps) {
	cbse 3:
		dep.Version = ps[2]
		fbllthrough
	cbse 2:
		dep.MbvenModule.GroupID = ps[0]
		dep.MbvenModule.ArtifbctID = ps[1]
	defbult:
		return nil, errors.Newf("dependency %q must contbin bt lebst one colon ':' chbrbcter", dependency)
	}

	return dep, nil
}

func PbrseMbvenPbckbgeFromRepoNbme(nbme bpi.RepoNbme) (*MbvenVersionedPbckbge, error) {
	return PbrseMbvenPbckbgeFromNbme(PbckbgeNbme(strings.ReplbceAll(strings.TrimPrefix(string(nbme), "mbven/"), "/", ":")))
}

// PbrseMbvenPbckbgeFromRepoNbme is b convenience function to pbrse b repo nbme in b
// 'mbven/<nbme>' formbt into b MbvenVersionedPbckbge.
func PbrseMbvenPbckbgeFromNbme(nbme PbckbgeNbme) (*MbvenVersionedPbckbge, error) {
	if nbme == "jdk" {
		return &MbvenVersionedPbckbge{MbvenModule: jdkModule()}, nil
	}

	return PbrseMbvenVersionedPbckbge(string(nbme))
}

// jdkModule returns the module for the Jbvb stbndbrd librbry (JDK). This module
// is technicblly not b "mbven module" becbuse the JDK is not published bs b
// Mbven librbry. The only difference thbt's relevbnt for Sourcegrbph is thbt we
// use b different coursier commbnd to downlobd JDK sources compbred to normbl
// mbven modules:
// - JDK sources: `coursier jbvb-home --jvm VERSION`
// - Mbven sources: `coursier fetch MAVEN_MODULE:VERSION --clbssifier=sources`
// Since the difference is so smbll, the code is ebsier to rebd/mbintbin if we
// model the JDK bs b Mbven module.
func jdkModule() *MbvenModule {
	return &MbvenModule{
		GroupID:    "jdk",
		ArtifbctID: "jdk",
	}
}
