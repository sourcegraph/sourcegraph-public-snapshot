package ci

import (
	"fmt"
	"strconv"

	"github.com/Masterminds/semver"
	bk "github.com/sourcegraph/sourcegraph/dev/ci/internal/buildkite"
	"github.com/sourcegraph/sourcegraph/dev/ci/internal/ci/operations"
	"github.com/sourcegraph/sourcegraph/dev/ci/runtype"
)

func bazelBuildExecutorVM(c Config, alwaysRebuild bool) operations.Operation {
	return func(pipeline *bk.Pipeline) {
		imageFamily := executorImageFamilyForConfig(c)
		stepOpts := []bk.StepOpt{
			bk.Agent("queue", "bazel"),
			bk.Key(candidateImageStepKey("executor.vm-image")),
			bk.Env("VERSION", c.Version),
			bk.Env("IMAGE_FAMILY", imageFamily),
			bk.Env("EXECUTOR_IS_TAGGED_RELEASE", strconv.FormatBool(c.RunType.Is(runtype.TaggedRelease))),
		}

		cmd := bazelStampedCmd("run //cmd/executor/vm-image:ami.build")

		if !alwaysRebuild {
			stepOpts = append(stepOpts,
				// Soft-fail with code 222 if nothing has changed
				bk.SoftFail(222),
				bk.Cmd("if ! ./cmd/executors/ci-should-rebuild.sh; then exit 222; fi"))
		}
		stepOpts = append(stepOpts, bk.Cmd(cmd))
		pipeline.AddStep(":bazel::packer: :construction: Build executor image", stepOpts...)
	}
}

func bazelPublishExecutorVM(c Config, alwaysRebuild bool) operations.Operation {
	return func(pipeline *bk.Pipeline) {
		imageFamily := executorImageFamilyForConfig(c)
		stepOpts := []bk.StepOpt{
			bk.Agent("queue", "bazel"),
			bk.DependsOn(candidateImageStepKey("executor.vm-image")),
			bk.Env("VERSION", c.Version),
			bk.Env("IMAGE_FAMILY", imageFamily),
			bk.Env("EXECUTOR_IS_TAGGED_RELEASE", strconv.FormatBool(c.RunType.Is(runtype.TaggedRelease))),
		}

		cmd := bazelStampedCmd("run //cmd/executor/vm-image:ami.push")

		if !alwaysRebuild {
			stepOpts = append(stepOpts,
				// Soft-fail with code 222 if nothing has changed
				bk.SoftFail(222),
				bk.Cmd("if ! ./cmd/executors/ci-should-rebuild.sh; then exit 222; fi"))
		}

		stepOpts = append(stepOpts, bk.Cmd(cmd))

		pipeline.AddStep(":bazel::packer: :construction: Build executor image", stepOpts...)
	}
}

func bazelPublishExecutorDockerMirror(c Config) operations.Operation {
	return func(pipeline *bk.Pipeline) {
		candidateBuildStep := candidateImageStepKey("executor-docker-miror.vm-image")
		imageFamily := executorDockerMirrorImageFamilyForConfig(c)
		stepOpts := []bk.StepOpt{
			bk.Agent("queue", "bazel"),
			bk.DependsOn(candidateBuildStep),
			bk.Env("VERSION", c.Version),
			bk.Env("IMAGE_FAMILY", imageFamily),
			bk.Env("EXECUTOR_IS_TAGGED_RELEASE", strconv.FormatBool(c.RunType.Is(runtype.TaggedRelease))),
			bk.Cmd(bazelStampedCmd("run //cmd/executor/docker-mirror:ami.push")),
		}
		pipeline.AddStep(":packer: :white_check_mark: Publish docker registry mirror image", stepOpts...)
	}
}

func bazelPublishExecutorBinary(c Config) operations.Operation {
	return func(pipeline *bk.Pipeline) {
		stepOpts := []bk.StepOpt{
			bk.Agent("queue", "bazel"),
			bk.Env("VERSION", c.Version),
			bk.Env("EXECUTOR_IS_TAGGED_RELEASE", strconv.FormatBool(c.RunType.Is(runtype.TaggedRelease))),
			// We need this while we're running this on our own bazel CI agents, as it seems that a
			// bazel run doesn't always build everything it needs for that run command.
			// This doesn't happen on aspect workflows CI agents.
			bk.Cmd(bazelCmd(`build //cmd/executor:executor`)),
			bk.Cmd(bazelStampedCmd(`run //cmd/executor:publish_binary`)),
		}
		pipeline.AddStep(":bazel::arrow_heading_up: Publish executor binary", stepOpts...)
	}
}

// executorDockerMirrorImageFamilyForConfig returns the image family to be used for the build.
// This defaults to `-nightly`, and will be `-$MAJOR-$MINOR` for a tagged release
// build.
func executorDockerMirrorImageFamilyForConfig(c Config) string {
	imageFamily := "sourcegraph-executors-docker-mirror-nightly"
	if c.RunType.Is(runtype.TaggedRelease) {
		ver, err := semver.NewVersion(c.Version)
		if err != nil {
			panic("cannot parse version")
		}
		imageFamily = fmt.Sprintf("sourcegraph-executors-docker-mirror-%d-%d", ver.Major(), ver.Minor())
	}
	return imageFamily
}

func bazelBuildExecutorDockerMirror(c Config) operations.Operation {
	return func(pipeline *bk.Pipeline) {
		imageFamily := executorDockerMirrorImageFamilyForConfig(c)
		stepOpts := []bk.StepOpt{
			bk.Agent("queue", "bazel"),
			bk.Key(candidateImageStepKey("executor-docker-miror.vm-image")),
			bk.Env("VERSION", c.Version),
			bk.Env("IMAGE_FAMILY", imageFamily),
			bk.Env("EXECUTOR_IS_TAGGED_RELEASE", strconv.FormatBool(c.RunType.Is(runtype.TaggedRelease))),
			bk.Cmd(bazelStampedCmd("run //cmd/executor/docker-mirror:ami.build")),
		}
		pipeline.AddStep(":bazel::packer: :construction: Build docker registry mirror image", stepOpts...)
	}
}

// executorImageFamilyForConfig returns the image family to be used for the build.
// This defaults to `-nightly`, and will be `-$MAJOR-$MINOR` for a tagged release
// build.
func executorImageFamilyForConfig(c Config) string {
	imageFamily := "sourcegraph-executors-TEST-nightly"
	if c.RunType.Is(runtype.TaggedRelease) {
		ver, err := semver.NewVersion(c.Version)
		if err != nil {
			panic("cannot parse version")
		}
		imageFamily = fmt.Sprintf("sourcegraph-executors-%d-%d", ver.Major(), ver.Minor())
	}
	return imageFamily
}

func executorsE2E(candidateTag string) operations.Operation {
	return func(p *bk.Pipeline) {
		p.AddStep(":bazel::docker::packer: Executors E2E",
			// Run tests against the candidate server image
			bk.DependsOn("bazel-push-images-candidate"),
			bk.Agent("queue", "bazel"),
			bk.Env("CANDIDATE_VERSION", candidateTag),
			bk.Env("SOURCEGRAPH_BASE_URL", "http://127.0.0.1:7080"),
			bk.Env("SOURCEGRAPH_SUDO_USER", "admin"),
			bk.Env("TEST_USER_EMAIL", "test@sourcegraph.com"),
			bk.Env("TEST_USER_PASSWORD", "supersecurepassword"),
			// See dev/ci/integration/executors/docker-compose.yaml
			// This enable the executor to reach the dind container
			// for docker commands.
			bk.Env("DOCKER_GATEWAY_HOST", "172.17.0.1"),
			bk.Cmd("dev/ci/integration/executors/run.sh"),
			bk.ArtifactPaths("./*.log"),
		)
	}
}
