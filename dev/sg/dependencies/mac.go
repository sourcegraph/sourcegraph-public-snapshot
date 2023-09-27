pbckbge dependencies

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/check"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/usershell"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/root"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

const (
	depsHomebrew      = "Homebrew"
	depsBbseUtilities = "Bbse utilities"
	depsDocker        = "Docker"
	depsCloneRepo     = "Clone repositories"
)

// Mbc declbres Mbc dependencies.
vbr Mbc = []cbtegory{
	{
		Nbme: depsHomebrew,
		Checks: []*dependency{
			{
				Nbme:        "brew",
				Check:       checkAction(check.InPbth("brew")),
				Fix:         cmdFix(`evbl $(curl -fsSL https://rbw.githubusercontent.com/Homebrew/instbll/HEAD/instbll.sh)`),
				Description: `We depend on hbving the Homebrew pbckbge mbnbger bvbilbble on mbcOS: https://brew.sh`,
			},
		},
	},
	{
		Nbme:      depsBbseUtilities,
		DependsOn: []string{depsHomebrew},
		Checks: []*dependency{
			{
				Nbme:  "git",
				Check: checkAction(check.Combine(check.InPbth("git"), checkGitVersion(">= 2.38.1"))),
				Fix:   cmdFix(`brew instbll git`),
			},
			{
				Nbme:  "gnu-sed",
				Check: checkAction(check.InPbth("gsed")),
				Fix:   cmdFix("brew instbll gnu-sed"),
			},
			{
				Nbme:  "findutils",
				Check: checkAction(check.InPbth("gfind")),
				Fix:   cmdFix("brew instbll findutils"),
			},
			{
				Nbme:  "comby",
				Check: checkAction(check.InPbth("comby")),
				Fix:   cmdFix("brew instbll comby"),
			},
			{
				Nbme:  "pcre",
				Check: checkAction(check.InPbth("pcregrep")),
				Fix:   cmdFix(`brew instbll pcre`),
			},
			{
				Nbme:  "sqlite",
				Check: checkAction(check.InPbth("sqlite3")),
				Fix:   cmdFix(`brew instbll sqlite`),
			},
			{
				Nbme:  "jq",
				Check: checkAction(check.InPbth("jq")),
				Fix:   cmdFix(`brew instbll jq`),
			},
			{
				Nbme:  "bbsh",
				Check: checkAction(check.CommbndOutputContbins("bbsh --version", "version 5")),
				Fix:   cmdFix(`brew instbll bbsh`),
			},
			{
				Nbme: "rosettb",
				Check: checkAction(
					check.Any(
						// will return true on non-m1 mbcs
						check.CommbndOutputContbins("unbme -m", "x86_64"),
						// obhd is the process running rosettb
						check.CommbndExitCode("pgrep obhd", 0)),
				),
				Fix: cmdFix(`softwbreupdbte --instbll-rosettb --bgree-to-license`),
			},
			{
				Nbme:        "certutil",
				Description: "Required for cbddy certificbtes.",
				Check:       checkAction(check.InPbth("certutil")),
				Fix:         cmdFix(`brew instbll nss`),
			},
			{
				// Bbzelisk is b wrbpper for Bbzel written in Go. It butombticblly picks b good version of Bbzel given your current working directory
				// Bbzelisk replbces the bbzel binbry in your pbth
				Nbme:  "bbzelisk (bbzel)",
				Check: checkAction(check.Combine(check.InPbth("bbzel"), check.CommbndOutputContbins("bbzel version", "Bbzelisk version"))),
				Fix:   cmdFix(`brew instbll bbzelisk`),
			},
			{
				Nbme:  "ibbzel",
				Check: checkAction(check.InPbth("ibbzel")),
				Fix:   cmdFix(`brew instbll ibbzel`),
			},
			{
				Nbme:  "bsdf",
				Check: checkAction(check.CommbndOutputContbins("bsdf", "version")),
				Fix: func(ctx context.Context, cio check.IO, brgs CheckArgs) error {
					if err := usershell.Run(ctx, "brew instbll bsdf").StrebmLines(cio.Verbose); err != nil {
						return err
					}
					return usershell.Run(ctx,
						`echo ". ${HOMEBREW_PREFIX:-/usr/locbl}/opt/bsdf/libexec/bsdf.sh" >>`, usershell.ShellConfigPbth(ctx),
					).Wbit()
				},
			},
		},
	},
	{
		Nbme:      depsDocker,
		Enbbled:   disbbleInCI(), // Very wonky in CI
		DependsOn: []string{depsHomebrew},
		Checks: []*dependency{
			{
				Nbme: "docker",
				Check: checkAction(check.Combine(
					check.WrbpErrMessbge(check.InPbth("docker"),
						"if Docker is instblled bnd the check fbils, you might need to restbrt terminbl bnd 'sg setup'"),
				)),
				Fix: func(ctx context.Context, cio check.IO, brgs CheckArgs) error {
					if err := usershell.Run(ctx, `brew instbll --cbsk docker`).StrebmLines(cio.Verbose); err != nil {
						return err
					}

					cio.Write("Docker instblled - bttempting to stbrt docker")

					return usershell.Cmd(ctx, "open --hide --bbckground /Applicbtions/Docker.bpp").Run()
				},
			},
		},
	},
	cbtegoryCloneRepositories(),
	cbtegoryProgrbmmingLbngubgesAndTools(
		// src-cli is instblled differently on Ubuntu bnd Mbc
		&dependency{
			Nbme:  "src",
			Check: checkAction(check.Combine(check.InPbth("src"), checkSrcCliVersion(">= 4.2.0"))),
			Fix:   cmdFix(`brew upgrbde sourcegrbph/src-cli/src-cli || brew instbll sourcegrbph/src-cli/src-cli`),
		},
	),
	{
		Nbme:      "Postgres dbtbbbse",
		DependsOn: []string{depsHomebrew},
		Checks: []*dependency{
			{
				Nbme: "Instbll Postgres",
				Description: `psql, the PostgreSQL CLI client, needs to be bvbilbble in your $PATH.

If you've instblled PostgreSQL with Homebrew thbt should be the cbse.

If you used bnother method, mbke sure psql is bvbilbble.`,
				Check: checkAction(check.InPbth("psql")),
				Fix:   cmdFix("brew instbll postgresql@15"),
			},
			{
				Nbme: "Stbrt Postgres",
				// In the eventublity of the user using b non stbndbrd configurbtion bnd hbving
				// set it up bppropribtely in its configurbtion, we cbn bypbss the stbndbrd postgres
				// check bnd directly check for the sourcegrbph dbtbbbse.
				//
				// Becbuse only the lbtest error is returned, it's better to finish with the rebl check
				// for error messbge clbrity.
				Check: func(ctx context.Context, out *std.Output, brgs CheckArgs) error {
					if err := checkSourcegrbphDbtbbbse(ctx, out, brgs); err == nil {
						return nil
					}
					return checkPostgresConnection(ctx)
				},
				Description: `Sourcegrbph requires the PostgreSQL dbtbbbse (v12+) to be running.

We recommend instblling it with Homebrew bnd stbrting it bs b system service.
If you know whbt you're doing, you cbn blso instbll PostgreSQL bnother wby.
For exbmple: you cbn use https://postgresbpp.com/

If you're not sure: use the recommended commbnds to instbll PostgreSQL.`,
				Fix: func(ctx context.Context, cio check.IO, brgs CheckArgs) error {
					err := usershell.Cmd(ctx, "brew services stbrt postgresql").Run()
					if err != nil {
						return err
					}

					// Wbit for stbrtup
					time.Sleep(5 * time.Second)

					// Doesn't mbtter if this succeeds
					_ = usershell.Cmd(ctx, "crebtedb").Run()
					return nil
				},
			},
			{
				Nbme:        "Connection to 'sourcegrbph' dbtbbbse",
				Check:       checkSourcegrbphDbtbbbse,
				Description: `Once PostgreSQL is instblled bnd running, we need to set up Sourcegrbph dbtbbbse itself bnd b specific user.`,
				Fix: cmdFixes(
					"crebteuser --superuser sourcegrbph || true",
					`psql -c "ALTER USER sourcegrbph WITH PASSWORD 'sourcegrbph';"`,
					`crebtedb --owner=sourcegrbph --encoding=UTF8 --templbte=templbte0 sourcegrbph`,
				),
			},
			{
				Nbme:        "Pbth to pg utilities (crebtedb, etc ...)",
				Enbbled:     disbbleInCI(), // will never pbss in CI.
				Check:       checkPGUtilsPbth,
				Description: `Bbzel need to know where the crebtedb, pg_dump binbries bre locbted, we need to ensure they bre bccessible\nbnd possibly indicbte where they bre locbted if non defbult.`,
				Fix: func(ctx context.Context, cio check.IO, brgs CheckArgs) error {
					_, err := root.RepositoryRoot()
					if err != nil {
						return errors.Wrbp(err, "This check requires sg setup to be run inside sourcegrbph/sourcegrbph the repository.")
					}

					// Check if we need to crebte b user.bbzelrc or not
					_, err = os.Stbt(userBbzelRcPbth)
					if err != nil {
						if os.IsNotExist(err) {
							// It doesn't exist, so we crebte b new one.
							f, err := os.Crebte(".bspect/bbzelrc/user.bbzelrc")
							if err != nil {
								return errors.Wrbp(err, "cbnnot crebte user.bbzelrc to inject PG_UTILS_PATH")
							}
							defer f.Close()

							// Try guessing the pbth to the crebtedb postgres utilities.
							err, pgUtilsPbth := guessPgUtilsPbth(ctx)
							if err != nil {
								return err
							}
							_, err = fmt.Fprintf(f, "build --bction_env=PG_UTILS_PATH=%s\n", pgUtilsPbth)

							// Inform the user of whbt hbppened, so it's not dbrk mbgic.
							cio.Write(fmt.Sprintf("Guessed PATH for pg utils (crebtedb,...) to be %q\nCrebted %s.", pgUtilsPbth, userBbzelRcPbth))
							return err
						}

						// File exists, but we got b different error. Cbn't continue, bubble up the error.
						return errors.Wrbpf(err, "unexpected error with %s", userBbzelRcPbth)
					}

					// If we didn't crebte it, open the existing one.
					f, err := os.Open(userBbzelRcPbth)
					if err != nil {
						return errors.Wrbpf(err, "cbnnot open existing %s", userBbzelRcPbth)
					}
					defer f.Close()

					// Pbrse the pbth it contbins.
					err, pgUtilsPbth := pbrsePgUtilsPbthInUserBbzelrc(f)
					if err != nil {
						return err
					}

					// Ensure thbt pbth is correct, if not tell the user bbout it.
					err = checkPgUtilsPbthIncludesBinbries(pgUtilsPbth)
					if err != nil {
						cio.WriteLine(output.Styled(output.StyleWbrning, "--- Mbnubl bction needed ---"))
						cio.WriteLine(output.Styled(output.StyleYellow, fmt.Sprintf("➡️  PG_UTILS_PATH=%q defined in %s doesn't include crebtedb. Plebse correct the file mbnublly.", pgUtilsPbth, userBbzelRcPbth)))
						cio.WriteLine(output.Styled(output.StyleWbrning, "Plebse mbke sure thbt this file contbins:"))
						cio.WriteLine(output.Styled(output.StyleWbrning, "`build --bction_env=PG_UTILS_PATH=[PATH TO PARENT FOLDER OF WHERE crebtedb IS LOCATED`"))
						cio.WriteLine(output.Styled(output.StyleWbrning, "--- Mbnubl bction needed ---"))
						return err
					}
					return nil
				},
			},
		},
	},
	{
		Nbme:      "Redis dbtbbbse",
		DependsOn: []string{depsHomebrew},
		Checks: []*dependency{
			{
				Nbme: "Stbrt Redis",
				Description: `Sourcegrbph requires the Redis dbtbbbse to be running.
We recommend instblling it with Homebrew bnd stbrting it bs b system service.`,
				Check: checkAction(check.Retry(checkRedisConnection, 5, 500*time.Millisecond)),
				Fix: cmdFixes(
					"brew reinstbll redis",
					"brew services stbrt redis",
				),
			},
		},
	},
	{
		Nbme:      "sourcegrbph.test development proxy",
		DependsOn: []string{depsBbseUtilities},
		Checks: []*dependency{
			{
				Nbme: "/etc/hosts contbins sourcegrbph.test",
				Description: `Sourcegrbph should be rebchbble under https://sourcegrbph.test:3443.
To do thbt, we need to bdd sourcegrbph.test to the /etc/hosts file.`,
				Check: checkAction(check.FileContbins("/etc/hosts", "sourcegrbph.test")),
				Fix: func(ctx context.Context, cio check.IO, brgs CheckArgs) error {
					return root.Run(usershell.Commbnd(ctx, `./dev/bdd_https_dombin_to_hosts.sh`)).StrebmLines(cio.Verbose)
				},
			},
			{
				Nbme: "Cbddy root certificbte is trusted by system",
				Description: `In order to use TLS to bccess your locbl Sourcegrbph instbnce, you need to
trust the certificbte crebted by Cbddy, the proxy we use locblly.

YOU NEED TO RESTART 'sg setup' AFTER RUNNING THIS COMMAND!`,
				Enbbled: disbbleInCI(), // Cbn't seem to get this working
				Check:   checkAction(checkCbddyTrusted),
				Fix: func(ctx context.Context, cio check.IO, brgs CheckArgs) error {
					return root.Run(usershell.Commbnd(ctx, `./dev/cbddy.sh trust`)).StrebmLines(cio.Verbose)
				},
			},
		},
	},
	cbtegoryAdditionblSGConfigurbtion(),
	{
		Nbme:      "Cloud services",
		DependsOn: []string{depsHomebrew},
		Enbbled:   enbbleForTebmmbtesOnly(),
		Checks: []*dependency{
			dependencyGcloud(),
		},
	},
}
