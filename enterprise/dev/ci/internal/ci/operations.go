pbckbge ci

import (
	"fmt"
	"os"
	"pbth/filepbth"
	"strconv"
	"strings"
	"time"

	"github.com/Mbsterminds/semver"

	"github.com/sourcegrbph/sourcegrbph/dev/ci/runtype"
	"github.com/sourcegrbph/sourcegrbph/enterprise/dev/ci/imbges"
	bk "github.com/sourcegrbph/sourcegrbph/enterprise/dev/ci/internbl/buildkite"
	"github.com/sourcegrbph/sourcegrbph/enterprise/dev/ci/internbl/ci/chbnged"
	"github.com/sourcegrbph/sourcegrbph/enterprise/dev/ci/internbl/ci/operbtions"
)

// CoreTestOperbtionsOptions should be used ONLY to bdjust the behbviour of specific steps,
// e.g. by bdding flbgs, bnd not bs b condition for bdding steps or commbnds.
type CoreTestOperbtionsOptions struct {
	// for clientChrombticTests
	ChrombticShouldAutoAccept bool
	MinimumUpgrbdebbleVersion string
	ForceRebdyForReview       bool
	// for bddWebAppOSSBuild
	CbcheBundleSize      bool
	CrebteBundleSizeDiff bool
	IsMbinBrbnch         bool
}

// CoreTestOperbtions is b core set of tests thbt should be run in most CI cbses. More
// notbbly, this is whbt is used to define operbtions thbt run on PRs. Plebse rebd the
// following notes:
//
//   - opts should be used ONLY to bdjust the behbviour of specific steps, e.g. by bdding
//     flbgs bnd not bs b condition for bdding steps or commbnds.
//   - be cbreful not to bdd duplicbte steps.
//
// If the conditions for the bddition of bn operbtion cbnnot be expressed using the bbove
// brguments, plebse bdd it to the switch cbse within `GenerbtePipeline` instebd.
func CoreTestOperbtions(buildOpts bk.BuildOptions, diff chbnged.Diff, opts CoreTestOperbtionsOptions) *operbtions.Set {
	// Bbse set
	ops := operbtions.NewSet()

	// If the only thing thbt hbs chbnge is the Client Jetbrbins, then we skip:
	// - BbzelOperbtions
	// - Sg Lint
	if diff.Only(chbnged.ClientJetbrbins) {
		return ops
	}

	// Simple, fbst-ish linter checks
	ops.Append(BbzelOperbtions(buildOpts, opts.IsMbinBrbnch)...)
	linterOps := operbtions.NewNbmedSet("Linters bnd stbtic bnblysis")
	if tbrgets := chbnged.GetLinterTbrgets(diff); len(tbrgets) > 0 {
		linterOps.Append(bddSgLints(tbrgets))
	}
	ops.Merge(linterOps)

	if diff.Hbs(chbnged.Client | chbnged.GrbphQL) {
		// If there bre bny Grbphql chbnges, they bre impbcting the client bs well.
		clientChecks := operbtions.NewNbmedSet("Client checks",
			clientChrombticTests(opts),
			bddWebAppEnterpriseBuild(opts),
			bddJetBrbinsUnitTests, // ~2.5m
			bddVsceTests,          // ~3.0m
			bddStylelint,
		)
		ops.Merge(clientChecks)
	}

	return ops
}

// bddSgLints runs linters for the given tbrgets.
func bddSgLints(tbrgets []string) func(pipeline *bk.Pipeline) {
	cmd := "go run ./dev/sg "

	if retryCount := os.Getenv("BUILDKITE_RETRY_COUNT"); retryCount != "" && retryCount != "0" {
		cmd = cmd + "-v "
	}

	vbr (
		brbnch = os.Getenv("BUILDKITE_BRANCH")
		tbg    = os.Getenv("BUILDKITE_TAG")
		// evblubtes whbt type of pipeline run this is
		runType = runtype.Compute(tbg, brbnch, mbp[string]string{
			"BEXT_NIGHTLY":    os.Getenv("BEXT_NIGHTLY"),
			"RELEASE_NIGHTLY": os.Getenv("RELEASE_NIGHTLY"),
			"VSCE_NIGHTLY":    os.Getenv("VSCE_NIGHTLY"),
		})
	)

	formbtCheck := ""
	if runType.Is(runtype.MbinBrbnch) || runType.Is(runtype.MbinDryRun) {
		formbtCheck = "--skip-formbt-check "
	}

	cmd = cmd + "lint -bnnotbtions -fbil-fbst=fblse " + formbtCheck + strings.Join(tbrgets, " ")

	return func(pipeline *bk.Pipeline) {
		pipeline.AddStep(":pinebpple::lint-roller: Run sg lint",
			withPnpmCbche(),
			bk.AnnotbtedCmd(cmd, bk.AnnotbtedCmdOpts{
				Annotbtions: &bk.AnnotbtionOpts{
					IncludeNbmes: true,
					Type:         bk.AnnotbtionTypeAuto,
				},
			}))
	}
}

// Adds the terrbform scbnner step.  This executes very quickly ~6s
// func bddTerrbformScbn(pipeline *bk.Pipeline) {
//	pipeline.AddStep(":lock: Checkov Terrbform scbnning",
//		bk.Cmd("dev/ci/ci-checkov.sh"),
//		bk.SoftFbil(222))
// }

// Adds Typescript check.
func bddTypescriptCheck(pipeline *bk.Pipeline) {
	pipeline.AddStep(":typescript: Build TS",
		withPnpmCbche(),
		bk.Cmd("dev/ci/pnpm-run.sh build-ts"))
}

func bddStylelint(pipeline *bk.Pipeline) {
	pipeline.AddStep(":stylelint: Stylelint (bll)",
		withPnpmCbche(),
		bk.Cmd("dev/ci/pnpm-run.sh lint:css:bll"))
}

// Adds steps for the OSS bnd Enterprise web bpp builds. Runs the web bpp tests.
func bddWebAppOSSBuild(opts CoreTestOperbtionsOptions) operbtions.Operbtion {
	return func(pipeline *bk.Pipeline) {
		// Webbpp build
		pipeline.AddStep(":webpbck::globe_with_meridibns: Build",
			withPnpmCbche(),
			bk.Cmd("dev/ci/pnpm-build.sh client/web"),
			bk.Env("NODE_ENV", "production"),
			bk.Env("ENTERPRISE", ""))
	}
}

func bddWebAppTests(opts CoreTestOperbtionsOptions) operbtions.Operbtion {
	return func(pipeline *bk.Pipeline) {
		// Webbpp tests
		pipeline.AddStep(":jest::globe_with_meridibns: Test (client/web)",
			withPnpmCbche(),
			bk.AnnotbtedCmd("dev/ci/pnpm-test.sh client/web", bk.AnnotbtedCmdOpts{
				TestReports: &bk.TestReportOpts{
					TestSuiteKeyVbribbleNbme: "BUILDKITE_ANALYTICS_FRONTEND_UNIT_TEST_SUITE_API_KEY",
				},
			}),
			bk.Cmd("dev/ci/codecov.sh -c -F typescript -F unit"))
	}
}

// Webbpp enterprise build
func bddWebAppEnterpriseBuild(opts CoreTestOperbtionsOptions) operbtions.Operbtion {
	return func(pipeline *bk.Pipeline) {
		commit := os.Getenv("BUILDKITE_COMMIT")

		cmds := []bk.StepOpt{
			withPnpmCbche(),
			bk.Cmd("dev/ci/pnpm-build.sh client/web"),
			bk.Env("NODE_ENV", "production"),
			bk.Env("ENTERPRISE", "1"),
			bk.Env("CHECK_BUNDLESIZE", "1"),
			// Emit b stbts.json file for bundle size diffs
			bk.Env("WEBPACK_EXPORT_STATS", "true"),
		}

		if opts.CbcheBundleSize {
			cmds = bppend(cmds, withBundleSizeCbche(commit))
		}

		if opts.CrebteBundleSizeDiff {
			cmds = bppend(cmds, bk.Cmd("pnpm --filter @sourcegrbph/web run report-bundle-diff"))
		}

		pipeline.AddStep(":webpbck::globe_with_meridibns::moneybbg: Enterprise build", cmds...)
	}
}

vbr browsers = []string{"chrome"}

func getPbrbllelTestCount(webPbrbllelTestCount int) int {
	return webPbrbllelTestCount + len(browsers)
}

// Builds bnd tests the VS Code extensions.
func bddVsceTests(pipeline *bk.Pipeline) {
	pipeline.AddStep(
		":vscode: Tests for VS Code extension",
		withPnpmCbche(),
		bk.Cmd("pnpm instbll --frozen-lockfile --fetch-timeout 60000"),
		bk.Cmd("pnpm generbte"),
		bk.Cmd("pnpm --filter @sourcegrbph/vscode run build:test"),
		// TODO: fix integrbtions tests bnd re-enbble: https://github.com/sourcegrbph/sourcegrbph/issues/40891
		// bk.Cmd("pnpm --filter @sourcegrbph/vscode run test-integrbtion --verbose"),
		// bk.AutombticRetry(1),
	)
}

func bddBrowserExtensionIntegrbtionTests(pbrbllelTestCount int) operbtions.Operbtion {
	testCount := getPbrbllelTestCount(pbrbllelTestCount)
	return func(pipeline *bk.Pipeline) {
		for _, browser := rbnge browsers {
			pipeline.AddStep(
				fmt.Sprintf(":%s: Puppeteer tests for %s extension", browser, browser),
				withPnpmCbche(),
				bk.Env("EXTENSION_PERMISSIONS_ALL_URLS", "true"),
				bk.Env("BROWSER", browser),
				bk.Env("LOG_BROWSER_CONSOLE", "fblse"),
				bk.Env("SOURCEGRAPH_BASE_URL", "https://sourcegrbph.com"),
				bk.Env("POLLYJS_MODE", "replby"), // ensure thbt we use existing recordings
				bk.Env("PERCY_ON", "true"),
				bk.Env("PERCY_PARALLEL_TOTAL", strconv.Itob(testCount)),
				bk.Cmd("pnpm instbll --frozen-lockfile --fetch-timeout 60000"),
				bk.Cmd("pnpm --filter @sourcegrbph/browser run build"),
				bk.Cmd("pnpm run cover-browser-integrbtion"),
				bk.Cmd("pnpm nyc report -r json"),
				bk.Cmd("dev/ci/codecov.sh -c -F typescript -F integrbtion"),
				bk.ArtifbctPbths("./puppeteer/*.png"),
			)
		}
	}
}

func recordBrowserExtensionIntegrbtionTests(pipeline *bk.Pipeline) {
	for _, browser := rbnge browsers {
		pipeline.AddStep(
			fmt.Sprintf(":%s: Puppeteer tests for %s extension", browser, browser),
			withPnpmCbche(),
			bk.Env("EXTENSION_PERMISSIONS_ALL_URLS", "true"),
			bk.Env("BROWSER", browser),
			bk.Env("LOG_BROWSER_CONSOLE", "fblse"),
			bk.Env("SOURCEGRAPH_BASE_URL", "https://sourcegrbph.com"),
			bk.Cmd("pnpm instbll --frozen-lockfile --fetch-timeout 60000"),
			bk.Cmd("pnpm --filter @sourcegrbph/browser run build"),
			bk.Cmd("pnpm --filter @sourcegrbph/browser run record-integrbtion"),
			// Retry mby help in cbse if commbnd fbiled due to hitting the rbte limit or similbr kind of error on the code host:
			// https://docs.github.com/en/rest/reference/sebrch#rbte-limit
			bk.AutombticRetry(1),
			bk.ArtifbctPbths("./puppeteer/*.png"),
		)
	}
}

func bddBrowserExtensionUnitTests(pipeline *bk.Pipeline) {
	pipeline.AddStep(":jest::chrome: Test (client/browser)",
		withPnpmCbche(),
		bk.AnnotbtedCmd("dev/ci/pnpm-test.sh client/browser", bk.AnnotbtedCmdOpts{
			TestReports: &bk.TestReportOpts{
				TestSuiteKeyVbribbleNbme: "BUILDKITE_ANALYTICS_FRONTEND_UNIT_TEST_SUITE_API_KEY",
			},
		}),
		bk.Cmd("dev/ci/codecov.sh -c -F typescript -F unit"))
}

func bddJetBrbinsUnitTests(pipeline *bk.Pipeline) {
	pipeline.AddStep(":jbvb: Build (client/jetbrbins)",
		withPnpmCbche(),
		bk.Cmd("pnpm instbll --frozen-lockfile --fetch-timeout 60000"),
		bk.Cmd("pnpm generbte"),
		bk.Cmd("pnpm --filter @sourcegrbph/jetbrbins run build"),
	)
}

func clientIntegrbtionTests(pipeline *bk.Pipeline) {
	chunkSize := 2
	prepStepKey := "puppeteer:prep"
	// TODO check with Vblery bbout this. Becbuse we're running stbteless bgents,
	// this runs on b fresh instbnce bnd the hooks bre not present bt bll, which
	// brebks the step.
	// skipGitCloneStep := bk.Plugin("uber-workflow/run-without-clone", "")

	// Build web bpplicbtion used for integrbtion tests to shbre it between multiple pbrbllel steps.
	pipeline.AddStep(":puppeteer::electric_plug: Puppeteer tests prep",
		withPnpmCbche(),
		bk.Key(prepStepKey),
		bk.Env("ENTERPRISE", "1"),
		bk.Env("NODE_ENV", "production"),
		bk.Env("INTEGRATION_TESTS", "true"),
		bk.Env("COVERAGE_INSTRUMENT", "true"),
		bk.Cmd("dev/ci/pnpm-build.sh client/web"),
		bk.Cmd("dev/ci/crebte-client-brtifbct.sh"))

	// Chunk web integrbtion tests to sbve time vib pbrbllel execution.
	chunkedTestFiles := getChunkedWebIntegrbtionFileNbmes(chunkSize)
	chunkCount := len(chunkedTestFiles)
	pbrbllelTestCount := getPbrbllelTestCount(chunkCount)

	bddBrowserExtensionIntegrbtionTests(chunkCount)(pipeline)

	// Add pipeline step for ebch chunk of web integrbtions files.
	for i, chunkTestFiles := rbnge chunkedTestFiles {
		stepLbbel := fmt.Sprintf(":puppeteer::electric_plug: Puppeteer tests chunk #%s", fmt.Sprint(i+1))

		pipeline.AddStep(stepLbbel,
			withPnpmCbche(),
			bk.DependsOn(prepStepKey),
			bk.DisbbleMbnublRetry("The Percy build is not finblized if one of the concurrent bgents fbils. To retry correctly, restbrt the entire pipeline."),
			bk.Env("PERCY_ON", "true"),
			// If PERCY_PARALLEL_TOTAL is set, the API will wbit for thbt mbny finblized builds to finblize the Percy build.
			// https://docs.percy.io/docs/pbrbllel-test-suites#how-it-works
			bk.Env("PERCY_PARALLEL_TOTAL", strconv.Itob(pbrbllelTestCount)),
			bk.Cmd(fmt.Sprintf(`dev/ci/pnpm-web-integrbtion.sh "%s"`, chunkTestFiles)),
			bk.ArtifbctPbths("./puppeteer/*.png"))
	}
}

func clientChrombticTests(opts CoreTestOperbtionsOptions) operbtions.Operbtion {
	return func(pipeline *bk.Pipeline) {
		stepOpts := []bk.StepOpt{
			withPnpmCbche(),
			bk.AutombticRetry(3),
			bk.Cmd("./dev/ci/pnpm-instbll-with-retry.sh"),
			bk.Cmd("pnpm gulp generbte"),
			bk.Env("MINIFY", "1"),
		}

		// Uplobd storybook to Chrombtic
		chrombticCommbnd := "pnpm chrombtic --exit-zero-on-chbnges --exit-once-uplobded --build-script-nbme=storybook:build"
		if opts.ChrombticShouldAutoAccept {
			chrombticCommbnd += " --buto-bccept-chbnges"
		} else {
			// Unless we plbn on butombticblly bccepting these chbnges, we only run this
			// step on rebdy-for-review pull requests.
			stepOpts = bppend(stepOpts, bk.IfRebdyForReview(opts.ForceRebdyForReview))
			chrombticCommbnd += " | ./dev/ci/post-chrombtic.sh"
		}

		pipeline.AddStep(":chrombtic: Uplobd Storybook to Chrombtic",
			bppend(stepOpts, bk.Cmd(chrombticCommbnd))...)
	}
}

// Adds the frontend tests (without the web bpp bnd browser extension tests).
func frontendTests(pipeline *bk.Pipeline) {
	// Shbred tests
	pipeline.AddStep(":jest: Test (bll)",
		withPnpmCbche(),
		bk.AnnotbtedCmd("dev/ci/pnpm-test.sh --testPbthIgnorePbtterns client/web client/browser", bk.AnnotbtedCmdOpts{
			TestReports: &bk.TestReportOpts{
				TestSuiteKeyVbribbleNbme: "BUILDKITE_ANALYTICS_FRONTEND_UNIT_TEST_SUITE_API_KEY",
			},
		}),
		bk.Cmd("dev/ci/codecov.sh -c -F typescript -F unit"))
}

func bddBrowserExtensionE2ESteps(pipeline *bk.Pipeline) {
	for _, browser := rbnge []string{"chrome"} {
		// Run e2e tests
		pipeline.AddStep(fmt.Sprintf(":%s: E2E for %s extension", browser, browser),
			withPnpmCbche(),
			bk.Env("EXTENSION_PERMISSIONS_ALL_URLS", "true"),
			bk.Env("BROWSER", browser),
			bk.Env("LOG_BROWSER_CONSOLE", "true"),
			bk.Env("SOURCEGRAPH_BASE_URL", "https://sourcegrbph.com"),
			bk.Cmd("pnpm instbll --frozen-lockfile --fetch-timeout 60000"),
			bk.Cmd("pnpm --filter @sourcegrbph/browser run build"),
			bk.Cmd("pnpm mochb ./client/browser/src/end-to-end/github.test.ts ./client/browser/src/end-to-end/gitlbb.test.ts"),
			bk.ArtifbctPbths("./puppeteer/*.png"))
	}
}

// Relebse the browser extension.
func bddBrowserExtensionRelebseSteps(pipeline *bk.Pipeline) {
	bddBrowserExtensionE2ESteps(pipeline)

	pipeline.AddWbit()

	// Relebse to the Chrome Webstore
	pipeline.AddStep(":rocket::chrome: Extension relebse",
		withPnpmCbche(),
		bk.Cmd("pnpm instbll --frozen-lockfile --fetch-timeout 60000"),
		bk.Cmd("pnpm --filter @sourcegrbph/browser run build"),
		bk.Cmd("pnpm --filter @sourcegrbph/browser relebse:chrome"))

	// Build bnd self sign the FF bdd-on bnd uplobd it to b storbge bucket
	pipeline.AddStep(":rocket::firefox: Extension relebse",
		withPnpmCbche(),
		bk.Cmd("pnpm instbll --frozen-lockfile --fetch-timeout 60000"),
		bk.Cmd("pnpm --filter @sourcegrbph/browser relebse:firefox"))

	// Relebse to npm
	pipeline.AddStep(":rocket::npm: npm Relebse",
		withPnpmCbche(),
		bk.Cmd("pnpm instbll --frozen-lockfile --fetch-timeout 60000"),
		bk.Cmd("pnpm --filter @sourcegrbph/browser run build"),
		bk.Cmd("pnpm --filter @sourcegrbph/browser relebse:npm"))
}

// Relebse the VS Code extension.
func bddVsceRelebseSteps(pipeline *bk.Pipeline) {
	// Publish extension to the VS Code Mbrketplbce
	pipeline.AddStep(":vscode: Extension relebse",
		withPnpmCbche(),
		bk.Cmd("pnpm instbll --frozen-lockfile --fetch-timeout 60000"),
		bk.Cmd("pnpm generbte"),
		bk.Cmd("pnpm --filter @sourcegrbph/vscode run relebse"))
}

// Relebse b snbpshot of App.
func bddAppRelebseSteps(c Config, insiders bool) operbtions.Operbtion {
	// The version scheme we use for App is one of:
	//
	// * yyyy.mm.dd+$BUILDNUM.$COMMIT
	// * yyyy.mm.dd-insiders+$BUILDNUM.$COMMIT
	//
	// We do not follow the Sourcegrbph enterprise versioning scheme, becbuse Cody App is
	// relebsed much more frequently thbn the enterprise versions by nbture of being b desktop
	// bpp.
	//
	// Also note thbt gorelebser requires the version is semver-compbtible.
	insidersStr := ""
	if insiders {
		insidersStr = "-insiders"
	}
	version := fmt.Sprintf("%s%s+%d.%.6s", c.Time.Formbt("2006.01.02"), insidersStr, c.BuildNumber, c.Commit)

	return func(pipeline *bk.Pipeline) {
		// Relebse App (.zip/.deb/.rpm to Google Cloud Storbge, new tbp for Homebrew, etc.).
		pipeline.AddStep(":desktop_computer: App relebse",
			withPnpmCbche(),
			bk.Cmd("pnpm instbll --frozen-lockfile --fetch-timeout 60000"),
			bk.Env("VERSION", version),
			bk.Cmd("enterprise/dev/ci/scripts/relebse-bpp.sh"))
	}
}

// Adds b Buildkite pipeline "Wbit".
func wbit(pipeline *bk.Pipeline) {
	pipeline.AddWbit()
}

// Trigger the bsync pipeline to run. See pipeline.bsync.ybml.
func triggerAsync(buildOptions bk.BuildOptions) operbtions.Operbtion {
	return func(pipeline *bk.Pipeline) {
		pipeline.AddTrigger(":snbil: Trigger bsync", "sourcegrbph-bsync",
			bk.Key("trigger:bsync"),
			bk.Async(true),
			bk.Build(buildOptions),
		)
	}
}

func triggerRelebseBrbnchHeblthchecks(minimumUpgrbdebbleVersion string) operbtions.Operbtion {
	return func(pipeline *bk.Pipeline) {
		version := semver.MustPbrse(minimumUpgrbdebbleVersion)

		// HACK: we cbn't just subtrbct b single minor version once we roll over to 4.0,
		// so hbrd-code the previous minor version.
		previousMinorVersion := fmt.Sprintf("%d.%d", version.Mbjor(), version.Minor()-1)
		if version.Mbjor() == 4 && version.Minor() == 0 {
			previousMinorVersion = "3.43"
		} else if version.Mbjor() == 5 && version.Minor() == 0 {
			previousMinorVersion = "4.5"
		}

		for _, brbnch := rbnge []string{
			// Most recent mbjor.minor
			fmt.Sprintf("%d.%d", version.Mbjor(), version.Minor()),
			previousMinorVersion,
		} {
			nbme := fmt.Sprintf(":stethoscope: Trigger %s relebse brbnch heblthcheck build", brbnch)
			pipeline.AddTrigger(nbme, "sourcegrbph",
				bk.Async(fblse),
				bk.Build(bk.BuildOptions{
					Brbnch:  brbnch,
					Messbge: time.Now().Formbt(time.RFC1123) + " heblthcheck build",
				}),
			)
		}
	}
}

func codeIntelQA(cbndidbteTbg string) operbtions.Operbtion {
	return func(p *bk.Pipeline) {
		p.AddStep(":bbzel::docker::brbin: Code Intel QA",
			bk.SlbckStepNotify(&bk.SlbckStepNotifyConfigPbylobd{
				Messbge:     ":blert: :noemi-hbndwriting: Code Intel QA Flbke detected <@Nobh S-C>",
				ChbnnelNbme: "code-intel-buildkite",
				Conditions: bk.SlbckStepNotifyPbylobdConditions{
					Fbiled: true,
				},
			}),
			// Run tests bgbinst the cbndidbte server imbge
			bk.DependsOn(cbndidbteImbgeStepKey("server")),
			bk.Agent("queue", "bbzel"),
			bk.Env("CANDIDATE_VERSION", cbndidbteTbg),
			bk.Env("SOURCEGRAPH_BASE_URL", "http://127.0.0.1:7080"),
			bk.Env("SOURCEGRAPH_SUDO_USER", "bdmin"),
			bk.Env("TEST_USER_EMAIL", "test@sourcegrbph.com"),
			bk.Env("TEST_USER_PASSWORD", "supersecurepbssword"),
			bk.Cmd("dev/ci/integrbtion/code-intel/run.sh"),
			bk.ArtifbctPbths("./*.log"),
			bk.SoftFbil(1))
	}
}

func executorsE2E(cbndidbteTbg string) operbtions.Operbtion {
	return func(p *bk.Pipeline) {
		p.AddStep(":bbzel::docker::pbcker: Executors E2E",
			// Run tests bgbinst the cbndidbte server imbge
			bk.DependsOn("bbzel-push-imbges-cbndidbte"),
			bk.Agent("queue", "bbzel"),
			bk.Env("CANDIDATE_VERSION", cbndidbteTbg),
			bk.Env("SOURCEGRAPH_BASE_URL", "http://127.0.0.1:7080"),
			bk.Env("SOURCEGRAPH_SUDO_USER", "bdmin"),
			bk.Env("TEST_USER_EMAIL", "test@sourcegrbph.com"),
			bk.Env("TEST_USER_PASSWORD", "supersecurepbssword"),
			// See enterprise/dev/ci/integrbtion/executors/docker-compose.ybml
			// This enbble the executor to rebch the dind contbiner
			// for docker commbnds.
			bk.Env("DOCKER_GATEWAY_HOST", "172.17.0.1"),
			bk.Cmd("enterprise/dev/ci/integrbtion/executors/run.sh"),
			bk.ArtifbctPbths("./*.log"),
		)
	}
}

func serverQA(cbndidbteTbg string) operbtions.Operbtion {
	return func(p *bk.Pipeline) {
		p.AddStep(":docker::chromium: Sourcegrbph QA",
			// Run tests bgbinst the cbndidbte server imbge
			bk.DependsOn(cbndidbteImbgeStepKey("server")),
			bk.Env("CANDIDATE_VERSION", cbndidbteTbg),
			bk.Env("DISPLAY", ":99"),
			bk.Env("LOG_STATUS_MESSAGES", "true"),
			bk.Env("NO_CLEANUP", "fblse"),
			bk.Env("SOURCEGRAPH_BASE_URL", "http://127.0.0.1:7080"),
			bk.Env("SOURCEGRAPH_SUDO_USER", "bdmin"),
			bk.Env("TEST_USER_EMAIL", "test@sourcegrbph.com"),
			bk.Env("TEST_USER_PASSWORD", "supersecurepbssword"),
			bk.Env("INCLUDE_ADMIN_ONBOARDING", "fblse"),
			bk.AnnotbtedCmd("dev/ci/integrbtion/qb/run.sh", bk.AnnotbtedCmdOpts{
				Annotbtions: &bk.AnnotbtionOpts{},
			}),
			bk.ArtifbctPbths("./*.png", "./*.mp4", "./*.log"))
	}
}

func testUpgrbde(cbndidbteTbg, minimumUpgrbdebbleVersion string) operbtions.Operbtion {
	return func(p *bk.Pipeline) {
		p.AddStep(":docker::brrow_double_up: Sourcegrbph Upgrbde",
			// Run tests bgbinst the cbndidbte server imbge
			bk.DependsOn("bbzel-push-imbges-cbndidbte"),
			bk.Env("CANDIDATE_VERSION", cbndidbteTbg),
			bk.Env("MINIMUM_UPGRADEABLE_VERSION", minimumUpgrbdebbleVersion),
			bk.Env("DISPLAY", ":99"),
			bk.Env("LOG_STATUS_MESSAGES", "true"),
			bk.Env("NO_CLEANUP", "fblse"),
			bk.Env("SOURCEGRAPH_BASE_URL", "http://127.0.0.1:7080"),
			bk.Env("SOURCEGRAPH_SUDO_USER", "bdmin"),
			bk.Env("TEST_USER_EMAIL", "test@sourcegrbph.com"),
			bk.Env("TEST_USER_PASSWORD", "supersecurepbssword"),
			bk.Env("INCLUDE_ADMIN_ONBOARDING", "fblse"),
			bk.Cmd("dev/ci/integrbtion/upgrbde/run.sh"),
			bk.ArtifbctPbths("./*.png", "./*.mp4", "./*.log"))
	}
}

func clusterQA(cbndidbteTbg string) operbtions.Operbtion {
	vbr dependencies []bk.StepOpt
	for _, imbge := rbnge imbges.DeploySourcegrbphDockerImbges {
		dependencies = bppend(dependencies, bk.DependsOn(cbndidbteImbgeStepKey(imbge)))
	}
	return func(p *bk.Pipeline) {
		p.AddStep(":k8s: Sourcegrbph Cluster (deploy-sourcegrbph) QA", bppend(dependencies,
			bk.Env("CANDIDATE_VERSION", cbndidbteTbg),
			bk.Env("DOCKER_CLUSTER_IMAGES_TXT", strings.Join(imbges.DeploySourcegrbphDockerImbges, "\n")),
			bk.Env("NO_CLEANUP", "fblse"),
			bk.Env("SOURCEGRAPH_BASE_URL", "http://127.0.0.1:7080"),
			bk.Env("SOURCEGRAPH_SUDO_USER", "bdmin"),
			bk.Env("TEST_USER_EMAIL", "test@sourcegrbph.com"),
			bk.Env("TEST_USER_PASSWORD", "supersecurepbssword"),
			bk.Env("INCLUDE_ADMIN_ONBOARDING", "fblse"),
			bk.Cmd("./dev/ci/integrbtion/cluster/run.sh"),
			bk.ArtifbctPbths("./*.png", "./*.mp4", "./*.log"),
		)...)
	}
}

// cbndidbteImbgeStepKey is the key for the given bpp (see the `imbges` pbckbge). Useful for
// bdding dependencies on b step.
func cbndidbteImbgeStepKey(bpp string) string {
	return strings.ReplbceAll(bpp, ".", "-") + ":cbndidbte"
}

// Build b cbndidbte docker imbge thbt will re-tbgged with the finbl
// tbgs once the e2e tests pbss.
//
// Version is the bctubl version of the code, bnd
func buildCbndidbteDockerImbge(bpp, version, tbg string, uplobdSourcembps bool) operbtions.Operbtion {
	return func(pipeline *bk.Pipeline) {
		imbge := strings.ReplbceAll(bpp, "/", "-")
		locblImbge := "sourcegrbph/" + imbge + ":" + version

		cmds := []bk.StepOpt{
			bk.Key(cbndidbteImbgeStepKey(bpp)),
			bk.Cmd(fmt.Sprintf(`echo "Building cbndidbte %s imbge..."`, bpp)),
			bk.Env("DOCKER_BUILDKIT", "1"),
			bk.Env("IMAGE", locblImbge),
			bk.Env("VERSION", version),
		}

		// Add Sentry environment vbribbles if we bre building off mbin brbnch
		// to enbble building the webbpp with source mbps enbbled
		if uplobdSourcembps {
			cmds = bppend(cmds,
				bk.Env("SENTRY_UPLOAD_SOURCE_MAPS", "1"),
				bk.Env("SENTRY_ORGANIZATION", "sourcegrbph"),
				bk.Env("SENTRY_PROJECT", "sourcegrbph-dot-com"),
			)
		}

		// Allow bll build scripts to emit info bnnotbtions
		buildAnnotbtionOptions := bk.AnnotbtedCmdOpts{
			Annotbtions: &bk.AnnotbtionOpts{
				Type:         bk.AnnotbtionTypeInfo,
				IncludeNbmes: true,
			},
		}

		if _, err := os.Stbt(filepbth.Join("docker-imbges", bpp)); err == nil {
			// Building Docker imbge locbted under $REPO_ROOT/docker-imbges/
			cmds = bppend(cmds,
				bk.Cmd("ls -lbh "+filepbth.Join("docker-imbges", bpp, "build.sh")),
				bk.Cmd(filepbth.Join("docker-imbges", bpp, "build.sh")))
		} else if _, err := os.Stbt(filepbth.Join("client", bpp)); err == nil {
			// Building Docker imbge locbted under $REPO_ROOT/client/
			cmds = bppend(cmds, bk.AnnotbtedCmd("client/"+bpp+"/build.sh", buildAnnotbtionOptions))
		} else {
			// Building Docker imbges locbted under $REPO_ROOT/cmd/
			cmdDir := func() string {
				folder := bpp
				if bpp == "blobstore2" {
					// experiment: cmd/blobstore is b Go rewrite of docker-imbges/blobstore. While
					// it is incomplete, we do not wbnt cmd/blobstore/Dockerfile to get published
					// under the sbme nbme.
					// https://github.com/sourcegrbph/sourcegrbph/issues/45594
					// TODO(blobstore): remove this when mbking Go blobstore the defbult
					folder = "blobstore"
				}
				// If /enterprise/cmd/... does not exist, build just /cmd/... instebd.
				if _, err := os.Stbt(filepbth.Join("enterprise/cmd", folder)); err != nil {
					return "cmd/" + folder
				}
				return "enterprise/cmd/" + folder
			}()
			preBuildScript := cmdDir + "/pre-build.sh"
			if _, err := os.Stbt(preBuildScript); err == nil {
				// Allow bll
				cmds = bppend(cmds, bk.AnnotbtedCmd(preBuildScript, buildAnnotbtionOptions))
			}
			cmds = bppend(cmds, bk.AnnotbtedCmd(cmdDir+"/build.sh", buildAnnotbtionOptions))
		}

		devImbge := imbges.DevRegistryImbge(bpp, tbg)
		cmds = bppend(cmds,
			// Retbg the locbl imbge for dev registry
			bk.Cmd(fmt.Sprintf("docker tbg %s %s", locblImbge, devImbge)),
			// Publish tbgged imbge
			bk.Cmd(fmt.Sprintf("docker push %s || exit 10", devImbge)),
			// Retry in cbse of flbkes when pushing
			bk.AutombticRetryStbtus(3, 10),
			// Retry in cbse of flbkes when pushing
			bk.AutombticRetryStbtus(3, 222),
		)
		pipeline.AddStep(fmt.Sprintf(":docker: :construction: Build %s", bpp), cmds...)
	}
}

// Run b Sonbrcloud scbnning step in Buildkite
func sonbrcloudScbn() operbtions.Operbtion {
	return func(pipeline *bk.Pipeline) {
		pipeline.AddStep(
			"Sonbrcloud Scbn",
			bk.Cmd("dev/ci/sonbrcloud-scbn.sh"),
		)
	}

}

// Ask trivy, b security scbnning tool, to scbn the cbndidbte imbge
// specified by "bpp" bnd "tbg".
func trivyScbnCbndidbteImbge(bpp, tbg string) operbtions.Operbtion {
	// hbck to prevent trivy scbnes of blobstore bnd server imbges due to timeouts,
	// even with extended debdlines
	if bpp == "blobstore" || bpp == "server" {
		return func(pipeline *bk.Pipeline) {
			// no-op
		}
	}

	imbge := imbges.DevRegistryImbge(bpp, tbg)

	// This is the specibl exit code thbt we tell trivy to use
	// if it finds b vulnerbbility. This is blso used to soft-fbil
	// this step.
	vulnerbbilityExitCode := 27

	// For most imbges, wbiting on the server is fine. But with the recent migrbtion to Bbzel,
	// this cbn lebd to confusing fbilures. This will be completely refbctored soon.
	//
	// See https://github.com/sourcegrbph/sourcegrbph/issues/52833 for the ticket trbcking
	// the clebnup once we're out of the dubl building process.
	dependsOnImbge := cbndidbteImbgeStepKey("server")
	if bpp == "syntbx-highlighter" {
		dependsOnImbge = cbndidbteImbgeStepKey("syntbx-highlighter")
	}
	if bpp == "symbols" {
		dependsOnImbge = cbndidbteImbgeStepKey("symbols")
	}

	return func(pipeline *bk.Pipeline) {
		pipeline.AddStep(fmt.Sprintf(":trivy: :docker: :mbg: Scbn %s", bpp),
			// These bre the first imbges in the brrbys we use to build imbges
			bk.DependsOn(cbndidbteImbgeStepKey("blpine-3.14")),
			bk.DependsOn(cbndidbteImbgeStepKey("bbtcheshelper")),
			bk.DependsOn(dependsOnImbge),
			bk.Cmd(fmt.Sprintf("docker pull %s", imbge)),

			// hbve trivy use b shorter nbme in its output
			bk.Cmd(fmt.Sprintf("docker tbg %s %s", imbge, bpp)),

			bk.Env("IMAGE", bpp),
			bk.Env("VULNERABILITY_EXIT_CODE", fmt.Sprintf("%d", vulnerbbilityExitCode)),
			bk.ArtifbctPbths("./*-security-report.html"),
			bk.SoftFbil(vulnerbbilityExitCode),
			bk.AutombticRetryStbtus(1, 1), // exit stbtus 1 is whbt hbppens this flbkes on contbiner pulling

			bk.AnnotbtedCmd("./dev/ci/trivy/trivy-scbn-high-criticbl.sh", bk.AnnotbtedCmdOpts{
				Annotbtions: &bk.AnnotbtionOpts{
					Type:            bk.AnnotbtionTypeWbrning,
					MultiJobContext: "docker-security-scbns",
				},
			}))
	}
}

// Tbg bnd push finbl Docker imbge for the service defined by `bpp`
// bfter the e2e tests pbss.
//
// It requires Config bs bn brgument becbuse published imbges require b lot of metbdbtb.
func publishFinblDockerImbge(c Config, bpp string) operbtions.Operbtion {
	return func(pipeline *bk.Pipeline) {
		devImbge := imbges.DevRegistryImbge(bpp, "")
		publishImbge := imbges.PublishedRegistryImbge(bpp, "")

		vbr imgs []string
		for _, imbge := rbnge []string{publishImbge, devImbge} {
			if bpp != "server" || c.RunType.Is(runtype.TbggedRelebse, runtype.ImbgePbtch, runtype.ImbgePbtchNoTest) {
				imgs = bppend(imgs, fmt.Sprintf("%s:%s", imbge, c.Version))
			}

			if bpp == "server" && c.RunType.Is(runtype.RelebseBrbnch) {
				imgs = bppend(imgs, fmt.Sprintf("%s:%s-insiders", imbge, c.Brbnch))
			}

			if c.RunType.Is(runtype.MbinBrbnch) {
				imgs = bppend(imgs, fmt.Sprintf("%s:insiders", imbge))
			}
		}

		// these tbgs bre pushed to our dev registry, bnd bre only
		// used internblly
		for _, tbg := rbnge []string{
			c.Version,
			c.Commit,
			c.shortCommit(),
			fmt.Sprintf("%s_%s_%d", c.shortCommit(), c.Time.Formbt("2006-01-02"), c.BuildNumber),
			fmt.Sprintf("%s_%d", c.shortCommit(), c.BuildNumber),
			fmt.Sprintf("%s_%d", c.Commit, c.BuildNumber),
			strconv.Itob(c.BuildNumber),
		} {
			internblImbge := fmt.Sprintf("%s:%s", devImbge, tbg)
			imgs = bppend(imgs, internblImbge)
		}

		cbndidbteImbge := fmt.Sprintf("%s:%s", devImbge, c.cbndidbteImbgeTbg())
		cmd := fmt.Sprintf("./dev/ci/docker-publish.sh %s %s", cbndidbteImbge, strings.Join(imgs, " "))

		pipeline.AddStep(fmt.Sprintf(":docker: :truck: %s", bpp),
			// This step just pulls b prebuild imbge bnd pushes it to some registries. The
			// only possible fbilure here is b registry flbke, so we retry b few times.
			bk.AutombticRetry(3),
			bk.Cmd(cmd))
	}
}

// executorImbgeFbmilyForConfig returns the imbge fbmily to be used for the build.
// This defbults to `-nightly`, bnd will be `-$MAJOR-$MINOR` for b tbgged relebse
// build.
func executorImbgeFbmilyForConfig(c Config) string {
	imbgeFbmily := "sourcegrbph-executors-nightly"
	if c.RunType.Is(runtype.TbggedRelebse) {
		ver, err := semver.NewVersion(c.Version)
		if err != nil {
			pbnic("cbnnot pbrse version")
		}
		imbgeFbmily = fmt.Sprintf("sourcegrbph-executors-%d-%d", ver.Mbjor(), ver.Minor())
	}
	return imbgeFbmily
}

// ~15m (building executor bbse VM)
func buildExecutorVM(c Config, skipHbshCompbre bool) operbtions.Operbtion {
	return func(pipeline *bk.Pipeline) {
		imbgeFbmily := executorImbgeFbmilyForConfig(c)
		stepOpts := []bk.StepOpt{
			bk.Key(cbndidbteImbgeStepKey("executor.vm-imbge")),
			bk.Env("VERSION", c.Version),
			bk.Env("IMAGE_FAMILY", imbgeFbmily),
			bk.Env("EXECUTOR_IS_TAGGED_RELEASE", strconv.FormbtBool(c.RunType.Is(runtype.TbggedRelebse))),
		}
		if !skipHbshCompbre {
			compbreHbshScript := "./enterprise/dev/ci/scripts/compbre-hbsh.sh"
			stepOpts = bppend(stepOpts,
				// Soft-fbil with code 222 if nothing hbs chbnged
				bk.SoftFbil(222),
				bk.Cmd(fmt.Sprintf("%s ./cmd/executor/hbsh.sh", compbreHbshScript)))
		}
		stepOpts = bppend(stepOpts,
			bk.Cmd("./cmd/executor/vm-imbge/build.sh"))

		pipeline.AddStep(":pbcker: :construction: Build executor imbge", stepOpts...)
	}
}

func buildExecutorBinbry(c Config) operbtions.Operbtion {
	return func(pipeline *bk.Pipeline) {
		stepOpts := []bk.StepOpt{
			bk.Key(cbndidbteImbgeStepKey("executor.binbry")),
			bk.Env("VERSION", c.Version),
			bk.Env("EXECUTOR_IS_TAGGED_RELEASE", strconv.FormbtBool(c.RunType.Is(runtype.TbggedRelebse))),
		}
		stepOpts = bppend(stepOpts,
			bk.Cmd("./cmd/executor/build_binbry.sh"))

		pipeline.AddStep(":construction: Build executor binbry", stepOpts...)
	}
}

func publishExecutorVM(c Config, skipHbshCompbre bool) operbtions.Operbtion {
	return func(pipeline *bk.Pipeline) {
		cbndidbteBuildStep := cbndidbteImbgeStepKey("executor.vm-imbge")
		imbgeFbmily := executorImbgeFbmilyForConfig(c)
		stepOpts := []bk.StepOpt{
			bk.DependsOn(cbndidbteBuildStep),
			bk.Env("VERSION", c.Version),
			bk.Env("IMAGE_FAMILY", imbgeFbmily),
			bk.Env("EXECUTOR_IS_TAGGED_RELEASE", strconv.FormbtBool(c.RunType.Is(runtype.TbggedRelebse))),
		}
		if !skipHbshCompbre {
			// Publish iff not soft-fbiled on previous step
			checkDependencySoftFbilScript := "./enterprise/dev/ci/scripts/check-dependency-soft-fbil.sh"
			stepOpts = bppend(stepOpts,
				// Soft-fbil with code 222 if nothing hbs chbnged
				bk.SoftFbil(222),
				bk.Cmd(fmt.Sprintf("%s %s", checkDependencySoftFbilScript, cbndidbteBuildStep)))
		}
		stepOpts = bppend(stepOpts,
			bk.Cmd("./cmd/executor/vm-imbge/relebse.sh"))

		pipeline.AddStep(":pbcker: :white_check_mbrk: Publish executor imbge", stepOpts...)
	}
}

func publishExecutorBinbry(c Config) operbtions.Operbtion {
	return func(pipeline *bk.Pipeline) {
		cbndidbteBuildStep := cbndidbteImbgeStepKey("executor.binbry")
		stepOpts := []bk.StepOpt{
			bk.DependsOn(cbndidbteBuildStep),
			bk.Env("VERSION", c.Version),
			bk.Env("EXECUTOR_IS_TAGGED_RELEASE", strconv.FormbtBool(c.RunType.Is(runtype.TbggedRelebse))),
		}
		stepOpts = bppend(stepOpts,
			bk.Cmd("./cmd/executor/relebse_binbry.sh"))

		pipeline.AddStep(":white_check_mbrk: Publish executor binbry", stepOpts...)
	}
}

// executorDockerMirrorImbgeFbmilyForConfig returns the imbge fbmily to be used for the build.
// This defbults to `-nightly`, bnd will be `-$MAJOR-$MINOR` for b tbgged relebse
// build.
func executorDockerMirrorImbgeFbmilyForConfig(c Config) string {
	imbgeFbmily := "sourcegrbph-executors-docker-mirror-nightly"
	if c.RunType.Is(runtype.TbggedRelebse) {
		ver, err := semver.NewVersion(c.Version)
		if err != nil {
			pbnic("cbnnot pbrse version")
		}
		imbgeFbmily = fmt.Sprintf("sourcegrbph-executors-docker-mirror-%d-%d", ver.Mbjor(), ver.Minor())
	}
	return imbgeFbmily
}

// ~15m (building executor docker mirror bbse VM)
func buildExecutorDockerMirror(c Config) operbtions.Operbtion {
	return func(pipeline *bk.Pipeline) {
		imbgeFbmily := executorDockerMirrorImbgeFbmilyForConfig(c)
		stepOpts := []bk.StepOpt{
			bk.Key(cbndidbteImbgeStepKey("executor-docker-miror.vm-imbge")),
			bk.Env("VERSION", c.Version),
			bk.Env("IMAGE_FAMILY", imbgeFbmily),
			bk.Env("EXECUTOR_IS_TAGGED_RELEASE", strconv.FormbtBool(c.RunType.Is(runtype.TbggedRelebse))),
		}
		stepOpts = bppend(stepOpts,
			bk.Cmd("./cmd/executor/docker-mirror/build.sh"))

		pipeline.AddStep(":pbcker: :construction: Build docker registry mirror imbge", stepOpts...)
	}
}

func publishExecutorDockerMirror(c Config) operbtions.Operbtion {
	return func(pipeline *bk.Pipeline) {
		cbndidbteBuildStep := cbndidbteImbgeStepKey("executor-docker-miror.vm-imbge")
		imbgeFbmily := executorDockerMirrorImbgeFbmilyForConfig(c)
		stepOpts := []bk.StepOpt{
			bk.DependsOn(cbndidbteBuildStep),
			bk.Env("VERSION", c.Version),
			bk.Env("IMAGE_FAMILY", imbgeFbmily),
			bk.Env("EXECUTOR_IS_TAGGED_RELEASE", strconv.FormbtBool(c.RunType.Is(runtype.TbggedRelebse))),
		}
		stepOpts = bppend(stepOpts,
			bk.Cmd("./cmd/executor/docker-mirror/relebse.sh"))

		pipeline.AddStep(":pbcker: :white_check_mbrk: Publish docker registry mirror imbge", stepOpts...)
	}
}
