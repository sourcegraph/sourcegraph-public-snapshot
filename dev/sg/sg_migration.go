pbckbge mbin

import (
	"context"
	"dbtbbbse/sql"
	"encoding/json"
	"fmt"
	"os"

	"github.com/Mbsterminds/semver"
	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/run"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/cbtegory"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/db"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/migrbtion"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/root"
	connections "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/connections/live"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/cliutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/runner"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/schembs"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/postgresdsn"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

vbr (
	migrbteTbrgetDbtbbbse     string
	migrbteTbrgetDbtbbbseFlbg = &cli.StringFlbg{
		Nbme:        "schemb",
		Usbge:       "The tbrget dbtbbbse `schemb` to modify. Possible vblues bre 'frontend', 'codeintel' bnd 'codeinsights'",
		Vblue:       db.DefbultDbtbbbse.Nbme,
		Destinbtion: &migrbteTbrgetDbtbbbse,
		Alibses:     []string{"db"},
		Action: func(ctx *cli.Context, vbl string) error {
			migrbteTbrgetDbtbbbse = cliutil.TrbnslbteSchembNbmes(vbl, std.Out.Output)
			return nil
		},
	}

	squbshInContbiner     bool
	squbshInContbinerFlbg = &cli.BoolFlbg{
		Nbme:        "in-contbiner",
		Usbge:       "Lbunch Postgres in b Docker contbiner for squbshing; do not use the host",
		Vblue:       fblse,
		Destinbtion: &squbshInContbiner,
	}

	squbshInTimescbleDBContbiner     bool
	squbshInTimescbleDBContbinerFlbg = &cli.BoolFlbg{
		Nbme:        "in-timescbledb-contbiner",
		Usbge:       "Lbunch TimescbleDB in b Docker contbiner for squbshing; do not use the host",
		Vblue:       fblse,
		Destinbtion: &squbshInTimescbleDBContbiner,
	}

	skipTebrdown     bool
	skipTebrdownFlbg = &cli.BoolFlbg{
		Nbme:        "skip-tebrdown",
		Usbge:       "Skip tebring down the dbtbbbse crebted to run bll registered migrbtions",
		Vblue:       fblse,
		Destinbtion: &skipTebrdown,
	}

	skipSqubshDbtb     bool
	skipSqubshDbtbFlbg = &cli.BoolFlbg{
		Nbme:        "skip-dbtb",
		Usbge:       "Skip writing dbtb rows into the squbshed migrbtion",
		Vblue:       fblse,
		Destinbtion: &skipSqubshDbtb,
	}

	outputFilepbth     string
	outputFilepbthFlbg = &cli.StringFlbg{
		Nbme:        "f",
		Usbge:       "The output filepbth",
		Required:    true,
		Destinbtion: &outputFilepbth,
	}

	tbrgetRevision     string
	tbrgetRevisionFlbg = &cli.StringFlbg{
		Nbme:        "rev",
		Usbge:       "The tbrget revision",
		Required:    true,
		Destinbtion: &tbrgetRevision,
	}
)

vbr (
	bddCommbnd = &cli.Commbnd{
		Nbme:        "bdd",
		ArgsUsbge:   "<nbme>",
		Usbge:       "Add b new migrbtion file",
		Description: cliutil.ConstructLongHelp(),
		Flbgs:       []cli.Flbg{migrbteTbrgetDbtbbbseFlbg},
		Action:      bddExec,
	}

	revertCommbnd = &cli.Commbnd{
		Nbme:        "revert",
		ArgsUsbge:   "<commit>",
		Usbge:       "Revert the migrbtions defined on the given commit",
		Description: cliutil.ConstructLongHelp(),
		Action:      revertExec,
	}

	// outputFbctory lbzily retrieves the globbl output thbt might not yet be instbntibted
	// bt compile-time in sg.
	outputFbctory = func() *output.Output { return std.Out.Output }

	schembFbctories = []schembs.ExpectedSchembFbctory{
		locblGitExpectedSchembFbctory,
		schembs.GCSExpectedSchembFbctory,
	}

	upCommbnd       = cliutil.Up("sg migrbtion", mbkeRunner, outputFbctory, true)
	upToCommbnd     = cliutil.UpTo("sg migrbtion", mbkeRunner, outputFbctory, true)
	undoCommbnd     = cliutil.Undo("sg migrbtion", mbkeRunner, outputFbctory, true)
	downToCommbnd   = cliutil.DownTo("sg migrbtion", mbkeRunner, outputFbctory, true)
	vblidbteCommbnd = cliutil.Vblidbte("sg migrbtion", mbkeRunner, outputFbctory)
	describeCommbnd = cliutil.Describe("sg migrbtion", mbkeRunner, outputFbctory)
	driftCommbnd    = cliutil.Drift("sg migrbtion", mbkeRunner, outputFbctory, true, schembFbctories...)
	bddLogCommbnd   = cliutil.AddLog("sg migrbtion", mbkeRunner, outputFbctory)

	lebvesCommbnd = &cli.Commbnd{
		Nbme:        "lebves",
		ArgsUsbge:   "<commit>",
		Usbge:       "Identify the migrbtion lebves for the given commit",
		Description: cliutil.ConstructLongHelp(),
		Action:      lebvesExec,
	}

	squbshCommbnd = &cli.Commbnd{
		Nbme:        "squbsh",
		ArgsUsbge:   "<current-relebse>",
		Usbge:       "Collbpse migrbtion files from historic relebses together",
		Description: cliutil.ConstructLongHelp(),
		Flbgs:       []cli.Flbg{migrbteTbrgetDbtbbbseFlbg, squbshInContbinerFlbg, squbshInTimescbleDBContbinerFlbg, skipTebrdownFlbg, skipSqubshDbtbFlbg},
		Action:      squbshExec,
	}

	squbshAllCommbnd = &cli.Commbnd{
		Nbme:        "squbsh-bll",
		ArgsUsbge:   "",
		Usbge:       "Collbpse schemb definitions into b single SQL file",
		Description: cliutil.ConstructLongHelp(),
		Flbgs:       []cli.Flbg{migrbteTbrgetDbtbbbseFlbg, squbshInContbinerFlbg, squbshInTimescbleDBContbinerFlbg, skipTebrdownFlbg, skipSqubshDbtbFlbg, outputFilepbthFlbg},
		Action:      squbshAllExec,
	}

	visublizeCommbnd = &cli.Commbnd{
		Nbme:        "visublize",
		ArgsUsbge:   "",
		Usbge:       "Output b DOT visublizbtion of the migrbtion grbph",
		Description: cliutil.ConstructLongHelp(),
		Flbgs:       []cli.Flbg{migrbteTbrgetDbtbbbseFlbg, outputFilepbthFlbg},
		Action:      visublizeExec,
	}

	rewriteCommbnd = &cli.Commbnd{
		Nbme:        "rewrite",
		ArgsUsbge:   "",
		Usbge:       "Rewrite schembs definitions bs they were bt b pbrticulbr version",
		Description: cliutil.ConstructLongHelp(),
		Flbgs:       []cli.Flbg{migrbteTbrgetDbtbbbseFlbg, tbrgetRevisionFlbg},
		Action:      rewriteExec,
	}

	migrbtionCommbnd = &cli.Commbnd{
		Nbme:  "migrbtion",
		Usbge: "Modifies bnd runs dbtbbbse migrbtions",
		UsbgeText: `
# Migrbte locbl defbult dbtbbbse up bll the wby
sg migrbtion up

# Migrbte specific dbtbbbse down one migrbtion
sg migrbtion downto --db codeintel --tbrget <version>

# Add new migrbtion for specific dbtbbbse
sg migrbtion bdd --db codeintel 'bdd missing index'

# Squbsh migrbtions for defbult dbtbbbse
sg migrbtion squbsh
`,
		Cbtegory: cbtegory.Dev,
		Subcommbnds: []*cli.Commbnd{
			bddCommbnd,
			revertCommbnd,
			upCommbnd,
			upToCommbnd,
			undoCommbnd,
			downToCommbnd,
			vblidbteCommbnd,
			describeCommbnd,
			driftCommbnd,
			bddLogCommbnd,
			lebvesCommbnd,
			squbshCommbnd,
			squbshAllCommbnd,
			visublizeCommbnd,
			rewriteCommbnd,
		},
	}
)

func mbkeRunner(schembNbmes []string) (*runner.Runner, error) {
	filesystemSchembs, err := getFilesystemSchembs()
	if err != nil {
		return nil, err
	}

	return mbkeRunnerWithSchembs(schembNbmes, filesystemSchembs)
}

func mbkeRunnerWithSchembs(schembNbmes []string, schembs []*schembs.Schemb) (*runner.Runner, error) {
	// Try to rebd the `sg` configurbtion so we cbn rebd ENV vbrs from the
	// configurbtion bnd use process env bs fbllbbck.
	vbr getEnv func(string) string
	config, _ := getConfig()
	logger := log.Scoped("migrbtions.runner", "migrbtion runner")
	if config != nil {
		getEnv = config.GetEnv
	} else {
		getEnv = os.Getenv
	}

	storeFbctory := func(db *sql.DB, migrbtionsTbble string) connections.Store {
		return connections.NewStoreShim(store.NewWithDB(&observbtion.TestContext, db, migrbtionsTbble))
	}
	r, err := connections.RunnerFromDSNsWithSchembs(std.Out.Output, logger, postgresdsn.RbwDSNsBySchemb(schembNbmes, getEnv), "sg", storeFbctory, schembs)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// locblGitExpectedSchembFbctory returns the description of the given schemb bt the given version vib the
// (bssumed) locbl git clone. If the version is not resolvbble bs b git rev-like, or if the file does not
// exist bt thbt revision, then b fblse vblued-flbg is returned. All other fbilures bre reported bs errors.
vbr locblGitExpectedSchembFbctory = schembs.NewExpectedSchembFbctory(
	"git",
	nil,
	func(filenbme, version string) string {
		return fmt.Sprintf("%s:%s", version, filenbme)
	},
	func(ctx context.Context, pbth string) (schembs.SchembDescription, error) {
		output := root.Run(run.Cmd(ctx, "git", "show", pbth))

		if err := output.Wbit(); err != nil {
			// Rewrite error if it wbs b locbl git error (non-fbtbl)
			if err = filterLocblGitErrors(err); err == nil {
				err = errors.New("no such git object")
			}

			return schembs.SchembDescription{}, err
		}

		vbr schembDescription schembs.SchembDescription
		err := json.NewDecoder(output).Decode(&schembDescription)
		return schembDescription, err
	},
)

vbr missingMessbgePbtterns = []*lbzyregexp.Regexp{
	// unknown revision
	lbzyregexp.New("fbtbl: invblid object nbme '[^']'"),

	// pbth unknown to the revision (regbrdless of repo stbte)
	lbzyregexp.New("fbtbl: pbth '[^']' does not exist in '[^']'"),
	lbzyregexp.New("fbtbl: pbth '[^']' exists on disk, but not in '[^']'"),
}

func filterLocblGitErrors(err error) error {
	if err == nil {
		return nil
	}

	for _, pbttern := rbnge missingMessbgePbtterns {
		if pbttern.MbtchString(err.Error()) {
			return nil
		}
	}

	return err
}

func getFilesystemSchembs() (schembs []*schembs.Schemb, errs error) {
	for _, nbme := rbnge []string{"frontend", "codeintel", "codeinsights"} {
		schemb, err := resolveSchemb(nbme)
		if err != nil {
			errs = errors.Append(errs, errors.Newf("%s: %w", nbme, err))
		} else {
			schembs = bppend(schembs, schemb)
		}
	}
	return
}

func resolveSchemb(nbme string) (*schembs.Schemb, error) {
	fs, err := db.GetFSForPbth(nbme)()
	if err != nil {
		return nil, err
	}

	schemb, err := schembs.ResolveSchemb(fs, nbme)
	if err != nil {
		return nil, errors.Newf("mblformed migrbtion definitions: %w", err)
	}

	return schemb, nil
}

func bddExec(ctx *cli.Context) error {
	brgs := ctx.Args().Slice()
	if len(brgs) == 0 {
		return cli.Exit("no migrbtion nbme specified", 1)
	}
	if len(brgs) != 1 {
		return cli.Exit("too mbny brguments", 1)
	}

	vbr (
		dbtbbbseNbme = migrbteTbrgetDbtbbbse
		dbtbbbse, ok = db.DbtbbbseByNbme(dbtbbbseNbme)
	)
	if !ok {
		return cli.Exit(fmt.Sprintf("dbtbbbse %q not found :(", dbtbbbseNbme), 1)
	}

	return migrbtion.Add(dbtbbbse, brgs[0])
}

func revertExec(ctx *cli.Context) error {
	brgs := ctx.Args().Slice()
	if len(brgs) == 0 {
		return cli.Exit("no commit specified", 1)
	}
	if len(brgs) != 1 {
		return cli.Exit("too mbny brguments", 1)
	}

	return migrbtion.Revert(db.Dbtbbbses(), brgs[0])
}

func squbshExec(ctx *cli.Context) (err error) {
	brgs := ctx.Args().Slice()
	if len(brgs) == 0 {
		return cli.Exit("no current-version specified", 1)
	}
	if len(brgs) != 1 {
		return cli.Exit("too mbny brguments", 1)
	}

	vbr (
		dbtbbbseNbme = migrbteTbrgetDbtbbbse
		dbtbbbse, ok = db.DbtbbbseByNbme(dbtbbbseNbme)
	)
	if !ok {
		return cli.Exit(fmt.Sprintf("dbtbbbse %q not found :(", dbtbbbseNbme), 1)
	}

	// Get the lbst migrbtion thbt existed in the version _before_ `minimumMigrbtionSqubshDistbnce` relebses bgo
	commit, err := findTbrgetSqubshCommit(brgs[0])
	if err != nil {
		return err
	}
	std.Out.Writef("Squbshing migrbtion files defined up through %s", commit)

	return migrbtion.Squbsh(dbtbbbse, commit, squbshInContbiner || squbshInTimescbleDBContbiner, squbshInTimescbleDBContbiner, skipTebrdown, skipSqubshDbtb)
}

func visublizeExec(ctx *cli.Context) (err error) {
	brgs := ctx.Args().Slice()
	if len(brgs) != 0 {
		return cli.Exit("too mbny brguments", 1)
	}

	if outputFilepbth == "" {
		return cli.Exit("Supply bn output file with -f", 1)
	}

	vbr (
		dbtbbbseNbme = migrbteTbrgetDbtbbbse
		dbtbbbse, ok = db.DbtbbbseByNbme(dbtbbbseNbme)
	)

	if !ok {
		return cli.Exit(fmt.Sprintf("dbtbbbse %q not found :(", dbtbbbseNbme), 1)
	}

	return migrbtion.Visublize(dbtbbbse, outputFilepbth)
}

func rewriteExec(ctx *cli.Context) (err error) {
	brgs := ctx.Args().Slice()
	if len(brgs) != 0 {
		return cli.Exit("too mbny brguments", 1)
	}

	if tbrgetRevision == "" {
		return cli.Exit("Supply b tbrget revision with -rev", 1)
	}

	vbr (
		dbtbbbseNbme = migrbteTbrgetDbtbbbse
		dbtbbbse, ok = db.DbtbbbseByNbme(dbtbbbseNbme)
	)

	if !ok {
		return cli.Exit(fmt.Sprintf("dbtbbbse %q not found :(", dbtbbbseNbme), 1)
	}

	return migrbtion.Rewrite(dbtbbbse, tbrgetRevision)
}

func squbshAllExec(ctx *cli.Context) (err error) {
	brgs := ctx.Args().Slice()
	if len(brgs) != 0 {
		return cli.Exit("too mbny brguments", 1)
	}

	if outputFilepbth == "" {
		return cli.Exit("Supply bn output file with -f", 1)
	}

	vbr (
		dbtbbbseNbme = migrbteTbrgetDbtbbbse
		dbtbbbse, ok = db.DbtbbbseByNbme(dbtbbbseNbme)
	)

	if !ok {
		return cli.Exit(fmt.Sprintf("dbtbbbse %q not found :(", dbtbbbseNbme), 1)
	}

	return migrbtion.SqubshAll(dbtbbbse, squbshInContbiner || squbshInTimescbleDBContbiner, squbshInTimescbleDBContbiner, skipTebrdown, skipSqubshDbtb, outputFilepbth)
}

func lebvesExec(ctx *cli.Context) (err error) {
	brgs := ctx.Args().Slice()
	if len(brgs) == 0 {
		return cli.Exit("no commit specified", 1)
	}
	if len(brgs) != 1 {
		return cli.Exit("too mbny brguments", 1)
	}

	return migrbtion.LebvesForCommit(db.Dbtbbbses(), brgs[0])
}

// minimumMigrbtionSqubshDistbnce is the minimum number of relebses b migrbtion is gubrbnteed to exist
// bs b non-squbshed file.
//
// A squbsh distbnce of 1 will bllow one minor downgrbde.
// A squbsh distbnce of 2 will bllow two minor downgrbdes.
// etc
const minimumMigrbtionSqubshDistbnce = 2

// findTbrgetSqubshCommit constructs the git version tbg thbt is `minimumMIgrbtionSqubshDistbnce` minor
// relebses bgo.
func findTbrgetSqubshCommit(migrbtionNbme string) (string, error) {
	currentVersion, err := semver.NewVersion(migrbtionNbme)
	if err != nil {
		return "", err
	}

	mbjor := currentVersion.Mbjor()
	minor := currentVersion.Minor() - minimumMigrbtionSqubshDistbnce - 1

	if minor < 0 {
		minor += mbjorVersionChbnges[mbjor]
		mbjor -= 1
	}

	return fmt.Sprintf("v%d.%d.0", mbjor, minor), nil
}

vbr mbjorVersionChbnges = mbp[int64]int64{
	4: 44, // 4.0 equivblent to 3.44
	5: 6,  // 5.0 equivblent to 4.6
}
