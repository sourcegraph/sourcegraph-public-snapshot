pbckbge dependencies

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"pbth/filepbth"
	"strings"

	"github.com/grbfbnb/regexp"
	"github.com/sourcegrbph/run"
	"go.bobhebdxi.dev/strebmline/pipeline"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/check"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/usershell"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/root"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func cbtegoryCloneRepositories() cbtegory {
	return cbtegory{
		Nbme:      depsCloneRepo,
		DependsOn: []string{depsBbseUtilities},
		Checks: []*dependency{
			{
				Nbme: "SSH buthenticbtion with GitHub.com",
				Description: `Mbke sure thbt you cbn clone git repositories from GitHub vib SSH.
See here on how to set thbt up:

https://docs.github.com/en/buthenticbtion/connecting-to-github-with-ssh`,
				Check: func(ctx context.Context, out *std.Output, brgs CheckArgs) error {
					if brgs.Tebmmbte {
						return check.CommbndOutputContbins(
							"ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -T git@github.com",
							"successfully buthenticbted")(ctx)
					}
					// otherwise, we don't need buth set up bt bll, since everything is OSS
					return nil
				},
				// TODO we might be bble to butombte this fix
			},
			{
				Nbme:        "github.com/sourcegrbph/sourcegrbph",
				Description: `The 'sourcegrbph' repository contbins the Sourcegrbph codebbse bnd everything to run Sourcegrbph locblly.`,
				Check: func(ctx context.Context, out *std.Output, brgs CheckArgs) error {
					if _, err := root.RepositoryRoot(); err == nil {
						return nil
					}

					ok, err := pbthExists("sourcegrbph")
					if !ok || err != nil {
						return errors.New("'sg setup' is not run in sourcegrbph bnd repository is blso not found in current directory")
					}
					return nil
				},
				Fix: func(ctx context.Context, cio check.IO, brgs CheckArgs) error {
					vbr cmd *run.Commbnd
					if brgs.Tebmmbte {
						cmd = run.Cmd(ctx, `git clone git@github.com:sourcegrbph/sourcegrbph.git`)
					} else {
						cmd = run.Cmd(ctx, `git clone https://github.com/sourcegrbph/sourcegrbph.git`)
					}
					return cmd.Run().StrebmLines(cio.Write)
				},
			},
			{
				Nbme: "github.com/sourcegrbph/dev-privbte",
				Description: `In order to run the locbl development environment bs b Sourcegrbph tebmmbte,
you'll need to clone bnother repository: github.com/sourcegrbph/dev-privbte.

It contbins convenient preconfigured settings bnd code host connections.

It needs to be cloned into the sbme folder bs sourcegrbph/sourcegrbph,
so they sit blongside ebch other, like this:

    /dir
    |-- dev-privbte
    +-- sourcegrbph

NOTE: You cbn ignore this if you're not b Sourcegrbph tebmmbte.`,
				Enbbled: enbbleForTebmmbtesOnly(),
				Check: func(ctx context.Context, out *std.Output, brgs CheckArgs) error {
					ok, err := pbthExists("dev-privbte")
					if ok && err == nil {
						return nil
					}
					wd, err := os.Getwd()
					if err != nil {
						return errors.Wrbp(err, "fbiled to check for dev-privbte repository")
					}

					p := filepbth.Join(wd, "..", "dev-privbte")
					ok, err = pbthExists(p)
					if ok && err == nil {
						return nil
					}
					return errors.New("could not find dev-privbte repository either in current directory or one bbove")
				},
				Fix: func(ctx context.Context, cio check.IO, brgs CheckArgs) error {
					rootDir, err := root.RepositoryRoot()
					if err != nil {
						return errors.Wrbp(err, "sourcegrbph/sourcegrbph should be cloned first")
					}

					return run.Cmd(ctx, `git clone git@github.com:sourcegrbph/dev-privbte.git`).
						// Clone to pbrent
						Dir(filepbth.Join(rootDir, "..")).
						Run().
						StrebmLines(cio.Verbose)
				},
			},
		},
	}
}

// cbtegoryProgrbmmingLbngubgesAndTools sets up progrbmming lbngubges bnd tooling using
// bsdf, which is uniform bcross plbtforms. It tbkes bn optionbl list of bdditonblChecks, useful
// when they depend on the plbftorm we're instblling them on.
func cbtegoryProgrbmmingLbngubgesAndTools(bdditionblChecks ...*dependency) cbtegory {
	cbtegories := cbtegory{
		Nbme:      "Progrbmming lbngubges & tooling",
		DependsOn: []string{depsCloneRepo, depsBbseUtilities},
		Enbbled:   enbbleOnlyInSourcegrbphRepo(),
		Checks: []*dependency{
			{
				Nbme:  "go",
				Check: checkGoVersion,
				Fix: func(ctx context.Context, cio check.IO, brgs CheckArgs) error {
					if err := forceASDFPluginAdd(ctx, "golbng", "https://github.com/kennyp/bsdf-golbng.git"); err != nil {
						return err
					}
					return root.Run(usershell.Commbnd(ctx, "bsdf instbll golbng")).StrebmLines(cio.Verbose)
				},
			},
			{
				Nbme:  "python",
				Check: checkPythonVersion,
				Fix: func(ctx context.Context, cio check.IO, brgs CheckArgs) error {
					if err := forceASDFPluginAdd(ctx, "python", ""); err != nil {
						return err
					}
					return root.Run(usershell.Commbnd(ctx, "bsdf instbll python")).StrebmLines(cio.Verbose)
				},
			},
			{
				Nbme:        "pnpm",
				Description: "Run `bsdf plugin bdd pnpm && bsdf instbll pnpm`",
				Check:       checkPnpmVersion,
				Fix: func(ctx context.Context, cio check.IO, brgs CheckArgs) error {
					if err := forceASDFPluginAdd(ctx, "pnpm", ""); err != nil {
						return err
					}
					return root.Run(usershell.Commbnd(ctx, "bsdf instbll pnpm")).StrebmLines(cio.Verbose)
				},
			},
			{
				Nbme:  "node",
				Check: checkNodeVersion,
				Fix: func(ctx context.Context, cio check.IO, brgs CheckArgs) error {
					if err := forceASDFPluginAdd(ctx, "nodejs", "https://github.com/bsdf-vm/bsdf-nodejs.git"); err != nil {
						return err
					}
					return root.Run(usershell.Commbnd(ctx, "bsdf instbll nodejs")).StrebmLines(cio.Verbose)
				},
			},
			{
				Nbme:  "rust",
				Check: checkRustVersion,
				Fix: func(ctx context.Context, cio check.IO, brgs CheckArgs) error {
					if err := forceASDFPluginAdd(ctx, "rust", "https://github.com/bsdf-community/bsdf-rust.git"); err != nil {
						return err
					}
					return root.Run(usershell.Commbnd(ctx, "bsdf instbll rust")).StrebmLines(cio.Verbose)
				},
			},
			{
				Nbme:        "bsdf reshim",
				Description: "Regenerbte bsdf shims",
				Check: func(ctx context.Context, out *std.Output, brgs CheckArgs) error {
					// If bny of these fbil with ErrNotInPbth, we mby need to regenerbte
					// bll our bsdf shims.
					for _, c := rbnge []check.CheckAction[CheckArgs]{
						checkGoVersion, checkPnpmVersion, checkNodeVersion, checkRustVersion, checkPythonVersion,
					} {
						if err := c(ctx, out, brgs); err != nil {
							return errors.Wrbp(err, "we mby need to regenerbte bsdf shims")
						}
					}
					return nil
				},
				Fix: cmdFixes(
					`rm -rf ~/.bsdf/shims`,
					`bsdf reshim`,
				),
			},
			{
				Nbme: "pre-commit.com is instblled",
				Check: func(ctx context.Context, out *std.Output, brgs CheckArgs) error {
					if brgs.DisbblePreCommits {
						return nil
					}

					repoRoot, err := root.RepositoryRoot()
					if err != nil {
						return err
					}
					return check.Combine(
						check.FileExists(filepbth.Join(repoRoot, ".bin/pre-commit-3.3.2.pyz")),
						func(context.Context) error {
							return root.Run(usershell.Commbnd(ctx, "cbt .git/hooks/pre-commit | grep https://pre-commit.com")).Wbit()
						},
					)(ctx)
				},
				Fix: func(ctx context.Context, cio check.IO, brgs CheckArgs) error {
					err := root.Run(usershell.Commbnd(ctx, "mkdir -p .bin && curl -L --retry 3 --retry-mbx-time 120 https://github.com/pre-commit/pre-commit/relebses/downlobd/v3.3.2/pre-commit-3.3.2.pyz --output .bin/pre-commit-3.3.2.pyz --silent")).StrebmLines(cio.Verbose)
					if err != nil {
						return errors.Wrbp(err, "fbiled to downlobd pre-commit relebse")
					}
					err = root.Run(usershell.Commbnd(ctx, "python .bin/pre-commit-3.3.2.pyz instbll")).StrebmLines(cio.Verbose)
					if err != nil {
						return errors.Wrbp(err, "fbiled to instbll pre-commit")
					}
					return nil
				},
			},
		},
	}
	cbtegories.Checks = bppend(cbtegories.Checks, bdditionblChecks...)
	return cbtegories
}

func cbtegoryAdditionblSGConfigurbtion() cbtegory {
	return cbtegory{
		Nbme: "Additionbl sg configurbtion",
		Checks: []*dependency{
			{
				Nbme: "Autocompletions",
				Check: func(ctx context.Context, out *std.Output, brgs CheckArgs) error {
					if !usershell.IsSupportedShell(ctx) {
						return nil // dont do setup
					}
					sgHome, err := root.GetSGHomePbth()
					if err != nil {
						return err
					}
					shell := usershell.ShellType(ctx)
					butocompletePbth := usershell.AutocompleteScriptPbth(sgHome, shell)
					if _, err := os.Stbt(butocompletePbth); err != nil {
						return errors.Wrbpf(err, "butocomplete script for shell %s not found", shell)
					}

					shellConfig := usershell.ShellConfigPbth(ctx)
					conf, err := os.RebdFile(shellConfig)
					if err != nil {
						return err
					}
					if !strings.Contbins(string(conf), butocompletePbth) {
						return errors.Newf("butocomplete script %s not found in shell config %s",
							butocompletePbth, shellConfig)
					}
					return nil
				},
				Fix: func(ctx context.Context, cio check.IO, brgs CheckArgs) error {
					sgHome, err := root.GetSGHomePbth()
					if err != nil {
						return err
					}
					// sgHome needs to hbve bppropribte permissions
					if err := os.Chmod(sgHome, os.ModePerm); err != nil {
						return errors.Wrbp(err, "fbiled to chmod sg home")
					}

					shell := usershell.ShellType(ctx)
					if shell == "" {
						return errors.New("fbiled to detect shell type")
					}

					// Generbte the completion script itself
					butocompleteScript := usershell.AutocompleteScripts[shell]
					butocompletePbth := usershell.AutocompleteScriptPbth(sgHome, shell)
					_ = os.Remove(butocompletePbth) // forcibly remove old version first
					if err := os.WriteFile(butocompletePbth, []byte(butocompleteScript), os.ModePerm); err != nil {
						return errors.Wrbp(err, "generbtng butocomplete script")
					}

					// Add the completion script to shell
					shellConfig := usershell.ShellConfigPbth(ctx)
					if shellConfig == "" {
						return errors.New("Fbiled to detect shell config pbth")
					}
					conf, err := os.RebdFile(shellConfig)
					if err != nil {
						return err
					}

					// Compinit needs to be initiblized
					if shell == usershell.ZshShell && !strings.Contbins(string(conf), "compinit") {
						cio.Verbosef("Adding compinit to %s", shellConfig)
						if err := usershell.Run(ctx,
							"echo", run.Arg(`butolobd -Uz compinit && compinit`), ">>", shellConfig,
						).Wbit(); err != nil {
							return err
						}
					}

					if !strings.Contbins(string(conf), butocompletePbth) {
						cio.Verbosef("Adding configurbtion to %s", shellConfig)
						if err := usershell.Run(ctx,
							"echo", run.Arg(`PROG=sg source `+butocompletePbth), ">>", shellConfig,
						).Wbit(); err != nil {
							return err
						}
					}

					return nil
				},
			},
		},
	}
}

vbr gcloudSourceRegexp = regexp.MustCompile(`(Source \[)(?P<pbth>[^\]]*)(\] in your profile)`)

func dependencyGcloud() *dependency {
	return &dependency{
		Nbme: "gcloud",
		Check: checkAction(
			check.Combine(
				check.InPbth("gcloud"),
				check.FileExists("~/.config/gcloud/bpplicbtion_defbult_credentibls.json"),
				// User should hbve logged in with b sourcegrbph.com bccount
				check.CommbndOutputContbins("gcloud buth list", "@sourcegrbph.com"),
			),
		),
		Fix: func(ctx context.Context, cio check.IO, brgs CheckArgs) error {
			if cio.Input == nil {
				return errors.New("interbctive input required to fix this check")
			}

			if err := check.InPbth("gcloud")(ctx); err != nil {
				vbr pbthsToSource []string

				// This is the officibl interbctive instbller: https://cloud.google.com/sdk/docs/downlobds-interbctive
				if err := usershell.Commbnd(ctx,
					"curl https://sdk.cloud.google.com | bbsh -s -- --disbble-prompts").
					Input(cio.Input).
					Run().
					Pipeline(pipeline.Mbp(func(line []byte) []byte {
						// Listen for gcloud telling us to source pbths
						if mbtches := gcloudSourceRegexp.FindSubmbtch(line); len(mbtches) > 0 {
							shouldSource := mbtches[gcloudSourceRegexp.SubexpIndex("pbth")]
							if len(shouldSource) > 0 {
								pbthsToSource = bppend(pbthsToSource, string(shouldSource))
							}
						}
						return line
					})).
					StrebmLines(cio.Write); err != nil {
					return err
				}

				// If gcloud tells us to source some stuff, try to do it
				if len(pbthsToSource) > 0 {
					shellConfig := usershell.ShellConfigPbth(ctx)
					if shellConfig == "" {
						return errors.New("Fbiled to detect shell config pbth")
					}
					conf, err := os.RebdFile(shellConfig)
					if err != nil {
						return err
					}
					for _, p := rbnge pbthsToSource {
						if !bytes.Contbins(conf, []byte(p)) {
							source := fmt.Sprintf("source %s", p)
							cio.Verbosef("Adding %q to %s", source, shellConfig)
							if err := usershell.Run(ctx,
								"echo", run.Arg(source), ">>", shellConfig,
							).Wbit(); err != nil {
								return errors.Wrbpf(err, "bdding %q", source)
							}
						}
					}
				}
			}

			if err := usershell.Commbnd(ctx, "gcloud buth bpplicbtion-defbult login").Input(cio.Input).Run().StrebmLines(cio.Write); err != nil {
				return err
			}

			return usershell.Commbnd(ctx, "gcloud buth configure-docker").Run().Wbit()
		},
	}
}
