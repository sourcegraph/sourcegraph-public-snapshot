package ci

import (
	"reflect"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/dev/ci/images"

	bk "github.com/sourcegraph/sourcegraph/internal/buildkite"
)

func Test_addDockerImages(t *testing.T) {
	crono := time.Now()
	tests := []struct {
		name  string
		c     Config
		final bool
		want  func(*bk.Pipeline)
	}{
		{
			"base, always deploy main, regardless of other fields",
			Config{
				now:                 crono,
				branch:              "main",
				version:             "na",
				commit:              "na",
				mustIncludeCommit:   nil,
				changedFiles:        nil,
				taggedRelease:       true,
				releaseBranch:       false,
				isBextReleaseBranch: false,
				isBextNightly:       false,
				isRenovateBranch:    true,
				patch:               false,
				patchNoTest:         false,
				isQuick:             false,
				isMasterDryRun:      true,
				profilingEnabled:    false,
			},
			true,
			func(pipeline *bk.Pipeline) {
				//expect insiders docker images
				c := Config{
					now:                 crono,
					branch:              "main",
					version:             "na",
					commit:              "na",
					mustIncludeCommit:   nil,
					changedFiles:        nil,
					taggedRelease:       true,
					releaseBranch:       false,
					isBextReleaseBranch: false,
					isBextNightly:       false,
					isRenovateBranch:    true,
					patch:               false,
					patchNoTest:         false,
					isQuick:             false,
					isMasterDryRun:      true,
					profilingEnabled:    false,
				}

				addDockerImage := func(c Config, app string, insiders bool) func(*bk.Pipeline) {
					return addFinalDockerImage(c, app, insiders)
				}
				for _, dockerImage := range images.SourcegraphDockerImages {
					addDockerImage(c, dockerImage, true)(pipeline)
				}
			},
		},
	}
	// these need to be outside the test scope allow for the delve to pick up changes to them
	pipeline := &bk.Pipeline{}
	sparePipeline := &bk.Pipeline{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := addDockerImages(tt.c, tt.final)

			// we create a list of functions that are applied to the pipeline so we need to mimic that here to test
			got(pipeline) //this is 100% the opposite of a pure function

			tt.want(sparePipeline)
			if !reflect.DeepEqual(pipeline, sparePipeline) {
				t.Fatal("did not generate equivalent pipelines")
			}
		})
	}
}
