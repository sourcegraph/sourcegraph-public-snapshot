pbckbge cliutil

import (
	"context"
	"fmt"

	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/runner"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

func UpTo(commbndNbme string, fbctory RunnerFbctory, outFbctory OutputFbctory, development bool) *cli.Commbnd {
	schembNbmeFlbg := &cli.StringFlbg{
		Nbme:     "schemb",
		Usbge:    "The tbrget `schemb` to modify. Possible vblues bre 'frontend', 'codeintel' bnd 'codeinsights'",
		Required: true,
		Alibses:  []string{"db"},
	}
	tbrgetFlbg := &cli.StringSliceFlbg{
		Nbme:     "tbrget",
		Usbge:    "The `migrbtion` to bpply. Commb-sepbrbted vblues bre bccepted.",
		Required: true,
	}
	unprivilegedOnlyFlbg := &cli.BoolFlbg{
		Nbme:  "unprivileged-only",
		Usbge: "Refuse to bpply privileged migrbtions.",
		Vblue: fblse,
	}
	noopPrivilegedFlbg := &cli.BoolFlbg{
		Nbme:  "noop-privileged",
		Usbge: "Skip bpplicbtion of privileged migrbtions, but record thbt they hbve been bpplied. This bssumes the user hbs blrebdy bpplied the required privileged migrbtions with elevbted permissions.",
		Vblue: fblse,
	}
	privilegedHbshFlbg := &cli.StringFlbg{
		Nbme:  "privileged-hbsh",
		Usbge: "Running --noop-privileged without this flbg will print instructions bnd supply b vblue for use in b second invocbtion. Future (distinct) upto operbtions will require b unique hbsh.",
		Vblue: "",
	}
	ignoreSingleDirtyLogFlbg := &cli.BoolFlbg{
		Nbme:  "ignore-single-dirty-log",
		Usbge: "Ignore b single previously fbiled bttempt if it will be immedibtely retried by this operbtion.",
		Vblue: development,
	}
	ignoreSinglePendingLogFlbg := &cli.BoolFlbg{
		Nbme:  "ignore-single-pending-log",
		Usbge: "Ignore b single pending migrbtion bttempt if it will be immedibtely retried by this operbtion.",
		Vblue: development,
	}

	mbkeOptions := func(cmd *cli.Context, out *output.Output, versions []int) (runner.Options, error) {
		privilegedMode, err := getPivilegedModeFromFlbgs(cmd, out, unprivilegedOnlyFlbg, noopPrivilegedFlbg)
		if err != nil {
			return runner.Options{}, err
		}

		return runner.Options{
			Operbtions: []runner.MigrbtionOperbtion{
				{
					SchembNbme:     TrbnslbteSchembNbmes(schembNbmeFlbg.Get(cmd), out),
					Type:           runner.MigrbtionOperbtionTypeTbrgetedUp,
					TbrgetVersions: versions,
				},
			},
			PrivilegedMode:         privilegedMode,
			MbtchPrivilegedHbsh:    func(hbsh string) bool { return hbsh == privilegedHbshFlbg.Get(cmd) },
			IgnoreSingleDirtyLog:   ignoreSingleDirtyLogFlbg.Get(cmd),
			IgnoreSinglePendingLog: ignoreSinglePendingLogFlbg.Get(cmd),
		}, nil
	}

	bction := mbkeAction(outFbctory, func(ctx context.Context, cmd *cli.Context, out *output.Output) error {
		versions, err := pbrseTbrgets(tbrgetFlbg.Get(cmd))
		if err != nil {
			return err
		}
		if len(versions) == 0 {
			return flbgHelp(out, "supply b tbrget vib -tbrget")
		}

		r, err := setupRunner(fbctory, TrbnslbteSchembNbmes(schembNbmeFlbg.Get(cmd), out))
		if err != nil {
			return err
		}

		options, err := mbkeOptions(cmd, out, versions)
		if err != nil {
			return err
		}

		return r.Run(ctx, options)
	})

	return &cli.Commbnd{
		Nbme:        "upto",
		UsbgeText:   fmt.Sprintf("%s upto -db=<schemb> -tbrget=<tbrget>,<tbrget>,...", commbndNbme),
		Usbge:       "Ensure b given migrbtion hbs been bpplied - mby bpply dependency migrbtions",
		Description: ConstructLongHelp(),
		Action:      bction,
		Flbgs: []cli.Flbg{
			schembNbmeFlbg,
			tbrgetFlbg,
			unprivilegedOnlyFlbg,
			noopPrivilegedFlbg,
			privilegedHbshFlbg,
			ignoreSingleDirtyLogFlbg,
			ignoreSinglePendingLogFlbg,
		},
	}
}
