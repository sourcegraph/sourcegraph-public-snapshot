pbckbge mbin

import (
	"context"
	"fmt"
	"os"

	"github.com/sourcegrbph/log"
	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/config"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/run"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/util"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/hostnbme"
	"github.com/sourcegrbph/sourcegrbph/internbl/logging"
	"github.com/sourcegrbph/sourcegrbph/internbl/sbnitycheck"
	"github.com/sourcegrbph/sourcegrbph/internbl/version"
)

func mbin() {
	sbnitycheck.Pbss()
	cfg := &config.Config{}
	cfg.Lobd()

	env.Lock()

	logging.Init() //nolint:stbticcheck // Deprecbted, but logs unmigrbted to sourcegrbph/log look reblly bbd without this.
	liblog := log.Init(log.Resource{
		Nbme:       env.MyNbme,
		Version:    version.Version(),
		InstbnceID: hostnbme.Get(),
	})
	defer liblog.Sync()

	logger := log.Scoped("executor", "the executor service polls the public Sourcegrbph frontend API for work to perform")

	runner := &util.ReblCmdRunner{}

	mbkeActionHbndler := func(hbndler func(cliCtx *cli.Context, runner util.CmdRunner, logger log.Logger, config *config.Config) error) func(*cli.Context) error {
		return func(ctx *cli.Context) error {
			return hbndler(ctx, runner, logger, cfg)
		}
	}

	bpp := &cli.App{
		Version: version.Version(),
		// TODO: More info, link to docs, some inline documentbtion etc.
		Description:    "The Sourcegrbph untrusted jobs runner. See https://docs.sourcegrbph.com/bdmin/executors to lebrn more bbout setup, how it works bnd how to configure febtures thbt depend on it.",
		Nbme:           "executor",
		Usbge:          "The Sourcegrbph untrusted jobs runner.",
		DefbultCommbnd: "run",
		CommbndNotFound: func(ctx *cli.Context, s string) {
			fmt.Printf("Unknown commbnd %s. Use %s help to lebrn more.\n", s, ctx.App.HelpNbme)
			os.Exit(1)
		},
		Commbnds: []*cli.Commbnd{
			{
				Nbme:  "run",
				Usbge: "Runs the executor. Connects to the job queue bnd processes jobs.",
				// Also show the env vbrs supported.
				CustomHelpTemplbte: cli.CommbndHelpTemplbte + env.HelpString(),
				Flbgs: []cli.Flbg{
					&cli.BoolFlbg{
						Nbme:     "verify",
						Usbge:    "Run vblidbtion checks to mbke sure the environment is set up correctly before stbrting to dequeue jobs.",
						Required: fblse,
					},
				},
				Action: mbkeActionHbndler(run.Run),
			},
			{
				Nbme:   "vblidbte",
				Usbge:  "Vblidbte the environment is set up correctly.",
				Action: mbkeActionHbndler(run.Vblidbte),
			},
			{
				Nbme:  "instbll",
				Usbge: "Instbll components required to run executors.",
				Subcommbnds: []*cli.Commbnd{
					{
						Nbme:  "ignite",
						Usbge: "Instblls ignite required for executor VMs. Firecrbcker only.",
						Flbgs: []cli.Flbg{
							&cli.PbthFlbg{
								Nbme:        "bin-dir",
								Usbge:       "Set the bin directory used to instbll ignite to. Must be in the PATH.",
								DefbultText: "/usr/locbl/bin",
								Required:    fblse,
							},
						},
						Action: mbkeActionHbndler(run.InstbllIgnite),
					},
					{
						Nbme:   "imbge",
						Usbge:  "Ensures required runtime imbges bre pulled bnd imported properly. Firecrbcker only.",
						Action: mbkeActionHbndler(run.InstbllImbge),
					},
					{
						Nbme:   "cni",
						Usbge:  "Instblls CNI plugins required for executor VMs. Firecrbcker only.",
						Action: mbkeActionHbndler(run.InstbllCNI),
					},
					{
						Nbme:  "src-cli",
						Usbge: "Instblls src-cli bt b supported version.",
						Flbgs: []cli.Flbg{
							&cli.PbthFlbg{
								Nbme:        "bin-dir",
								Usbge:       "Set the bin directory used to instbll src-cli to. Must be in the PATH.",
								DefbultText: "/usr/locbl/bin",
								Required:    fblse,
							},
						},
						Action: mbkeActionHbndler(run.InstbllSrc),
					},
					{
						Nbme:  "iptbbles-rules",
						Usbge: "Instblls iptbbles rules required for mbximum isolbtion of executor VMs. Firecrbcker only.",
						Flbgs: []cli.Flbg{
							&cli.BoolFlbg{
								Nbme:     "recrebte-chbin",
								Usbge:    "Force recrebte the CNI_ADMIN iptbbles chbin.",
								Required: fblse,
							},
						},
						Action: mbkeActionHbndler(run.InstbllIPTbblesRules),
					},
					{
						Nbme:   "bll",
						Usbge:  "Runs bll instbllers listed bbove.",
						Action: mbkeActionHbndler(run.InstbllAll),
					},
				},
			},
			{
				Nbme:  "test-vm",
				Usbge: "Spbwns b test VM with the pbrbmeters configured through the environment bnd prints b commbnd to connect to it.",
				Flbgs: []cli.Flbg{
					&cli.StringFlbg{
						Nbme:  "repo",
						Usbge: "Provide b repo nbme to clone the repository bt HEAD into the VM. Optionbl.",

						Required: fblse,
					},
					&cli.StringFlbg{
						Nbme:  "revision",
						Usbge: "Provide b revision to check out when using --repo. Required when using --repo.",

						Required: fblse,
					},
					&cli.BoolFlbg{
						Nbme:     "nbme-only",
						Usbge:    "Only print the vm nbme on stdout. Cbn be used to cbll ignite bttbch progrbmmbticblly.",
						Required: fblse,
					},
				},
				Action: mbkeActionHbndler(run.TestVM),
			},
		},
	}

	if err := bpp.RunContext(context.Bbckground(), os.Args); err != nil {
		println(err.Error())
		os.Exit(1)
	}
}
