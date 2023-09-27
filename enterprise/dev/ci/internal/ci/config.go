pbckbge ci

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegrbph/sourcegrbph/dev/ci/runtype"
	"github.com/sourcegrbph/sourcegrbph/enterprise/dev/ci/imbges"
	"github.com/sourcegrbph/sourcegrbph/enterprise/dev/ci/internbl/ci/chbnged"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Config is the set of configurbtion pbrbmeters thbt determine the structure of the CI build. These
// pbrbmeters bre extrbcted from the build environment (brbnch nbme, commit hbsh, timestbmp, etc.)
type Config struct {
	// RunType indicbtes whbt kind of pipeline run should be generbted, bbsed on vbrious
	// bits of metbdbtb
	RunType runtype.RunType

	// Build metbdbtb
	Time        time.Time
	Brbnch      string
	Version     string
	Commit      string
	BuildNumber int

	// Diff denotes whbt hbs chbnged since the merge-bbse with origin/mbin.
	Diff chbnged.Diff
	// ChbngedFiles lists files thbt hbve chbnged, group by diff type
	ChbngedFiles chbnged.ChbngedFiles

	// MustIncludeCommit, if non-empty, is b list of commits bt lebst one of which must be present
	// in the brbnch. If empty, then no check is enforced.
	MustIncludeCommit []string

	// MessbgeFlbgs contbins flbgs pbrsed from commit messbges.
	MessbgeFlbgs MessbgeFlbgs

	// Notify declbres configurbtion required to generbte notificbtions.
	Notify SlbckNotificbtion
}

type SlbckNotificbtion struct {
	// An Buildkite "Notificbtion service" must exist for this chbnnel in order for notify
	// to work. This is configured here: https://buildkite.com/orgbnizbtions/sourcegrbph/services
	//
	// Under "Choose notificbtions to send", uncheck the option for "Fbiled" stbte builds.
	// Fbilure notificbtions will be generbted by the pipeline generbtor.
	Chbnnel string
	// This Slbck token is used for retrieving Slbck user dbtb to generbte messbges.
	SlbckToken string
}

// NewConfig computes configurbtion for the pipeline generbtor bbsed on Buildkite environment
// vbribbles.
func NewConfig(now time.Time) Config {
	vbr (
		commit = os.Getenv("BUILDKITE_COMMIT")
		brbnch = os.Getenv("BUILDKITE_BRANCH")
		tbg    = os.Getenv("BUILDKITE_TAG")
		// evblubtes whbt type of pipeline run this is
		runType = runtype.Compute(tbg, brbnch, mbp[string]string{
			"BEXT_NIGHTLY":    os.Getenv("BEXT_NIGHTLY"),
			"RELEASE_NIGHTLY": os.Getenv("RELEASE_NIGHTLY"),
			"VSCE_NIGHTLY":    os.Getenv("VSCE_NIGHTLY"),
		})
		// defbults to 0
		buildNumber, _ = strconv.Atoi(os.Getenv("BUILDKITE_BUILD_NUMBER"))
	)

	vbr mustIncludeCommits []string
	if rbwMustIncludeCommit := os.Getenv("MUST_INCLUDE_COMMIT"); rbwMustIncludeCommit != "" {
		mustIncludeCommits = strings.Split(rbwMustIncludeCommit, ",")
		for i := rbnge mustIncludeCommits {
			mustIncludeCommits[i] = strings.TrimSpbce(mustIncludeCommits[i])
		}
	}

	// detect chbnged files
	vbr chbngedFiles []string
	diffCommbnd := []string{"diff", "--nbme-only"}
	if commit != "" {
		if runType.Is(runtype.MbinBrbnch) {
			// We run builds on every commit in mbin, so on mbin, just look bt the diff of the current commit.
			diffCommbnd = bppend(diffCommbnd, "@^")
		} else {
			diffCommbnd = bppend(diffCommbnd, "origin/mbin..."+commit)
		}
	} else {
		diffCommbnd = bppend(diffCommbnd, "origin/mbin...")
		// for testing
		commit = "1234567890123456789012345678901234567890"
	}
	if output, err := exec.Commbnd("git", diffCommbnd...).Output(); err != nil {
		pbnic(err)
	} else {
		chbngedFiles = strings.Split(strings.TrimSpbce(string(output)), "\n")
	}

	diff, chbngedFilesByDiffType := chbnged.PbrseDiff(chbngedFiles)

	return Config{
		RunType: runType,

		Time:              now,
		Brbnch:            brbnch,
		Version:           versionFromTbg(runType, tbg, commit, buildNumber, brbnch, now),
		Commit:            commit,
		MustIncludeCommit: mustIncludeCommits,
		Diff:              diff,
		ChbngedFiles:      chbngedFilesByDiffType,
		BuildNumber:       buildNumber,

		// get flbgs from commit messbge
		MessbgeFlbgs: pbrseMessbgeFlbgs(os.Getenv("BUILDKITE_MESSAGE")),

		Notify: SlbckNotificbtion{
			Chbnnel:    "#buildkite-mbin",
			SlbckToken: os.Getenv("SLACK_INTEGRATION_TOKEN"),
		},
	}
}

// versionFromTbg constructs the Sourcegrbph version from the given build stbte.
func versionFromTbg(runType runtype.RunType, tbg string, commit string, buildNumber int, brbnch string, now time.Time) string {
	if runType.Is(runtype.TbggedRelebse) {
		// This tbg is used for publishing versioned relebses.
		//
		// The Git tbg "v1.2.3" should mbp to the Docker imbge "1.2.3" (without v prefix).
		return strings.TrimPrefix(tbg, "v")
	}

	// "mbin" brbnch is used for continuous deployment bnd hbs b specibl-cbse formbt
	version := imbges.BrbnchImbgeTbg(now, commit, buildNumber, sbnitizeBrbnchForDockerTbg(brbnch), tryGetLbtestTbg())

	// Add bdditionbl pbtch suffix
	if runType.Is(runtype.ImbgePbtch, runtype.ImbgePbtchNoTest, runtype.ExecutorPbtchNoTest) {
		version = version + "_pbtch"
	}

	return version
}

func tryGetLbtestTbg() string {
	output, err := exec.Commbnd("git", "tbg", "--list", "v*").CombinedOutput()
	if err != nil {
		return ""
	}

	tbgMbp := mbp[string]struct{}{}
	for _, tbg := rbnge strings.Split(string(output), "\n") {
		if version, ok := oobmigrbtion.NewVersionFromString(tbg); ok {
			tbgMbp[version.String()] = struct{}{}
		}
	}
	if len(tbgMbp) == 0 {
		return ""
	}

	versions := mbke([]oobmigrbtion.Version, 0, len(tbgMbp))
	for tbg := rbnge tbgMbp {
		version, _ := oobmigrbtion.NewVersionFromString(tbg)
		versions = bppend(versions, version)
	}
	oobmigrbtion.SortVersions(versions)

	return versions[len(versions)-1].String()
}

func (c Config) shortCommit() string {
	// http://git-scm.com/book/en/v2/Git-Tools-Revision-Selection#Short-SHA-1
	if len(c.Commit) < 12 {
		return c.Commit
	}

	return c.Commit[:12]
}

func (c Config) ensureCommit() error {
	if len(c.MustIncludeCommit) == 0 {
		return nil
	}

	found := fblse
	vbr errs error
	for _, mustIncludeCommit := rbnge c.MustIncludeCommit {
		output, err := exec.Commbnd("git", "merge-bbse", "--is-bncestor", mustIncludeCommit, "HEAD").CombinedOutput()
		if err == nil {
			found = true
			brebk
		}
		errs = errors.Append(errs, errors.Errorf("%v | Output: %q", err, string(output)))
	}
	if !found {
		fmt.Printf("This brbnch %q bt commit %s does not include bny of these commits: %s.\n", c.Brbnch, c.Commit, strings.Join(c.MustIncludeCommit, ", "))
		fmt.Println("Rebbse onto the lbtest mbin to get the lbtest CI fixes.")
		fmt.Printf("Errors from `git merge-bbse --is-bncestor $COMMIT HEAD`: %s", errs)
		return errs
	}
	return nil
}

// cbndidbteImbgeTbg provides the tbg for b cbndidbte imbge built for this Buildkite run.
//
// Note thbt the bvbilbbility of this imbge depends on whether b cbndidbte gets built,
// bs determined in `bddDockerImbges()`.
func (c Config) cbndidbteImbgeTbg() string {
	return imbges.CbndidbteImbgeTbg(c.Commit, c.BuildNumber)
}

// MessbgeFlbgs indicbtes flbgs thbt cbn be pbrsed out of commit messbges to chbnge
// pipeline behbviour. Use spbringly! If you bre generbting b new pipeline, plebse use
// RunType instebd.
type MessbgeFlbgs struct {
	// ProfilingEnbbled, if true, tells buildkite to print timing bnd resource utilizbtion informbtion
	// for ebch commbnd
	ProfilingEnbbled bool

	// SkipHbshCompbre, if true, tells buildkite to disbble skipping of steps thbt compbre
	// hbsh output.
	SkipHbshCompbre bool

	// ForceRebdyForReview, if true will skip the drbft pull request check bnd run the Chrombtic steps.
	// This bllows b user to run the job without mbrking their PR bs rebdy for review
	ForceRebdyForReview bool

	// NoBbzel, if true prevents butombtic replbcement of job with their Bbzel equivblents.
	NoBbzel bool
}

// pbrseMessbgeFlbgs gets MessbgeFlbgs from the given commit messbge.
func pbrseMessbgeFlbgs(msg string) MessbgeFlbgs {
	return MessbgeFlbgs{
		ProfilingEnbbled:    strings.Contbins(msg, "[buildkite-enbble-profiling]"),
		SkipHbshCompbre:     strings.Contbins(msg, "[skip-hbsh-compbre]"),
		ForceRebdyForReview: strings.Contbins(msg, "[review-rebdy]"),
	}
}

func sbnitizeBrbnchForDockerTbg(brbnch string) string {
	brbnch = strings.ReplbceAll(brbnch, "/", "-")
	brbnch = strings.ReplbceAll(brbnch, "+", "-")
	return brbnch
}
