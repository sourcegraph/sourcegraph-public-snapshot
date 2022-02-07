// gen-pipeline.go generates a Buildkite YAML file that tests the entire
// Sourcegraph application and writes it to stdout.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/buildkite"
	"github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/ci"
)

var preview bool
var wantYaml bool

func init() {
	flag.BoolVar(&preview, "preview", false, "Preview the pipeline steps")
	flag.BoolVar(&wantYaml, "yaml", false, "Use YAML instead of JSON")
}

func main() {
	flag.Parse()

	config := ci.NewConfig(time.Now())

	// For the time being, we are running main builds in // of the normal builds in
	// the stateless agents queue, in order to observe its stability.
	if buildkite.FeatureFlags.StatelessBuild {
		// We do not want to trigger any deployment.
		config.RunType = ci.MainDryRun
	}

	pipeline, err := ci.GeneratePipeline(config)
	if err != nil {
		panic(err)
	}

	if preview {
		previewPipeline(os.Stdout, config, pipeline)
		return
	}

	if wantYaml {
		_, err = pipeline.WriteYAMLTo(os.Stdout)
	} else {
		_, err = pipeline.WriteJSONTo(os.Stdout)
	}
	if err != nil {
		panic(err)
	}
}

func previewPipeline(w io.Writer, c ci.Config, pipeline *buildkite.Pipeline) {
	fmt.Fprintf(w, "Detected run type:\n\t%s\n", c.RunType.String())
	fmt.Fprintf(w, "Detected diffs:\n\t%s\n", c.Diff.String())
	fmt.Fprintf(w, "Computed build steps:\n")
	printPipeline(w, "", pipeline)
}

func printPipeline(w io.Writer, prefix string, pipeline *buildkite.Pipeline) {
	if pipeline.Group.Group != "" {
		fmt.Fprintf(w, "%s%s\n", prefix, pipeline.Group.Group)
	}
	for _, raw := range pipeline.Steps {
		switch v := raw.(type) {
		case *buildkite.Step:
			printStep(w, prefix, v)
		case *buildkite.Pipeline:
			printPipeline(w, prefix+"\t", v)
		}
	}
}

func printStep(w io.Writer, prefix string, step *buildkite.Step) {
	fmt.Fprintf(w, "%s\t%s\n", prefix, step.Label)
	switch {
	case len(step.DependsOn) > 5:
		fmt.Fprintf(w, "%s\t\t→ depends on %s, ... (%d more steps)\n", prefix, strings.Join(step.DependsOn[0:5], ", "), len(step.DependsOn)-5)
	case len(step.DependsOn) > 0:
		fmt.Fprintf(w, "%s\t\t→ depends on %s\n", prefix, strings.Join(step.DependsOn, " "))
	}
}
