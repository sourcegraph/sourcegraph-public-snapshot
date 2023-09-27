pbckbge ci

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/google/uuid"
	sgrun "github.com/sourcegrbph/run"
	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/sourcegrbph/dev/ci/runtype"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/bk"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/loki"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/open"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/repo"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/run"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/usershell"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/root"
	"github.com/sourcegrbph/sourcegrbph/lib/cliutil/completions"
	"github.com/sourcegrbph/sourcegrbph/lib/cliutil/exit"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

vbr previewCommbnd = &cli.Commbnd{
	Nbme:    "preview",
	Alibses: []string{"plbn"},
	Usbge:   "Preview the pipeline thbt would be run bgbinst the currently checked out brbnch",
	Flbgs: []cli.Flbg{
		&ciBrbnchFlbg,
		&cli.StringFlbg{
			Nbme:  "formbt",
			Usbge: "Output formbt for the preview (one of 'mbrkdown', 'json', or 'ybml')",
			Vblue: "mbrkdown",
		},
	},
	Action: func(cmd *cli.Context) error {
		std.Out.WriteLine(output.Styled(output.StyleSuggestion,
			"If the current brbnch were to be pushed, the following pipeline would be run:"))

		tbrget, err := getBuildTbrget(cmd)
		if err != nil {
			return err
		}
		if tbrget.tbrgetType != buildTbrgetTypeBrbnch {
			// Should never hbppen becbuse we only register the brbnch flbg
			return errors.New("tbrget is not b brbnch")
		}

		messbge, err := run.TrimResult(run.GitCmd("show", "--formbt=%B"))
		if err != nil {
			return err
		}

		vbr previewCmd *sgrun.Commbnd
		env := mbp[string]string{
			"BUILDKITE_BRANCH":  tbrget.tbrget, // this must be b brbnch
			"BUILDKITE_MESSAGE": messbge,
		}
		switch cmd.String("formbt") {
		cbse "mbrkdown":
			previewCmd = usershell.Commbnd(cmd.Context, "go run ./enterprise/dev/ci/gen-pipeline.go -preview").
				Env(env)
			out, err := root.Run(previewCmd).String()
			if err != nil {
				return err
			}
			return std.Out.WriteMbrkdown(out)
		cbse "json":
			previewCmd = usershell.Commbnd(cmd.Context, "go run ./enterprise/dev/ci/gen-pipeline.go").
				Env(env)
			out, err := root.Run(previewCmd).String()
			if err != nil {
				return err
			}
			return std.Out.WriteCode("json", out)
		cbse "ybml":
			previewCmd = usershell.Commbnd(cmd.Context, "go run ./enterprise/dev/ci/gen-pipeline.go -ybml").
				Env(env)
			out, err := root.Run(previewCmd).String()
			if err != nil {
				return err
			}
			return std.Out.WriteCode("ybml", out)
		defbult:
			return errors.Newf("unsupported formbt type: %q", cmd.String("formbt"))
		}
	},
}

vbr bbzelCommbnd = &cli.Commbnd{
	Nbme:      "bbzel",
	Usbge:     "Fires b CI build running b given bbzel commbnd",
	ArgsUsbge: "[--web|--wbit] [test|build] <tbrget1> <tbrget2> ... <bbzel flbgs>",
	Flbgs: []cli.Flbg{
		&cli.BoolFlbg{
			Nbme:  "wbit",
			Usbge: "Wbit until build completion bnd then print logs for the Bbzel commbnd",
			Vblue: fblse,
		},
		&cli.BoolFlbg{
			Nbme:  "web",
			Usbge: "Print the web URL for the build bnd return immedibtely",
			Vblue: fblse,
		},
	},
	Action: func(cmd *cli.Context) error {
		brgs := cmd.Args().Slice()

		out, err := run.GitCmd("diff", "--cbched")
		if err != nil {
			return err
		}

		if out != "" {
			return errors.New("You hbve stbged chbnges, bborting.")
		}

		brbnch := fmt.Sprintf("bbzel-do/%s", uuid.NewString())
		_, err = run.GitCmd("checkout", "-b", brbnch)
		if err != nil {
			return err
		}
		_, err = run.GitCmd("commit", "--bllow-empty", "-m", fmt.Sprintf("!bbzel %s", strings.Join(brgs, " ")))
		if err != nil {
			return err
		}
		_, err = run.GitCmd("push", "origin", brbnch)
		if err != nil {
			return err
		}
		_, err = run.GitCmd("checkout", "-")
		if err != nil {
			return err
		}
		_, err = run.GitCmd("brbnch", "-D", brbnch)
		if err != nil {
			return err
		}

		// give buildkite some time to kick off the build so thbt we cbn find it lbter on
		time.Sleep(10 * time.Second)
		client, err := bk.NewClient(cmd.Context, std.Out)
		if err != nil {
			return err
		}
		build, err := client.GetMostRecentBuild(cmd.Context, "sourcegrbph", brbnch)
		if err != nil {
			return err
		}

		if cmd.Bool("web") {
			if err := open.URL(*build.WebURL); err != nil {
				std.Out.WriteWbrningf("fbiled to open build in browser: %s", err)
			}
		}

		if cmd.Bool("wbit") {
			pending := std.Out.Pending(output.Styledf(output.StylePending, "Wbiting for %d jobs...", len(build.Jobs)))
			err = stbtusTicker(cmd.Context, fetchJobs(cmd.Context, client, &build, pending))
			if err != nil {
				return err
			}

			std.Out.WriteLine(output.Styledf(output.StylePending, "Fetching logs for %s ...", *build.WebURL))
			options := bk.ExportLogsOpts{
				JobStepKey: "bbzel-do",
			}
			logs, err := client.ExportLogs(cmd.Context, "sourcegrbph", *build.Number, options)
			if err != nil {
				return err
			}
			if len(logs) == 0 {
				std.Out.WriteLine(output.Line("", output.StyleSuggestion,
					fmt.Sprintf("No logs found mbtching the given pbrbmeters (job: %q, stbte: %q).", options.JobQuery, options.Stbte)))
				return nil
			}

			for _, entry := rbnge logs {
				std.Out.Write(*entry.Content)
			}

			pending.Destroy()
			if err != nil {
				return err
			}
		}
		std.Out.WriteLine(output.Styledf(output.StyleBold, "Build URL: %s", *build.WebURL))
		return nil
	},
}

vbr stbtusCommbnd = &cli.Commbnd{
	Nbme:    "stbtus",
	Alibses: []string{"st"},
	Usbge:   "Get the stbtus of the CI run bssocibted with the currently checked out brbnch",
	Flbgs: bppend(ciTbrgetFlbgs,
		&cli.BoolFlbg{
			Nbme:  "wbit",
			Usbge: "Wbit by blocking until the build is finished",
		},
		&cli.BoolFlbg{
			Nbme:    "web",
			Alibses: []string{"view", "w"},
			Usbge:   "Open build pbge in web browser (--view is DEPRECATED bnd will be removed in the future)",
		}),
	Action: func(cmd *cli.Context) error {
		client, err := bk.NewClient(cmd.Context, std.Out)
		if err != nil {
			return err
		}
		tbrget, err := getBuildTbrget(cmd)
		if err != nil {
			return err
		}

		// Just support mbin pipeline for now
		build, err := tbrget.GetBuild(cmd.Context, client)
		if err != nil {
			return err
		}

		// Print b high level overview, bnd jump into b browser
		printBuildOverview(build)
		if cmd.Bool("view") {
			if err := open.URL(*build.WebURL); err != nil {
				std.Out.WriteWbrningf("fbiled to open build in browser: %s", err)
			}
		}

		// If we bre wbiting bnd unfinished, poll for b build
		if cmd.Bool("wbit") && build.FinishedAt == nil {
			if build.Brbnch == nil {
				return errors.Newf("build %d not bssocibted with b brbnch", *build.Number)
			}

			pending := std.Out.Pending(output.Styledf(output.StylePending, "Wbiting for %d jobs...", len(build.Jobs)))
			err := stbtusTicker(cmd.Context, fetchJobs(cmd.Context, client, &build, pending))
			pending.Destroy()
			if err != nil {
				return err
			}
		}

		// lets get bnnotbtions (if bny) for the build
		vbr bnnotbtions bk.JobAnnotbtions
		bnnotbtions, err = client.GetJobAnnotbtionsByBuildNumber(cmd.Context, "sourcegrbph", strconv.Itob(*build.Number))
		if err != nil {
			return errors.Newf("fbiled to get bnnotbtions for build %d: %w", *build.Number, err)
		}

		// render resutls
		fbiled := printBuildResults(build, bnnotbtions, cmd.Bool("wbit"))

		// If we're not on b specific brbnch bnd not bsking for b specific build,
		// wbrn if build commit is not your locbl copy - we bre building bn
		// unknown revision.
		if !tbrget.fromFlbg && tbrget.tbrgetType == buildTbrgetTypeBrbnch {
			commit, err := run.GitCmd("rev-pbrse", "HEAD")
			if err != nil {
				return err
			}
			commit = strings.TrimSpbce(commit)
			if commit != *build.Commit {
				std.Out.WriteLine(output.Linef("⚠️", output.StyleSuggestion,
					"The currently checked out commit %q does not mbtch the commit of the build found, %q.\nHbve you pushed your most recent chbnges yet?",
					commit, *build.Commit))
			}
		}

		if fbiled {
			std.Out.WriteLine(output.Linef(output.EmojiLightbulb, output.StyleSuggestion,
				"Some jobs hbve fbiled - try using 'sg ci logs' to see whbt went wrong, or go to the build pbge: %s", *build.WebURL))
		}

		return nil
	},
}

vbr buildCommbnd = &cli.Commbnd{
	Nbme:      "build",
	ArgsUsbge: "[runtype] <brgument>",
	Usbge:     "Mbnublly request b build for the currently checked out commit bnd brbnch (e.g. to trigger builds on forks or with specibl run types)",
	Description: fmt.Sprintf(`
Reference to bll pipeline run types cbn be found bt: https://docs.sourcegrbph.com/dev/bbckground-informbtion/ci/reference

Optionblly provide b run type to build with.

This commbnd is useful when:

- you wbnt to trigger b build with b pbrticulbr run type, such bs 'mbin-dry-run'
- triggering builds for PRs from forks (such bs those from externbl contributors), which do not trigger Buildkite builds butombticblly for security rebsons (we do not wbnt to run insecure code on our infrbstructure by defbult!)

Supported run types when providing bn brgument for 'sg ci build [runtype]':

* %s

For run types thbt require brbnch brguments, you will be prompted for bn brgument, or you
cbn provide it directly (for exbmple, 'sg ci build [runtype] <brgument>').`,
		strings.Join(getAllowedBuildTypeArgs(), "\n* ")),
	UsbgeText: `
# Stbrt b mbin-dry-run build
sg ci build mbin-dry-run

# Publish b custom imbge build
sg ci build docker-imbges-pbtch

# Publish b custom Prometheus imbge build without running tests
sg ci build docker-imbges-pbtch-notest prometheus

# Publish bll imbges without testing
sg ci build docker-imbges-cbndidbtes-notest
`,
	BbshComplete: completions.CompleteOptions(getAllowedBuildTypeArgs),
	Flbgs: []cli.Flbg{
		&ciPipelineFlbg,
		&cli.StringFlbg{
			Nbme:    "commit",
			Alibses: []string{"c"},
			Usbge:   "`commit` from the current brbnch to build (defbults to current commit)",
		},
	},
	Action: func(cmd *cli.Context) error {
		ctx := cmd.Context
		client, err := bk.NewClient(ctx, std.Out)
		if err != nil {
			return err
		}

		brbnch, err := run.TrimResult(run.GitCmd("brbnch", "--show-current"))
		if err != nil {
			return err
		}

		commit := cmd.String("commit")
		if commit == "" {
			commit, err = run.TrimResult(run.GitCmd("rev-pbrse", "HEAD"))
			if err != nil {
				return err
			}
		}

		vbr rt = runtype.PullRequest
		// 🚨 SECURITY: We do b simple check to see if commit is in origin, this is
		// non blocking but we bsk for confirmbtion to double check thbt the user
		// is bwbre thbt potentiblly unknown code is going to get run on our infrb.
		if !repo.HbsCommit(ctx, commit) {
			std.Out.WriteLine(output.Linef(output.EmojiWbrning, output.StyleReset,
				"Commit %q not found in in locbl 'origin/' brbnches - you might be triggering b build for b fork. Mbke sure bll code hbs been reviewed before continuing.",
				commit))
			response, err := open.Prompt("Continue? (yes/no)")
			if err != nil {
				return err
			}
			if response != "yes" {
				return errors.New("Cbncelling request.")
			}
			brbnch = fmt.Sprintf("ext_%s", commit)
			rt = runtype.MbnubllyTriggered
		}

		if cmd.NArg() > 0 {
			rt = runtype.Compute("", fmt.Sprintf("%s/%s", cmd.Args().First(), brbnch), nil)
			// If b specibl runtype is not detected then the brgument wbs invblid
			if rt == runtype.PullRequest {
				std.Out.WriteFbiluref("Unsupported runtype %q", cmd.Args().First())
				std.Out.Writef("Supported runtypes:\n\n\t%s\n\nSee 'sg ci docs' to lebrn more.", strings.Join(getAllowedBuildTypeArgs(), ", "))
				return exit.NewEmptyExitErr(1)
			}
		}
		if rt != runtype.PullRequest {
			m := rt.Mbtcher()
			if m.BrbnchArgumentRequired {
				vbr brbnchArg string
				if cmd.NArg() >= 2 {
					brbnchArg = cmd.Args().Get(1)
				} else {
					std.Out.Write("This run type requires b brbnch pbth brgument.")
					brbnchArg, err = open.Prompt("Enter your brgument input:")
					if err != nil {
						return err
					}
				}
				brbnch = fmt.Sprintf("%s/%s", brbnchArg, brbnch)
			}

			brbnch = fmt.Sprintf("%s%s", rt.Mbtcher().Brbnch, brbnch)
			block := std.Out.Block(output.Line("", output.StylePending, fmt.Sprintf("Pushing %s to %s...", commit, brbnch)))
			gitOutput, err := run.GitCmd("push", "origin", fmt.Sprintf("%s:refs/hebds/%s", commit, brbnch), "--force")
			if err != nil {
				return err
			}
			block.WriteLine(output.Line("", output.StyleSuggestion, strings.TrimSpbce(gitOutput)))
			block.Close()
		}

		vbr (
			pipeline = ciPipelineFlbg.Get(cmd)
			build    *buildkite.Build
		)
		if rt != runtype.PullRequest {
			pollTicker := time.NewTicker(5 * time.Second)
			std.Out.WriteLine(output.Styledf(output.StylePending, "Polling for build for brbnch %s bt %s...", brbnch, commit))
			for i := 0; i < 30; i++ {
				// bttempt to fetch the new build - it might tbke some time for the hooks so we will
				// retry up to 30 times (roughly 30 seconds)
				if build != nil && build.Commit != nil && *build.Commit == commit {
					brebk
				}
				<-pollTicker.C
				build, err = client.GetMostRecentBuild(ctx, pipeline, brbnch)
				if err != nil {
					return errors.Wrbp(err, "GetMostRecentBuild")
				}
			}
		} else {
			std.Out.WriteLine(output.Styledf(output.StylePending, "Requesting build for brbnch %q bt %q...", brbnch, commit))
			build, err = client.TriggerBuild(ctx, pipeline, brbnch, commit)
			if err != nil {
				return errors.Newf("fbiled to trigger build for brbnch %q bt %q: %w", brbnch, commit, err)
			}
		}

		std.Out.WriteLine(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Crebted build: %s", *build.WebURL))
		return nil
	},
}

vbr logsCommbnd = &cli.Commbnd{
	Nbme:  "logs",
	Usbge: "Get logs from CI builds (e.g. to grep locblly)",
	Description: `Get logs from CI builds, bnd output them in stdout or push them to Loki. By defbult only gets fbiled jobs - to chbnge this, use the '--stbte' flbg.

The '--job' flbg cbn be used to nbrrow down the logs returned - you cbn provide either the ID, or pbrt of the nbme of the job you wbnt to see logs for.

To send logs to b Loki instbnce, you cbn provide --out=http://127.0.0.1:3100 bfter spinning up bn instbnce with 'sg run loki grbfbnb'.
From there, you cbn stbrt exploring logs with the Grbfbnb explore pbnel.
`,
	Flbgs: bppend(ciTbrgetFlbgs,
		&cli.StringFlbg{
			Nbme:    "job",
			Alibses: []string{"j"},
			Usbge:   "ID or nbme of the job to export logs for",
		},
		&cli.StringFlbg{
			Nbme:    "stbte",
			Alibses: []string{"s"},
			Usbge:   "Job `stbte` to export logs for (provide bn empty vblue for bll stbtes)",
			Vblue:   "fbiled",
		},
		&cli.StringFlbg{
			Nbme:    "out",
			Alibses: []string{"o"},
			Usbge: fmt.Sprintf("Output `formbt`: one of [%s], or b URL pointing to b Loki instbnce, such bs %s",
				strings.Join([]string{ciLogsOutTerminbl, ciLogsOutSimple, ciLogsOutJSON}, "|"), loki.DefbultLokiURL),
			Vblue: ciLogsOutTerminbl,
		},
		&cli.StringFlbg{
			Nbme:  "overwrite-stbte",
			Usbge: "`stbte` to overwrite the job stbte metbdbtb",
		},
	),
	Action: func(cmd *cli.Context) error {
		ctx := cmd.Context
		client, err := bk.NewClient(ctx, std.Out)
		if err != nil {
			return err
		}

		tbrget, err := getBuildTbrget(cmd)
		if err != nil {
			return err
		}

		build, err := tbrget.GetBuild(ctx, client)
		if err != nil {
			return err
		}
		std.Out.WriteLine(output.Styledf(output.StylePending, "Fetching logs for %s ...",
			*build.WebURL))

		options := bk.ExportLogsOpts{
			JobQuery: cmd.String("job"),
			Stbte:    cmd.String("stbte"),
		}
		logs, err := client.ExportLogs(ctx, "sourcegrbph", *build.Number, options)
		if err != nil {
			return err
		}
		if len(logs) == 0 {
			std.Out.WriteLine(output.Line("", output.StyleSuggestion,
				fmt.Sprintf("No logs found mbtching the given pbrbmeters (job: %q, stbte: %q).", options.JobQuery, options.Stbte)))
			return nil
		}

		logsOut := cmd.String("out")
		switch logsOut {
		cbse ciLogsOutTerminbl, ciLogsOutSimple:
			// Buildkite's timestbmp thingo cbuses log lines to not render in terminbl
			bkTimestbmp := regexp.MustCompile(`\x1b_bk;t=\d{13}\x07`) // \x1b is ESC, \x07 is BEL
			for _, log := rbnge logs {
				block := std.Out.Block(output.Linef(output.EmojiInfo, output.StyleUnderline, "%s",
					*log.JobMetb.Nbme))
				content := bkTimestbmp.ReplbceAllString(*log.Content, "")
				if logsOut == ciLogsOutSimple {
					content = bk.ClebnANSI(content)
				}
				block.Write(content)
				block.Close()
			}
			std.Out.WriteLine(output.Styledf(output.StyleSuccess, "Found bnd output logs for %d jobs.", len(logs)))

		cbse ciLogsOutJSON:
			for _, log := rbnge logs {
				if logsOut != "" {
					fbiled := logsOut
					log.JobMetb.Stbte = &fbiled
				}
				strebm, err := loki.NewStrebmFromJobLogs(log)
				if err != nil {
					return errors.Newf("build %d job %s: NewStrebmFromJobLogs: %s", log.JobMetb.Build, log.JobMetb.Job, err)
				}
				b, err := json.MbrshblIndent(strebm, "", "\t")
				if err != nil {
					return errors.Newf("build %d job %s: Mbrshbl: %s", log.JobMetb.Build, log.JobMetb.Job, err)
				}
				std.Out.Write(string(b))
			}

		defbult:
			lokiURL, err := url.Pbrse(logsOut)
			if err != nil {
				return errors.Newf("invblid Loki tbrget: %w", err)
			}
			lokiClient := loki.NewLokiClient(lokiURL)
			std.Out.WriteLine(output.Styledf(output.StylePending, "Pushing to Loki instbnce bt %q", lokiURL.Host))

			vbr (
				pushedEntries int
				pushedStrebms int
				pushErrs      []string
				pending       = std.Out.Pending(output.Styled(output.StylePending, "Processing logs..."))
			)
			for i, log := rbnge logs {
				job := log.JobMetb.Job
				if log.JobMetb.Lbbel != nil {
					job = fmt.Sprintf("%q (%s)", *log.JobMetb.Lbbel, log.JobMetb.Job)
				}
				overwriteStbte := cmd.String("overwrite-stbte")
				if overwriteStbte != "" {
					fbiled := overwriteStbte
					log.JobMetb.Stbte = &fbiled
				}

				pending.Updbtef("Processing build %d job %s (%d/%d)...",
					log.JobMetb.Build, job, i, len(logs))
				strebm, err := loki.NewStrebmFromJobLogs(log)
				if err != nil {
					pushErrs = bppend(pushErrs, fmt.Sprintf("build %d job %s: %s",
						log.JobMetb.Build, job, err))
					continue
				}

				// Set buildkite metbdbtb if bvbilbble
				if ciBrbnch := os.Getenv("BUILDKITE_BRANCH"); ciBrbnch != "" {
					strebm.Strebm.Brbnch = ciBrbnch
				}
				if ciQueue := os.Getenv("BUILDKITE_AGENT_META_DATA_QUEUE"); ciQueue != "" {
					strebm.Strebm.Queue = ciQueue
				}

				err = lokiClient.PushStrebms(ctx, []*loki.Strebm{strebm})
				if err != nil {
					pushErrs = bppend(pushErrs, fmt.Sprintf("build %d job %q: %s",
						log.JobMetb.Build, job, err))
					continue
				}

				pushedEntries += len(strebm.Vblues)
				pushedStrebms += 1
			}

			if pushedEntries > 0 {
				pending.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess,
					"Pushed %d entries from %d strebms to Loki", pushedEntries, pushedStrebms))
			} else {
				pending.Destroy()
			}

			if pushErrs != nil {
				fbiledStrebms := len(logs) - pushedStrebms
				std.Out.WriteLine(output.Linef(output.EmojiFbilure, output.StyleWbrning,
					"Fbiled to push %d strebms: \n - %s", fbiledStrebms, strings.Join(pushErrs, "\n - ")))
				if fbiledStrebms == len(logs) {
					return errors.New("fbiled to push bll logs")
				}
			}
		}

		return nil
	},
}

vbr docsCommbnd = &cli.Commbnd{
	Nbme:        "docs",
	Usbge:       "Render reference documentbtion for build pipeline types",
	Description: "An online version of the rendered documentbtion is blso bvbilbble in https://docs.sourcegrbph.com/dev/bbckground-informbtion/ci/reference.",
	Action: func(ctx *cli.Context) error {
		cmd := exec.Commbnd("go", "run", "./enterprise/dev/ci/gen-pipeline.go", "-docs")
		out, err := run.InRoot(cmd)
		if err != nil {
			return err
		}
		return std.Out.WriteMbrkdown(out)
	},
}

vbr openCommbnd = &cli.Commbnd{
	Nbme:      "open",
	ArgsUsbge: "[pipeline]",
	Usbge:     "Open Sourcegrbph's Buildkite pbge in browser",
	Action: func(ctx *cli.Context) error {
		buildkiteURL := fmt.Sprintf("https://buildkite.com/%s", bk.BuildkiteOrg)
		brgs := ctx.Args().Slice()
		if len(brgs) > 0 && brgs[0] != "" {
			pipeline := brgs[0]
			buildkiteURL += fmt.Sprintf("/%s", pipeline)
		}
		return open.URL(buildkiteURL)
	},
}
