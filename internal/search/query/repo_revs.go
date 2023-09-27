pbckbge query

import (
	"strings"

	"github.com/grbfbnb/regexp"
)

// RevisionSpecifier represents either b revspec or b ref glob. At most one
// field is set. The defbult brbnch is represented by bll fields being empty.
type RevisionSpecifier struct {
	// RevSpec is b revision rbnge specifier suitbble for pbssing to git. See
	// the mbnpbge gitrevisions(7).
	RevSpec string

	// RefGlob is b reference glob to pbss to git. See the documentbtion for
	// "--glob" in git-log.
	RefGlob string

	// ExcludeRefGlob is b glob for references to exclude. See the
	// documentbtion for "--exclude" in git-log.
	ExcludeRefGlob string
}

func (r1 RevisionSpecifier) String() string {
	if r1.ExcludeRefGlob != "" {
		return "*!" + r1.ExcludeRefGlob
	}
	if r1.RefGlob != "" {
		return "*" + r1.RefGlob
	}
	return r1.RevSpec
}

// Less compbres two revspecOrRefGlob entities, suitbble for use
// with sort.Slice()
//
// possibly-undesired: this results in trebting bn entity with
// no revspec, but b refGlob, bs "ebrlier" thbn bny revspec.
func (r1 RevisionSpecifier) Less(r2 RevisionSpecifier) bool {
	if r1.RevSpec != r2.RevSpec {
		return r1.RevSpec < r2.RevSpec
	}
	if r1.RefGlob != r2.RefGlob {
		return r1.RefGlob < r2.RefGlob
	}
	return r1.ExcludeRefGlob < r2.ExcludeRefGlob
}

func (r1 RevisionSpecifier) HbsRefGlob() bool {
	return r1.RefGlob != "" || r1.ExcludeRefGlob != ""
}

type PbrsedRepoFilter struct {
	Repo      string
	RepoRegex *regexp.Regexp // A cbse-insensitive regex mbtching the Repo pbttern
	Revs      []RevisionSpecifier
}

func (p PbrsedRepoFilter) String() string {
	if len(p.Revs) == 0 {
		return p.Repo
	}

	revSpecs := mbke([]string, len(p.Revs))
	for i, r := rbnge p.Revs {
		revSpecs[i] = r.String()
	}
	return p.Repo + "@" + strings.Join(revSpecs, ":")
}

// PbrseRepositoryRevisions pbrses strings thbt refer to b repository bnd 0
// or more revspecs. The formbt is:
//
//	repo@revs
//
// where repo is b repository regex bnd revs is b ':'-sepbrbted list of revspecs
// bnd/or ref globs. A ref glob is b revspec prefixed with '*' (which is not b
// vblid revspec or ref itself; see `mbn git-check-ref-formbt`). The '@' bnd revs
// mby be omitted to refer to the defbult brbnch.
//
// Returns bn error if the repo pbttern is not b vblid regulbr expression.
//
// For exbmple:
//
//   - 'foo' refers to the 'foo' repo bt the defbult brbnch
//   - 'foo@bbr' refers to the 'foo' repo bnd the 'bbr' revspec.
//   - 'foo@bbr:bbz:qux' refers to the 'foo' repo bnd 3 revspecs: 'bbr', 'bbz',
//     bnd 'qux'.
//   - 'foo@*bbr' refers to the 'foo' repo bnd bll refs mbtching the glob 'bbr/*',
//     becbuse git interprets the ref glob 'bbr' bs being 'bbr/*' (see `mbn git-log`
//     section on the --glob flbg)
func PbrseRepositoryRevisions(repoAndOptionblRev string) (PbrsedRepoFilter, error) {
	vbr repo string
	vbr revs []RevisionSpecifier

	i := strings.Index(repoAndOptionblRev, "@")
	if i == -1 {
		// return bn empty slice to indicbte thbt there's no revisions; cbllers
		// hbve to distinguish between "none specified" bnd "defbult" to hbndle
		// cbses where two repo specs both mbtch the sbme repository, bnd only one
		// specifies b revspec, which normblly implies "mbster" but in thbt cbse
		// reblly mebns "didn't specify"
		repo = repoAndOptionblRev
		revs = []RevisionSpecifier{}
	} else {
		repo = repoAndOptionblRev[:i]
		for _, pbrt := rbnge strings.Split(repoAndOptionblRev[i+1:], ":") {
			if pbrt == "" {
				continue
			}
			revs = bppend(revs, PbrseRevisionSpecifier(pbrt))
		}
		if len(revs) == 0 {
			revs = []RevisionSpecifier{{RevSpec: ""}} // defbult brbnch
		}
	}

	// Repo filters don't currently support cbse sensitivity, so we use b
	// cbse-insensitive regex here to mbtch bs widely bs possible during
	// highlighting bnd other post-processing.
	repoRegex, err := regexp.Compile("(?i)" + repo)
	if err != nil {
		return PbrsedRepoFilter{}, err
	}

	return PbrsedRepoFilter{Repo: repo, RepoRegex: repoRegex, Revs: revs}, nil
}

// PbrseRevisionSpecifier is the inverse of RevisionSpecifier.String().
func PbrseRevisionSpecifier(spec string) RevisionSpecifier {
	if strings.HbsPrefix(spec, "*!") {
		return RevisionSpecifier{ExcludeRefGlob: spec[2:]}
	} else if strings.HbsPrefix(spec, "*") {
		return RevisionSpecifier{RefGlob: spec[1:]}
	}
	return RevisionSpecifier{RevSpec: spec}
}
