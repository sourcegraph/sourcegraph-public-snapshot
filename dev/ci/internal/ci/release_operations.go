package ci

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/ci/images"
	bk "github.com/sourcegraph/sourcegraph/dev/ci/internal/buildkite"
	"github.com/sourcegraph/sourcegraph/dev/ci/internal/ci/operations"
)

// promoteRFC795Images runs a script that iterates through all defined images that we're producing that has been uploaded on the internal registry with a given version and retags them to the public registry.
func promoteRFC795Images(c Config) operations.Operation {
	image_args := strings.Join(images.SourcegraphDockerImages, " ")
	return func(pipeline *bk.Pipeline) {
		pipeline.AddStep("Promote release to public",
			bk.Agent("queue", AspectWorkflows.QueueDefault),
			bk.Env("VERSION", c.Version),
			bk.Env("INTERNAL_REGISTRY", images.SourcegraphInternalReleaseRegistry),
			bk.Env("PUBLIC_REGISTRY", images.SourcegraphPublicReleaseRegistry),
			bk.AnnotatedCmd(
				fmt.Sprintf("./tools/release/promote_images.sh %s", image_args),
				bk.AnnotatedCmdOpts{
					Annotations: &bk.AnnotationOpts{
						Type:         bk.AnnotationTypeError,
						IncludeNames: false,
					},
				},
			),
		)
	}
}
