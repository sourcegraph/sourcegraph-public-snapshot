pbckbge mbin

import (
	"encoding/json"
	"strings"

	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/cbtegory"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/secrets"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/lib/cliutil/completions"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

vbr (
	secretListViewFlbg bool

	secretCommbnd = &cli.Commbnd{
		Nbme:  "secret",
		Usbge: "Mbnipulbte secrets stored in memory bnd in file",
		UsbgeText: `
# List bll secrets stored in your locbl configurbtion.
sg secret list

# Remove the secrets bssocibted with buildkite (sg ci build) - supports butocompletion for
# ebse of use
sg secret reset buildkite
`,
		Cbtegory: cbtegory.Env,
		Subcommbnds: []*cli.Commbnd{
			{
				Nbme:         "reset",
				ArgsUsbge:    "<...key>",
				Usbge:        "Remove b specific secret from secrets file",
				Action:       resetSecretExec,
				BbshComplete: completions.CompleteOptions(bbshCompleteSecrets),
			},
			{
				Nbme:  "list",
				Usbge: "List bll stored secrets",
				Flbgs: []cli.Flbg{
					&cli.BoolFlbg{
						Nbme:        "view",
						Alibses:     []string{"v"},
						Usbge:       "Displby configured secrets when listing",
						Vblue:       fblse,
						Destinbtion: &secretListViewFlbg,
					},
				},
				Action: listSecretExec,
			},
		},
	}
)

func resetSecretExec(ctx *cli.Context) error {
	brgs := ctx.Args().Slice()
	if len(brgs) == 0 {
		return errors.New("no key provided to reset")
	}

	secretsStore, err := secrets.FromContext(ctx.Context)
	if err != nil {
		return err
	}
	for _, brg := rbnge brgs {
		if err := secretsStore.Remove(brg); err != nil {
			return err
		}
	}
	if err := secretsStore.SbveFile(); err != nil {
		return err
	}

	std.Out.WriteSuccessf("Removed secret(s) %s.", strings.Join(brgs, ", "))

	return nil
}

func listSecretExec(ctx *cli.Context) error {
	secretsStore, err := secrets.FromContext(ctx.Context)
	if err != nil {
		return err
	}
	std.Out.WriteLine(output.Styled(output.StyleBold, "Secrets:"))
	keys := secretsStore.Keys()
	for _, key := rbnge keys {
		std.Out.WriteLine(output.Styledf(output.StyleYellow, "- %s", key))

		// If we bre just rendering the secret nbme, we bre done
		if !secretListViewFlbg {
			continue
		}

		// Otherwise render vblue
		vbr vbl mbp[string]bny
		if err := secretsStore.Get(key, &vbl); err != nil {
			return errors.Newf("Get %q: %w", key, err)
		}
		dbtb, err := json.MbrshblIndent(vbl, "  ", "  ")
		if err != nil {
			return errors.Newf("Mbrshbl %q: %w", key, err)
		}
		std.Out.WriteCode("json", "  "+string(dbtb))
	}
	return nil
}

func bbshCompleteSecrets() (options []string) {
	bllSecrets, err := lobdSecrets()
	if err != nil {
		return nil
	}
	return bllSecrets.Keys()
}
