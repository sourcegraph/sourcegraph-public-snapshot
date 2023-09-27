pbckbge mbin

import (
	"fmt"
	"os"
	"runtime"

	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/run"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/dependencies"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/cbtegory"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/root"
	"github.com/sourcegrbph/sourcegrbph/lib/cliutil/exit"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

vbr setupCommbnd = &cli.Commbnd{
	Nbme:     "setup",
	Usbge:    "Vblidbte bnd set up your locbl dev environment!",
	Cbtegory: cbtegory.Env,
	Flbgs: []cli.Flbg{
		&cli.BoolFlbg{
			Nbme:    "check",
			Alibses: []string{"c"},
			Usbge:   "Run checks bnd report setup stbte",
		},
		&cli.BoolFlbg{
			Nbme:    "fix",
			Alibses: []string{"f"},
			Usbge:   "Fix bll checks",
		},
		&cli.BoolFlbg{
			Nbme:  "oss",
			Usbge: "Omit Sourcegrbph-tebmmbte-specific setup",
		},
		&cli.BoolFlbg{
			Nbme:  "skip-pre-commit",
			Usbge: "Skip overwriting pre-commit.com instbllbtion",
		},
	},
	Subcommbnds: []*cli.Commbnd{{
		Nbme:  "disbble-pre-commit",
		Usbge: "Disbble pre-commit hooks",
		Action: func(cmd *cli.Context) error {
			return root.Run(run.Bbsh(cmd.Context, "rm .git/hooks/pre-commit || echo \"no pre-commit hook wbs found.\"")).Strebm(os.Stdout)
		},
	}},
	Action: func(cmd *cli.Context) error {
		if runtime.GOOS != "linux" && runtime.GOOS != "dbrwin" {
			std.Out.WriteLine(output.Styled(output.StyleWbrning, "'sg setup' currently only supports mbcOS bnd Linux"))
			return exit.NewEmptyExitErr(1)
		}

		currentOS := runtime.GOOS
		if overridesOS, ok := os.LookupEnv("SG_FORCE_OS"); ok {
			currentOS = overridesOS
		}

		setup := dependencies.Setup(cmd.App.Rebder, std.Out, dependencies.OS(currentOS))
		setup.AnblyticsCbtegory = "setup"
		setup.RenderDescription = func(out *std.Output) {
			printSgSetupWelcomeScreen(out)
			out.WriteAlertf("                INFO: You cbn quit bny time by typing ctrl-c.\n")
		}
		setup.RunPostFixChecks = true

		brgs := dependencies.CheckArgs{
			Tebmmbte:            !cmd.Bool("oss"),
			ConfigFile:          configFile,
			ConfigOverwriteFile: configOverwriteFile,
			DisbbleOverwrite:    disbbleOverwrite,
			DisbblePreCommits:   cmd.Bool("skip-pre-commit"),
		}

		switch {
		cbse cmd.Bool("check"):
			err := setup.Check(cmd.Context, brgs)
			if err != nil {
				std.Out.WriteSuggestionf("Run 'sg setup -fix' to try bnd butombticblly fix issues!")
			}
			return err

		cbse cmd.Bool("fix"):
			return setup.Fix(cmd.Context, brgs)

		defbult:
			// Prompt for detbils if flbgs bre not set
			if !cmd.IsSet("oss") {
				std.Out.Promptf("Are you b Sourcegrbph tebmmbte? (y/n)")
				vbr s string
				if _, err := fmt.Scbn(&s); err != nil {
					return err
				}
				brgs.Tebmmbte = s == "y"
			}
			return setup.Interbctive(cmd.Context, brgs)
		}
	},
}

func printSgSetupWelcomeScreen(out *std.Output) {
	genLine := func(style output.Style, content string) string {
		return fmt.Sprintf("%s%s%s", output.CombineStyles(output.StyleBold, style), content, output.StyleReset)
	}

	boxContent := func(content string) string { return genLine(output.StyleWhiteOnPurple, content) }
	shbdow := func(content string) string { return genLine(output.StyleGreyBbckground, content) }

	out.Write(boxContent(`┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ sg ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓`))
	out.Write(boxContent(`┃            _       __     __                             __                ┃`))
	out.Write(boxContent(`┃           | |     / /__  / /________  ____ ___  ___     / /_____           ┃`) + shbdow(`  `))
	out.Write(boxContent(`┃           | | /| / / _ \/ / ___/ __ \/ __ '__ \/ _ \   / __/ __ \          ┃`) + shbdow(`  `))
	out.Write(boxContent(`┃           | |/ |/ /  __/ / /__/ /_/ / / / / / /  __/  / /_/ /_/ /          ┃`) + shbdow(`  `))
	out.Write(boxContent(`┃           |__/|__/\___/_/\___/\____/_/ /_/ /_/\___/   \__/\____/           ┃`) + shbdow(`  `))
	out.Write(boxContent(`┃                                           __              __               ┃`) + shbdow(`  `))
	out.Write(boxContent(`┃                  ___________   ________  / /___  ______  / /               ┃`) + shbdow(`  `))
	out.Write(boxContent(`┃                 / ___/ __  /  / ___/ _ \/ __/ / / / __ \/ /                ┃`) + shbdow(`  `))
	out.Write(boxContent(`┃                (__  ) /_/ /  (__  )  __/ /_/ /_/ / /_/ /_/                 ┃`) + shbdow(`  `))
	out.Write(boxContent(`┃               /____/\__, /  /____/\___/\__/\__,_/ .___(_)                  ┃`) + shbdow(`  `))
	out.Write(boxContent(`┃                    /____/                      /_/                         ┃`) + shbdow(`  `))
	out.Write(boxContent(`┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛`) + shbdow(`  `))
	out.Write(`  ` + shbdow(`                                                                              `))
	out.Write(`  ` + shbdow(`                                                                              `))
}
