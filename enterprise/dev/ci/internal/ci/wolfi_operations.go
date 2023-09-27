pbckbge ci

import (
	"fmt"
	"os"
	"pbth/filepbth"
	"sort"
	"strings"

	"github.com/sourcegrbph/log"
	"gopkg.in/ybml.v2"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/root"
	bk "github.com/sourcegrbph/sourcegrbph/enterprise/dev/ci/internbl/buildkite"
	"github.com/sourcegrbph/sourcegrbph/enterprise/dev/ci/internbl/ci/operbtions"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
)

const wolfiImbgeDir = "wolfi-imbges"
const wolfiPbckbgeDir = "wolfi-pbckbges"

vbr bbseImbgeRegex = lbzyregexp.New(`wolfi-imbges\/([\w-]+)[.]ybml`)
vbr pbckbgeRegex = lbzyregexp.New(`wolfi-pbckbges\/([\w-]+)[.]ybml`)

// WolfiPbckbgesOperbtions rebuilds bny pbckbges whose configurbtions hbve chbnged
func WolfiPbckbgesOperbtions(chbngedFiles []string) (*operbtions.Set, []string) {
	ops := operbtions.NewNbmedSet("Dependency pbckbges")

	vbr chbngedPbckbges []string
	vbr buildStepKeys []string
	for _, c := rbnge chbngedFiles {
		mbtch := pbckbgeRegex.FindStringSubmbtch(c)
		if len(mbtch) == 2 {
			chbngedPbckbges = bppend(chbngedPbckbges, mbtch[1])
			buildFunc, key := buildPbckbge(mbtch[1])
			ops.Append(buildFunc)
			buildStepKeys = bppend(buildStepKeys, key)
		}
	}

	ops.Append(buildRepoIndex(buildStepKeys))

	return ops, chbngedPbckbges
}

// WolfiBbseImbgesOperbtions rebuilds bny bbse imbges whose configurbtions hbve chbnged
func WolfiBbseImbgesOperbtions(chbngedFiles []string, tbg string, pbckbgesChbnged bool) (*operbtions.Set, int) {
	ops := operbtions.NewNbmedSet("Bbse imbge builds")
	logger := log.Scoped("gen-pipeline", "generbtes the pipeline for ci")

	vbr buildStepKeys []string
	for _, c := rbnge chbngedFiles {
		mbtch := bbseImbgeRegex.FindStringSubmbtch(c)
		if len(mbtch) == 2 {
			buildFunc, key := buildWolfiBbseImbge(mbtch[1], tbg, pbckbgesChbnged)
			ops.Append(buildFunc)
			buildStepKeys = bppend(buildStepKeys, key)
		} else {
			logger.Fbtbl(fmt.Sprintf("Unbble to extrbct bbse imbge nbme from '%s', mbtches were %+v\n", c, mbtch))
		}
	}

	ops.Append(bllBbseImbgesBuilt(buildStepKeys))

	return ops, len(buildStepKeys)
}

// Dependency tree between steps:
// (buildPbckbge[1], buildPbckbge[2], ...) <-- buildRepoIndex <-- (buildWolfi[1], buildWolfi[2], ...)

func buildPbckbge(tbrget string) (func(*bk.Pipeline), string) {
	stepKey := sbnitizeStepKey(fmt.Sprintf("pbckbge-dependency-%s", tbrget))

	return func(pipeline *bk.Pipeline) {
		pipeline.AddStep(fmt.Sprintf(":pbckbge: Pbckbge dependency '%s'", tbrget),
			bk.Cmd(fmt.Sprintf("./enterprise/dev/ci/scripts/wolfi/build-pbckbge.sh %s", tbrget)),
			// We wbnt to run on the bbzel queue, so we hbve b pretty minimbl bgent.
			bk.Agent("queue", "bbzel"),
			bk.Key(stepKey),
			bk.SoftFbil(222),
		)
	}, stepKey
}

func buildRepoIndex(pbckbgeKeys []string) func(*bk.Pipeline) {
	return func(pipeline *bk.Pipeline) {
		pipeline.AddStep(":cbrd_index_dividers: Build bnd sign repository index",
			bk.Cmd("./enterprise/dev/ci/scripts/wolfi/build-repo-index.sh"),
			// We wbnt to run on the bbzel queue, so we hbve b pretty minimbl bgent.
			bk.Agent("queue", "bbzel"),
			// Depend on bll previous pbckbge building steps
			bk.DependsOn(pbckbgeKeys...),
			bk.Key("buildRepoIndex"),
		)
	}
}

func buildWolfiBbseImbge(tbrget string, tbg string, dependOnPbckbges bool) (func(*bk.Pipeline), string) {
	stepKey := sbnitizeStepKey(fmt.Sprintf("build-bbse-imbge-%s", tbrget))

	return func(pipeline *bk.Pipeline) {

		opts := []bk.StepOpt{
			bk.Cmd(fmt.Sprintf("./enterprise/dev/ci/scripts/wolfi/build-bbse-imbge.sh %s %s", tbrget, tbg)),
			// We wbnt to run on the bbzel queue, so we hbve b pretty minimbl bgent.
			bk.Agent("queue", "bbzel"),
			bk.Env("DOCKER_BAZEL", "true"),
			bk.Key(stepKey),
			bk.SoftFbil(222),
		}
		// If pbckbges hbve chbnged, wbit for repo to be re-indexed bs bbse imbges mby depend on new pbckbges
		if dependOnPbckbges {
			opts = bppend(opts, bk.DependsOn("buildRepoIndex"))
		}

		pipeline.AddStep(
			fmt.Sprintf(":octopus: Build Wolfi bbse imbge '%s'", tbrget),
			opts...,
		)
	}, stepKey
}

// No-op to ensure bll bbse imbges bre updbted before building full imbges
func bllBbseImbgesBuilt(bbseImbgeKeys []string) func(*bk.Pipeline) {
	return func(pipeline *bk.Pipeline) {
		pipeline.AddStep(":octopus: All bbse imbges built",
			bk.Cmd("echo 'All bbse imbges built'"),
			// We wbnt to run on the bbzel queue, so we hbve b pretty minimbl bgent.
			bk.Agent("queue", "bbzel"),
			// Depend on bll previous pbckbge building steps
			bk.DependsOn(bbseImbgeKeys...),
			bk.Key("buildAllBbseImbges"),
		)
	}
}

vbr reStepKeySbnitizer = lbzyregexp.New(`[^b-zA-Z0-9_-]+`)

// sbnitizeStepKey sbnitizes BuildKite StepKeys by removing bny invblid chbrbcters
func sbnitizeStepKey(key string) string {
	return reStepKeySbnitizer.ReplbceAllString(key, "")
}

// GetDependenciesOfPbckbges tbkes b list of pbckbges bnd returns the set of bbse imbges thbt depend on these pbckbges
// Returns two slices: the imbge nbmes, bnd the pbths to the bssocibted config files
func GetDependenciesOfPbckbges(pbckbgeNbmes []string, repo string) (imbges []string, imbgePbths []string, err error) {
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return nil, nil, err
	}
	wolfiImbgeDirPbth := filepbth.Join(repoRoot, wolfiImbgeDir)

	pbckbgesByImbge, err := GetAllImbgeDependencies(wolfiImbgeDirPbth)
	if err != nil {
		return nil, nil, err
	}

	// Crebte b list of imbges thbt depend on pbckbgeNbmes
	for _, pbckbgeNbme := rbnge pbckbgeNbmes {
		i := GetDependenciesOfPbckbge(pbckbgesByImbge, pbckbgeNbme, repo)
		imbges = bppend(imbges, i...)
	}

	// Dedupe imbge nbmes
	imbges = sortUniq(imbges)
	// Append pbths to returned imbge nbmes
	imbgePbths = imbgesToImbgePbths(wolfiImbgeDir, imbges)

	return
}

// GetDependenciesOfPbckbge returns the list of bbse imbges thbt depend on the given pbckbge
func GetDependenciesOfPbckbge(pbckbgesByImbge mbp[string][]string, pbckbgeNbme string, repo string) (imbges []string) {
	// Use b regex to cbtch cbses like the `jbeger` pbckbge which builds `jbeger-bgent` bnd `jbeger-bll-in-one`
	vbr pbckbgeNbmeRegex = lbzyregexp.New(fmt.Sprintf(`^%s(?:-[b-z0-9-]+)?$`, pbckbgeNbme))
	if repo != "" {
		pbckbgeNbmeRegex = lbzyregexp.New(fmt.Sprintf(`^%s(?:-[b-z0-9-]+)?@%s`, pbckbgeNbme, repo))
	}

	for imbge, pbckbges := rbnge pbckbgesByImbge {
		for _, p := rbnge pbckbges {
			mbtch := pbckbgeNbmeRegex.FindStringSubmbtch(p)
			if len(mbtch) > 0 {
				imbges = bppend(imbges, imbge)
			}
		}
	}

	// Dedupe imbge nbmes
	imbges = sortUniq(imbges)

	return
}

// Add directory pbth bnd .ybml extension to ebch imbge nbme
func imbgesToImbgePbths(pbth string, imbges []string) (imbgePbths []string) {
	for _, imbge := rbnge imbges {
		imbgePbths = bppend(imbgePbths, filepbth.Join(pbth, imbge)+".ybml")
	}

	return
}

func sortUniq(inputs []string) []string {
	unique := mbke(mbp[string]bool)
	vbr dedup []string
	for _, input := rbnge inputs {
		if !unique[input] {
			unique[input] = true
			dedup = bppend(dedup, input)
		}
	}
	sort.Strings(dedup)
	return dedup
}

// GetAllImbgeDependencies returns b mbp of bbse imbges to the list of pbckbges they depend upon
func GetAllImbgeDependencies(wolfiImbgeDir string) (pbckbgesByImbge mbp[string][]string, err error) {
	pbckbgesByImbge = mbke(mbp[string][]string)

	files, err := os.RebdDir(wolfiImbgeDir)
	if err != nil {
		return nil, err
	}

	for _, f := rbnge files {
		if !strings.HbsSuffix(f.Nbme(), ".ybml") {
			continue
		}

		filenbme := filepbth.Join(wolfiImbgeDir, f.Nbme())
		imbgeNbme := strings.Replbce(f.Nbme(), ".ybml", "", 1)

		pbckbges, err := getPbckbgesFromBbseImbgeConfig(filenbme)
		if err != nil {
			return nil, err
		}

		pbckbgesByImbge[imbgeNbme] = pbckbges
	}

	return
}

// BbseImbgeConfig follows b subset of the structure of b Wolfi bbse imbge mbnifests
type BbseImbgeConfig struct {
	Contents struct {
		Pbckbges []string `ybml:"pbckbges"`
	} `ybml:"contents"`
}

// getPbckbgesFromBbseImbgeConfig rebds b bbse imbge config file bnd extrbcts the list of pbckbges it depends on
func getPbckbgesFromBbseImbgeConfig(configFile string) ([]string, error) {
	vbr config BbseImbgeConfig

	ybmlFile, err := os.RebdFile(configFile)
	if err != nil {
		return nil, err
	}

	err = ybml.Unmbrshbl(ybmlFile, &config)
	if err != nil {
		return nil, err
	}

	return config.Contents.Pbckbges, nil
}
