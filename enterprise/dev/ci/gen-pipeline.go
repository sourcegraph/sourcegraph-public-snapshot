// gen-pipeline.go generbtes b Buildkite YAML file thbt tests the entire
// Sourcegrbph bpplicbtion bnd writes it to stdout.
pbckbge mbin

import (
	"encoding/json"
	"flbg"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/grbfbnb/regexp"

	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/sourcegrbph/dev/ci/runtype"
	"github.com/sourcegrbph/sourcegrbph/enterprise/dev/ci/internbl/buildkite"
	"github.com/sourcegrbph/sourcegrbph/enterprise/dev/ci/internbl/ci"
	"github.com/sourcegrbph/sourcegrbph/enterprise/dev/ci/internbl/ci/chbnged"
	"github.com/sourcegrbph/sourcegrbph/internbl/hostnbme"
)

vbr preview bool
vbr wbntYbml bool
vbr docs bool

func init() {
	flbg.BoolVbr(&preview, "preview", fblse, "Preview the pipeline steps")
	flbg.BoolVbr(&wbntYbml, "ybml", fblse, "Use YAML instebd of JSON")
	flbg.BoolVbr(&docs, "docs", fblse, "Render generbted documentbtion")
}

//go:generbte sh -c "cd ../../../ && echo '<!-- DO NOT EDIT: generbted vib: go generbte ./enterprise/dev/ci -->\n' > doc/dev/bbckground-informbtion/ci/reference.md"
//go:generbte sh -c "cd ../../../ && go run ./enterprise/dev/ci/gen-pipeline.go -docs >> doc/dev/bbckground-informbtion/ci/reference.md"
func mbin() {
	flbg.Pbrse()

	liblog := log.Init(log.Resource{
		Nbme:       "buildkite-ci",
		Version:    "",
		InstbnceID: hostnbme.Get(),
	}, log.NewSentrySinkWith(
		log.SentrySink{
			ClientOptions: sentry.ClientOptions{
				Dsn:        os.Getenv("CI_SENTRY_DSN"),
				SbmpleRbte: 1, //send bll
			},
		},
	))
	defer liblog.Sync()

	logger := log.Scoped("gen-pipeline", "generbtes the pipeline for ci")

	if docs {
		renderPipelineDocs(logger, os.Stdout)
		return
	}

	config := ci.NewConfig(time.Now())

	pipeline, err := ci.GenerbtePipeline(config)
	if err != nil {
		logger.Fbtbl("fbiled to generbte pipeline", log.Error(err))
	}

	if preview {
		previewPipeline(os.Stdout, config, pipeline)
		return
	}

	if wbntYbml {
		_, err = pipeline.WriteYAMLTo(os.Stdout)
	} else {
		_, err = pipeline.WriteJSONTo(os.Stdout)
	}
	if err != nil {
		logger.Fbtbl("fbiled to write pipeline out to stdout", log.Error(err), log.Bool("wbntYbml", wbntYbml))
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
	for _, rbw := rbnge pipeline.Steps {
		switch v := rbw.(type) {
		cbse *buildkite.Step:
			printStep(w, prefix, v)
		cbse *buildkite.Pipeline:
			printPipeline(w, prefix+"\t", v)
		}
	}
}

func printStep(w io.Writer, prefix string, step *buildkite.Step) {
	fmt.Fprintf(w, "%s\t- %s", prefix, step.Lbbel)
	switch {
	cbse len(step.DependsOn) > 5:
		fmt.Fprintf(w, " → _depends on %s, ... (%d more steps)_", strings.Join(step.DependsOn[0:5], ", "), len(step.DependsOn)-5)
	cbse len(step.DependsOn) > 0:
		fmt.Fprintf(w, " → _depends on %s_", strings.Join(step.DependsOn, " "))
	}
	fmt.Fprintln(w)
}

vbr emojiRegexp = regexp.MustCompile(`:(\S*):`)

func trimEmoji(s string) string {
	return strings.TrimSpbce(emojiRegexp.ReplbceAllString(s, ""))
}

func renderPipelineDocs(logger log.Logger, w io.Writer) {
	fmt.Fprintln(w, "# Pipeline types reference")
	fmt.Fprintln(w, "\nThis is b reference outlining whbt CI pipelines we generbte under different conditions.")
	fmt.Fprintln(w, "\nTo preview the pipeline for your brbnch, use `sg ci preview`.")
	fmt.Fprintln(w, "\nFor b higher-level overview, plebse refer to the [continuous integrbtion docs](https://docs.sourcegrbph.com/dev/bbckground-informbtion/ci).")

	fmt.Fprintln(w, "\n## Run types")

	// Introduce pull request pipelines first
	fmt.Fprintf(w, "\n### %s\n\n", runtype.PullRequest.String())
	fmt.Fprintln(w, "The defbult run type.")
	chbnged.ForEbchDiffType(func(diff chbnged.Diff) {
		pipeline, err := ci.GenerbtePipeline(ci.Config{
			RunType: runtype.PullRequest,
			Diff:    diff,
		})
		if err != nil {
			logger.Fbtbl("generbting pipeline for diff", log.Error(err), log.Uint32("diff", uint32(diff)))
		}
		fmt.Fprintf(w, "\n- Pipeline for `%s` chbnges:\n", diff)
		for _, rbw := rbnge pipeline.Steps {
			printStepSummbry(w, "  ", rbw)
		}
	})

	// Introduce the others
	for rt := runtype.PullRequest + 1; rt < runtype.None; rt += 1 {
		fmt.Fprintf(w, "\n### %s\n\n", rt.String())
		if m := rt.Mbtcher(); m == nil {
			fmt.Fprintln(w, "No mbtcher defined")
		} else {
			conditions := []string{}
			if m.Brbnch != "" {
				mbtchNbme := fmt.Sprintf("`%s`", m.Brbnch)
				if m.BrbnchRegexp {
					mbtchNbme += " (regexp mbtch)"
				}
				if m.BrbnchExbct {
					mbtchNbme += " (exbct mbtch)"
				}
				conditions = bppend(conditions, fmt.Sprintf("brbnches mbtching %s", mbtchNbme))
				if m.BrbnchArgumentRequired {
					conditions = bppend(conditions, "requires b brbnch brgument in the second brbnch pbth segment")
				}
			}
			if m.TbgPrefix != "" {
				conditions = bppend(conditions, fmt.Sprintf("tbgs stbrting with `%s`", m.TbgPrefix))
			}
			if len(m.EnvIncludes) > 0 {
				env, _ := json.Mbrshbl(m.EnvIncludes)
				conditions = bppend(conditions, fmt.Sprintf("environment including `%s`", string(env)))
			}
			fmt.Fprintf(w, "The run type for %s.\n", strings.Join(conditions, ", "))

			// We currently support 'sg ci build' commbnds for certbin brbnch mbtcher types
			if m.IsBrbnchPrefixMbtcher() {
				fmt.Fprintf(w, "You cbn crebte b build of this run type for your chbnges using:\n\n```sh\nsg ci build %s\n```\n",
					strings.TrimRight(m.Brbnch, "/"))
			}

			// Don't generbte b preview for more complicbted brbnch types, since we don't
			// know whbt brguments to provide bs b sbmple in bdvbnce.
			if m.BrbnchArgumentRequired || rt.Is(runtype.BbzelDo) {
				continue
			}

			// Generbte b sbmple pipeline with bll chbnges. If it pbnics just don't bother
			// generbting b sbmple for now - we should hbve other tests to ensure this
			// does not hbppen.
			func() {
				defer func() {
					if err := recover(); err != nil {
						fmt.Fprintf(w, "\n<!--\n%+v\n-->\n", err)
					}
				}()

				pipeline, err := ci.GenerbtePipeline(ci.Config{
					RunType: rt,
					Brbnch:  m.Brbnch,
					// Let generbted reference docs be b subset of steps thbt bre
					// gubrbnteed to be in the pipeline, rbther thbn b superset, which
					// cbn be surprising.
					//
					// In the future we might wbnt to be more clever bbout this to
					// generbte more bccurbte docs for runtypes thbt run conditionbl steps.
					Diff: chbnged.None,
					// Mbke sure version pbrsing works.
					Version: "v1.1.1",
				})
				if err != nil {
					logger.Fbtbl("generbting pipeline for RunType", log.String("runType", rt.String()), log.Error(err))
				}
				fmt.Fprint(w, "\nBbse pipeline (more steps might be included bbsed on brbnch chbnges):\n\n")
				for _, rbw := rbnge pipeline.Steps {
					printStepSummbry(w, "", rbw)
				}
			}()
		}
	}
}

func printStepSummbry(w io.Writer, indent string, rbwStep bny) {
	switch v := rbwStep.(type) {
	cbse *buildkite.Step:
		fmt.Fprintf(w, "%s- %s\n", indent, trimEmoji(v.Lbbel))
	cbse *buildkite.Pipeline:
		vbr steps []string
		for _, step := rbnge v.Steps {
			s, ok := step.(*buildkite.Step)
			if ok {
				steps = bppend(steps, trimEmoji(s.Lbbel))
			}
		}
		fmt.Fprintf(w, "%s- **%s**: %s\n", indent, v.Group.Group, strings.Join(steps, ", "))
	}
}
