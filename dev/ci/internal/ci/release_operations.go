package ci

import (
	"fmt"
	"strings"

	bk "github.com/sourcegraph/sourcegraph/dev/ci/internal/buildkite"
	"github.com/sourcegraph/sourcegraph/dev/ci/runtype"

	"github.com/sourcegraph/sourcegraph/dev/ci/images"
	"github.com/sourcegraph/sourcegraph/dev/ci/internal/ci/operations"
)

// checkSecurityApproval checks whether the specified release has release approval from the Security Team.
func checkSecurityApproval(c Config) operations.Operation {
	return func(pipeline *bk.Pipeline) {
		pipeline.AddStep(":nodesecurity: Check security approval",
			bk.Agent("queue", AspectWorkflows.QueueDefault),
			bk.Env("VERSION", c.Version),
			bk.AnnotatedCmd(
				"./tools/release/check_security_approval.sh",
				bk.AnnotatedCmdOpts{
					Annotations: &bk.AnnotationOpts{
						Type:         bk.AnnotationTypeInfo,
						IncludeNames: false,
					},
				},
			),
		)
	}
}

// releasePromoteImages runs a script that iterates through all defined images that we're producing that has been uploaded
// on the internal registry with a given version and retags them to the public registry.
func releasePromoteImages(c Config) operations.Operation {
	additionalProdRegistries := strings.Join([]string{images.SourcegraphArtifactRegistryPublicRegistry, images.CloudEphemeralRegistry}, " ")
	image_args := strings.Join(images.SourcegraphDockerImages, " ")
	return func(pipeline *bk.Pipeline) {
		pipeline.AddStep("Promote release to public",
			bk.Agent("queue", AspectWorkflows.QueueDefault),
			bk.Env("VERSION", c.Version),
			bk.Env("INTERNAL_REGISTRY", images.SourcegraphInternalReleaseRegistry),
			bk.Env("PUBLIC_REGISTRY", images.SourcegraphDockerPublishRegistry),
			bk.Env("ADDITIONAL_PROD_REGISTRIES", additionalProdRegistries),
			bk.AnnotatedCmd(
				fmt.Sprintf("./tools/release/promote_images.sh %s", image_args),
				bk.AnnotatedCmdOpts{
					Annotations: &bk.AnnotationOpts{
						Type:         bk.AnnotationTypeInfo,
						IncludeNames: false,
					},
				},
			),
		)
	}
}

// releaseTestOperations runs the script defined in release.yaml that tests the release.
func releaseTestOperation(c Config) operations.Operation {
	devRegistry := images.SourcegraphDockerDevRegistry
	prodRegistry := images.SourcegraphDockerPublishRegistry

	if c.RunType.Is(runtype.InternalRelease) {
		prodRegistry = images.SourcegraphInternalReleaseRegistry
	}

	return func(pipeline *bk.Pipeline) {
		stepOpts := []bk.StepOpt{
			bk.Agent("queue", AspectWorkflows.QueueDefault),
			bk.Env("DEV_REGISTRY", devRegistry),
			bk.Env("PROD_REGISTRY", prodRegistry),
			bk.Env("VERSION", c.Version),
			bk.AnnotatedCmd(
				bazelCmd(`run --run_under="cd $$PWD &&" //dev/sg -- release run test --branch $$BUILDKITE_BRANCH --version $$VERSION --development=$$IS_DEVELOPMENT_RELEASE`),
				bk.AnnotatedCmdOpts{
					Annotations: &bk.AnnotationOpts{
						Type:         bk.AnnotationTypeInfo,
						IncludeNames: true,
					},
				},
			),
		}

		// If we're on a internal release, the release tests cannot run without
		// having pushed the images, but if we're on a promote release, we don't need
		// to wait for the push images job.
		if c.RunType.Is(runtype.InternalRelease) {
			stepOpts = append(stepOpts,
				bk.DependsOn("bazel-push-images"),
			)
		}

		pipeline.AddStep("Release tests", stepOpts...)
	}
}

// releaseFinalizeOperation runs the script defined in release.yaml that finalizes the release. It picks
// the variant (internal or public) based on the run type.
//
// Important: this helper doesn't inject the `wait` step, it's on the calling side to handle that. This is
// necessary because by definition, you want to call finalize only when everything else succeeded.
func releaseFinalizeOperation(c Config) operations.Operation {
	label := "Finalize internal release"
	command := "internal"
	if c.RunType.Is(runtype.PromoteRelease) {
		label = "Final release promotion"
		command = "promote-to-public"
	}

	return func(pipeline *bk.Pipeline) {
		pipeline.AddStep(label,
			bk.Agent("queue", AspectWorkflows.QueueDefault),
			bk.Env("VERSION", c.Version),
			bk.AnnotatedCmd(
				bazelCmd(fmt.Sprintf(`run --run_under="cd $$PWD &&" //dev/sg -- release run %s finalize --branch $$BUILDKITE_BRANCH --version $$VERSION --development=$$IS_DEVELOPMENT_RELEASE`, command)),
				bk.AnnotatedCmdOpts{
					Annotations: &bk.AnnotationOpts{
						Type:         bk.AnnotationTypeInfo,
						IncludeNames: true,
					},
				},
			))
	}
}
