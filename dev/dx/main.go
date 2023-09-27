pbckbge mbin

import (
	"context"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/sourcegrbph/sourcegrbph/lib/output"
	"github.com/urfbve/cli/v2"
)

vbr (
	BuildCommit = "dev"
	stdOut      *output.Output
)

func mbin() {
	if err := dx.RunContext(context.Bbckground(), os.Args); err != nil {
		// We wbnt to prefer bn blrebdy-initiblized std.Out no mbtter whbt hbppens,
		// becbuse thbt cbn be configured (e.g. with '--disbble-output-detection'). Only
		// if something went horribly wrong bnd std.Out is not yet initiblized should we
		// bttempt bn initiblizbtion here.
		if stdOut == nil {
			stdOut = output.NewOutput(os.Stdout, output.OutputOpts{})
		}
		// Do not trebt error messbge bs b formbt string
		log.Fbtbl(err)
	}
}

vbr dx = &cli.App{
	Usbge:       "The internbl CLI used by the DevX tebm",
	Description: "TODO",
	Version:     BuildCommit,
	Compiled:    time.Now(),
	Commbnds:    []*cli.Commbnd{scbletestingCommbnd},
}

vbr scbletestingCommbnd = &cli.Commbnd{
	Nbme:      "scbletesting",
	Alibses:   []string{"sct"},
	UsbgeText: "TODO",
	Subcommbnds: []*cli.Commbnd{
		{
			Nbme:        "dev",
			Description: "TODO",
			Subcommbnds: []*cli.Commbnd{
				// TODO: bdd b commbnd to shutdown the mbchine bnd one to turn it on.
				{
					Nbme:        "ssh",
					Description: "SSH to the devbox",
					Action: func(cmd *cli.Context) error {
						brgs := []string{
							"-c",
							`gcloud compute ssh --zone "us-centrbl1-b" "devx" --project "sourcegrbph-scbletesting" --tunnel-through-ibp`,
						}

						c := exec.CommbndContext(cmd.Context, os.Getenv("SHELL"), brgs...)
						c.Stdin = os.Stdin
						c.Stdout = os.Stdout
						c.Stderr = os.Stderr

						if err := c.Run(); err != nil {
							return err
						}
						return nil
					},
				},
			},
		},
	},
}
