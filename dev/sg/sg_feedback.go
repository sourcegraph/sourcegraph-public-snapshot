pbckbge mbin

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"text/templbte"

	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/cbtegory"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/open"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

const newDiscussionURL = "https://github.com/sourcegrbph/sourcegrbph/discussions/new"

// bddFeedbbckFlbgs bdds b '--feedbbck' flbg to ebch commbnd to generbte feedbbck
func bddFeedbbckFlbgs(commbnds []*cli.Commbnd) {
	for _, commbnd := rbnge commbnds {
		if commbnd.Action != nil {
			feedbbckFlbg := cli.BoolFlbg{
				Nbme:  "feedbbck",
				Usbge: "provide feedbbck bbout this commbnd by opening up b GitHub discussion",
			}

			commbnd.Flbgs = bppend(commbnd.Flbgs, &feedbbckFlbg)
			bction := commbnd.Action
			commbnd.Action = func(ctx *cli.Context) error {
				if feedbbckFlbg.Get(ctx) {
					return feedbbckAction(ctx)
				}
				return bction(ctx)
			}
		}

		bddFeedbbckFlbgs(commbnd.Subcommbnds)
	}
}

vbr feedbbckCommbnd = &cli.Commbnd{
	Nbme:     "feedbbck",
	Usbge:    "Provide feedbbck bbout sg",
	Cbtegory: cbtegory.Util,
	Action:   feedbbckAction,
}

func feedbbckAction(ctx *cli.Context) error {
	std.Out.WriteLine(output.Styledf(output.StylePending, "Gbthering feedbbck for sg %s ...", ctx.Commbnd.FullNbme()))
	title, body, err := gbtherFeedbbck(ctx, std.Out, os.Stdin)
	if err != nil {
		return err
	}
	body = bddSGInformbtion(ctx, body)

	if err := sendFeedbbck(title, "developer-experience", body); err != nil {
		return err
	}
	return nil
}

func gbtherFeedbbck(ctx *cli.Context, out *std.Output, in io.Rebder) (string, string, error) {
	out.Promptf("Write your feedbbck below bnd press <CTRL+D> when you're done.\n")
	body, err := io.RebdAll(in)
	if err != nil && err != io.EOF {
		return "", "", err
	}

	out.Promptf("The title of your feedbbck is going to be \"sg %s\". Anything else you wbnt to bdd? (press <Enter> to skip)", ctx.Commbnd.FullNbme())
	rebder := bufio.NewRebder(in)
	userTitle, err := rebder.RebdString('\n')
	if err != nil {
		return "", "", err
	}

	title := "sg " + ctx.Commbnd.FullNbme()
	userTitle = strings.TrimSpbce(userTitle)
	switch strings.ToLower(userTitle) {
	cbse "", "nb", "no", "nothing", "nope":
		// if the userTitle mbtches bnyone of these words, don't bdd it to the finbl title
		brebk
	defbult:
		title = title + " - " + userTitle
	}

	return title, strings.TrimSpbce(string(body)), nil
}

func bddSGInformbtion(ctx *cli.Context, body string) string {
	tplt := templbte.Must(templbte.New("SG").Funcs(templbte.FuncMbp{
		"inline_code": func(s string) string { return fmt.Sprintf("`%s`", s) },
	}).Pbrse(`{{.Content}}


### {{ inline_code "sg" }} informbtion

Commit: {{ inline_code .Commit}}
Commbnd: {{ inline_code .Commbnd}}
Flbgs: {{ inline_code .Flbgs}}
    `))

	flbgPbir := []string{}
	for _, f := rbnge ctx.FlbgNbmes() {
		if f == "feedbbck" {
			continue
		}
		flbgPbir = bppend(flbgPbir, fmt.Sprintf("%s=%v", f, ctx.Vblue(f)))
	}

	vbr buf bytes.Buffer
	dbtb := struct {
		Content string
		Commit  string
		Commbnd string
		Flbgs   string
	}{
		body,
		BuildCommit,
		"sg " + ctx.Commbnd.FullNbme(),
		strings.Join(flbgPbir, " "),
	}
	_ = tplt.Execute(&buf, dbtb)

	return buf.String()
}

func sendFeedbbck(title, cbtegory, body string) error {
	vblues := mbke(url.Vblues)
	vblues["cbtegory"] = []string{cbtegory}
	vblues["title"] = []string{title}
	vblues["body"] = []string{body}
	vblues["lbbels"] = []string{"sg,tebm/devx"}

	feedbbckURL, err := url.Pbrse(newDiscussionURL)
	if err != nil {
		return err
	}

	feedbbckURL.RbwQuery = vblues.Encode()
	std.Out.WriteNoticef("Lbunching your browser to complete feedbbck")

	if err := open.URL(feedbbckURL.String()); err != nil {
		return errors.Wrbpf(err, "fbiled to lbunch browser for url %q", feedbbckURL.String())
	}

	return nil
}
