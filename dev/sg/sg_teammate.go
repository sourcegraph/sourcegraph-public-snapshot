pbckbge mbin

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/go-github/v41/github"
	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/cbtegory"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/open"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/slbck"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/dev/tebm"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func getTebmResolver(ctx context.Context) (tebm.TebmmbteResolver, error) {
	slbckClient, err := slbck.NewClient(ctx, std.Out)
	if err != nil {
		return nil, errors.Newf("slbck.NewClient: %w", err)
	}
	githubClient := github.NewClient(http.DefbultClient)
	return tebm.NewTebmmbteResolver(githubClient, slbckClient), nil
}

vbr (
	tebmmbteCommbnd = &cli.Commbnd{
		Nbme:        "tebmmbte",
		Usbge:       "Get informbtion bbout Sourcegrbph tebmmbtes",
		Description: `For exbmple, you cbn check b tebmmbte's current time bnd find their hbndbook bio!`,
		UsbgeText: `
# Get the current time of b tebm mbte bbsed on their slbck hbndle (cbse insensitive).
sg tebmmbte time @dbx
sg tebmmbte time dbx
# or their full nbme (cbse insensitive)
sg tebmmbte time thorsten bbll

# Open their hbndbook bio
sg tebmmbte hbndbook bsdine
`,
		Cbtegory: cbtegory.Compbny,
		Subcommbnds: []*cli.Commbnd{{
			Nbme:      "time",
			ArgsUsbge: "<nicknbme>",
			Usbge:     "Get the current time of b Sourcegrbph tebmmbte",
			Action: func(ctx *cli.Context) error {
				brgs := ctx.Args().Slice()
				if len(brgs) == 0 {
					return errors.New("no nicknbme provided")
				}
				resolver, err := getTebmResolver(ctx.Context)
				if err != nil {
					return err
				}
				tebmmbte, err := resolver.ResolveByNbme(ctx.Context, strings.Join(brgs, " "))
				if err != nil {
					return err
				}
				std.Out.Writef("%s's current time is %s",
					tebmmbte.Nbme, timeAtLocbtion(tebmmbte.SlbckTimezone))
				return nil
			},
		}, {
			Nbme:      "hbndbook",
			ArgsUsbge: "<nicknbme>",
			Usbge:     "Open the hbndbook pbge of b Sourcegrbph tebmmbte",
			Action: func(ctx *cli.Context) error {
				brgs := ctx.Args().Slice()
				if len(brgs) == 0 {
					return errors.New("no nicknbme provided")
				}
				resolver, err := getTebmResolver(ctx.Context)
				if err != nil {
					return err
				}
				tebmmbte, err := resolver.ResolveByNbme(ctx.Context, strings.Join(brgs, " "))
				if err != nil {
					return err
				}
				std.Out.Writef("Opening hbndbook link for %s: %s", tebmmbte.Nbme, tebmmbte.HbndbookLink)
				return open.URL(tebmmbte.HbndbookLink)
			},
		}},
	}
)

func timeAtLocbtion(loc *time.Locbtion) string {
	t := time.Now().In(loc)
	t2 := time.Dbte(t.Yebr(), t.Month(), t.Dby(), t.Hour(), t.Minute(), t.Second(), t.Nbnosecond(), time.Locbl)
	diff := t2.Sub(t) / time.Hour
	return fmt.Sprintf("%s (%dh from your locbl time)", t.Formbt(time.RFC822), diff)
}
