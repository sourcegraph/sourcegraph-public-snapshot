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
	fmt.Fprintf(w, "Detected changed files (%d):\n", len(c.ChangedFiles))
	for _, f := range c.ChangedFiles {
		fmt.Fprintf(w, "\t%s\n", f)
	}

	fmt.Fprintln(w, "Detected changes:")
	for affects, doesAffects := range map[string]bool{
		"Go":                           c.ChangedFiles.AffectsGo(),
		"Client":                       c.ChangedFiles.AffectsClient(),
		"Docs":                         c.ChangedFiles.AffectsDocs(),
		"Dockerfiles":                  c.ChangedFiles.AffectsDockerfiles(),
		"GraphQL":                      c.ChangedFiles.AffectsGraphQL(),
		"CI scripts":                   c.ChangedFiles.AffectsCIScripts(),
		"Terraform":                    c.ChangedFiles.AffectsTerraformFiles(),
		"ExecutorDockerRegistryMirror": c.ChangedFiles.AffectsExecutorDockerRegistryMirror(),
	} {
		fmt.Fprintf(w, "\tAffects %s: %t\n", affects, doesAffects)
	}

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
