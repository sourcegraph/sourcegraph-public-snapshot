package ci

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/ci/images"
	bk "github.com/sourcegraph/sourcegraph/dev/ci/internal/buildkite"
	"github.com/sourcegraph/sourcegraph/dev/ci/internal/ci/operations"
	"github.com/sourcegraph/sourcegraph/dev/ci/runtype"
)

// candidateImageStepKey is the key for the given app (see the `images` package). Useful for
// adding dependencies on a step.
func candidateImageStepKey(app string) string {
	return strings.ReplaceAll(app, ".", "-") + ":candidate"
}

// Tag and push final Docker image for the service defined by `app`
// after the e2e tests pass.
//
// It requires Config as an argument because published images require a lot of metadata.
func publishFinalDockerImage(c Config, app string) operations.Operation {
	return func(pipeline *bk.Pipeline) {
		devImage := images.DevRegistryImage(app, "")
		publishImage := images.PublishedRegistryImage(app, "")

		var imgs []string
		for _, image := range []string{publishImage, devImage} {
			if app != "server" || c.RunType.Is(runtype.TaggedRelease, runtype.ImagePatch, runtype.ImagePatchNoTest) {
				imgs = append(imgs, fmt.Sprintf("%s:%s", image, c.Version))
			}

			if app == "server" && c.RunType.Is(runtype.ReleaseBranch) {
				imgs = append(imgs, fmt.Sprintf("%s:%s-insiders", image, c.Branch))
			}

			if c.RunType.Is(runtype.MainBranch) {
				imgs = append(imgs, fmt.Sprintf("%s:insiders", image))
			}
		}

		// these tags are pushed to our dev registry, and are only
		// used internally
		for _, tag := range []string{
			c.Version,
			c.Commit,
			c.shortCommit(),
			fmt.Sprintf("%s_%s_%d", c.shortCommit(), c.Time.Format("2006-01-02"), c.BuildNumber),
			fmt.Sprintf("%s_%d", c.shortCommit(), c.BuildNumber),
			fmt.Sprintf("%s_%d", c.Commit, c.BuildNumber),
			strconv.Itoa(c.BuildNumber),
		} {
			internalImage := fmt.Sprintf("%s:%s", devImage, tag)
			imgs = append(imgs, internalImage)
		}

		candidateImage := fmt.Sprintf("%s:%s", devImage, c.candidateImageTag())
		cmd := fmt.Sprintf("./dev/ci/docker-publish.sh %s %s", candidateImage, strings.Join(imgs, " "))

		pipeline.AddStep(fmt.Sprintf(":docker: :truck: %s", app),
			// This step just pulls a prebuild image and pushes it to some registries. The
			// only possible failure here is a registry flake, so we retry a few times.
			bk.AutomaticRetry(3),
			bk.Cmd(cmd))
	}
}

// Used in default run type
func bazelPushImagesCandidates(version string, isAspectBuild bool) func(*bk.Pipeline) {
	depKey := "bazel-tests"
	if isAspectBuild {
		depKey = "__main__::test"
	}
	return bazelPushImagesCmd(version, true, depKey)
}

// Used in default run type
func bazelPushImagesFinal(version string, isAspectBuild bool) func(*bk.Pipeline) {
	depKey := "bazel-tests"
	if isAspectBuild {
		depKey = "__main__::test"
	}
	return bazelPushImagesCmd(version, false, depKey)
}

// Used in CandidateNoTest run type
func bazelPushImagesNoTest(version string) func(*bk.Pipeline) {
	return bazelPushImagesCmd(version, false, "pipeline-gen")
}

func bazelPushImagesCmd(version string, isCandidate bool, depKey string) func(*bk.Pipeline) {
	stepName := ":bazel::docker: Push final images"
	stepKey := "bazel-push-images"
	candidate := ""

	if isCandidate {
		stepName = ":bazel::docker: Push candidate Images"
		stepKey = stepKey + "-candidate"
		candidate = "true"
	}

	return func(pipeline *bk.Pipeline) {
		pipeline.AddStep(stepName,
			bk.Agent("queue", "bazel"),
			bk.DependsOn(depKey),
			bk.Key(stepKey),
			bk.Env("PUSH_VERSION", version),
			bk.Env("CANDIDATE_ONLY", candidate),
			bazelApplyPrecheckChanges(),
			bk.Cmd(bazelStampedCmd(`build $$(bazel query 'kind("oci_push rule", //...)')`)),
			bk.Cmd("./dev/ci/push_all.sh"),
		)
	}
}

// Tag and push final Docker image for the service defined by `app`
// after the e2e tests pass.
//
// It requires Config as an argument because published images require a lot of metadata.
func bazelPublishFinalDockerImage(c Config, apps []string) operations.Operation {
	return func(pipeline *bk.Pipeline) {
		cmds := []bk.StepOpt{}
		cmds = append(cmds, bk.Agent("queue", "bazel"))

		for _, app := range apps {

			devImage := images.DevRegistryImage(app, "")
			publishImage := images.PublishedRegistryImage(app, "")

			var imgs []string
			for _, image := range []string{publishImage, devImage} {
				if app != "server" || c.RunType.Is(runtype.TaggedRelease, runtype.ImagePatch, runtype.ImagePatchNoTest) {
					imgs = append(imgs, fmt.Sprintf("%s:%s", image, c.Version))
				}

				if app == "server" && c.RunType.Is(runtype.ReleaseBranch) {
					imgs = append(imgs, fmt.Sprintf("%s:%s-insiders", image, c.Branch))
				}

				if c.RunType.Is(runtype.MainBranch) {
					imgs = append(imgs, fmt.Sprintf("%s:insiders", image))
				}
			}

			// these tags are pushed to our dev registry, and are only
			// used internally
			for _, tag := range []string{
				c.Version,
				c.Commit,
				c.shortCommit(),
				fmt.Sprintf("%s_%s_%d", c.shortCommit(), c.Time.Format("2006-01-02"), c.BuildNumber),
				fmt.Sprintf("%s_%d", c.shortCommit(), c.BuildNumber),
				fmt.Sprintf("%s_%d", c.Commit, c.BuildNumber),
				strconv.Itoa(c.BuildNumber),
			} {
				internalImage := fmt.Sprintf("%s:%s", devImage, tag)
				imgs = append(imgs, internalImage)
			}

			candidateImage := fmt.Sprintf("%s:%s", devImage, c.candidateImageTag())
			cmds = append(cmds, bk.Cmd(fmt.Sprintf("./dev/ci/docker-publish.sh %s %s", candidateImage, strings.Join(imgs, " "))))
		}
		pipeline.AddStep(":docker: :truck: Publish images", cmds...)
		// This step just pulls a prebuild image and pushes it to some registries. The
		// only possible failure here is a registry flake, so we retry a few times.
		bk.AutomaticRetry(3)
	}
}
