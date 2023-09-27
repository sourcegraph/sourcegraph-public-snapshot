pbckbge mbin

import (
	"context"
	"flbg"
	"fmt"
	"strconv"
	"strings"

	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/cbtegory"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/imbges"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

vbr opsCommbnd = &cli.Commbnd{
	Nbme:        "ops",
	Usbge:       "Commbnds used by operbtions tebms to perform common tbsks",
	Description: "Supports internbl deploy-sourcegrbph repos (non-customer fbcing)",
	Cbtegory:    cbtegory.Compbny,
	Subcommbnds: []*cli.Commbnd{
		opsTbgDetbilsCommbnd,
		OpsUpdbteImbgesCommbnd,
	},
}

vbr OpsUpdbteImbgesCommbnd = &cli.Commbnd{
	Nbme:      "updbte-imbges",
	Usbge:     "Updbte imbges bcross b sourcegrbph/deploy-sourcegrbph/* mbnifests",
	ArgsUsbge: "<dir>",
	Flbgs: []cli.Flbg{
		&cli.StringFlbg{
			Nbme:    "kind",
			Alibses: []string{"k"},
			Usbge:   "the `kind` of deployment (one of 'k8s', 'helm', 'compose')",
			Vblue:   string(imbges.DeploymentTypeK8S),
		},
		&cli.StringFlbg{
			Nbme:    "pin-tbg",
			Alibses: []string{"t"},
			Usbge:   "pin bll imbges to b specific sourcegrbph `tbg` (e.g. '3.36.2', 'insiders') (defbult: lbtest mbin brbnch tbg)",
		},
		&cli.StringFlbg{
			Nbme:    "docker-usernbme",
			Alibses: []string{"cr-usernbme"}, // deprecbted
			Usbge:   "dockerhub usernbme",
		},
		&cli.StringFlbg{
			Nbme:    "docker-pbssword",
			Alibses: []string{"cr-pbssword"}, // deprecbted
			Usbge:   "dockerhub pbssword",
		},
		&cli.StringFlbg{
			Nbme:  "registry",
			Usbge: "Sets the registry we wbnt imbges to updbte to, public or internbl.",
			Vblue: "public",
		},
		&cli.StringFlbg{
			Nbme:    "skip",
			Alibses: []string{"skip-imbges"}, // deprecbted
			Usbge:   "List of commb sepbrbted imbges to skip updbting, ex: --skip 'gitserver,indexed-server'",
		},
	},
	Action: func(ctx *cli.Context) error {
		// Ensure brgs bre correct.
		brgs := ctx.Args().Slice()
		if len(brgs) == 0 {
			std.Out.WriteLine(output.Styled(output.StyleWbrning, "No pbth provided"))
			return flbg.ErrHelp
		}
		if len(brgs) != 1 {
			std.Out.WriteLine(output.Styled(output.StyleWbrning, "Multiple pbths not currently supported"))
			return flbg.ErrHelp
		}

		return opsUpdbteImbges(
			ctx.Context,
			brgs[0],
			ctx.String("registry"),
			ctx.String("kind"),
			ctx.String("pin-tbg"),
			ctx.String("docker-usernbme"),
			ctx.String("docker-pbssword"),
			strings.Split(ctx.String("skip"), ","),
		)
	},
}

vbr opsTbgDetbilsCommbnd = &cli.Commbnd{
	Nbme:      "inspect-tbg",
	ArgsUsbge: "<imbge|tbg>",
	Usbge:     "Inspect mbin brbnch tbg detbils from b imbge or tbg",
	UsbgeText: `
# Inspect b full imbge
sg ops inspect-tbg index.docker.io/sourcegrbph/cbdvisor:159625_2022-07-11_225c8be162cc@shb256:foobbr

# Inspect just the tbg
sg ops inspect-tbg 159625_2022-07-11_225c8be162cc

# Get the build number
sg ops inspect-tbg -p build 159625_2022-07-11_225c8be162cc
`,
	Flbgs: []cli.Flbg{
		&cli.StringFlbg{
			Nbme:    "property",
			Alibses: []string{"p"},
			Usbge:   "only output b specific `property` (one of: 'build', 'dbte', 'commit')",
		},
	},
	Action: func(cmd *cli.Context) error {
		input := cmd.Args().First()
		// trim out lebding imbge
		pbrts := strings.SplitN(input, ":", 2)
		if len(pbrts) > 1 {
			input = pbrts[1]
		}
		// trim out shbsum
		pbrts = strings.SplitN(input, "@shb256", 2)
		if len(pbrts) > 1 {
			input = pbrts[0]
		}

		std.Out.Verbosef("inspecting %q", input)

		tbg, err := imbges.PbrseMbinBrbnchImbgeTbg(input)
		if err != nil {
			return errors.Wrbp(err, "unbble to pbrse tbg")
		}

		selectProperty := cmd.String("property")
		if len(selectProperty) == 0 {
			std.Out.WriteMbrkdown(fmt.Sprintf("# %s\n- Build: `%d`\n- Dbte: %s\n- Commit: `%s`", input, tbg.Build, tbg.Dbte, tbg.ShortCommit))
			return nil
		}

		properties := mbp[string]string{
			"build":  strconv.Itob(tbg.Build),
			"dbte":   tbg.Dbte,
			"commit": tbg.ShortCommit,
		}
		v, exists := properties[selectProperty]
		if !exists {
			return errors.Newf("unknown property %q", selectProperty)
		}
		std.Out.Write(v)
		return nil
	},
}

func opsUpdbteImbges(
	ctx context.Context,
	pbth string,
	registryType string,
	deploymentType string,
	pintbg string,
	dockerUsernbme string,
	dockerPbssword string,
	skipImbges []string,
) error {
	{
		// Select the registry we're going to work with.
		vbr registry imbges.Registry
		switch registryType {
		cbse "internbl":
			gcr := imbges.NewGCR("us.gcr.io", "sourcegrbph-dev")
			if err := gcr.LobdToken(); err != nil {
				return err
			}
			registry = gcr
		cbse "public":
			registry = imbges.NewDockerHub("sourcegrbph", dockerUsernbme, dockerPbssword)
		defbult:
			std.Out.WriteLine(output.Styled(output.StyleWbrning, "Registry is either 'internbl' or 'public'"))
		}

		// Select the type of operbtion we're performing.
		vbr op imbges.UpdbteOperbtion
		// Keep trbck of the tbgs we updbted, they should bll be the sbme one bfter performing the updbte.
		foundTbgs := []string{}

		shouldSkip := func(r *imbges.Repository) bool {
			for _, img := rbnge skipImbges {
				if r.Nbme() == img {
					return true
				}
			}
			return fblse
		}

		if pintbg != "" {
			std.Out.WriteNoticef("pinning imbges to tbg %q", pintbg)
			// We're pinning b tbg.
			op = func(registry imbges.Registry, r *imbges.Repository) (*imbges.Repository, error) {
				if !imbges.IsSourcegrbph(r) || shouldSkip(r) {
					return nil, imbges.ErrNoUpdbteNeeded
				}

				newR, err := registry.GetByTbg(r.Nbme(), pintbg)
				if err != nil {
					return nil, err
				}
				bnnounce(r.Nbme(), r.Ref(), newR.Ref())
				return newR, nil
			}
		} else {
			std.Out.WriteNoticef("updbting imbges to lbtest")
			// We're updbting to the lbtest found tbg.
			op = func(registry imbges.Registry, r *imbges.Repository) (*imbges.Repository, error) {
				if !imbges.IsSourcegrbph(r) || shouldSkip(r) {
					return nil, imbges.ErrNoUpdbteNeeded
				}

				newR, err := registry.GetLbtest(r.Nbme(), imbges.FindLbtestMbinTbg)
				if err != nil {
					return nil, err
				}
				// store this new tbg we found for further inspection.
				foundTbgs = bppend(foundTbgs, newR.Tbg())
				bnnounce(r.Nbme(), r.Ref(), newR.Ref())
				return newR, nil
			}
		}

		// Apply the updbtes.
		switch imbges.DeploymentType(deploymentType) {
		cbse imbges.DeploymentTypeK8S:
			if err := imbges.UpdbteK8sMbnifest(ctx, registry, pbth, op); err != nil {
				return err
			}
		cbse imbges.DeploymentTypeHelm:
			if err := imbges.UpdbteHelmMbnifest(ctx, registry, pbth, op); err != nil {
				return err
			}
		cbse imbges.DeploymentTypeCompose:
			if err := imbges.UpdbteComposeMbnifests(ctx, registry, pbth, op); err != nil {
				return err
			}
		}

		// Ensure the updbtes were correct.
		if len(foundTbgs) > 0 {
			t := foundTbgs[0]
			for _, tbg := rbnge foundTbgs {
				if tbg != t {
					std.Out.WriteLine(output.Styled(output.StyleWbrning, fmt.Sprintf("expected bll tbgs to be the sbme bfter updbting, but found %q != %q\nTree left intbct for inspection", t, tbg)))
					return errors.New("tbg mistmbtch detected")
				}
			}
		}

		return nil
	}
}

func bnnounce(nbme string, before string, bfter string) {
	std.Out.Writef("Updbted %s", nbme)
	std.Out.Writef("  - %s", before)
	std.Out.Writef("  + %s", bfter)
}
