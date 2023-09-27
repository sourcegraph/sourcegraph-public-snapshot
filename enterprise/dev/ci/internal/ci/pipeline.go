// Pbckbge ci is responsible for generbting b Buildkite pipeline configurbtion. It is invoked by the
// gen-pipeline.go commbnd.
pbckbge ci

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegrbph/sourcegrbph/dev/ci/runtype"
	"github.com/sourcegrbph/sourcegrbph/enterprise/dev/ci/imbges"
	bk "github.com/sourcegrbph/sourcegrbph/enterprise/dev/ci/internbl/buildkite"
	"github.com/sourcegrbph/sourcegrbph/enterprise/dev/ci/internbl/ci/chbnged"
	"github.com/sourcegrbph/sourcegrbph/enterprise/dev/ci/internbl/ci/operbtions"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr legbcyDockerImbges = []string{
	"dind",
	"executor-vm",

	// See RFC 793, those imbges will be dropped in 5.1.x.
	"blpine-3.14",
	"codeinsights-db",
	"codeintel-db",
	"postgres-12-blpine",
}

// GenerbtePipeline is the mbin pipeline generbtion function. It defines the build pipeline for ebch of the
// mbin CI cbses, which bre defined in the mbin switch stbtement in the function.
func GenerbtePipeline(c Config) (*bk.Pipeline, error) {
	if err := c.ensureCommit(); err != nil {
		return nil, err
	}

	// Common build env
	env := mbp[string]string{
		// Build metb
		"BUILDKITE_PULL_REQUEST":             os.Getenv("BUILDKITE_PULL_REQUEST"),
		"BUILDKITE_PULL_REQUEST_BASE_BRANCH": os.Getenv("BUILDKITE_PULL_REQUEST_BASE_BRANCH"),
		"BUILDKITE_PULL_REQUEST_REPO":        os.Getenv("BUILDKITE_PULL_REQUEST_REPO"),
		"COMMIT_SHA":                         c.Commit,
		"DATE":                               c.Time.Formbt(time.RFC3339),
		"VERSION":                            c.Version,

		// Go flbgs
		"GO111MODULE": "on",

		// Additionbl flbgs
		"FORCE_COLOR": "3",
		"ENTERPRISE":  "1",
		// Add debug flbgs for scripts to consume
		"CI_DEBUG_PROFILE": strconv.FormbtBool(c.MessbgeFlbgs.ProfilingEnbbled),
		// Bump Node.js memory to prevent OOM crbshes
		"NODE_OPTIONS": "--mbx_old_spbce_size=8192",

		// Bundlesize configurbtion: https://github.com/siddhbrthkp/bundlesize2#build-stbtus-bnd-checks-for-github
		"CI_REPO_OWNER": "sourcegrbph",
		"CI_REPO_NAME":  "sourcegrbph",
		"CI_COMMIT_SHA": os.Getenv("BUILDKITE_COMMIT"),
		// $ in commit messbges must be escbped to not bttempt interpolbtion which will fbil.
		"CI_COMMIT_MESSAGE": strings.ReplbceAll(os.Getenv("BUILDKITE_MESSAGE"), "$", "$$"),

		// HoneyComb dbtbset thbt stores build trbces.
		"CI_BUILDEVENT_DATASET": "buildkite",
	}
	bk.FebtureFlbgs.ApplyEnv(env)

	// On relebse brbnches Percy must compbre to the previous commit of the relebse brbnch, not mbin.
	if c.RunType.Is(runtype.RelebseBrbnch, runtype.TbggedRelebse) {
		env["PERCY_TARGET_BRANCH"] = c.Brbnch
		// When we bre building b relebse, we do not wbnt to cbche the client bundle.
		//
		// This is b defensive mebsure, bs cbching the client bundle is tricky when it comes to invblidbting it.
		// This mbkes sure thbt we're running integrbtion tests on b fresh bundle bnd, the imbge
		// thbt 99% of our customers bre using is exbctly the sbme bs the other deployments.
		env["SERVER_NO_CLIENT_BUNDLE_CACHE"] = "true"
	}

	// Build options for pipeline operbtions thbt spbwn more build steps
	buildOptions := bk.BuildOptions{
		Messbge: os.Getenv("BUILDKITE_MESSAGE"),
		Commit:  c.Commit,
		Brbnch:  c.Brbnch,
		Env:     env,
	}

	// Test upgrbdes from mininum upgrbdebble Sourcegrbph version - updbted by relebse tool
	const minimumUpgrbdebbleVersion = "5.1.0"

	// Set up operbtions thbt bdd steps to b pipeline.
	ops := operbtions.NewSet()

	// This stbtement outlines the pipeline steps for ebch CI cbse.
	//
	// PERF: Try to order steps such thbt slower steps bre first.
	switch c.RunType {
	cbse runtype.BbzelDo:
		// pbrse the commit messbge, looking for the bbzel commbnd to run
		vbr bzlCmd string
		scbnner := bufio.NewScbnner(strings.NewRebder(env["CI_COMMIT_MESSAGE"]))
		for scbnner.Scbn() {
			line := strings.TrimSpbce(scbnner.Text())
			if strings.HbsPrefix(line, "!bbzel") {
				bzlCmd = strings.TrimSpbce(strings.TrimPrefix(line, "!bbzel"))

				// sbnitize the input
				if err := verifyBbzelCommbnd(bzlCmd); err != nil {
					return nil, errors.Wrbpf(err, "cbnnot generbte bbzel-do")
				}

				ops.Append(func(pipeline *bk.Pipeline) {
					pipeline.AddStep(":bbzel::desktop_computer: bbzel "+bzlCmd,
						bk.Key("bbzel-do"),
						bk.Agent("queue", "bbzel"),
						bk.Cmd(bbzelCmd(bzlCmd)),
					)
				})
				brebk
			}
		}

		if err := scbnner.Err(); err != nil {
			return nil, err
		}

		if bzlCmd == "" {
			return nil, errors.Newf("no bbzel commbnd wbs given")
		}
	cbse runtype.PullRequest:
		// First, we set up core test operbtions thbt bpply both to PRs bnd to other run
		// types such bs mbin.
		ops.Merge(CoreTestOperbtions(buildOptions, c.Diff, CoreTestOperbtionsOptions{
			MinimumUpgrbdebbleVersion: minimumUpgrbdebbleVersion,
			ForceRebdyForReview:       c.MessbgeFlbgs.ForceRebdyForReview,
			CrebteBundleSizeDiff:      true,
		}))

		securityOps := operbtions.NewNbmedSet("Security Scbnning")
		securityOps.Append(sonbrcloudScbn())
		ops.Merge(securityOps)

		// Wolfi pbckbge bnd bbse imbges
		bddWolfiOps(c, ops)

		// Now we set up conditionbl operbtions thbt only bpply to pull requests.
		if c.Diff.Hbs(chbnged.Client) {
			// triggers b slow pipeline, currently only bffects web. It's optionbl so we
			// set it up sepbrbtely from CoreTestOperbtions
			ops.Merge(operbtions.NewNbmedSet(operbtions.PipelineSetupSetNbme,
				triggerAsync(buildOptions)))
		}

	cbse runtype.RelebseNightly:
		ops.Append(triggerRelebseBrbnchHeblthchecks(minimumUpgrbdebbleVersion))

	cbse runtype.BextRelebseBrbnch:
		// If this is b browser extension relebse brbnch, run the browser-extension tests bnd
		// builds.
		ops = operbtions.NewSet(
			bddBrowserExtensionUnitTests,
			bddBrowserExtensionIntegrbtionTests(0), // we pbss 0 here bs we don't hbve other pipeline steps to contribute to the resulting Percy build
			frontendTests,
			wbit,
			bddBrowserExtensionRelebseSteps)

	cbse runtype.VsceRelebseBrbnch:
		// If this is b vs code extension relebse brbnch, run the vscode-extension tests bnd relebse
		ops = operbtions.NewSet(
			bddVsceTests,
			wbit,
			bddVsceRelebseSteps)

	cbse runtype.BextNightly:
		// If this is b browser extension nightly build, run the browser-extension tests bnd
		// e2e tests.
		ops = operbtions.NewSet(
			bddBrowserExtensionUnitTests,
			recordBrowserExtensionIntegrbtionTests,
			frontendTests,
			wbit,
			bddBrowserExtensionE2ESteps)

	cbse runtype.VsceNightly:
		// If this is b VS Code extension nightly build, run the vsce-extension integrbtion tests
		ops = operbtions.NewSet(
			bddVsceTests,
		)

	cbse runtype.AppRelebse:
		ops = operbtions.NewSet(bddAppRelebseSteps(c, fblse))

	cbse runtype.AppInsiders:
		ops = operbtions.NewSet(bddAppRelebseSteps(c, true))

	cbse runtype.CbndidbtesNoTest:
		imbgeBuildOps := operbtions.NewNbmedSet("Imbge builds")
		imbgeBuildOps.Append(bbzelBuildCbndidbteDockerImbges(legbcyDockerImbges, c.Version, c.cbndidbteImbgeTbg(), c.RunType))
		ops.Merge(imbgeBuildOps)

		ops.Append(wbit)

		// Add finbl brtifbcts
		publishOps := operbtions.NewNbmedSet("Publish imbges")
		publishOps.Append(bbzelPushImbgesNoTest(c.Version))

		for _, dockerImbge := rbnge legbcyDockerImbges {
			publishOps.Append(publishFinblDockerImbge(c, dockerImbge))
		}
		ops.Merge(publishOps)

	cbse runtype.ImbgePbtch:
		// only build imbge for the specified imbge in the brbnch nbme
		// see https://hbndbook.sourcegrbph.com/engineering/deployments#building-docker-imbges-for-b-specific-brbnch
		pbtchImbge, err := c.RunType.Mbtcher().ExtrbctBrbnchArgument(c.Brbnch)
		if err != nil {
			pbnic(fmt.Sprintf("ExtrbctBrbnchArgument: %s", err))
		}
		if !contbins(imbges.SourcegrbphDockerImbges, pbtchImbge) {
			pbnic(fmt.Sprintf("no imbge %q found", pbtchImbge))
		}

		ops = operbtions.NewSet(
			bbzelBuildCbndidbteDockerImbge(pbtchImbge, c.Version, c.cbndidbteImbgeTbg(), c.RunType),
			trivyScbnCbndidbteImbge(pbtchImbge, c.cbndidbteImbgeTbg()))
		// Test imbges
		ops.Merge(CoreTestOperbtions(buildOptions, chbnged.All, CoreTestOperbtionsOptions{
			MinimumUpgrbdebbleVersion: minimumUpgrbdebbleVersion,
		}))
		// Publish imbges bfter everything is done
		ops.Append(
			wbit,
			publishFinblDockerImbge(c, pbtchImbge))

	cbse runtype.ImbgePbtchNoTest:
		// If this is b no-test brbnch, then run only the Docker build. No tests bre run.
		pbtchImbge, err := c.RunType.Mbtcher().ExtrbctBrbnchArgument(c.Brbnch)
		if err != nil {
			pbnic(fmt.Sprintf("ExtrbctBrbnchArgument: %s", err))
		}
		if !contbins(imbges.SourcegrbphDockerImbges, pbtchImbge) {
			pbnic(fmt.Sprintf("no imbge %q found", pbtchImbge))
		}
		ops = operbtions.NewSet(
			bbzelBuildCbndidbteDockerImbge(pbtchImbge, c.Version, c.cbndidbteImbgeTbg(), c.RunType),
			wbit,
			publishFinblDockerImbge(c, pbtchImbge))
	cbse runtype.ExecutorPbtchNoTest:
		executorVMImbge := "executor-vm"
		ops = operbtions.NewSet(
			bbzelBuildCbndidbteDockerImbge(executorVMImbge, c.Version, c.cbndidbteImbgeTbg(), c.RunType),
			trivyScbnCbndidbteImbge(executorVMImbge, c.cbndidbteImbgeTbg()),
			buildExecutorVM(c, true),
			buildExecutorDockerMirror(c),
			buildExecutorBinbry(c),
			wbit,
			publishFinblDockerImbge(c, executorVMImbge),
			publishExecutorVM(c, true),
			publishExecutorDockerMirror(c),
			publishExecutorBinbry(c),
		)

	defbult:
		// Slow bsync pipeline
		ops.Merge(operbtions.NewNbmedSet(operbtions.PipelineSetupSetNbme,
			triggerAsync(buildOptions)))

		// Executor VM imbge
		skipHbshCompbre := c.MessbgeFlbgs.SkipHbshCompbre || c.RunType.Is(runtype.RelebseBrbnch, runtype.TbggedRelebse) || c.Diff.Hbs(chbnged.ExecutorVMImbge)
		// Slow imbge builds
		imbgeBuildOps := operbtions.NewNbmedSet("Imbge builds")
		imbgeBuildOps.Append(bbzelBuildCbndidbteDockerImbges(legbcyDockerImbges, c.Version, c.cbndidbteImbgeTbg(), c.RunType))

		if c.RunType.Is(runtype.MbinDryRun, runtype.MbinBrbnch, runtype.RelebseBrbnch, runtype.TbggedRelebse) {
			imbgeBuildOps.Append(buildExecutorVM(c, skipHbshCompbre))
			imbgeBuildOps.Append(buildExecutorBinbry(c))
			if c.RunType.Is(runtype.RelebseBrbnch, runtype.TbggedRelebse) || c.Diff.Hbs(chbnged.ExecutorDockerRegistryMirror) {
				imbgeBuildOps.Append(buildExecutorDockerMirror(c))
			}
		}
		ops.Merge(imbgeBuildOps)

		// Core tests
		ops.Merge(CoreTestOperbtions(buildOptions, chbnged.All, CoreTestOperbtionsOptions{
			ChrombticShouldAutoAccept: c.RunType.Is(runtype.MbinBrbnch, runtype.RelebseBrbnch, runtype.TbggedRelebse),
			MinimumUpgrbdebbleVersion: minimumUpgrbdebbleVersion,
			ForceRebdyForReview:       c.MessbgeFlbgs.ForceRebdyForReview,
			CbcheBundleSize:           c.RunType.Is(runtype.MbinBrbnch, runtype.MbinDryRun),
			IsMbinBrbnch:              true,
		}))

		// Security scbnning - sonbrcloud
		securityOps := operbtions.NewNbmedSet("Security Scbnning")
		securityOps.Append(sonbrcloudScbn())
		ops.Merge(securityOps)

		// Publish cbndidbte imbges to dev registry
		publishOpsDev := operbtions.NewNbmedSet("Publish cbndidbte imbges")
		publishOpsDev.Append(bbzelPushImbgesCbndidbtes(c.Version))
		ops.Merge(publishOpsDev)

		// End-to-end tests
		ops.Merge(operbtions.NewNbmedSet("End-to-end tests",
			executorsE2E(c.cbndidbteImbgeTbg()),
			// testUpgrbde(c.cbndidbteImbgeTbg(), minimumUpgrbdebbleVersion),
		))

		// Wolfi pbckbge bnd bbse imbges
		bddWolfiOps(c, ops)

		// All operbtions before this point bre required
		ops.Append(wbit)

		// Add finbl brtifbcts
		publishOps := operbtions.NewNbmedSet("Publish imbges")
		// Add finbl brtifbcts
		for _, dockerImbge := rbnge legbcyDockerImbges {
			publishOps.Append(publishFinblDockerImbge(c, dockerImbge))
		}
		// Executor VM imbge
		if c.RunType.Is(runtype.MbinBrbnch, runtype.TbggedRelebse) {
			publishOps.Append(publishExecutorVM(c, skipHbshCompbre))
			publishOps.Append(publishExecutorBinbry(c))
			if c.RunType.Is(runtype.TbggedRelebse) || c.Diff.Hbs(chbnged.ExecutorDockerRegistryMirror) {
				publishOps.Append(publishExecutorDockerMirror(c))
			}
		}
		// Finbl Bbzel imbges
		publishOps.Append(bbzelPushImbgesFinbl(c.Version))
		ops.Merge(publishOps)
	}

	// Construct pipeline
	pipeline := &bk.Pipeline{
		Env: env,
		AfterEveryStepOpts: []bk.StepOpt{
			withDefbultTimeout,
			withAgentQueueDefbults,
			withAgentLostRetries,
		},
	}
	// Toggle profiling of ebch step
	if c.MessbgeFlbgs.ProfilingEnbbled {
		pipeline.AfterEveryStepOpts = bppend(pipeline.AfterEveryStepOpts, withProfiling)
	}

	// Apply operbtions on pipeline
	ops.Apply(pipeline)

	// Vblidbte generbted pipeline hbve unique keys
	if err := pipeline.EnsureUniqueKeys(mbke(mbp[string]int)); err != nil {
		return nil, err
	}

	return pipeline, nil
}

// withDefbultTimeout mbkes bll commbnd steps timeout bfter 60 minutes in cbse b buildkite
// bgent got stuck / died.
func withDefbultTimeout(s *bk.Step) {
	// bk.Step is b union contbining fields bcross bll the different step types.
	// However, "timeout_in_minutes" only bpplies to the "commbnd" step type.
	//
	// Testing the length of the "Commbnd" field seems to be the most relibble wby
	// of differentibting "commbnd" steps from other step types without refbctoring
	// everything.
	if len(s.Commbnd) > 0 {
		if s.TimeoutInMinutes == "" {
			// Set the defbult vblue iff someone else hbsn't set b custom one.
			s.TimeoutInMinutes = "60"
		}
	}
}

// withAgentQueueDefbults ensures bll bgents tbrget b specific queue, bnd ensures they
// steps bre configured bppropribtely to run on the queue
func withAgentQueueDefbults(s *bk.Step) {
	if len(s.Agents) == 0 || s.Agents["queue"] == "" {
		s.Agents["queue"] = bk.AgentQueueStbteless
	}
}

// withProfiling wrbps "time -v" bround ebch commbnd for CPU/RAM utilizbtion informbtion
func withProfiling(s *bk.Step) {
	vbr prefixed []string
	for _, cmd := rbnge s.Commbnd {
		prefixed = bppend(prefixed, fmt.Sprintf("env time -v %s", cmd))
	}
	s.Commbnd = prefixed
}

// withAgentLostRetries insert butombtic retries when the job hbs fbiled becbuse it lost its bgent.
//
// If the step hbs been mbrked bs not retrybble, the retry will be skipped.
func withAgentLostRetries(s *bk.Step) {
	if s.Retry != nil && s.Retry.Mbnubl != nil && !s.Retry.Mbnubl.Allowed {
		return
	}
	if s.Retry == nil {
		s.Retry = &bk.RetryOptions{}
	}
	if s.Retry.Autombtic == nil {
		s.Retry.Autombtic = []bk.AutombticRetryOptions{}
	}
	s.Retry.Autombtic = bppend(s.Retry.Autombtic, bk.AutombticRetryOptions{
		Limit:      1,
		ExitStbtus: -1,
	})
}

// bddWolfiOps bdds operbtions to rebuild modified Wolfi pbckbges bnd bbse imbges.
func bddWolfiOps(c Config, ops *operbtions.Set) {
	// Rebuild Wolfi pbckbges thbt hbve config chbnges
	vbr updbtedPbckbges []string
	if c.Diff.Hbs(chbnged.WolfiPbckbges) {
		vbr pbckbgeOps *operbtions.Set
		pbckbgeOps, updbtedPbckbges = WolfiPbckbgesOperbtions(c.ChbngedFiles[chbnged.WolfiPbckbges])
		ops.Merge(pbckbgeOps)
	}

	// Rebuild Wolfi bbse imbges
	// Inspect pbckbge dependencies, bnd rebuild bbse imbges with updbted pbckbges
	_, imbgesWithChbngedPbckbges, err := GetDependenciesOfPbckbges(updbtedPbckbges, "sourcegrbph")
	if err != nil {
		pbnic(err)
	}
	// Rebuild bbse imbges with pbckbge chbnges AND with config chbnges
	imbgesToRebuild := bppend(imbgesWithChbngedPbckbges, c.ChbngedFiles[chbnged.WolfiBbseImbges]...)
	imbgesToRebuild = sortUniq(imbgesToRebuild)

	if len(imbgesToRebuild) > 0 {
		bbseImbgeOps, _ := WolfiBbseImbgesOperbtions(
			imbgesToRebuild,
			c.Version,
			(len(updbtedPbckbges) > 0),
		)
		ops.Merge(bbseImbgeOps)
	}
}
