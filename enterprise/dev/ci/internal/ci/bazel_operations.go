package ci

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/sourcegraph/enterprise/dev/ci/images"
	bk "github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/buildkite"
	"github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/ci/operations"
)

const bazelRemoteCacheURL = "https://storage.googleapis.com/sourcegraph_bazel_cache"

func BazelOperations(optional bool) *operations.Set {
	ops := operations.NewNamedSet("Bazel")
	ops.Append(bazelConfigure(optional))
	ops.Append(bazelTest(optional, "//..."))
	return ops
}

func bazelRawCmd(args ...string) string {
	pre := []string{
		"bazel",
		"--bazelrc=.bazelrc",
		"--bazelrc=.aspect/bazelrc/ci.bazelrc",
		"--bazelrc=.aspect/bazelrc/ci.sourcegraph.bazelrc",
	}

	post := []string{
		"--remote_cache=$$CI_BAZEL_REMOTE_CACHE",
		"--google_credentials=/mnt/gcloud-service-account/gcloud-service-account.json",
	}

	rawCmd := append(pre, args...)
	rawCmd = append(rawCmd, post...)
	return strings.Join(rawCmd, " ")
}

// bazelAnalysisPhase only runs the analasys phase, ensure that the buildfiles
// are correct, but do not actually build anything.
func bazelAnalysisPhase(optional bool) func(*bk.Pipeline) {
	// We run :gazelle since 'configure' causes issues on CI, where it doesn't have the go path available
	cmd := bazelRawCmd(
		"build",
		"--nobuild", // this is the key flag to enable this.
		"//...",
	)

	cmds := []bk.StepOpt{
		bk.Key("bazel-analysis"),
		bk.Env("CI_BAZEL_REMOTE_CACHE", bazelRemoteCacheURL),
		bk.Agent("queue", "bazel"),
		bk.RawCmd(cmd),
	}

	return func(pipeline *bk.Pipeline) {
		if optional {
			cmds = append(cmds, bk.SoftFail())
		}

		pipeline.AddStep(":bazel: Analysis phase",
			cmds...,
		)
	}
}
func bazelConfigure(optional bool) func(*bk.Pipeline) {
	// We run :gazelle since 'configure' causes issues on CI, where it doesn't have the go path available
	cmds := []bk.StepOpt{
		bk.Key("bazel-configure"),
		bk.Env("CI_BAZEL_REMOTE_CACHE", bazelRemoteCacheURL),
		bk.Agent("queue", "bazel"),
		bk.AnnotatedCmd("dev/ci/bazel-configure.sh", bk.AnnotatedCmdOpts{
			Annotations: &bk.AnnotationOpts{
				Type:         bk.AnnotationTypeWarning,
				IncludeNames: false,
			},
		}),
	}

	return func(pipeline *bk.Pipeline) {
		if optional {
			cmds = append(cmds, bk.SoftFail())
		}
		pipeline.AddStep(":bazel: Ensure buildfiles are up to date",
			cmds...,
		)
	}
}

// bazelBuildAndTest will perform a build and test on the same agent, which is the preferred method
// over running them in two separate jobs, so we don't build the same code twice.
func bazelBuildAndTest(optional bool, targets ...string) func(*bk.Pipeline) {
	cmds := []bk.StepOpt{
		bk.DependsOn("bazel-configure"),
		bk.Env("CI_BAZEL_REMOTE_CACHE", bazelRemoteCacheURL),
		bk.Agent("queue", "bazel"),
	}

	ts := strings.Join(targets, " ")
	bazelBuildCmd := bazelRawCmd(fmt.Sprintf("build %s", ts))
	bazelTestCmd := bazelRawCmd(
		fmt.Sprintf("test %s", ts),
		"--remote_cache=$$CI_BAZEL_REMOTE_CACHE",
		"--google_credentials=/mnt/gcloud-service-account/gcloud-service-account.json",
	)

	cmds = append(
		cmds,
		bk.RawCmd(bazelBuildCmd),
		bk.RawCmd(bazelTestCmd),
	)

	return func(pipeline *bk.Pipeline) {
		if optional {
			cmds = append(cmds, bk.SoftFail())
		}
		pipeline.AddStep(":bazel: Build && Test",
			cmds...,
		)
	}
}

func bazelTest(optional bool, targets ...string) func(*bk.Pipeline) {
	cmds := []bk.StepOpt{
		bk.Env("CI_BAZEL_REMOTE_CACHE", bazelRemoteCacheURL),
		bk.DependsOn("bazel-configure"),
		bk.Agent("queue", "bazel"),
	}

	bazelRawCmd := bazelRawCmd(fmt.Sprintf("test %s", strings.Join(targets, " ")))
	cmds = append(cmds, bk.RawCmd(bazelRawCmd))

	return func(pipeline *bk.Pipeline) {
		if optional {
			cmds = append(cmds, bk.SoftFail())
		}
		pipeline.AddStep(":bazel: Tests",
			cmds...,
		)
	}
}

func bazelTestWithDepends(optional bool, dependsOn string, targets ...string) func(*bk.Pipeline) {
	cmds := []bk.StepOpt{
		bk.Env("CI_BAZEL_REMOTE_CACHE", bazelRemoteCacheURL),
		bk.Agent("queue", "bazel"),
	}

	bazelRawCmd := bazelRawCmd(fmt.Sprintf("test %s", strings.Join(targets, " ")))
	cmds = append(cmds, bk.RawCmd(bazelRawCmd))
	cmds = append(cmds, bk.DependsOn(dependsOn))

	return func(pipeline *bk.Pipeline) {
		if optional {
			cmds = append(cmds, bk.SoftFail())
		}
		pipeline.AddStep(":bazel: Tests",
			cmds...,
		)
	}
}

func bazelBuild(optional bool, targets ...string) func(*bk.Pipeline) {
	cmds := []bk.StepOpt{
		bk.Env("CI_BAZEL_REMOTE_CACHE", bazelRemoteCacheURL),
		bk.Agent("queue", "bazel"),
	}
	bazelRawCmd := bazelRawCmd(fmt.Sprintf("build %s", strings.Join(targets, " ")))
	cmds = append(cmds, bk.RawCmd(bazelRawCmd))

	return func(pipeline *bk.Pipeline) {
		if optional {
			cmds = append(cmds, bk.SoftFail())
		}
		pipeline.AddStep(":bazel: Build ...",
			cmds...,
		)
	}
}

func bazelBuildCandidateDockerImages(apps []string, version string, tag string, uploadSourcemaps bool) operations.Operation {
	return func(pipeline *bk.Pipeline) {
		cmds := []bk.StepOpt{}

		var key string
		for _, app := range apps {
			name := strings.ReplaceAll(app, ".", "-")
			key = key + "_" + name
		}
		key = key + ":candidate"
		cmds = append(cmds,
			bk.Key(key),
			bk.Env("DOCKER_BAZEL", "true"),
			bk.Env("VERSION", version),
			bk.Agent("queue", "bazel"),
		)

		// Add Sentry environment variables if we are building off main branch
		// to enable building the webapp with source maps enabled
		if uploadSourcemaps {
			cmds = append(cmds,
				bk.Env("SENTRY_UPLOAD_SOURCE_MAPS", "1"),
				bk.Env("SENTRY_ORGANIZATION", "sourcegraph"),
				bk.Env("SENTRY_PROJECT", "sourcegraph-dot-com"),
			)
		}

		// Allow all build scripts to emit info annotations
		// TODO(JH) probably remove
		buildAnnotationOptions := bk.AnnotatedCmdOpts{
			Annotations: &bk.AnnotationOpts{
				Type:         bk.AnnotationTypeInfo,
				IncludeNames: true,
			},
		}

		for _, app := range apps {
			image := strings.ReplaceAll(app, "/", "-")
			localImage := "sourcegraph/" + image + ":" + version

			cmds = append(cmds,
				bk.Cmd("export IMAGE "+localImage),
				bk.Cmd(fmt.Sprintf(`echo "Building candidate %s image..."`, app)),
			)

			if _, err := os.Stat(filepath.Join("docker-images", app)); err == nil {
				// Building Docker image located under $REPO_ROOT/docker-images/
				buildScriptPath := filepath.Join("docker-images", app, "build.sh")
				_, err := os.Stat(filepath.Join("docker-images", app, "build-bazel.sh"))
				if err == nil {
					// If the file exists.
					buildScriptPath = filepath.Join("docker-images", app, "build-bazel.sh")
				}

				cmds = append(cmds,
					bk.Cmd("ls -lah "+buildScriptPath),
					bk.Cmd(buildScriptPath),
				)
			} else {
				// Building Docker images located under $REPO_ROOT/cmd/
				cmdDir := func() string {
					folder := app
					if app == "blobstore2" {
						// experiment: cmd/blobstore is a Go rewrite of docker-images/blobstore. While
						// it is incomplete, we do not want cmd/blobstore/Dockerfile to get published
						// under the same name.
						// https://github.com/sourcegraph/sourcegraph/issues/45594
						// TODO(blobstore): remove this when making Go blobstore the default
						folder = "blobstore"
					}
					// If /enterprise/cmd/... does not exist, build just /cmd/... instead.
					if _, err := os.Stat(filepath.Join("enterprise/cmd", folder)); err != nil {
						return "cmd/" + folder
					}
					return "enterprise/cmd/" + folder
				}()
				// TODO(JH) we don't need this anymore.
				// preBuildScript := cmdDir + "/pre-build.sh"
				// if _, err := os.Stat(preBuildScript); err == nil {
				// 	// Allow all
				// 	cmds = append(cmds, bk.AnnotatedCmd(preBuildScript, buildAnnotationOptions))
				// }
				buildScriptPath := filepath.Join(cmdDir, "build.sh")
				_, err := os.Stat(filepath.Join(cmdDir, "build-bazel.sh"))
				if err == nil {
					// If the file exists.
					buildScriptPath = filepath.Join(cmdDir, "build-bazel.sh")
				}
				cmds = append(cmds, bk.AnnotatedCmd(buildScriptPath, buildAnnotationOptions))
			}

			devImage := images.DevRegistryImage(app, tag)
			cmds = append(cmds,
				// Retag the local image for dev registry
				bk.Cmd(fmt.Sprintf("docker tag %s %s", localImage, devImage)),
				// Publish tagged image
				bk.Cmd(fmt.Sprintf("docker push %s || exit 10", devImage)),
				// Retry in case of flakes when pushing
				// bk.AutomaticRetryStatus(3, 10),
				// Retry in case of flakes when pushing
				// bk.AutomaticRetryStatus(3, 222),
			)
		}
		pipeline.AddStep(fmt.Sprintf(":docker: :construction: Build %s", strings.Join(apps, " ")), cmds...)
	}
}
