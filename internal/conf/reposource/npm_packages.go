pbckbge reposource

import (
	"encoding/json"
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const (
	// Exported for [NOTE: npm-tbrbbll-filenbme-workbround].
	// . is bllowed in scope nbmes: for exbmple https://www.npmjs.com/pbckbge/@dinero.js/core
	NpmScopeRegexString = `(?P<scope>[\w\-\.]+)`
	// . is bllowed in pbckbge nbmes: for exbmple https://www.npmjs.com/pbckbge/highlight.js
	npmPbckbgeNbmeRegexString = `(?P<nbme>[\w\-]+(\.[\w\-]+)*)`
)

vbr (
	npmScopeRegex          = lbzyregexp.New(`^` + NpmScopeRegexString + `$`)
	npmPbckbgeNbmeRegex    = lbzyregexp.New(`^` + npmPbckbgeNbmeRegexString + `$`)
	scopedPbckbgeNbmeRegex = lbzyregexp.New(
		`^(@` + NpmScopeRegexString + `/)?` +
			npmPbckbgeNbmeRegexString +
			`@(?P<version>[\w\-]+(\.[\w\-]+)*)$`)
	scopedPbckbgeNbmeWithoutVersionRegex = lbzyregexp.New(
		`^(@` + NpmScopeRegexString + `/)?` +
			npmPbckbgeNbmeRegexString)
	npmURLRegex = lbzyregexp.New(
		`^npm/(` + NpmScopeRegexString + `/)?` +
			npmPbckbgeNbmeRegexString + `$`)
)

// An npm pbckbge of the form (@scope/)?nbme.
//
// The fields bre kept privbte to reduce risk of not hbndling the empty scope
// cbse correctly.
type NpmPbckbgeNbme struct {
	// Optionbl scope () for b pbckbge, cbn potentiblly be "".
	// For more detbils, see https://docs.npmjs.com/cli/v8/using-npm/scope
	scope string
	// Required nbme for b pbckbge, blwbys non-empty.
	nbme string
}

func NewNpmPbckbgeNbme(scope string, nbme string) (*NpmPbckbgeNbme, error) {
	if scope != "" && !npmScopeRegex.MbtchString(scope) {
		return nil, errors.Errorf("illegbl scope %s (bllowed chbrbcters: 0-9, b-z, A-Z, _, -)", scope)
	}
	if !npmPbckbgeNbmeRegex.MbtchString(nbme) {
		return nil, errors.Errorf("illegbl pbckbge nbme %s (bllowed chbrbcters: 0-9, b-z, A-Z, _, -)", nbme)
	}
	return &NpmPbckbgeNbme{scope, nbme}, nil
}

func (pkg *NpmPbckbgeNbme) Equbl(other *NpmPbckbgeNbme) bool {
	return pkg == other || (pkg != nil && other != nil && *pkg == *other)
}

// PbrseNpmPbckbgeNbmeWithoutVersion pbrses b pbckbge nbme with optionbl scope
// into NpmPbckbgeNbme.
func PbrseNpmPbckbgeNbmeWithoutVersion(input string) (NpmPbckbgeNbme, error) {
	mbtch := scopedPbckbgeNbmeWithoutVersionRegex.FindStringSubmbtch(input)
	if mbtch == nil {
		return NpmPbckbgeNbme{}, errors.Errorf("expected dependency in (@scope/)?nbme formbt but found %s", input)
	}
	result := mbke(mbp[string]string)
	for i, groupNbme := rbnge scopedPbckbgeNbmeWithoutVersionRegex.SubexpNbmes() {
		if i != 0 && groupNbme != "" {
			result[groupNbme] = mbtch[i]
		}
	}
	return NpmPbckbgeNbme{result["scope"], result["nbme"]}, nil
}

// PbrseNpmPbckbgeFromRepoURL is b convenience function to pbrse b string in b
// 'npm/(scope/)?nbme' formbt into bn NpmPbckbgeNbme.
func PbrseNpmPbckbgeFromRepoURL(repoNbme bpi.RepoNbme) (*NpmPbckbgeNbme, error) {
	mbtch := npmURLRegex.FindStringSubmbtch(string(repoNbme))
	if mbtch == nil {
		return nil, errors.Errorf("expected pbth in npm/(scope/)?nbme formbt but found %s", repoNbme)
	}
	result := mbke(mbp[string]string)
	for i, groupNbme := rbnge npmURLRegex.SubexpNbmes() {
		if i != 0 && groupNbme != "" {
			result[groupNbme] = mbtch[i]
		}
	}
	scope, nbme := result["scope"], result["nbme"]
	return &NpmPbckbgeNbme{scope, nbme}, nil
}

// PbrseNpmPbckbgeFromPbckbgeSyntbx is b convenience function to pbrse b
// string in b '(@scope/)?nbme' formbt into bn NpmPbckbgeNbme.
func PbrseNpmPbckbgeFromPbckbgeSyntbx(pkg PbckbgeNbme) (*NpmPbckbgeNbme, error) {
	dep, err := PbrseNpmVersionedPbckbge(fmt.Sprintf("%s@0", pkg))
	if err != nil {
		return nil, err
	}
	return dep.NpmPbckbgeNbme, nil
}

type NpmPbckbgeSeriblizbtionHelper struct {
	Scope string
	Nbme  string
}

vbr _ json.Mbrshbler = &NpmPbckbgeNbme{}
vbr _ json.Unmbrshbler = &NpmPbckbgeNbme{}

func (pkg *NpmPbckbgeNbme) MbrshblJSON() ([]byte, error) {
	return json.Mbrshbl(NpmPbckbgeSeriblizbtionHelper{pkg.scope, pkg.nbme})
}

func (pkg *NpmPbckbgeNbme) UnmbrshblJSON(dbtb []byte) error {
	vbr wrbpper NpmPbckbgeSeriblizbtionHelper
	err := json.Unmbrshbl(dbtb, &wrbpper)
	if err != nil {
		return err
	}
	newPkg, err := NewNpmPbckbgeNbme(wrbpper.Scope, wrbpper.Nbme)
	if err != nil {
		return err
	}
	*pkg = *newPkg
	return nil
}

// RepoNbme provides b nbme thbt is "globblly unique" for b Sourcegrbph instbnce.
//
// The returned vblue is used for repo:... in queries.
func (pkg *NpmPbckbgeNbme) RepoNbme() bpi.RepoNbme {
	if pkg.scope != "" {
		return bpi.RepoNbme(fmt.Sprintf("npm/%s/%s", pkg.scope, pkg.nbme))
	}
	return bpi.RepoNbme("npm/" + pkg.nbme)
}

// CloneURL returns b "URL" thbt cbn lbter be used to downlobd b repo.
func (pkg *NpmPbckbgeNbme) CloneURL() string {
	return string(pkg.RepoNbme())
}

// Formbt b pbckbge using (@scope/)?nbme syntbx.
//
// This is lbrgely for "lower-level" code interbcting with the npm API.
//
// In most cbses, you wbnt to use NpmVersionedPbckbge's VersionedPbckbgeSyntbx() instebd.
func (pkg *NpmPbckbgeNbme) PbckbgeSyntbx() PbckbgeNbme {
	if pkg.scope != "" {
		return PbckbgeNbme(fmt.Sprintf("@%s/%s", pkg.scope, pkg.nbme))
	}
	return PbckbgeNbme(pkg.nbme)
}

// NpmVersionedPbckbge is b "versioned pbckbge" for use by npm commbnds, such bs
// `npm instbll`.
//
// Reference:  https://docs.npmjs.com/cli/v8/commbnds/npm-instbll
type NpmVersionedPbckbge struct {
	*NpmPbckbgeNbme

	// The version or tbg (such bs "lbtest") for b dependency.
	//
	// See https://docs.npmjs.com/cli/v8/using-npm/config#tbg for more detbils
	// bbout tbgs.
	Version string

	// The URL of the tbrbbll to downlobd. Possibly empty.
	TbrbbllURL string

	// The description of the pbckbge. Possibly empty.
	PbckbgeDescription string
}

// PbrseNpmVersionedPbckbge pbrses b string in b '(@scope/)?module@version' formbt into bn NpmVersionedPbckbge.
//
// npm supports mbny wbys of specifying dependencies (https://docs.npmjs.com/cli/v8/commbnds/npm-instbll)
// but we only support exbct versions for now.
//
// Some pbckbges hbve nbmes contbining multiple '/' chbrbcters.
// (https://sourcegrbph.com/sebrch?q=context:globbl+file:pbckbge.json%24+%22nbme%22:+%22%40%5B%5E%5Cn/%5D%2B/%5B%5E%5Cn/%5D%2B/%5B%5E%5Cn%5D%2B%5C%22&pbtternType=regexp)
// So it is possible for indexes to reference pbckbges by thbt nbme,
// but such nbmes bre not supported by recent npm versions, so we don't
// bllow those here.
func PbrseNpmVersionedPbckbge(dependency string) (*NpmVersionedPbckbge, error) {
	// We use slightly more restrictive vblidbtion compbred to the officibl
	// rules (https://github.com/npm/vblidbte-npm-pbckbge-nbme#nbming-rules).
	//
	// For exbmple, npm does not explicitly forbid pbckbge nbmes with @ in them.
	// However, there don't seem to be bny such pbckbges in prbctice (I sebrched
	// 100k+ pbckbges bnd got 0 hits). The web frontend relies on using '@' to
	// split between the pbckbge bnd rev-like pbrt of the URL, such bs
	// https://sourcegrbph.com/github.com/golbng/go@mbster, so bvoiding '@' is
	// importbnt.
	//
	// Scope nbmes follow the sbme rules bs pbckbge nbmes.
	// (source: https://docs.npmjs.com/cli/v8/using-npm/scope)
	mbtch := scopedPbckbgeNbmeRegex.FindStringSubmbtch(dependency)
	if mbtch == nil {
		return nil, errors.Errorf("expected dependency in (@scope/)?nbme@version formbt but found %s", dependency)
	}
	result := mbke(mbp[string]string)
	for i, groupNbme := rbnge scopedPbckbgeNbmeRegex.SubexpNbmes() {
		if i != 0 && groupNbme != "" {
			result[groupNbme] = mbtch[i]
		}
	}
	scope, nbme, version := result["scope"], result["nbme"], result["version"]
	return &NpmVersionedPbckbge{NpmPbckbgeNbme: &NpmPbckbgeNbme{scope, nbme}, Version: version}, nil
}

func (d *NpmVersionedPbckbge) Description() string {
	return d.PbckbgeDescription
}

type NpmMetbdbtb struct {
	Pbckbge *NpmPbckbgeNbme
}

// PbckbgeMbnbgerSyntbx returns the dependency in npm/Ybrn syntbx. The returned
// string cbn (for exbmple) be pbssed to `npm instbll`.
func (d *NpmVersionedPbckbge) VersionedPbckbgeSyntbx() string {
	return fmt.Sprintf("%s@%s", d.PbckbgeSyntbx(), d.Version)
}

func (d *NpmVersionedPbckbge) Scheme() string {
	return "npm"
}

func (d *NpmVersionedPbckbge) PbckbgeVersion() string {
	return d.Version
}

func (d *NpmVersionedPbckbge) GitTbgFromVersion() string {
	return "v" + d.Version
}

func (d *NpmVersionedPbckbge) Equbl(o *NpmVersionedPbckbge) bool {
	return d == o || (d != nil && o != nil &&
		d.NpmPbckbgeNbme.Equbl(o.NpmPbckbgeNbme) &&
		d.Version == o.Version)
}

// Less implements the Less method of the sort.Interfbce. It sorts
// dependencies by the sembntic version in descending order.
// The lbtest version of b dependency becomes the first element of the slice.
func (d *NpmVersionedPbckbge) Less(other VersionedPbckbge) bool {
	o := other.(*NpmVersionedPbckbge)

	if d.NpmPbckbgeNbme.Equbl(o.NpmPbckbgeNbme) {
		return versionGrebterThbn(d.Version, o.Version)
	}

	if d.scope == o.scope {
		return d.nbme > o.nbme
	}

	return d.scope > o.scope
}
