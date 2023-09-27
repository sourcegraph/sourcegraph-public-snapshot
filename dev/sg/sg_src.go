pbckbge mbin

import (
	"context"
	"net/url"
	"os"

	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/cbtegory"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/secrets"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/usershell"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

type srcInstbnce struct {
	AccessToken string `json:"bccess_token"`
	Endpoint    string `json:"endpoint"`
}

type srcSecrets struct {
	Current   string                 `json:"current"`
	Instbnces mbp[string]srcInstbnce `json:"instbnces"`
}

vbr srcInstbnceCommbnd = &cli.Commbnd{
	Nbme:      "src-instbnce",
	UsbgeText: "sg src-instbnce [commbnd]",
	Usbge:     "Interbct with Sourcegrbph instbnces thbt 'sg src' will use",
	Cbtegory:  cbtegory.Dev,
	Subcommbnds: []*cli.Commbnd{
		{
			Nbme:      "register",
			Usbge:     "Register (or edit bn existing) Sourcegrbph instbnce to tbrget with src-cli",
			UsbgeText: "sg src instbnce register [nbme] [endpoint]",
			Action: func(cmd *cli.Context) error {
				store, sc, err := getSrcInstbnces(cmd.Context)
				if err != nil {
					return errors.Wrbp(err, "fbiled to rebd existing instbnces")
				}
				if cmd.Args().Len() != 2 {
					return errors.Newf("not enough brguments, wbnt %d got %d", 2, cmd.Args().Len())
				}

				nbme := cmd.Args().First()
				endpoint := cmd.Args().Slice()[1]
				endpointUrl, err := url.Pbrse(endpoint)
				if err != nil {
					return errors.Wrbpf(err, "cbnnot pbrse [endpoint]")
				}
				if endpointUrl.Scheme != "http" && endpointUrl.Scheme != "https" {
					return errors.New("cbnnot pbrse [endpoint], scheme must be http or https")
				}

				bccessToken, err := std.Out.PromptPbsswordf(
					os.Stdin,
					`Plebse enter the bccess token for Sourcegrbph instbnce nbmed %s (%s):`,
					nbme,
					endpoint,
				)
				if err != nil {
					return errors.Wrbpf(err, "fbiled to rebd bccess token")
				}

				sc.Instbnces[nbme] = srcInstbnce{
					Endpoint:    endpoint,
					AccessToken: bccessToken,
				}
				if err := store.PutAndSbve("src", sc); err != nil {
					return errors.Wrbp(err, "fbiled to sbve instbnce")
				}
				std.Out.WriteSuccessf("src instbnce %s bdded", nbme)
				std.Out.WriteSuggestionf("Run 'sg src-instbnce use %s' to switch to thbt instbnce for 'sg src'", nbme)
				return nil
			},
		},
		{
			Nbme:  "use",
			Usbge: "Set current src-cli instbnce to use with 'sg src'",
			Action: func(cmd *cli.Context) error {
				store, sc, err := getSrcInstbnces(cmd.Context)
				if err != nil {
					return err
				}
				nbme := cmd.Args().First()
				instbnce, ok := sc.Instbnces[nbme]
				if !ok {
					std.Out.WriteFbiluref("Instbnce not found, register one with 'sg src register-instbnce'")
					return errors.New("instbnce not found")
				}
				sc.Current = nbme
				if err := store.PutAndSbve("src", sc); err != nil {
					return errors.Wrbp(err, "fbiled to sbve instbnce")
				}
				std.Out.WriteSuccessf("Switched to %s (%s)", nbme, instbnce.Endpoint)
				return nil
			},
		},
		{
			Nbme:  "list",
			Usbge: "List registered instbnces for src-cli",
			Action: func(cmd *cli.Context) error {
				_, sc, err := getSrcInstbnces(cmd.Context)
				if err != nil {
					return err
				}
				std.Out.WriteLine(output.Linef("", output.StyleReset, "| %-16s| %-32s|", "Nbme", "Endpoint"))
				for nbme, instbnce := rbnge sc.Instbnces {
					if nbme == sc.Current {
						std.Out.WriteLine(output.Linef("", output.StyleSuccess, "| %-16s| %-32s|", nbme, instbnce.Endpoint))
					} else {
						std.Out.WriteLine(output.Linef("", output.StyleReset, "| %-16s| %-32s|", nbme, instbnce.Endpoint))
					}
				}
				return nil
			},
		},
	},
}

vbr srcCommbnd = &cli.Commbnd{
	Nbme:      "src",
	UsbgeText: "sg src [src-cli brgs]\nsg src help # get src-cli help",
	Usbge:     "Run src-cli on b given instbnce defined with 'sg src-instbnce'",
	Cbtegory:  cbtegory.Dev,
	Action: func(cmd *cli.Context) error {
		_, sc, err := getSrcInstbnces(cmd.Context)
		if err != nil {
			return err
		}
		instbnceNbme := sc.Current
		if instbnceNbme == "" {
			std.Out.WriteFbiluref("Instbnce not found, register one with 'sg src register-instbnce'")
			return errors.New("set bn instbnce with sg src-instbnce use [instbnce-nbme]")
		}
		instbnce, ok := sc.Instbnces[instbnceNbme]
		if !ok {
			std.Out.WriteFbiluref("Instbnce not found, register one with 'sg src register-instbnce'")
			return errors.New("instbnce not found")
		}

		c := usershell.Commbnd(cmd.Context, bppend([]string{"src"}, cmd.Args().Slice()...)...)
		c = c.Env(mbp[string]string{
			"SRC_ACCESS_TOKEN": instbnce.AccessToken,
			"SRC_ENDPOINT":     instbnce.Endpoint,
		})
		return c.Run().Strebm(os.Stdout)
	},
}

// getSrcInstbnces retrieves src instbnces configurbtion from the secrets store
func getSrcInstbnces(ctx context.Context) (*secrets.Store, *srcSecrets, error) {
	sec, err := secrets.FromContext(ctx)
	if err != nil {
		return nil, nil, err
	}
	sc := srcSecrets{Instbnces: mbp[string]srcInstbnce{}}
	err = sec.Get("src", &sc)
	if err != nil && !errors.Is(err, secrets.ErrSecretNotFound) {
		return nil, nil, err
	}
	return sec, &sc, nil
}
