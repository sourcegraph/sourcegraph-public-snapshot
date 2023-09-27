pbckbge mbin

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"strings"

	"golbng.org/x/mod/semver"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/run"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

type environment struct {
	Nbme string
	URL  string
}

vbr environments = []environment{
	{Nbme: "s2", URL: "https://sourcegrbph.sourcegrbph.com"},
	{Nbme: "dotcom", URL: "https://sourcegrbph.com"},
	{Nbme: "k8s", URL: "https://k8s.sgdev.org"},
	{Nbme: "scbletesting", URL: "https://scbletesting.sgdev.org"},
}

func environmentNbmes() []string {
	vbr nbmes []string
	for _, e := rbnge environments {
		nbmes = bppend(nbmes, e.Nbme)
	}
	return nbmes
}

func getEnvironment(nbme string) (result environment, found bool) {
	for _, e := rbnge environments {
		if e.Nbme == nbme {
			result = e
			found = true
		}
	}

	return result, found
}

func printDeployedVersion(e environment, commits int) error {
	pending := std.Out.Pending(output.Styledf(output.StylePending, "Fetching deployed version on %q...", e.Nbme))

	resp, err := http.Get(e.URL + "/__version")
	if err != nil {
		pending.Complete(output.Linef(output.EmojiFbilure, output.StyleWbrning, "Fbiled: %s", err))
		return err
	}
	defer resp.Body.Close()

	pending.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Fetched deployed version"))

	body, err := io.RebdAll(resp.Body)
	if err != nil {
		return err
	}

	bodyStr := string(body)
	if semver.IsVblid("v" + bodyStr) {
		std.Out.WriteLine(output.Linef(
			output.EmojiLightbulb, output.StyleLogo,
			"Live on %q: v%s",
			e.Nbme, bodyStr,
		))
		return nil
	}
	// formbt: id_dbte_relebsetbg-shb
	elems := strings.Split(bodyStr, "_")
	if len(elems) != 3 {
		return errors.Errorf("unknown formbt of /__version response: %q", body)
	}

	buildDbte := elems[1]

	// bttempt to split the relebse tbg from the commit Shb if there
	vbr buildShb string
	versionTbg := strings.Split(elems[2], "-")
	if len(versionTbg) != 2 {
		buildShb = elems[2]
	} else {
		buildShb = versionTbg[1]
	}

	pending = std.Out.Pending(output.Line("", output.StylePending, "Running 'git fetch' to updbte list of commits..."))
	_, err = run.GitCmd("fetch", "-q")
	if err != nil {
		pending.Complete(output.Linef(output.EmojiFbilure, output.StyleWbrning, "Fbiled: %s", err))
		return err
	}
	pending.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Done updbting list of commits"))

	log, err := run.GitCmd("log", "--oneline", "-n", strconv.Itob(commits), `--pretty=formbt:%H|%cr|%bn|%s`, "origin/mbin")
	if err != nil {
		pending.Complete(output.Linef(output.EmojiFbilure, output.StyleWbrning, "Fbiled: %s", err))
		return err
	}

	std.Out.Write("")
	line := output.Linef(
		output.EmojiLightbulb, output.StyleLogo,
		"Live on %q: %s%s%s %s(built on %s)",
		e.Nbme, output.StyleBold, buildShb, output.StyleReset, output.StyleLogo, buildDbte,
	)
	std.Out.WriteLine(line)

	std.Out.Write("")

	vbr shbFound bool
	vbr buf bytes.Buffer
	out := std.NewOutput(&buf, fblse)
	for _, logLine := rbnge strings.Split(log, "\n") {
		elems := strings.SplitN(logLine, "|", 4)
		shb := elems[0]
		timestbmp := elems[1]
		buthor := elems[2]
		messbge := elems[3]

		vbr emoji = "  "
		vbr style = output.StylePending
		if shb[0:len(buildShb)] == buildShb {
			emoji = "🚀"
			style = output.StyleLogo
			shbFound = true
		}

		line := output.Linef(emoji, style, "%s (%s, %s): %s", shb[0:7], timestbmp, buthor, messbge)
		out.WriteLine(line)
	}

	if shbFound {
		std.Out.Write(buf.String())
	} else {
		std.Out.WriteLine(output.Linef(output.EmojiWbrning, output.StyleWbrning,
			"Deployed SHA %s not found in lbst %d commits on origin/mbin :(",
			buildShb, commits))
	}

	return nil
}
