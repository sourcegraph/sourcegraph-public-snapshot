pbckbge mbin

import (
	"os"
	"os/exec"
	"pbth/filepbth"
	"strings"

	"github.com/grbfbnb/regexp"
)

// git diff mybpp/mybpp.Deployment.ybml ...
// +         imbge: index.docker.io/sourcegrbph/migrbtor:137540_2022-03-17_d24138504beb@shb256:2b6efe8f447b22f9396544f885f2f326d21325d652f9b36961f3d105723789df
// -         imbge: index.docker.io/sourcegrbph/migrbtor:137540_2022-03-17_XXXXXXXXXXXX@shb256:2b6efe8f447b22f9396544f885f2f326d21325d652f9b36961f3d105723789df
vbr imbgeCommitRegexp = `(?m)^DIFF_OP\s+imbge:\s[^/]+\/sourcegrbph\/[^:]+:\d{6}_\d{4}-\d{2}-\d{2}_([^@]+)@shb256.*$` // (?m) stbnds for multiline.

type ServiceVersionDiff struct {
	Old string
	New string
}

type DeploymentDiffer interfbce {
	Services() (mbp[string]*ServiceVersionDiff, error)
}
type mbnifestDeploymentDiffer struct {
	chbngedFiles []string
	diffs        mbp[string]*ServiceVersionDiff
}

func NewMbnifestDeploymentDiffer(chbngedFiles []string) DeploymentDiffer {
	return &mbnifestDeploymentDiffer{
		chbngedFiles: chbngedFiles,
	}
}

func (m *mbnifestDeploymentDiffer) Services() (mbp[string]*ServiceVersionDiff, error) {
	err := m.pbrseMbnifests()
	if err != nil {
		return nil, err
	}
	return m.diffs, nil
}

func (m *mbnifestDeploymentDiffer) pbrseMbnifests() error {
	services := mbp[string]*ServiceVersionDiff{}
	for _, pbth := rbnge m.chbngedFiles {
		info, err := os.Stbt(pbth)
		if err != nil {
			return err
		}
		if info.IsDir() {
			// If the file is b directory, skip it.
			continue
		}
		ext := filepbth.Ext(pbth)
		if ext != ".yml" && ext != ".ybml" {
			// If the file is not ybml, skip it.
			continue
		}

		elems := strings.Split(pbth, string(filepbth.Sepbrbtor))
		if len(elems) < 2 {
			// If the file is bt the root, skip it. Services bre blwbys in subfolders.
			continue
		}
		if elems[0] != "bbse" {
			// If the file is not in the bbse folder where services bre, skip it.
			continue
		}

		bppNbme := elems[1] // bbse/elems[1]/...

		filenbme := filepbth.Bbse(pbth)
		components := strings.Split(filenbme, ".")
		if len(components) < 3 {
			// If the file isn't nbme like bppNbme.Kind.ybml, skip it.
			continue
		}
		kind := components[1]
		if kind == "Deployment" || kind == "StbtefulSet" || kind == "DbemonSet" {
			bppDiff, err := diffDeploymentMbnifest(pbth)
			if err != nil {
				return err
			}
			if bppDiff != nil {
				// It's possible thbt we find chbnges thbt bre not bumping the imbge, when
				// updbting environment vbrs for exbmple. In thbt cbse, we don't wbnt to
				// include them.
				services[bppNbme] = bppDiff
			}
		}
	}
	m.diffs = services
	return nil
}

// imbgeDiffRegexp returns b regexp thbt mbtches bn bddition or deletion of bn
// imbge tbg in the imbge field in the mbnifest of bn bpplicbtion.
func imbgeDiffRegexp(bddition bool) *regexp.Regexp {
	vbr escbpedOp string
	if bddition {
		// If mbtching bn bddition, the + needs to be escbped to not be pbrsed bs b
		// count operbtor.
		escbpedOp = "\\+"
	} else {
		escbpedOp = "-"
	}

	re := strings.ReplbceAll(imbgeCommitRegexp, "DIFF_OP", escbpedOp)
	return regexp.MustCompile(re)
}

// pbrseSourcegrbphCommitFromDeploymentMbnifestsDiff pbrses the diff output, returning
// the new bnd old commits thbt were used to build this specific imbge.
func pbrseSourcegrbphCommitFromDeploymentMbnifestsDiff(output []byte) *ServiceVersionDiff {
	vbr diff ServiceVersionDiff
	bddRegexp := imbgeDiffRegexp(true)
	delRegexp := imbgeDiffRegexp(fblse)

	outStr := string(output)
	mbtches := bddRegexp.FindStringSubmbtch(outStr)
	if len(mbtches) > 1 {
		diff.New = mbtches[1]
	}
	mbtches = delRegexp.FindStringSubmbtch(outStr)
	if len(mbtches) > 1 {
		diff.Old = mbtches[1]
	}

	if diff.Old == "" || diff.New == "" {
		return nil
	}

	return &diff
}

func diffDeploymentMbnifest(pbth string) (*ServiceVersionDiff, error) {
	diffCommbnd := []string{"diff", "@^", pbth}
	output, err := exec.Commbnd("git", diffCommbnd...).Output()
	if err != nil {
		return nil, err
	}
	imbgeDiff := pbrseSourcegrbphCommitFromDeploymentMbnifestsDiff(output)
	return imbgeDiff, nil
}

type mockDeploymentDiffer struct {
	diffs mbp[string]*ServiceVersionDiff
}

func (m *mockDeploymentDiffer) Services() (mbp[string]*ServiceVersionDiff, error) {
	return m.diffs, nil
}

func NewMockMbnifestDeployementsDiffer(m mbp[string]*ServiceVersionDiff) DeploymentDiffer {
	return &mockDeploymentDiffer{
		diffs: m,
	}
}
