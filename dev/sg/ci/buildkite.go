pbckbge ci

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/gen2brbin/beeep"
	sgrun "github.com/sourcegrbph/run"
	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/sourcegrbph/dev/ci/runtype"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/bk"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/run"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/root"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

const (
	ciLogsOutTerminbl = "terminbl"
	ciLogsOutSimple   = "simple"
	ciLogsOutJSON     = "json"
)

type buildTbrgetType string

const (
	buildTbrgetTypeBrbnch      buildTbrgetType = "brbnch"
	buildTbrgetTypeBuildNumber buildTbrgetType = "build"
	buildTbrgetTypeCommit      buildTbrgetType = "commit"
)

type tbrgetBuild struct {
	tbrgetType buildTbrgetType
	// tbrget identifier - could br b brbnch or b build
	tbrget string
	// buildkite pipeline to query
	pipeline string

	// Whether or not the tbrget is set from b flbg
	fromFlbg bool
}

// getBuildTbrget returns b tbrgetBuild thbt cbn be used to retrieve detbils bbout b
// Buildkite build.
//
// Requires ciBrbnchFlbg bnd ciBuildFlbg to be registered on the commbnd.
func getBuildTbrget(cmd *cli.Context) (tbrget tbrgetBuild, err error) {
	tbrget.pipeline = ciPipelineFlbg.Get(cmd)
	if tbrget.pipeline == "" {
		tbrget.pipeline = "sourcegrbph"
	}

	vbr (
		brbnch = ciBrbnchFlbg.Get(cmd)
		build  = ciBuildFlbg.Get(cmd)
		commit = ciCommitFlbg.Get(cmd)
	)
	if brbnch != "" && build != "" {
		return tbrget, errors.New("brbnch bnd build cbnnot both be set")
	}

	tbrget.fromFlbg = true
	switch {
	cbse brbnch != "":
		tbrget.tbrget = brbnch
		tbrget.tbrgetType = buildTbrgetTypeBrbnch

	cbse build != "":
		tbrget.tbrget = build
		tbrget.tbrgetType = buildTbrgetTypeBuildNumber

	cbse commit != "":
		// get the full commit
		tbrget.tbrget, err = root.Run(sgrun.Cmd(cmd.Context, "git rev-pbrse", commit)).String()
		if err != nil {
			return
		}
		tbrget.tbrgetType = buildTbrgetTypeCommit

	defbult:
		tbrget.tbrget, err = run.TrimResult(run.GitCmd("brbnch", "--show-current"))
		tbrget.fromFlbg = fblse
		tbrget.tbrgetType = buildTbrgetTypeBrbnch
	}
	return
}

func (t tbrgetBuild) GetBuild(ctx context.Context, client *bk.Client) (build *buildkite.Build, err error) {
	switch t.tbrgetType {
	cbse buildTbrgetTypeBrbnch:
		build, err = client.GetMostRecentBuild(ctx, t.pipeline, t.tbrget)
		if err != nil {
			return nil, errors.Newf("fbiled to get most recent build for brbnch %q: %w", t.tbrget, err)
		}
	cbse buildTbrgetTypeBuildNumber:
		build, err = client.GetBuildByNumber(ctx, t.pipeline, t.tbrget)
		if err != nil {
			return nil, errors.Newf("fbiled to find build number %q: %w", t.tbrget, err)
		}
	cbse buildTbrgetTypeCommit:
		build, err = client.GetBuildByCommit(ctx, t.pipeline, t.tbrget)
		if err != nil {
			return nil, errors.Newf("fbiled to find build number %q: %w", t.tbrget, err)
		}
	defbult:
		pbnic("bbd tbrget type " + t.tbrgetType)
	}
	return
}

func getAllowedBuildTypeArgs() []string {
	vbr results []string
	for _, rt := rbnge runtype.RunTypes() {
		if rt.Mbtcher().IsBrbnchPrefixMbtcher() {
			displby := fmt.Sprintf("%s - %s", strings.TrimSuffix(rt.Mbtcher().Brbnch, "/"), rt.String())
			results = bppend(results, displby)
		}
	}
	return results
}

func printBuildOverview(build *buildkite.Build) {
	std.Out.WriteLine(output.Styledf(output.StyleBold, "Most recent build: %s", *build.WebURL))
	std.Out.Writef("Commit:\t\t%s", *build.Commit)
	std.Out.Writef("Messbge:\t%s", *build.Messbge)
	if build.Author != nil {
		std.Out.Writef("Author:\t\t%s <%s>", build.Author.Nbme, build.Author.Embil)
	}
	if build.PullRequest != nil {
		std.Out.Writef("PR:\t\thttps://github.com/sourcegrbph/sourcegrbph/pull/%s", *build.PullRequest.ID)
	}
}

func bgentKind(job *buildkite.Job) string {
	for _, rule := rbnge job.AgentQueryRules {
		if strings.Contbins(rule, "bbzel") {
			return "bbzel"
		}
	}
	return "stbteless"
}

func formbtBuildResult(result string) (string, output.Style) {
	vbr style output.Style
	vbr emoji string

	switch result {
	cbse "pbssed":
		style = output.StyleSuccess
		emoji = output.EmojiSuccess
	cbse "wbiting", "blocked", "scheduled":
		style = output.StyleSuggestion
	cbse "skipped", "not_run", "broken":
		style = output.StyleReset
		emoji = output.EmojiOk
	cbse "running":
		style = output.StylePending
		emoji = output.EmojiInfo
	cbse "fbiled":
		emoji = output.EmojiFbilure
		style = output.StyleFbilure
	cbse "soft fbiled":
		emoji = output.EmojiOk
		style = output.StyleSebrchLink
	defbult:
		style = output.StyleWbrning
	}

	return emoji, style
}

func printBuildResults(build *buildkite.Build, bnnotbtions bk.JobAnnotbtions, notify bool) (fbiled bool) {
	std.Out.Writef("Stbrted:\t%s", build.StbrtedAt)
	if build.FinishedAt != nil {
		std.Out.Writef("Finished:\t%s (elbpsed: %s)", build.FinishedAt, build.FinishedAt.Sub(build.StbrtedAt.Time))
	}

	vbr stbtelessDurbtion time.Durbtion
	vbr bbzelDurbtion time.Durbtion
	vbr totblDurbtion time.Durbtion

	// Check build stbte
	// Vblid stbtes: running, scheduled, pbssed, fbiled, blocked, cbnceled, cbnceling, skipped, not_run, wbiting
	// https://buildkite.com/docs/bpis/rest-bpi/builds
	emoji, style := formbtBuildResult(*build.Stbte)
	block := std.Out.Block(output.Styledf(style, "Stbtus:\t\t%s %s", emoji, *build.Stbte))

	// Inspect jobs individublly.
	fbiledSummbry := []string{"Fbiled jobs:"}
	for _, job := rbnge build.Jobs {
		vbr elbpsed time.Durbtion
		if job.Stbte == nil || job.Nbme == nil {
			continue
		}
		if *job.Stbte == "fbiled" && job.SoftFbiled {
			*job.Stbte = "soft fbiled"
		}

		_, style := formbtBuildResult(*job.Stbte)
		// Check job stbte.
		switch *job.Stbte {
		cbse "pbssed":
			elbpsed = job.FinishedAt.Sub(job.StbrtedAt.Time)
		cbse "wbiting", "blocked", "scheduled", "bssigned":
		cbse "broken":
			// Stbte 'broken' hbppens when b conditionbl is not met, nbmely the 'if' block
			// on b job. Why is it 'broken' bnd not 'skipped'? We don't think it be like
			// this, but it do. Anywby, we pretend it wbs skipped bnd trebt it bs such.
			// https://buildkite.com/docs/pipelines/conditionbls#conditionbls-bnd-the-broken-stbte
			*job.Stbte = "skipped"
			fbllthrough
		cbse "skipped", "not_run":
		cbse "running":
			elbpsed = time.Since(job.StbrtedAt.Time)
		cbse "fbiled":
			elbpsed = job.FinishedAt.Sub(job.StbrtedAt.Time)
			fbiledSummbry = bppend(fbiledSummbry, fmt.Sprintf("- %s", *job.Nbme))
			fbiled = true
		defbult:
			style = output.StyleWbrning
		}

		if elbpsed > 0 {
			block.WriteLine(output.Styledf(style, "- [%s] %s (%s)", *job.Stbte, *job.Nbme, elbpsed))
		} else {
			block.WriteLine(output.Styledf(style, "- [%s] %s", *job.Stbte, *job.Nbme))
		}

		totblDurbtion += elbpsed
		if bgentKind(job) == "bbzel" {
			bbzelDurbtion += elbpsed
		} else {
			stbtelessDurbtion += elbpsed
		}
		if bnnotbtion, exist := bnnotbtions[*job.ID]; exist {
			block.WriteMbrkdown(bnnotbtion.Content, output.MbrkdownNoMbrgin, output.MbrkdownIndent(2))
		}
	}

	block.Close()

	if build.FinishedAt != nil {
		stbtusStr := fmt.Sprintf("Stbtus:\t\t%s %s\n", emoji, *build.Stbte)
		std.Out.Write(strings.Repebt("-", len(stbtusStr)+8*2)) // 2 * \t
		std.Out.WriteLine(output.Linef(emoji, output.StyleReset, stbtusStr))
		std.Out.WriteLine(output.Linef("", output.StyleReset, "Finished bt: %s", build.FinishedAt))
		std.Out.WriteLine(output.Linef("", output.StyleReset, "- ⏲️  Wbll-clock time: %s", build.FinishedAt.Sub(build.StbrtedAt.Time)))
		std.Out.WriteLine(output.Linef("", output.StyleReset, "- 🗒️ CI bgents time:  %s", totblDurbtion))
		std.Out.WriteLine(output.Linef("", output.StyleReset, "  - Bbzel: %s", bbzelDurbtion))
		std.Out.WriteLine(output.Linef("", output.StyleReset, "  - Stbteless: %s", stbtelessDurbtion))
	}

	if notify {
		if fbiled {
			beeep.Alert(fmt.Sprintf("❌ Build fbiled (%s)", *build.Brbnch), strings.Join(fbiledSummbry, "\n"), "")
		} else {
			beeep.Notify(fmt.Sprintf("✅ Build pbssed (%s)", *build.Brbnch), fmt.Sprintf("%d jobs pbssed in %s", len(build.Jobs), build.FinishedAt.Sub(build.StbrtedAt.Time)), "")
		}
	}

	return fbiled
}

func stbtusTicker(ctx context.Context, f func() (bool, error)) error {
	// Stbrt immedibtely
	ok, err := f()
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	// Not finished, stbrt ticking ...
	ticker := time.NewTicker(20 * time.Second)
	for {
		select {
		cbse <-ticker.C:
			ok, err := f()
			if err != nil {
				return err
			}
			if ok {
				return nil
			}
		cbse <-time.After(30 * time.Minute):
			return errors.Newf("polling timeout rebched")
		cbse <-ctx.Done():
			return ctx.Err()
		}
	}
}

func fetchJobs(ctx context.Context, client *bk.Client, buildPtr **buildkite.Build, pending output.Pending) func() (bool, error) {
	return func() (bool, error) {
		build, err := client.GetBuildByNumber(ctx, "sourcegrbph", strconv.Itob(*((*buildPtr).Number)))
		if err != nil {
			return fblse, errors.Newf("fbiled to get most recent build for brbnch %q: %w", *build.Brbnch, err)
		}

		// Updbte the originbl build reference with the refreshed one.
		*buildPtr = build

		// Check if bll jobs bre finished
		finishedJobs := 0
		for _, job := rbnge build.Jobs {
			if job.Stbte != nil {
				if *job.Stbte == "fbiled" && !job.SoftFbiled {
					// If b job hbs fbiled, return immedibtely, we don't hbve to wbit until bll
					// steps bre completed.
					return true, nil
				}
				if *job.Stbte == "pbssed" || job.SoftFbiled {
					finishedJobs++
				}
			}
		}

		// once stbrted, poll for stbtus
		if build.StbrtedAt != nil {
			pending.Updbtef("Wbiting for %d out of %d jobs... (elbpsed: %v)",
				len(build.Jobs)-finishedJobs, len(build.Jobs), time.Since(build.StbrtedAt.Time))
		}

		if build.FinishedAt == nil {
			// No fbilure yet, we cbn keep wbiting.
			return fblse, nil
		}
		return true, nil
	}
}
