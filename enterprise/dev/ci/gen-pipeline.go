// gen-pipeline.go generates a Buildkite YAML file that tests the entire
// Sourcegraph application and writes it to stdout.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/dev/ci/runtype"
	"github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/buildkite"
	"github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/ci"
	"github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/ci/changed"
)

var preview bool
var wantYaml bool
var docs bool

func init() {
	flag.BoolVar(&preview, "preview", false, "Preview the pipeline steps")
	flag.BoolVar(&wantYaml, "yaml", false, "Use YAML instead of JSON")
	flag.BoolVar(&docs, "docs", false, "Render generated documentation")
}

func main() {
	flag.Parse()

	if docs {
		renderPipelineDocs(os.Stdout)
		return
	}

	config := ci.NewConfig(time.Now())

	// For the time being, we are running main builds in // of the normal builds in
	// the stateless agents queue, in order to observe its stability.
	if buildkite.FeatureFlags.StatelessBuild {
		// We do not want to trigger any deployment.
		config.RunType = runtype.MainDryRun
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
	fmt.Fprintf(w, "- **Detected run type:** %s\n", c.RunType.String())
	fmt.Fprintf(w, "- **Detected diffs:** %s\n", c.Diff.String())
	fmt.Fprintf(w, "- **Computed build steps:**\n")
	printPipeline(w, "", pipeline)
}

func printPipeline(w io.Writer, prefix string, pipeline *buildkite.Pipeline) {
	if pipeline.Group.Group != "" {
		fmt.Fprintf(w, "%s- **%s**\n", prefix, pipeline.Group.Group)
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
	fmt.Fprintf(w, "%s\t- %s", prefix, step.Label)
	switch {
	case len(step.DependsOn) > 5:
		fmt.Fprintf(w, " → _depends on %s, ... (%d more steps)_", strings.Join(step.DependsOn[0:5], ", "), len(step.DependsOn)-5)
	case len(step.DependsOn) > 0:
		fmt.Fprintf(w, " → _depends on %s_", strings.Join(step.DependsOn, " "))
	}
	fmt.Fprintln(w)
}

var emojiRegexp = regexp.MustCompile(`:(\S*):`)

func trimEmoji(s string) string {
	return strings.TrimSpace(emojiRegexp.ReplaceAllString(s, ""))
}

func renderPipelineDocs(w io.Writer) {
	fmt.Fprintln(w, "# Pipeline reference")
	fmt.Fprintln(w, "\n## Run types")

	// Introduce pull request builds first
	fmt.Fprintf(w, "\n### %s\n\n", runtype.PullRequest.String())
	fmt.Fprintln(w, "The default run type.")
	changed.ForEachDiffType(func(diff changed.Diff) {
		pipeline, err := ci.GeneratePipeline(ci.Config{
			RunType: runtype.PullRequest,
			Diff:    diff,
		})
		if err != nil {
			log.Fatalf("Generating pipeline for diff %q: %s", diff, err)
		}
		fmt.Fprintf(w, "\n- Pipeline for `%s` changes:\n", diff)
		for _, raw := range pipeline.Steps {
			printStepSummary(w, raw)
		}
	})

	// Introduce the others
	for rt := runtype.PullRequest + 1; rt < runtype.None; rt += 1 {
		fmt.Fprintf(w, "\n### %s\n\n", rt.String())
		if m := rt.Matcher(); m == nil {
			fmt.Fprintln(w, "No matcher defined")
		} else {
			conditions := []string{}
			if m.Branch != "" {
				matchName := fmt.Sprintf("`%s`", m.Branch)
				if m.BranchRegexp {
					matchName += " (regexp)"
				}
				if m.BranchExact {
					matchName += " (exact)"
				}
				conditions = append(conditions, fmt.Sprintf("branches matching %s", matchName))
			}
			if m.TagPrefix != "" {
				conditions = append(conditions, fmt.Sprintf("tags starting with `%s`", m.TagPrefix))
			}
			if len(m.EnvIncludes) > 0 {
				env, _ := json.Marshal(m.EnvIncludes)
				conditions = append(conditions, fmt.Sprintf("environment including `%s`", string(env)))
			}
			fmt.Fprintf(w, "The run type for %s.\n", strings.Join(conditions, ", "))

			pipeline, err := ci.GeneratePipeline(ci.Config{
				RunType: runtype.PullRequest,
				Diff:    changed.All,
				Branch:  m.Branch,
			})
			if err != nil {
				log.Fatalf("Generating pipeline for RunType %q: %s", rt.String(), err)
			}
			fmt.Fprintln(w, "\n- Default pipeline:")
			for _, raw := range pipeline.Steps {
				printStepSummary(w, raw)
			}
		}
	}
}

func printStepSummary(w io.Writer, rawStep interface{}) {
	switch v := rawStep.(type) {
	case *buildkite.Step:
		fmt.Fprintf(w, "  - %s\n", trimEmoji(v.Label))
	case *buildkite.Pipeline:
		var steps []string
		for _, step := range v.Steps {
			s, ok := step.(*buildkite.Step)
			if ok {
				steps = append(steps, trimEmoji(s.Label))
			}
		}
		fmt.Fprintf(w, "  - **%s**: %s\n", v.Group.Group, strings.Join(steps, ", "))
	}
}
