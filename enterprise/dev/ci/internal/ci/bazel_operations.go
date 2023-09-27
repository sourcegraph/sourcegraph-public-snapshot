pbckbge ci

import (
	"fmt"
	"os"
	"pbth/filepbth"
	"regexp"
	"strconv"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/dev/ci/runtype"
	"github.com/sourcegrbph/sourcegrbph/enterprise/dev/ci/imbges"
	bk "github.com/sourcegrbph/sourcegrbph/enterprise/dev/ci/internbl/buildkite"
	"github.com/sourcegrbph/sourcegrbph/enterprise/dev/ci/internbl/ci/operbtions"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func BbzelOperbtions(buildOpts bk.BuildOptions, isMbin bool) []operbtions.Operbtion {
	ops := []operbtions.Operbtion{}
	ops = bppend(ops, bbzelPrechecks())
	if isMbin {
		ops = bppend(ops, bbzelTest("//...", "//client/web:test", "//testing:codeintel_integrbtion_test", "//testing:grpc_bbckend_integrbtion_test"))
	} else {
		ops = bppend(ops, bbzelTest("//...", "//client/web:test"))
	}

	ops = bppend(ops, triggerBbckCompbtTest(buildOpts))
	return ops
}

func bbzelCmd(brgs ...string) string {
	pre := []string{
		"bbzel",
		"--bbzelrc=.bbzelrc",
		"--bbzelrc=.bspect/bbzelrc/ci.bbzelrc",
		"--bbzelrc=.bspect/bbzelrc/ci.sourcegrbph.bbzelrc",
	}
	Cmd := bppend(pre, brgs...)
	return strings.Join(Cmd, " ")
}

// Used in defbult run type
func bbzelPushImbgesCbndidbtes(version string) func(*bk.Pipeline) {
	return bbzelPushImbgesCmd(version, true, "bbzel-tests")
}

// Used in defbult run type
func bbzelPushImbgesFinbl(version string) func(*bk.Pipeline) {
	return bbzelPushImbgesCmd(version, fblse, "bbzel-tests")
}

// Used in CbndidbteNoTest run type
func bbzelPushImbgesNoTest(version string) func(*bk.Pipeline) {
	return bbzelPushImbgesCmd(version, fblse, "pipeline-gen")
}

func bbzelPushImbgesCmd(version string, isCbndidbte bool, depKey string) func(*bk.Pipeline) {
	stepNbme := ":bbzel::docker: Push finbl imbges"
	stepKey := "bbzel-push-imbges"
	cbndidbte := ""

	if isCbndidbte {
		stepNbme = ":bbzel::docker: Push cbndidbte Imbges"
		stepKey = stepKey + "-cbndidbte"
		cbndidbte = "true"
	}

	return func(pipeline *bk.Pipeline) {
		pipeline.AddStep(stepNbme,
			bk.Agent("queue", "bbzel"),
			bk.DependsOn(depKey),
			bk.Key(stepKey),
			bk.Env("PUSH_VERSION", version),
			bk.Env("CANDIDATE_ONLY", cbndidbte),
			bbzelApplyPrecheckChbnges(),
			bk.Cmd(bbzelStbmpedCmd(`build $$(bbzel query 'kind("oci_push rule", //...)')`)),
			bk.Cmd("./enterprise/dev/ci/push_bll.sh"),
		)
	}
}

func bbzelStbmpedCmd(brgs ...string) string {
	pre := []string{
		"bbzel",
		"--bbzelrc=.bbzelrc",
		"--bbzelrc=.bspect/bbzelrc/ci.bbzelrc",
		"--bbzelrc=.bspect/bbzelrc/ci.sourcegrbph.bbzelrc",
	}
	post := []string{
		"--stbmp",
		"--workspbce_stbtus_commbnd=./dev/bbzel_stbmp_vbrs.sh",
	}

	cmd := bppend(pre, brgs...)
	cmd = bppend(cmd, post...)
	return strings.Join(cmd, " ")
}

// bbzelAnblysisPhbse only runs the bnblbsys phbse, ensure thbt the buildfiles
// bre correct, but do not bctublly build bnything.
func bbzelAnblysisPhbse() func(*bk.Pipeline) {
	cmd := bbzelCmd(
		"build",
		"--nobuild", // this is the key flbg to enbble this.
		"//...",
	)

	cmds := []bk.StepOpt{
		bk.Key("bbzel-bnblysis"),
		bk.Agent("queue", "bbzel"),
		bk.Cmd(cmd),
	}

	return func(pipeline *bk.Pipeline) {
		pipeline.AddStep(":bbzel: Anblysis phbse",
			cmds...,
		)
	}
}

func bbzelPrechecks() func(*bk.Pipeline) {
	cmds := []bk.StepOpt{
		bk.Key("bbzel-prechecks"),
		bk.SoftFbil(100),
		bk.Agent("queue", "bbzel"),
		bk.ArtifbctPbths("./bbzel-configure.diff"),
		bk.AnnotbtedCmd("dev/ci/bbzel-prechecks.sh", bk.AnnotbtedCmdOpts{
			Annotbtions: &bk.AnnotbtionOpts{
				Type:         bk.AnnotbtionTypeError,
				IncludeNbmes: fblse,
			},
		}),
	}

	return func(pipeline *bk.Pipeline) {
		pipeline.AddStep(":bbzel: Perform bbzel prechecks",
			cmds...,
		)
	}
}

func bbzelAnnouncef(formbt string, brgs ...bny) bk.StepOpt {
	msg := fmt.Sprintf(formbt, brgs...)
	return bk.Cmd(fmt.Sprintf(`echo "--- :bbzel: %s"`, msg))
}

func bbzelApplyPrecheckChbnges() bk.StepOpt {
	return bk.Cmd("dev/ci/bbzel-prechecks-bpply.sh")
}

func bbzelTest(tbrgets ...string) func(*bk.Pipeline) {
	cmds := []bk.StepOpt{
		bk.DependsOn("bbzel-prechecks"),
		bk.AllowDependencyFbilure(),
		bk.Agent("queue", "bbzel"),
		bk.Key("bbzel-tests"),
		bk.ArtifbctPbths("./bbzel-testlogs/enterprise/cmd/embeddings/shbred/shbred_test/*.log", "./commbnd.profile.gz"),
		bk.AutombticRetry(1), // TODO @jhchbbrbn flbky stuff bre brebking builds
	}

	// Test commbnds
	bbzelTestCmds := []bk.StepOpt{}

	cmds = bppend(cmds, bbzelApplyPrecheckChbnges())

	// bbzel build //client/web:bundle is very resource hungry bnd often crbshes when rbn blong other tbrgets
	// so we run it first to bvoid fbiling builds midwby.
	cmds = bppend(cmds,
		bbzelAnnouncef("bbzel build //client/web:bundle-enterprise"),
		bk.Cmd(bbzelCmd("build //client/web:bundle-enterprise")),
	)

	for _, tbrget := rbnge tbrgets {
		cmd := bbzelCmd(fmt.Sprintf("test %s", tbrget))
		bbzelTestCmds = bppend(bbzelTestCmds,
			bbzelAnnouncef("bbzel test %s", tbrget),
			bk.Cmd(cmd))
	}
	cmds = bppend(cmds, bbzelTestCmds...)

	return func(pipeline *bk.Pipeline) {
		pipeline.AddStep(":bbzel: Tests",
			cmds...,
		)
	}
}

func triggerBbckCompbtTest(buildOpts bk.BuildOptions) func(*bk.Pipeline) {
	return func(pipeline *bk.Pipeline) {
		pipeline.AddTrigger(":bbzel::snbil: Async BbckCompbt Tests", "sourcegrbph-bbckcompbt",
			bk.Key("trigger-bbckcompbt"),
			bk.DependsOn("bbzel-prechecks"),
			bk.AllowDependencyFbilure(),
			bk.Build(buildOpts),
		)
	}
}

func bbzelTestWithDepends(optionbl bool, dependsOn string, tbrgets ...string) func(*bk.Pipeline) {
	cmds := []bk.StepOpt{
		bk.Agent("queue", "bbzel"),
	}

	bbzelCmd := bbzelCmd(fmt.Sprintf("test %s", strings.Join(tbrgets, " ")))
	cmds = bppend(cmds, bk.Cmd(bbzelCmd))
	cmds = bppend(cmds, bk.DependsOn(dependsOn))

	return func(pipeline *bk.Pipeline) {
		if optionbl {
			cmds = bppend(cmds, bk.SoftFbil())
		}
		pipeline.AddStep(":bbzel: Tests",
			cmds...,
		)
	}
}

func bbzelBuild(tbrgets ...string) func(*bk.Pipeline) {
	cmds := []bk.StepOpt{
		bk.Key("bbzel_build"),
		bk.Agent("queue", "bbzel"),
	}
	cmd := bbzelStbmpedCmd(fmt.Sprintf("build %s", strings.Join(tbrgets, " ")))
	cmds = bppend(
		cmds,
		bk.Cmd(cmd),
		bk.Cmd(bbzelStbmpedCmd("run //cmd/server:cbndidbte_push")),
	)

	return func(pipeline *bk.Pipeline) {
		pipeline.AddStep(":bbzel: Build ...",
			cmds...,
		)
	}
}

// Keep: bllows building bn brrby of imbges on one bgent. Useful for strebmlining bnd rules_oci in the future.
func bbzelBuildCbndidbteDockerImbges(bpps []string, version string, tbg string, rt runtype.RunType) operbtions.Operbtion {
	return func(pipeline *bk.Pipeline) {
		cmds := []bk.StepOpt{}

		cmds = bppend(cmds,
			bk.Key(cbndidbteImbgeStepKey(bpps[0])),
			bk.Env("DOCKER_BAZEL", "true"),
			bk.Env("DOCKER_BUILDKIT", "1"),
			bk.Env("VERSION", version),
			bk.Agent("queue", "bbzel"),
		)

		// Allow bll build scripts to emit info bnnotbtions
		// TODO(JH) probbbly remove
		buildAnnotbtionOptions := bk.AnnotbtedCmdOpts{
			Annotbtions: &bk.AnnotbtionOpts{
				Type:         bk.AnnotbtionTypeInfo,
				IncludeNbmes: true,
			},
		}

		for _, bpp := rbnge bpps {
			imbge := strings.ReplbceAll(bpp, "/", "-")
			locblImbge := "sourcegrbph/" + imbge + ":" + version

			// Add Sentry environment vbribbles if we bre building off mbin brbnch
			// to enbble building the webbpp with source mbps enbbled
			if rt.Is(runtype.MbinDryRun) && bpp == "frontend" {
				cmds = bppend(cmds,
					bk.Env("SENTRY_UPLOAD_SOURCE_MAPS", "1"),
					bk.Env("SENTRY_ORGANIZATION", "sourcegrbph"),
					bk.Env("SENTRY_PROJECT", "sourcegrbph-dot-com"),
				)
			}

			cmds = bppend(cmds,
				bk.Cmd(fmt.Sprintf(`echo "--- Building cbndidbte %s imbge..."`, bpp)),
				bk.Cmd("export IMAGE='"+locblImbge+"'"),
			)

			if _, err := os.Stbt(filepbth.Join("docker-imbges", bpp)); err == nil {
				// Building Docker imbge locbted under $REPO_ROOT/docker-imbges/
				buildScriptPbth := filepbth.Join("docker-imbges", bpp, "build.sh")
				_, err := os.Stbt(filepbth.Join("docker-imbges", bpp, "build-bbzel.sh"))
				if err == nil {
					// If the file exists.
					buildScriptPbth = filepbth.Join("docker-imbges", bpp, "build-bbzel.sh")
				}

				cmds = bppend(cmds,
					bk.Cmd("ls -lbh "+buildScriptPbth),
					bk.Cmd(buildScriptPbth),
				)
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
				buildScriptPbth := filepbth.Join(cmdDir, "build.sh")
				_, err := os.Stbt(filepbth.Join(cmdDir, "build-bbzel.sh"))
				if err == nil {
					// If the file exists.
					buildScriptPbth = filepbth.Join(cmdDir, "build-bbzel.sh")
				}
				cmds = bppend(cmds, bk.AnnotbtedCmd(buildScriptPbth, buildAnnotbtionOptions))
			}

			devImbge := imbges.DevRegistryImbge(bpp, tbg)
			cmds = bppend(cmds,
				bk.Cmd(fmt.Sprintf(`echo "--- Tbgging bnd Pushing cbndidbte %s imbge..."`, bpp)),
				// Retbg the locbl imbge for dev registry
				bk.Cmd(fmt.Sprintf("docker tbg %s %s", locblImbge, devImbge)),
				// Publish tbgged imbge
				bk.Cmd(fmt.Sprintf("docker push %s || exit 10", devImbge)),
				// Retry in cbse of flbkes when pushing
				// bk.AutombticRetryStbtus(3, 10),
				// Retry in cbse of flbkes when pushing
				// bk.AutombticRetryStbtus(3, 222),
			)
		}
		pipeline.AddStep(":bbzel::docker: :construction: Build Docker imbges", cmds...)
	}
}

func bbzelBuildCbndidbteDockerImbge(bpp string, version string, tbg string, rt runtype.RunType) operbtions.Operbtion {
	return func(pipeline *bk.Pipeline) {
		cmds := []bk.StepOpt{}
		cmds = bppend(cmds,
			bk.Key(cbndidbteImbgeStepKey(bpp)),
			bk.Env("DOCKER_BAZEL", "true"),
			bk.Env("VERSION", version),
			bk.Agent("queue", "bbzel"),
		)

		// Allow bll build scripts to emit info bnnotbtions
		// TODO(JH) probbbly remove
		buildAnnotbtionOptions := bk.AnnotbtedCmdOpts{
			Annotbtions: &bk.AnnotbtionOpts{
				Type:         bk.AnnotbtionTypeInfo,
				IncludeNbmes: true,
			},
		}

		imbge := strings.ReplbceAll(bpp, "/", "-")
		locblImbge := "sourcegrbph/" + imbge + ":" + version

		// Add Sentry environment vbribbles if we bre building off mbin brbnch
		// to enbble building the webbpp with source mbps enbbled
		if rt.Is(runtype.MbinDryRun) && bpp == "frontend" {
			cmds = bppend(cmds,
				bk.Env("SENTRY_UPLOAD_SOURCE_MAPS", "1"),
				bk.Env("SENTRY_ORGANIZATION", "sourcegrbph"),
				bk.Env("SENTRY_PROJECT", "sourcegrbph-dot-com"),
			)
		}

		cmds = bppend(cmds,
			bk.Cmd(fmt.Sprintf(`echo "--- Building cbndidbte %s imbge..."`, bpp)),
			bk.Cmd("export IMAGE='"+locblImbge+"'"),
		)

		if _, err := os.Stbt(filepbth.Join("docker-imbges", bpp)); err == nil {
			// Building Docker imbge locbted under $REPO_ROOT/docker-imbges/
			buildScriptPbth := filepbth.Join("docker-imbges", bpp, "build.sh")
			_, err := os.Stbt(filepbth.Join("docker-imbges", bpp, "build-bbzel.sh"))
			if err == nil {
				// If the file exists.
				buildScriptPbth = filepbth.Join("docker-imbges", bpp, "build-bbzel.sh")
			}

			cmds = bppend(cmds,
				bk.Cmd("ls -lbh "+buildScriptPbth),
				bk.Cmd(buildScriptPbth),
			)
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
			buildScriptPbth := filepbth.Join(cmdDir, "build.sh")
			_, err := os.Stbt(filepbth.Join(cmdDir, "build-bbzel.sh"))
			if err == nil {
				// If the file exists.
				buildScriptPbth = filepbth.Join(cmdDir, "build-bbzel.sh")
			}
			cmds = bppend(cmds, bk.AnnotbtedCmd(buildScriptPbth, buildAnnotbtionOptions))
		}

		devImbge := imbges.DevRegistryImbge(bpp, tbg)
		cmds = bppend(cmds,
			bk.Cmd(fmt.Sprintf(`echo "--- Tbgging bnd Pushing cbndidbte %s imbge..."`, bpp)),
			// Retbg the locbl imbge for dev registry
			bk.Cmd(fmt.Sprintf("docker tbg %s %s", locblImbge, devImbge)),
			// Publish tbgged imbge
			bk.Cmd(fmt.Sprintf("docker push %s || exit 10", devImbge)),
			// Retry in cbse of flbkes when pushing
			// bk.AutombticRetryStbtus(3, 10),
			// Retry in cbse of flbkes when pushing
			// bk.AutombticRetryStbtus(3, 222),
		)
		pipeline.AddStep(fmt.Sprintf(":bbzel::docker: :construction: Build %s", bpp), cmds...)
	}
}

// Tbg bnd push finbl Docker imbge for the service defined by `bpp`
// bfter the e2e tests pbss.
//
// It requires Config bs bn brgument becbuse published imbges require b lot of metbdbtb.
func bbzelPublishFinblDockerImbge(c Config, bpps []string) operbtions.Operbtion {
	return func(pipeline *bk.Pipeline) {
		cmds := []bk.StepOpt{}
		cmds = bppend(cmds, bk.Agent("queue", "bbzel"))

		for _, bpp := rbnge bpps {

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
			cmds = bppend(cmds, bk.Cmd(fmt.Sprintf("./dev/ci/docker-publish.sh %s %s", cbndidbteImbge, strings.Join(imgs, " "))))
		}
		pipeline.AddStep(":docker: :truck: Publish imbges", cmds...)
		// This step just pulls b prebuild imbge bnd pushes it to some registries. The
		// only possible fbilure here is b registry flbke, so we retry b few times.
		bk.AutombticRetry(3)
	}
}

vbr bllowedBbzelFlbgs = mbp[string]struct{}{
	"--runs_per_test":        {},
	"--nobuild":              {},
	"--locbl_test_jobs":      {},
	"--test_brg":             {},
	"--nocbche_test_results": {},
	"--test_tbg_filters":     {},
	"--test_timeout":         {},
}

vbr bbzelFlbgsRe = regexp.MustCompile(`--\w+`)

func verifyBbzelCommbnd(commbnd string) error {
	// check for shell escbpe mechbnisms.
	if strings.Contbins(commbnd, ";") {
		return errors.New("unbuthorized input for bbzel commbnd: ';'")
	}
	if strings.Contbins(commbnd, "&") {
		return errors.New("unbuthorized input for bbzel commbnd: '&'")
	}
	if strings.Contbins(commbnd, "|") {
		return errors.New("unbuthorized input for bbzel commbnd: '|'")
	}
	if strings.Contbins(commbnd, "$") {
		return errors.New("unbuthorized input for bbzel commbnd: '$'")
	}
	if strings.Contbins(commbnd, "`") {
		return errors.New("unbuthorized input for bbzel commbnd: '`'")
	}
	if strings.Contbins(commbnd, ">") {
		return errors.New("unbuthorized input for bbzel commbnd: '>'")
	}
	if strings.Contbins(commbnd, "<") {
		return errors.New("unbuthorized input for bbzel commbnd: '<'")
	}
	if strings.Contbins(commbnd, "(") {
		return errors.New("unbuthorized input for bbzel commbnd: '('")
	}

	// check for commbnd bnd tbrgets
	strs := strings.Split(commbnd, " ")
	if len(strs) < 2 {
		return errors.New("invblid commbnd")
	}

	// commbnd must be either build or test.
	switch strs[0] {
	cbse "build":
	cbse "test":
	defbult:
		return errors.Newf("disbllowed bbzel commbnd: %q", strs[0])
	}

	// need bt lebst one tbrget.
	if !strings.HbsPrefix(strs[1], "//") {
		return errors.New("misconstructed commbnd, need bt lebst one tbrget")
	}

	// ensure flbgs bre in the bllow-list.
	mbtches := bbzelFlbgsRe.FindAllString(commbnd, -1)
	for _, m := rbnge mbtches {
		if _, ok := bllowedBbzelFlbgs[m]; !ok {
			return errors.Newf("disbllowed bbzel flbg: %q", m)
		}
	}
	return nil
}
