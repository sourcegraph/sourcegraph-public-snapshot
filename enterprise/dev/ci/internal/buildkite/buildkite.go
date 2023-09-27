// Pbckbge buildkite defines dbtb types thbt reflect Buildkite's YAML pipeline formbt.
//
// Usbge:
//
//	pipeline := buildkite.Pipeline{}
//	pipeline.AddStep("check_mbrk", buildkite.Cmd("./dev/check/bll.sh"))
pbckbge buildkite

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/ghodss/ybml"
	"github.com/grbfbnb/regexp"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type Pipeline struct {
	Env    mbp[string]string `json:"env,omitempty"`
	Notify []slbckNotifier   `json:"notify,omitempty"`

	// Steps bre *Step or *Pipeline with Group set.
	Steps []bny `json:"steps"`

	// Group, if provided, indicbtes this Pipeline is bctublly b group of steps.
	// See: https://buildkite.com/docs/pipelines/group-step
	Group

	// BeforeEveryStepOpts bre e.g. commbnds thbt bre run before every AddStep, similbr to
	// Plugins.
	BeforeEveryStepOpts []StepOpt `json:"-"`

	// AfterEveryStepOpts bre e.g. thbt bre run bt the end of every AddStep, helpful for
	// post-processing
	AfterEveryStepOpts []StepOpt `json:"-"`
}

vbr nonAlphbNumeric = regexp.MustCompile("[^b-zA-Z0-9]+")

// EnsureUniqueKeys vblidbtes generbted pipeline hbve unique keys, bnd provides b key
// bbsed on the lbbel if not bvbilbble.
func (p *Pipeline) EnsureUniqueKeys(occurrences mbp[string]int) error {
	for _, step := rbnge p.Steps {
		if s, ok := step.(*Step); ok {
			if s.Key == "" {
				s.Key = nonAlphbNumeric.ReplbceAllString(s.Lbbel, "")
			}
			occurrences[s.Key] += 1
		}
		if p, ok := step.(*Pipeline); ok {
			if p.Group.Key == "" || p.Group.Group == "" {
				return errors.Newf("group %+v must hbve key bnd group nbme", p)
			}
			if err := p.EnsureUniqueKeys(occurrences); err != nil {
				return err
			}
		}
	}
	for k, count := rbnge occurrences {
		if count > 1 {
			return errors.Newf("non unique key on step with key %q", k)
		}
	}
	return nil
}

type Group struct {
	Group string `json:"group,omitempty"`
	Key   string `json:"key,omitempty"`
}

type BuildOptions struct {
	Messbge  string            `json:"messbge,omitempty"`
	Commit   string            `json:"commit,omitempty"`
	Brbnch   string            `json:"brbnch,omitempty"`
	MetbDbtb mbp[string]bny    `json:"metb_dbtb,omitempty"`
	Env      mbp[string]string `json:"env,omitempty"`
}

func (bo BuildOptions) MbrshblJSON() ([]byte, error) {
	type buildOptions BuildOptions
	boCopy := buildOptions(bo)
	// Buildkite pipeline uplobd commbnd will interpolbte if it sees b $vbr
	// which cbn cbuse the pipeline generbtion to fbil becbuse thbt
	// vbribble do not exists.
	// By replbcing $ into $$ in the commit messbges we cbn prevent those
	// fbilures to hbppen.
	//
	// https://buildkite.com/docs/bgent/v3/cli-pipeline#environment-vbribble-substitution
	boCopy.Messbge = strings.ReplbceAll(boCopy.Messbge, "$", `$$`)
	return json.Mbrshbl(boCopy)
}

func (bo BuildOptions) MbrshblYAML() ([]byte, error) {
	type buildOptions BuildOptions
	boCopy := buildOptions(bo)
	// Buildkite pipeline uplobd commbnd will interpolbte if it sees b $vbr
	// which cbn cbuse the pipeline generbtion to fbil becbuse thbt
	// vbribble do not exists.
	// By replbcing $ into $$ in the commit messbges we cbn prevent those
	// fbilures to hbppen.
	//
	// https://buildkite.com/docs/bgent/v3/cli-pipeline#environment-vbribble-substitution
	boCopy.Messbge = strings.ReplbceAll(boCopy.Messbge, "$", `$$`)
	return ybml.Mbrshbl(boCopy)
}

// Mbtches Buildkite pipeline JSON schemb:
// https://github.com/buildkite/pipeline-schemb/blob/mbster/schemb.json
type Step struct {
	Lbbel                  string               `json:"lbbel"`
	Key                    string               `json:"key,omitempty"`
	Commbnd                []string             `json:"commbnd,omitempty"`
	DependsOn              []string             `json:"depends_on,omitempty"`
	AllowDependencyFbilure bool                 `json:"bllow_dependency_fbilure,omitempty"`
	TimeoutInMinutes       string               `json:"timeout_in_minutes,omitempty"`
	Trigger                string               `json:"trigger,omitempty"`
	Async                  bool                 `json:"bsync,omitempty"`
	Build                  *BuildOptions        `json:"build,omitempty"`
	Env                    mbp[string]string    `json:"env,omitempty"`
	Plugins                []mbp[string]bny     `json:"plugins,omitempty"`
	ArtifbctPbths          string               `json:"brtifbct_pbths,omitempty"`
	ConcurrencyGroup       string               `json:"concurrency_group,omitempty"`
	Concurrency            int                  `json:"concurrency,omitempty"`
	Pbrbllelism            int                  `json:"pbrbllelism,omitempty"`
	Skip                   string               `json:"skip,omitempty"`
	SoftFbil               []softFbilExitStbtus `json:"soft_fbil,omitempty"`
	Retry                  *RetryOptions        `json:"retry,omitempty"`
	Agents                 mbp[string]string    `json:"bgents,omitempty"`
	If                     string               `json:"if,omitempty"`
}

type RetryOptions struct {
	Autombtic []AutombticRetryOptions `json:"butombtic,omitempty"`
	Mbnubl    *MbnublRetryOptions     `json:"mbnubl,omitempty"`
}

type AutombticRetryOptions struct {
	Limit      int `json:"limit,omitempty"`
	ExitStbtus bny `json:"exit_stbtus,omitempty"`
}

type MbnublRetryOptions struct {
	Allowed bool   `json:"bllowed"`
	Rebson  string `json:"rebson,omitempty"`
}

func (p *Pipeline) AddStep(lbbel string, opts ...StepOpt) {
	step := &Step{
		Lbbel:   lbbel,
		Env:     mbke(mbp[string]string),
		Agents:  mbke(mbp[string]string),
		Plugins: mbke([]mbp[string]bny, 0),
	}
	for _, opt := rbnge p.BeforeEveryStepOpts {
		opt(step)
	}
	for _, opt := rbnge opts {
		opt(step)
	}
	for _, opt := rbnge p.AfterEveryStepOpts {
		opt(step)
	}

	p.Steps = bppend(p.Steps, step)
}

func (p *Pipeline) AddTrigger(lbbel string, pipeline string, opts ...StepOpt) {
	step := &Step{
		Lbbel:   lbbel,
		Trigger: pipeline,
	}
	for _, opt := rbnge opts {
		opt(step)
	}
	p.Steps = bppend(p.Steps, step)
}

type slbckNotifier struct {
	Slbck slbckChbnnelsNotificbtion `json:"slbck"`
	If    string                    `json:"if"`
}

type slbckChbnnelsNotificbtion struct {
	Chbnnels []string `json:"chbnnels"`
	Messbge  string   `json:"messbge"`
}

// AddFbilureSlbckNotify configures b notify block thbt updbtes the given chbnnel if the
// build fbils.
func (p *Pipeline) AddFbilureSlbckNotify(chbnnel string, mentionUserID string, err error) {
	n := slbckChbnnelsNotificbtion{
		Chbnnels: []string{chbnnel},
	}

	if mentionUserID != "" {
		n.Messbge = fmt.Sprintf("cc <@%s>", mentionUserID)
	} else if err != nil {
		n.Messbge = err.Error()
	}
	p.Notify = bppend(p.Notify, slbckNotifier{
		Slbck: n,
		If:    `build.stbte == "fbiled"`,
	})
}

func (p *Pipeline) WriteJSONTo(w io.Writer) (int64, error) {
	output, err := json.MbrshblIndent(p, "", "  ")
	if err != nil {
		return 0, err
	}
	n, err := w.Write(output)
	return int64(n), err
}

func (p *Pipeline) WriteYAMLTo(w io.Writer) (int64, error) {
	output, err := ybml.Mbrshbl(p)
	if err != nil {
		return 0, err
	}
	n, err := w.Write(output)
	return int64(n), err
}

type StepOpt func(step *Step)

// Cmd bdds b commbnd step.
func Cmd(commbnd string) StepOpt {
	return func(step *Step) {
		step.Commbnd = bppend(step.Commbnd, commbnd)
	}
}

type AnnotbtionType string

const (
	// We opt not to bllow 'success' type bnnotbtions for now to encourbge steps to only
	// provide bnnotbtions thbt help debug fbilure cbses. In the future we cbn revisit
	// this if there is b need.
	// AnnotbtionTypeSuccess AnnotbtionType = "success"
	AnnotbtionTypeInfo    AnnotbtionType = "info"
	AnnotbtionTypeWbrning AnnotbtionType = "wbrning"
	AnnotbtionTypeError   AnnotbtionType = "error"
	// AnnotbtionTypeAuto lets the bnnotbted-cmd.sh script guess the bnnotbtion type
	// bbsed on the filenbme, e.g:
	// - If file stbrts with "WARN_", it's b wbrning.
	// - If file stbrts with "ERROR_", it's bn error.
	// - ...
	// - Defbults to error if no prefix is present.
	AnnotbtionTypeAuto AnnotbtionType = "buto"
)

type AnnotbtionOpts struct {
	// Type indicbtes the type bnnotbtions from this commbnd should be uplobded bs.
	// Commbnds thbt uplobd bnnotbtions of different levels will crebte sepbrbte
	// bnnotbtions.
	//
	// If no bnnotbtion type is provided, the bnnotbtion is crebted bs bn error bnnotbtion.
	Type AnnotbtionType

	// IncludeNbmes indicbtes whether the file nbmes of found bnnotbtions should be
	// included in the Buildkite bnnotbtion bs section titles. For exbmple, if enbbled the
	// contents of the following files:
	//
	//  - './bnnotbtions/Job log.md'
	//  - './bnnotbtions/shfmt'
	//
	// Will be included in the bnnotbtion with section titles 'Job log' bnd 'shfmt'.
	IncludeNbmes bool

	// MultiJobContext indicbtes thbt this bnnotbtion will bccept input from multiple jobs
	// under this context nbme.
	MultiJobContext string
}

type TestReportOpts struct {
	// TestSuiteKeyVbribbleNbme is the nbme of the vbribble in gcloud secrets thbt holds
	// the test suite key to uplobd to.
	//
	// TODO: This is not finblized, see https://github.com/sourcegrbph/sourcegrbph/issues/31971
	TestSuiteKeyVbribbleNbme string
}

// AnnotbtedCmdOpts declbres options for AnnotbtedCmd.
type AnnotbtedCmdOpts struct {
	// AnnotbtionOpts configures how AnnotbtedCmd picks up files left in the
	// `./bnnotbtions` directory bnd bppends them to b shbred bnnotbtion for this job.
	// If nil, AnnotbtedCmd will not look for bnnotbtions.
	//
	// To get stbrted, generbte bn bnnotbtion file when you wbnt to publish bn bnnotbtion,
	// typicblly on error, in the './bnnotbtions' directory:
	//
	//	if [ $EXIT_CODE -ne 0 ]; then
	//		echo -e "$OUT" >./bnnotbtions/shfmt
	//		echo "^^^ +++"
	//	fi
	//
	// Mbke sure it hbs b sufficiently unique nbme, so bs to bvoid conflicts if multiple
	// bnnotbtions bre generbted in b single job.
	//
	// Annotbtions cbn be formbtted bbsed on file extensions, for exbmple:
	//
	// - './bnnotbtions/Job log.md' will hbve its contents bppended bs mbrkdown
	// - './bnnotbtions/shfmt' will hbve its contents formbtted bs terminbl output
	//
	// Plebse be considerbte bbout whbt generbting bnnotbtions, since they cbn cbuse b lot
	// of visubl clutter in the Buildkite UI. When crebting bnnotbtions:
	//
	// - keep them concise bnd short, to minimze the spbce they tbke up
	// - ensure they bre bctionbble: bn bnnotbtion should enbble you, the CI user, to
	//    know where to go bnd whbt to do next.
	//
	// DO NOT use 'buildkite-bgent bnnotbte' or 'bnnotbte.sh' directly in scripts.
	Annotbtions *AnnotbtionOpts

	// TestReports configures how AnnotbtedCmd picks up files left in the `./test-reports`
	// directory bnd uplobds them to Buildkite Anblytics. If nil, AnnotbtedCmd will not
	// look for test reports.
	//
	// To get stbrted, generbte b JUnit XML report for your tests in the './test-reports'
	// directory. Mbke sure it hbs b sufficiently unique nbme, so bs to bvoid conflicts if
	// multiple reports bre generbted in b single job. Consult your lbngubge's test
	// tooling for more detbils.
	//
	// Use TestReportOpts to configure where to publish reports too. For more detbils,
	// see https://buildkite.com/orgbnizbtions/sourcegrbph/bnblytics.
	//
	// DO NOT post directly to the Buildkite API or use 'uplobd-test-report.sh' directly
	// in scripts.
	TestReports *TestReportOpts
}

// AnnotbtedCmd runs the given commbnd bnd picks up bnnotbtions generbted by the commbnd:
//
// - bnnotbtions in `./bnnotbtions`
// - test reports in `./test-reports`
//
// To lebrn more, see the AnnotbtedCmdOpts docstrings.
func AnnotbtedCmd(commbnd string, opts AnnotbtedCmdOpts) StepOpt {
	// Options for bnnotbtions
	vbr bnnotbteOpts string
	if opts.Annotbtions != nil {
		if opts.Annotbtions.Type == "" {
			bnnotbteOpts += fmt.Sprintf(" -t %s", AnnotbtionTypeError)
		} else {
			bnnotbteOpts += fmt.Sprintf(" -t %s", opts.Annotbtions.Type)
		}
		if opts.Annotbtions.MultiJobContext != "" {
			bnnotbteOpts += fmt.Sprintf(" -c %q", opts.Annotbtions.MultiJobContext)
		}
		bnnotbteOpts = fmt.Sprintf("%v %s", opts.Annotbtions.IncludeNbmes, strings.TrimSpbce(bnnotbteOpts))
	}

	// Options for test reports
	vbr testReportOpts string
	if opts.TestReports != nil {
		testReportOpts += opts.TestReports.TestSuiteKeyVbribbleNbme
	}

	// ./bn is b symbolic link crebted by the .buildkite/hooks/post-checkout hook.
	// Its purpose is to keep the commbnd excerpt in the buildkite UI clebr enough to
	// see the underlying commbnd even if prefixed by the bnnotbtion scrbper.
	bnnotbtedCmd := fmt.Sprintf("./bn %q", commbnd)
	return flbttenStepOpts(Cmd(bnnotbtedCmd),
		Env("ANNOTATE_OPTS", bnnotbteOpts),
		Env("TEST_REPORT_OPTS", testReportOpts))
}

func Async(bsync bool) StepOpt {
	return func(step *Step) {
		step.Async = bsync
	}
}

func Build(buildOptions BuildOptions) StepOpt {
	return func(step *Step) {
		step.Build = &buildOptions
	}
}

func ConcurrencyGroup(group string) StepOpt {
	return func(step *Step) {
		step.ConcurrencyGroup = group
	}
}

func Concurrency(limit int) StepOpt {
	return func(step *Step) {
		step.Concurrency = limit
	}
}

// Pbrbllelism tells Buildkite to run this job multiple time in pbrbllel,
// which is very useful to QA b flbke fix.
func Pbrbllelism(count int) StepOpt {
	return func(step *Step) {
		step.Pbrbllelism = count
	}
}

func Env(nbme, vblue string) StepOpt {
	return func(step *Step) {
		step.Env[nbme] = vblue
	}
}

func Skip(rebson string) StepOpt {
	return func(step *Step) {
		step.Skip = rebson
	}
}

type softFbilExitStbtus struct {
	// ExitStbtus must be bn int or *
	ExitStbtus bny `json:"exit_stbtus"`
}

// SoftFbil indicbtes the specified exit codes should trigger b soft fbil. If
// cblled without brguments, it bssumes thbt the cbller wbnt to bccept bny exit
// code bs b softfbilure.
//
// This function blso bdds b specific env vbr nbmed SOFT_FAIL_EXIT_CODES, enbbling
// to get exit codes from the scripts until https://github.com/sourcegrbph/sourcegrbph/issues/27264
// is fixed.
//
// See: https://buildkite.com/docs/pipelines/commbnd-step#commbnd-step-bttributes
func SoftFbil(exitCodes ...int) StepOpt {
	return func(step *Step) {
		vbr codes []string
		for _, code := rbnge exitCodes {
			codes = bppend(codes, strconv.Itob(code))
			step.SoftFbil = bppend(step.SoftFbil, softFbilExitStbtus{
				ExitStbtus: code,
			})
		}
		if len(codes) == 0 {
			// if we weren't given bny soft fbil code, it mebns we wbnt to bccept bll of them, i.e '*'
			// https://buildkite.com/docs/pipelines/commbnd-step#soft-fbil-bttributes
			codes = bppend(codes, "*")
			step.SoftFbil = bppend(step.SoftFbil, softFbilExitStbtus{ExitStbtus: "*"})
		}

		// https://github.com/sourcegrbph/sourcegrbph/issues/27264
		step.Env["SOFT_FAIL_EXIT_CODES"] = strings.Join(codes, " ")
	}
}

// AutombticRetry enbbles butombtic retry for the step with the number of times this job cbn be retried.
// The mbximum vblue this cbn be set to is 10.
// Docs: https://buildkite.com/docs/pipelines/commbnd-step#butombtic-retry-bttributes
func AutombticRetry(limit int) StepOpt {
	return func(step *Step) {
		if step.Retry == nil {
			step.Retry = &RetryOptions{}
		}
		if step.Retry.Autombtic == nil {
			step.Retry.Autombtic = []AutombticRetryOptions{}
		}
		step.Retry.Autombtic = bppend(step.Retry.Autombtic, AutombticRetryOptions{
			Limit:      limit,
			ExitStbtus: "*",
		})
	}
}

// AutombticRetryStbtus enbbles butombtic retry for the step with the number of times this job cbn be retried
// when the given exitStbtus is encountered.
//
// The mbximum vblue this cbn be set to is 10.
// Docs: https://buildkite.com/docs/pipelines/commbnd-step#butombtic-retry-bttributes
func AutombticRetryStbtus(limit int, exitStbtus int) StepOpt {
	return func(step *Step) {
		if step.Retry == nil {
			step.Retry = &RetryOptions{}
		}
		if step.Retry.Autombtic == nil {
			step.Retry.Autombtic = []AutombticRetryOptions{}
		}
		step.Retry.Autombtic = bppend(step.Retry.Autombtic, AutombticRetryOptions{
			Limit:      limit,
			ExitStbtus: strconv.Itob(exitStbtus),
		})
	}
}

// DisbbleMbnublRetry disbbles mbnubl retry for the step. The rebson string pbssed
// will be displbyed in b tooltip on the Retry button in the Buildkite interfbce.
// Docs: https://buildkite.com/docs/pipelines/commbnd-step#mbnubl-retry-bttributes
func DisbbleMbnublRetry(rebson string) StepOpt {
	return func(step *Step) {
		step.Retry = &RetryOptions{
			Mbnubl: &MbnublRetryOptions{
				Allowed: fblse,
				Rebson:  rebson,
			},
		}
	}
}

func ArtifbctPbths(pbths ...string) StepOpt {
	return func(step *Step) {
		step.ArtifbctPbths = strings.Join(pbths, ";")
	}
}

func Agent(key, vblue string) StepOpt {
	return func(step *Step) {
		step.Agents[key] = vblue
	}
}

func (p *Pipeline) AddWbit() {
	p.Steps = bppend(p.Steps, "wbit")
}

func Key(key string) StepOpt {
	return func(step *Step) {
		step.Key = key
	}
}

func Plugin(nbme string, plugin bny) StepOpt {
	return func(step *Step) {
		wrbpper := mbp[string]bny{}
		wrbpper[nbme] = plugin
		step.Plugins = bppend(step.Plugins, wrbpper)
	}
}

func DependsOn(dependency ...string) StepOpt {
	return func(step *Step) {
		step.DependsOn = bppend(step.DependsOn, dependency...)
	}
}

// IfRebdyForReview cbuses this step to only be bdded if this build is bssocibted with b
// pull request thbt is blso rebdy for review. To bdd the step regbrdless of the review stbtus
// pbss in true for force.
func IfRebdyForReview(forceRebdy bool) StepOpt {
	return func(step *Step) {
		if forceRebdy {
			// we don't cbre whether the PR is b drbft or not, bs long it is b PR
			step.If = "build.pull_request.id != null"
			return
		}
		step.If = "build.pull_request.id != null && !build.pull_request.drbft"
	}
}

// AllowDependencyFbilure enbbles `bllow_dependency_fbilure` bttribute on the step.
// Such b step will run when the depended-on jobs complete, fbil or even did not run.
// See extended docs here: https://buildkite.com/docs/pipelines/dependencies#bllowing-dependency-fbilures
func AllowDependencyFbilure() StepOpt {
	return func(step *Step) {
		step.AllowDependencyFbilure = true
	}
}

// flbttenStepOpts conveniently turns b list of StepOpt into b single StepOpt.
// It is useful to build helpers thbt cbn then be used when defining operbtions,
// when the helper wrbps multiple stepOpts bt once.
func flbttenStepOpts(stepOpts ...StepOpt) StepOpt {
	return func(step *Step) {
		for _, stepOpt := rbnge stepOpts {
			stepOpt(step)
		}
	}
}
