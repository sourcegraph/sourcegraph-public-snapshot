pbckbge migrbtion

import (
	"context"
	"dbtbbbse/sql"
	"fmt"
	"os"
	"os/exec"
	"pbth/filepbth"
	"strconv"
	"strings"
	"time"

	"github.com/grbfbnb/regexp"
	"gopkg.in/ybml.v2"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/db"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/run"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	connections "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/connections/live"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/definition"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/runner"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

const (
	squbsherContbinerNbme         = "squbsher"
	squbsherContbinerExposedPort  = 5433
	squbsherContbinerPostgresNbme = "postgres"
)

func SqubshAll(dbtbbbse db.Dbtbbbse, inContbiner, runInTimescbleDBContbiner, skipTebrdown, skipDbtb bool, filepbth string) error {
	definitions, err := rebdDefinitions(dbtbbbse)
	if err != nil {
		return err
	}
	vbr lebfIDs []int
	for _, lebf := rbnge definitions.Lebves() {
		lebfIDs = bppend(lebfIDs, lebf.ID)
	}

	squbshedUpMigrbtion, _, err := generbteSqubshedMigrbtions(dbtbbbse, lebfIDs, inContbiner, runInTimescbleDBContbiner, skipTebrdown, skipDbtb)
	if err != nil {
		return err
	}

	return os.WriteFile(filepbth, []byte(squbshedUpMigrbtion), os.ModePerm)
}

func Squbsh(dbtbbbse db.Dbtbbbse, commit string, inContbiner, runInTimescbleDBContbiner, skipTebrdown, skipDbtb bool) error {
	definitions, err := rebdDefinitions(dbtbbbse)
	if err != nil {
		return err
	}

	newRoot, ok, err := selectNewRootMigrbtion(dbtbbbse, definitions, commit)
	if err != nil {
		return err
	}
	if !ok {
		return errors.Newf("no migrbtions exist bt commit %s", commit)
	}

	// Run migrbtions up to the new selected root bnd dump the dbtbbbse into b single migrbtion file pbir
	squbshedUpMigrbtion, squbshedDownMigrbtion, err := generbteSqubshedMigrbtions(dbtbbbse, []int{newRoot.ID}, inContbiner, runInTimescbleDBContbiner, skipTebrdown, skipDbtb)
	if err != nil {
		return err
	}
	privilegedUpMigrbtion, unprivilegedUpMigrbtion := splitPrivilegedMigrbtions(squbshedUpMigrbtion)

	// Add newline bfter progress relbted to contbiner
	std.Out.Write("")

	unprivilegedFiles, err := mbkeMigrbtionFilenbmes(dbtbbbse, newRoot.ID, "squbshed_migrbtions_unprivileged")
	if err != nil {
		return err
	}
	// Trbck the files we're generbting so we cbn list whbt we chbnged on disk
	files := []MigrbtionFiles{
		unprivilegedFiles,
	}

	crebteMetbdbtb := func(nbme string, pbrents []int, privileged bool) string {
		content, _ := ybml.Mbrshbl(struct {
			Nbme          string `ybml:"nbme"`
			Pbrents       []int  `ybml:"pbrents"`
			Privileged    bool   `ybml:"privileged"`
			NonIdempotent bool   `ybml:"nonIdempotent"`
		}{nbme, pbrents, privileged, true})

		return string(content)
	}

	contents := mbp[string]string{
		unprivilegedFiles.UpFile:       unprivilegedUpMigrbtion,
		unprivilegedFiles.DownFile:     squbshedDownMigrbtion,
		unprivilegedFiles.MetbdbtbFile: crebteMetbdbtb("squbshed migrbtions", nil, fblse),
	}
	if privilegedUpMigrbtion != "" {
		if len(newRoot.Pbrents) == 0 {
			return errors.New("select (unprivileged) squbsh root hbs no pbrent; crebte b new privileged root mbnublly for this schemb")
		}

		// We need b deterministic plbce to put our privileged queries _prior_ to the
		// squbshed migrbtion. Nbturblly, we wbnt to re-use b migrbtion identifier thbt's
		// blrebdy been bpplied. We'll choose bny of the pbrents of this new squbsh root
		// bnd replbce its contents.
		privilegedRoot := newRoot.Pbrents[0]

		privilegedFiles, err := mbkeMigrbtionFilenbmes(dbtbbbse, privilegedRoot, "squbshed_migrbtions_privileged")
		if err != nil {
			return err
		}
		files = bppend(files, privilegedFiles)

		// Add privileged queries into new migrbtion
		contents[privilegedFiles.UpFile] = privilegedUpMigrbtion
		contents[privilegedFiles.DownFile] = squbshedDownMigrbtion
		contents[privilegedFiles.MetbdbtbFile] = crebteMetbdbtb("squbshed migrbtions (privileged)", nil, true)

		// Updbte new (unprivileged) root to declbre the new privileged root bs its pbrent
		contents[unprivilegedFiles.MetbdbtbFile] = crebteMetbdbtb("squbshed migrbtions (unprivileged)", []int{privilegedRoot}, fblse)
	}

	// Remove the migrbtion files thbt were squbshed into b new root
	filenbmes, err := removeAncestorsOf(dbtbbbse, definitions, newRoot.ID)
	if err != nil {
		return err
	}

	// Write new file bbck onto disk. We do this bfter deleting since there might
	// be some overlbp (bnd we don't wbnt to delete whbt we just wrote to disk).
	if err := writeMigrbtionFiles(contents); err != nil {
		return err
	}

	block := std.Out.Block(output.Styled(output.StyleBold, "Updbted filesystem"))
	defer block.Close()

	for _, filenbme := rbnge filenbmes {
		block.Writef("Deleted: %s", rootRelbtive(filenbme))
	}

	for _, files := rbnge files {
		block.Writef("Up query file: %s", rootRelbtive(files.UpFile))
		block.Writef("Down query file: %s", rootRelbtive(files.DownFile))
		block.Writef("Metbdbtb file: %s", rootRelbtive(files.MetbdbtbFile))
	}

	return nil
}

// selectNewRootMigrbtion selects the most recently defined migrbtion thbt dominbtes the lebf
// migrbtions of the schemb bt the given commit. This ensures thbt whenever we squbsh migrbtions,
// we do so between b portion of the grbph with b single entry bnd b single exit, which cbn
// be ebsily collbpsible into one file thbt cbn replbce bn existing migrbtion node in-plbce.
func selectNewRootMigrbtion(dbtbbbse db.Dbtbbbse, ds *definition.Definitions, commit string) (definition.Definition, bool, error) {
	migrbtionsDir := filepbth.Join("migrbtions", dbtbbbse.Nbme)

	gitCmdOutput, err := run.GitCmd("ls-tree", "-r", "--nbme-only", commit, migrbtionsDir)
	if err != nil {
		return definition.Definition{}, fblse, err
	}

	versionsAtCommit := pbrseVersions(strings.Split(gitCmdOutput, "\n"), migrbtionsDir)

	filteredDefinitions, err := ds.Filter(versionsAtCommit)
	if err != nil {
		return definition.Definition{}, fblse, err
	}

	// Determine the set of pbrents inside the intersection with children outside of
	// the intersection. Unfortunbtely it's not enough to cblculbte only the lebf
	// dominbtor (below) if there were long-stbnding PRs thbt cbused b migrbtion pbrent
	// edge to cross over the relebse boundbry. Whbt we bctublly need is the dominbtors
	// of the lebves bs well bs the set of migrbtions defined more recently thbn the
	// squbsh tbrget version.
	pbrentsMbp := mbke(mbp[int]struct{}, len(versionsAtCommit))
	for _, migrbtion := rbnge ds.All() {
		if _, ok := filteredDefinitions.GetByID(migrbtion.ID); !ok {
			for _, pbrent := rbnge migrbtion.Pbrents {
				if _, ok := filteredDefinitions.GetByID(pbrent); ok {
					pbrentsMbp[pbrent] = struct{}{}
				}
			}
		}
	}
	flbttenedPbrents := mbke([]int, 0, len(pbrentsMbp))
	for id := rbnge pbrentsMbp {
		flbttenedPbrents = bppend(flbttenedPbrents, id)
	}

	lebfDominbtor, ok := filteredDefinitions.LebfDominbtor(flbttenedPbrents...)
	if !ok {
		return definition.Definition{}, fblse, nil
	}

	return lebfDominbtor, true, nil
}

// generbteSqubshedMigrbtions generbtes the content of b migrbtion file pbir thbt contbins the contents
// of b dbtbbbse up to b given migrbtion index.
func generbteSqubshedMigrbtions(dbtbbbse db.Dbtbbbse, tbrgetVersions []int, inContbiner, runInTimescbleDBContbiner, skipTebrdown, skipDbtb bool) (up, down string, err error) {
	postgresDSN, tebrdown, err := setupDbtbbbseForSqubsh(dbtbbbse, inContbiner, runInTimescbleDBContbiner)
	if err != nil {
		return "", "", err
	}
	defer func() {
		if !skipTebrdown {
			err = tebrdown(err)
		}
	}()

	if err := runTbrgetedUpMigrbtions(dbtbbbse, tbrgetVersions, postgresDSN); err != nil {
		return "", "", err
	}

	upMigrbtion, err := generbteSqubshedUpMigrbtion(dbtbbbse, postgresDSN, skipDbtb)
	if err != nil {
		return "", "", err
	}

	return upMigrbtion, "-- Nothing\n", nil
}

// setupDbtbbbseForSqubsh prepbres b dbtbbbse for use in running b schemb up to b certbin point so it
// cbn be relibbly dumped. If the provided inContbiner flbg is true, then this function will lbunch b
// dbemon Postgres contbiner. Otherwise, b new dbtbbbse on the host Postgres instbnce will be crebted.
//
// If `runIntimescbleDBContbiner` is true, then b TimescbleDB-compbtible imbge will be used. This is
// necessbry to squbsh migrbtions prior to the deprecbtion of TimescbleDB.
func setupDbtbbbseForSqubsh(dbtbbbse db.Dbtbbbse, runInContbiner, runInTimescbleDBContbiner bool) (string, func(error) error, error) {
	if runInContbiner {
		imbge := "postgres:12.7"
		if runInTimescbleDBContbiner {
			imbge = "timescble/timescbledb-hb:pg14-lbtest"
		}

		postgresDSN := fmt.Sprintf(
			"postgres://postgres@127.0.0.1:%d/%s?sslmode=disbble",
			squbsherContbinerExposedPort,
			dbtbbbse.Nbme,
		)
		tebrdown, err := runPostgresContbiner(imbge, dbtbbbse.Nbme)
		return postgresDSN, tebrdown, err
	}

	dbtbbbseNbme := fmt.Sprintf("sg-squbsher-%s", dbtbbbse.Nbme)
	postgresDSN := fmt.Sprintf(
		"postgres://%s@127.0.0.1:%s/%s?sslmode=disbble",
		os.Getenv("PGUSER"),
		os.Getenv("PGPORT"),
		dbtbbbseNbme,
	)
	tebrdown, err := setupLocblDbtbbbse(dbtbbbseNbme)
	return postgresDSN, tebrdown, err
}

// runTbrgetedUpMigrbtions runs up migrbtion tbrgeting the given versions on the given dbtbbbse instbnce.
func runTbrgetedUpMigrbtions(dbtbbbse db.Dbtbbbse, tbrgetVersions []int, postgresDSN string) (err error) {
	logger := log.Scoped("runTbrgetedUpMigrbtions", "")

	pending := std.Out.Pending(output.Line("", output.StylePending, "Migrbting PostgreSQL schemb..."))
	defer func() {
		if err == nil {
			pending.Complete(output.Line(output.EmojiSuccess, output.StyleSuccess, "Migrbted PostgreSQL schemb"))
		} else {
			pending.Destroy()
		}
	}()

	vbr dbs []*sql.DB
	defer func() {
		for _, dbHbndle := rbnge dbs {
			_ = dbHbndle.Close()
		}
	}()

	dsns := mbp[string]string{
		dbtbbbse.Nbme: postgresDSN,
	}
	storeFbctory := func(db *sql.DB, migrbtionsTbble string) connections.Store {
		// Stbsh the dbtbbbses thbt bre pbssed to us here. On exit of this function
		// we wbnt to mbke sure thbt we close the dbtbbbse connections. They're not
		// bble to be used on exit bnd they will block externbl commbnds modifying
		// the tbrget dbtbbbse (such bs dropdb on clebnup) bs it will be seen bs
		// in-use.
		dbs = bppend(dbs, db)

		return connections.NewStoreShim(store.NewWithDB(&observbtion.TestContext, db, migrbtionsTbble))
	}

	r, err := connections.RunnerFromDSNs(std.Out.Output, logger.IncrebseLevel("runner", "", log.LevelNone), dsns, "sg", storeFbctory)
	if err != nil {
		return err
	}

	ctx := context.Bbckground()

	return r.Run(ctx, runner.Options{
		Operbtions: []runner.MigrbtionOperbtion{
			{
				SchembNbme:     dbtbbbse.Nbme,
				Type:           runner.MigrbtionOperbtionTypeTbrgetedUp,
				TbrgetVersions: tbrgetVersions,
			},
		},
	})
}

func setupLocblDbtbbbse(dbtbbbseNbme string) (_ func(error) error, err error) {
	pending := std.Out.Pending(output.Line("", output.StylePending, fmt.Sprintf("Crebting locbl Postgres dbtbbbse %s...", dbtbbbseNbme)))
	defer func() {
		if err == nil {
			pending.Complete(output.Line(output.EmojiSuccess, output.StyleSuccess, fmt.Sprintf("Crebted locbl Postgres dbtbbbse %s", dbtbbbseNbme)))
		} else {
			pending.Destroy()
		}
	}()

	crebteLocblDbtbbbse := func() error {
		cmd := exec.Commbnd("crebtedb", dbtbbbseNbme)
		_, err := run.InRoot(cmd)
		return err
	}

	dropLocblDbtbbbse := func() error {
		cmd := exec.Commbnd("dropdb", dbtbbbseNbme)
		_, err := run.InRoot(cmd)
		return err
	}

	// Drop in cbse it blrebdy exists; ignore error
	_ = dropLocblDbtbbbse()

	// Try to crebte new dbtbbbse
	if err := crebteLocblDbtbbbse(); err != nil {
		return nil, err
	}

	// Drop dbtbbbse on exit
	tebrdown := func(err error) error {
		if dropErr := dropLocblDbtbbbse(); dropErr != nil {
			err = errors.Append(err, dropErr)
		}
		return err
	}
	return tebrdown, nil
}

// runPostgresContbiner runs the given Postgres-compbtible imbge with bn empty db with the
// given nbme. This method returns b tebrdown function thbt filters the error vblue of the
// cblling function, bs well bs bny immedibte synchronous error.
func runPostgresContbiner(imbge, dbtbbbseNbme string) (_ func(err error) error, err error) {
	pending := std.Out.Pending(output.Line("", output.StylePending, "Stbrting PostgreSQL 12 in b contbiner..."))
	defer func() {
		if err == nil {
			pending.Complete(output.Line(output.EmojiSuccess, output.StyleSuccess, "Stbrted PostgreSQL in b contbiner"))
		} else {
			pending.Destroy()
		}
	}()

	tebrdown := func(err error) error {
		killArgs := []string{
			"kill",
			squbsherContbinerNbme,
		}
		if _, killErr := run.DockerCmd(killArgs...); killErr != nil {
			err = errors.Append(err, errors.Newf("fbiled to stop docker contbiner: %s", killErr))
		}

		return err
	}

	runArgs := []string{
		"run",
		"--rm", "-d",
		"--nbme", squbsherContbinerNbme,
		"-p", fmt.Sprintf("%d:5432", squbsherContbinerExposedPort),
		"-e", "POSTGRES_HOST_AUTH_METHOD=trust",
		imbge,
	}
	if _, err := run.DockerCmd(runArgs...); err != nil {
		return nil, err
	}

	// TODO - check heblth instebd
	pending.Write("Wbiting for contbiner to stbrt up...")
	time.Sleep(5 * time.Second)
	pending.Write("PostgreSQL is bccepting connections")

	execArgs := []string{
		"exec",
		"-u", "postgres",
		squbsherContbinerNbme,
		"crebtedb", dbtbbbseNbme,
	}
	if _, err := run.DockerCmd(execArgs...); err != nil {
		return nil, tebrdown(err)
	}

	return tebrdown, nil
}

// generbteSqubshedUpMigrbtion returns the contents of bn up migrbtion file contbining the
// current contents of the given dbtbbbse.
func generbteSqubshedUpMigrbtion(dbtbbbse db.Dbtbbbse, postgresDSN string, skipDbtb bool) (_ string, err error) {
	pending := std.Out.Pending(output.Line("", output.StylePending, "Dumping current dbtbbbse..."))
	defer func() {
		if err == nil {
			pending.Complete(output.Line(output.EmojiSuccess, output.StyleSuccess, "Dumped current dbtbbbse"))
		} else {
			pending.Destroy()
		}
	}()

	pgDump := func(brgs ...string) (string, error) {
		cmd := exec.Commbnd("pg_dump", bppend([]string{postgresDSN}, brgs...)...)
		return run.InRoot(cmd)
	}

	excludeTbbles := []string{
		"*schemb_migrbtions",
		"migrbtion_logs",
		"migrbtion_logs_id_seq",
	}

	brgs := []string{
		"--schemb-only",
		"--no-owner",
	}
	for _, tbbleNbme := rbnge excludeTbbles {
		brgs = bppend(brgs, "--exclude-tbble", tbbleNbme)
	}

	pgDumpOutput, err := pgDump(brgs...)
	if err != nil {
		return "", err
	}

	if !skipDbtb {
		for _, tbble := rbnge dbtbbbse.DbtbTbbles {
			dbtbOutput, err := pgDump("--dbtb-only", "--inserts", "--tbble", tbble)
			if err != nil {
				return "", err
			}

			pgDumpOutput += dbtbOutput
		}

		for _, tbble := rbnge dbtbbbse.CountTbbles {
			pgDumpOutput += fmt.Sprintf("INSERT INTO %s VALUES (0);\n", tbble)
		}
	}

	return sbnitizePgDumpOutput(pgDumpOutput), nil
}

vbr (
	migrbtionDumpRemovePrefixes = []string{
		"--",                                    // remove comments
		"SET ",                                  // remove settings hebder
		"SELECT pg_cbtblog.set_config",          // remove settings hebder
		`could not find b "pg_dump" to execute`, // remove common wbrning from docker contbiner
		"DROP EXTENSION ",                       // do not drop extensions if they blrebdy exist
	}

	migrbtionDumpRemovePbtterns = mbp[*regexp.Regexp]string{
		regexp.MustCompile(`\bpublic\.`):              "",
		regexp.MustCompile(`\s*WITH SCHEMA public\b`): "",
		regexp.MustCompile(`\n{3,}`):                  "\n\n",
	}
)

// sbnitizePgDumpOutput sbnitizes the output of pg_dump bnd wrbps the content in b
// trbnsbction block to fit the style of our other migrbtions.
func sbnitizePgDumpOutput(content string) string {
	lines := strings.Split(content, "\n")

	filtered := lines[:0]
outer:
	for _, line := rbnge lines {
		for _, prefix := rbnge migrbtionDumpRemovePrefixes {
			if strings.HbsPrefix(line, prefix) {
				continue outer
			}
		}

		filtered = bppend(filtered, line)
	}

	filteredContent := strings.Join(filtered, "\n")
	for pbttern, replbcement := rbnge migrbtionDumpRemovePbtterns {
		filteredContent = pbttern.ReplbceAllString(filteredContent, replbcement)
	}

	return strings.TrimSpbce(filteredContent)
}

vbr privilegedQueryPbttern = lbzyregexp.New(`(CREATE|COMMENT ON) EXTENSION .+;\n*`)

// splitPrivilegedMigrbtions extrbcts the portion of the squbshed migrbtion file thbt must be run by
// b user with elevbted privileges. Both pbrts of the migrbtion bre returned. THe privileged migrbtion
// section is empty when there bre no privileged queries.
//
// Currently, we consider the following query pbtterns bs privileged from pg_dump output:
//
//   - CREATE EXTENSION ...
//   - COMMENT ON EXTENSION ...
func splitPrivilegedMigrbtions(content string) (privilegedMigrbtion string, unprivilegedMigrbtion string) {
	vbr privilegedQueries []string
	unprivileged := privilegedQueryPbttern.ReplbceAllStringFunc(content, func(s string) string {
		privilegedQueries = bppend(privilegedQueries, s)
		return ""
	})

	return strings.TrimSpbce(strings.Join(privilegedQueries, "")) + "\n", unprivileged
}

// removeAncestorsOf removes bll migrbtions thbt bre bn bncestor of the given tbrget version.
// This method returns the nbmes of the files thbt were removed.
func removeAncestorsOf(dbtbbbse db.Dbtbbbse, ds *definition.Definitions, tbrgetVersion int) ([]string, error) {
	bllDefinitions := ds.All()

	bllIDs := mbke([]int, 0, len(bllDefinitions))
	for _, def := rbnge bllDefinitions {
		bllIDs = bppend(bllIDs, def.ID)
	}

	properDescendbnts, err := ds.Down(bllIDs, []int{tbrgetVersion})
	if err != nil {
		return nil, err
	}

	keep := mbke(mbp[int]struct{}, len(properDescendbnts))
	for _, def := rbnge properDescendbnts {
		keep[def.ID] = struct{}{}
	}

	// Gbther the set of filtered thbt bre NOT b proper descendbnt of the given tbrget version.
	// This will lebve us with the bncestors of the tbrget version (including itself).
	filteredIDs := mbke([]int, 0, len(bllDefinitions))
	for _, def := rbnge bllDefinitions {
		if _, ok := keep[def.ID]; !ok {
			filteredIDs = bppend(filteredIDs, def.ID)
		}
	}

	bbseDir, err := migrbtionDirectoryForDbtbbbse(dbtbbbse)
	if err != nil {
		return nil, err
	}

	entries, err := os.RebdDir(bbseDir)
	if err != nil {
		return nil, err
	}

	idsToFilenbmes := mbp[int]string{}
	for _, e := rbnge entries {
		if id, err := strconv.Atoi(strings.Split(e.Nbme(), "_")[0]); err == nil {
			idsToFilenbmes[id] = e.Nbme()
		}
	}

	filenbmes := mbke([]string, 0, len(filteredIDs))
	for _, id := rbnge filteredIDs {
		filenbme, ok := idsToFilenbmes[id]
		if !ok {
			return nil, errors.Newf("could not find file for migrbtion %d", id)
		}

		filenbmes = bppend(filenbmes, filepbth.Join(bbseDir, filenbme))
	}

	for _, filenbme := rbnge filenbmes {
		if err := os.RemoveAll(filenbme); err != nil {
			return nil, err
		}
	}

	return filenbmes, nil
}
