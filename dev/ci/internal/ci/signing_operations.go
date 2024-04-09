package ci

import (
	bk "github.com/sourcegraph/sourcegraph/dev/ci/internal/buildkite"
	"github.com/sourcegraph/sourcegraph/dev/ci/internal/ci/operations"
)

func ContainerSigningOperations(changedFiles []string) *operations.Set {
	ops := operations.NewNamedSet("Container Signing")

	signFunc := SignContainerImages()
	ops.Append(signFunc)

	return ops
}

// Roughly how should this work?
// It makes sense to download a list of images that were pushed by the bazel-push-images step.
// Otherwise we'd need to run inside that step ourselves.

func SignContainerImages() func(*bk.Pipeline) {
	return func(pipeline *bk.Pipeline) {
		cmd := "./dev/ci/scripts/signing/sign-containers.sh"
		opts := []bk.StepOpt{
			bk.Cmd(cmd),
			bk.Agent("queue", AspectWorkflows.QueueDefault),
			bk.DependsOn("bazel-push-images"),
			bk.Key("sign-container-images"),
			bk.SoftFail(222),
		}

		pipeline.AddStep(
			"Container signing",
			opts...,
		)
	}
}
