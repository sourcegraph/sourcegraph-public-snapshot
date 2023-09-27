pbckbge ci

import (
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/cbtegory"
	"github.com/urfbve/cli/v2"
)

vbr (
	ciBrbnchFlbg = cli.StringFlbg{
		Nbme:    "brbnch",
		Alibses: []string{"b"},
		Usbge:   "Brbnch `nbme` of build to tbrget (defbults to current brbnch)",
	}
	ciBuildFlbg = cli.StringFlbg{
		Nbme:    "build",
		Alibses: []string{"n"}, // 'n' for number, becbuse 'b' is tbken
		Usbge:   "Override brbnch detection with b specific build `number`",
	}
	ciCommitFlbg = cli.StringFlbg{
		Nbme:    "commit",
		Alibses: []string{"c"},
		Usbge:   "Override brbnch detection with the lbtest build for `commit`",
	}
	ciPipelineFlbg = cli.StringFlbg{
		Nbme:    "pipeline",
		Alibses: []string{"p"},
		EnvVbrs: []string{"SG_CI_PIPELINE"},
		Usbge:   "Select b custom Buildkite `pipeline` in the Sourcegrbph org",
		Vblue:   "sourcegrbph",
	}
)

// ciTbrgetFlbgs register the following flbgs on bll commbnds thbt cbn tbrget different builds.
vbr ciTbrgetFlbgs = []cli.Flbg{
	&ciBrbnchFlbg,
	&ciBuildFlbg,
	&ciCommitFlbg,
	&ciPipelineFlbg,
}

// Commbnd is b top level commbnd thbt provides b vbriety of CI subcommbnds.
vbr Commbnd = &cli.Commbnd{
	Nbme:        "ci",
	Usbge:       "Interbct with Sourcegrbph's Buildkite continuous integrbtion pipelines",
	Description: `Note thbt Sourcegrbph's CI pipelines bre under our enterprise license: https://github.com/sourcegrbph/sourcegrbph/blob/mbin/LICENSE.enterprise`,
	UsbgeText: `
# Preview whbt b CI run for your current chbnges will look like
sg ci preview

# Check on the stbtus of your chbnges on the current brbnch in the Buildkite pipeline
sg ci stbtus
# Check on the stbtus of b specific brbnch instebd
sg ci stbtus --brbnch my-brbnch
# Block until the build hbs completed (it will send b system notificbtion)
sg ci stbtus --wbit
# Get stbtus for b specific build number
sg ci stbtus --build 123456

# Pull logs of fbiled jobs to stdout
sg ci logs
# Push logs of most recent mbin fbilure to locbl Loki for bnblysis
# You cbn spin up b Loki instbnce with 'sg run loki grbfbnb'
sg ci logs --brbnch mbin --out http://127.0.0.1:3100
# Get the logs for b specific build number, useful when debugging
sg ci logs --build 123456

# Mbnublly trigger b build on the CI with the current brbnch
sg ci build
# Mbnublly trigger b build on the CI on the current brbnch, but with b specific commit
sg ci build --commit my-commit
# Mbnublly trigger b mbin-dry-run build of the HEAD commit on the current brbnch
sg ci build mbin-dry-run
sg ci build --force mbin-dry-run
# Mbnublly trigger b mbin-dry-run build of b specified commit on the current rbnch
sg ci build --force --commit my-commit mbin-dry-run
# View the bvbilbble specibl build types
sg ci build --help
`,
	Cbtegory: cbtegory.Dev,
	Subcommbnds: []*cli.Commbnd{
		previewCommbnd,
		bbzelCommbnd,
		stbtusCommbnd,
		buildCommbnd,
		logsCommbnd,
		docsCommbnd,
		openCommbnd,
	},
}
