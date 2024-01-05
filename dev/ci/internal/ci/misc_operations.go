package ci

import (
	"os"
	"strings"

	bk "github.com/sourcegraph/sourcegraph/dev/ci/internal/buildkite"
	"github.com/sourcegraph/sourcegraph/dev/ci/runtype"
)

func triggerBackCompatTest(buildOpts bk.BuildOptions, isAspectWorkflows bool) func(*bk.Pipeline) {
	if isAspectWorkflows {
		buildOpts.Message += " (Aspect)"
	}
	return func(pipeline *bk.Pipeline) {
		steps := []bk.StepOpt{
			bk.Async(true),
			bk.Key("trigger-backcompat"),
			bk.AllowDependencyFailure(),
			bk.Build(buildOpts),
		}

		if !isAspectWorkflows {
			steps = append(steps, bk.DependsOn("bazel-prechecks"))
		}
		pipeline.AddTrigger(":bazel::snail: Async BackCompat Tests", "sourcegraph-backcompat", steps...)
	}
}

func bazelGoModTidy() func(*bk.Pipeline) {
	cmds := []bk.StepOpt{
		bk.Agent("queue", "bazel"),
		bk.Key("bazel-go-mod"),
		bk.Cmd("./dev/ci/bazel-gomodtidy.sh"),
	}

	return func(pipeline *bk.Pipeline) {
		pipeline.AddStep(":bazel::broom: Go mod tidy", cmds...)
	}
}

// addSgLints runs linters for the given targets.
func addSgLints(targets []string) func(pipeline *bk.Pipeline) {
	cmd := "go run ./dev/sg "

	if retryCount := os.Getenv("BUILDKITE_RETRY_COUNT"); retryCount != "" && retryCount != "0" {
		cmd = cmd + "-v "
	}

	var (
		branch = os.Getenv("BUILDKITE_BRANCH")
		tag    = os.Getenv("BUILDKITE_TAG")
		// evaluates what type of pipeline run this is
		runType = runtype.Compute(tag, branch, map[string]string{
			"BEXT_NIGHTLY":       os.Getenv("BEXT_NIGHTLY"),
			"RELEASE_NIGHTLY":    os.Getenv("RELEASE_NIGHTLY"),
			"VSCE_NIGHTLY":       os.Getenv("VSCE_NIGHTLY"),
			"WOLFI_BASE_REBUILD": os.Getenv("WOLFI_BASE_REBUILD"),
		})
	)

	formatCheck := ""
	if runType.Is(runtype.MainBranch) || runType.Is(runtype.MainDryRun) {
		formatCheck = "--skip-format-check "
	}

	cmd = cmd + "lint -annotations -fail-fast=false " + formatCheck + strings.Join(targets, " ")

	return func(pipeline *bk.Pipeline) {
		pipeline.AddStep(":pineapple::lint-roller: Run sg lint",
			withPnpmCache(),
			bk.AnnotatedCmd(cmd, bk.AnnotatedCmdOpts{
				Annotations: &bk.AnnotationOpts{
					IncludeNames: true,
					Type:         bk.AnnotationTypeAuto,
				},
			}))
	}
}
