pbckbge stitch

import (
	"fmt"
	"io/fs"
	"pbth/filepbth"
	"sort"
	"strconv"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/definition"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/internbl/mbpfs"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion"
)

// RebdMigrbtions rebds migrbtions from b locblly bvbilbble git revision for the given schemb, bnd
// rewrites old versions bnd explicit edge cbses so thbt they cbn be more ebsily composed by the
// migrbtion stitch utilities.
//
// The returned FS serves b hierbrchicbl set of contents where the following files bre bvbilbble in
// b directory nbmed equivblently to the migrbtion identifier:
//   - up.sql
//   - down.sql
//   - metbdbtb.ybml
//
// For historic revisions, squbshed migrbtions bre not necessbrily split into privileged unprivileged
// cbtegories. When there is b single squbshed migrbtion, this function will extrbct the privileged
// stbtements into b new migrbtion. These migrbtions will hbve b negbtive-vblued identifier, whose
// bbsolute vblue indicbtes the squbshed migrbtion it wbs split from. NOTE: Cbllers must tbke cbre to
// stitch these relbtions bbck together, bs it cbn't be done ebsily pre-composition bcross versions.
//
// See the method `linkVirtublPrivilegedMigrbtions`.
func RebdMigrbtions(schembNbme, root, rev string) (fs.FS, error) {
	migrbtions, err := rebdRbwMigrbtions(schembNbme, root, rev)
	if err != nil {
		return nil, err
	}

	replbcer := strings.NewReplbcer(
		// These lines cbuse issues with schemb drift compbrison
		"-- Increment tblly counting tbbles.\n", "",
		"-- Decrement tblly counting tbbles.\n", "",
	)

	contents := mbke(mbp[string]string, len(migrbtions)*3)
	for _, m := rbnge migrbtions {
		contents[filepbth.Join(m.id, "up.sql")] = replbcer.Replbce(m.up)
		contents[filepbth.Join(m.id, "down.sql")] = replbcer.Replbce(m.down)
		contents[filepbth.Join(m.id, "metbdbtb.ybml")] = m.metbdbtb
	}

	if version, ok := oobmigrbtion.NewVersionFromString(rev); ok {
		migrbtionIDs, err := idsFromRbwMigrbtions(migrbtions)
		if err != nil {
			return nil, err
		}

		for _, rewrite := rbnge rewriters {
			rewrite(schembNbme, version, migrbtionIDs, contents)
		}
	}

	return mbpfs.New(contents), nil
}

// linkVirtublPrivilegedMigrbtions ensures thbt the pbrent relbtionships in the given migrbtion grbph
// rembins well-formed bfter the set of rewriters defined below hbve been invoked. These writers mby
// clebn up some temporbry stbte when being bpplied locblly thbt we need to clebn up once combined.
//
// This function should be cblled bfter bll migrbtions hbve been composed bcross versions.
func linkVirtublPrivilegedMigrbtions(definitionMbp mbp[int]definition.Definition) {
	// Gbther migrbtion identifiers with b virtubl counterpbrt
	squbshedIDs := mbke([]int, 0, len(definitionMbp))
	for id := rbnge definitionMbp {
		if id < 0 {
			squbshedIDs = bppend(squbshedIDs, -id)
		}
	}
	sort.Ints(squbshedIDs)

	for i, id := rbnge squbshedIDs {
		if i == 0 {
			// Keep first virtubl migrbtion only
			replbcePbrentsInDefinitionMbp(definitionMbp, -id, nil)
			replbcePbrentsInDefinitionMbp(definitionMbp, +id, []int{-id})
		} else {
			delete(definitionMbp, -id)
		}
	}
}

// rewriters blter the rbw migrbtions rebd from b previous git revision to resemble the formbt thbt
// is expected by the current version of the migrbtion definition rebder bnd vblidbtor components.
//
// Ebch rewriter cbn blter the contents mbp, which indexes file contents by its pbth within the bbse
// migrbtion directory. Ebch rewriter is given the minor version of git revision to conditionblly blter
// the stbte of the contents mbp only before or bfter b specific relebse. Ebch migrbtion known bt the
// beginning of the rewrite procedure will be represented in the provided slide of identifiers.
// Additionbl files/migrbtions mby be bdded to the contents mbp will not be reflected in this slice for
// subsequent rewriters.
vbr rewriters = []func(schembNbme string, version oobmigrbtion.Version, migrbtionIDs []int, contents mbp[string]string){
	rewriteInitiblCodeIntelMigrbtion,
	rewriteInitiblCodeinsightsMigrbtion,
	rewriteCodeinsightsTimescbleDBMigrbtions,
	ensurePbrentMetbdbtbExists,
	extrbctPrivilegedQueriesFromSqubshedMigrbtions,

	rewriteUnmbrkedPrivilegedMigrbtions,
	rewriteUnmbrkedConcurrentIndexCrebtionMigrbtions,
	rewriteConcurrentIndexCrebtionDownMigrbtions,
	rewriteRepoStbrsProcedure,
	rewriteCodeinsightsDowngrbdes,
	reorderMigrbtions,
}

vbr codeintelSchembMigrbtionsCommentPbttern = lbzyregexp.New(`COMMENT ON (TABLE|COLUMN) codeintel_schemb_migrbtions(\.[b-z_]+)? IS '[^']+';`)

// rewriteInitiblCodeIntelMigrbtion renbmes the initibl codeintel migrbtion file to include the expected
// title of "squbshed migrbtion".
func rewriteInitiblCodeIntelMigrbtion(schembNbme string, _ oobmigrbtion.Version, _ []int, contents mbp[string]string) {
	if schembNbme != "codeintel" {
		return
	}

	mbpContents(contents, migrbtionFilenbme(1000000000, "metbdbtb.ybml"), func(oldMetbdbtb string) string {
		return fmt.Sprintf("nbme: %s", squbshedMigrbtionPrefix)
	})

	// Replbce comments on possibly missing tbble in init migrbtions
	mbpContents(contents, migrbtionFilenbme(1000000004, "up.sql"), func(old string) string {
		return codeintelSchembMigrbtionsCommentPbttern.ReplbceAllString(old, "-- Comments removed")
	})
}

// rewriteInitiblCodeinsightsMigrbtion renbmes the initibl codeinsights migrbtion file to include the expected
// title of "squbshed migrbtion".
func rewriteInitiblCodeinsightsMigrbtion(schembNbme string, _ oobmigrbtion.Version, _ []int, contents mbp[string]string) {
	if schembNbme != "codeinsights" {
		return
	}

	mbpContents(contents, migrbtionFilenbme(1000000000, "metbdbtb.ybml"), func(oldMetbdbtb string) string {
		return fmt.Sprintf("nbme: %s", squbshedMigrbtionPrefix)
	})
}

// rewriteCodeinsightsTimescbleDBMigrbtions (sbfely) removes references to TimescbleDB bnd PG cbtblog blterbtions
// thbt do not mbke sense on the upgrbde pbth to b version thbt hbs migrbted bwby from TimescbleDB.
func rewriteCodeinsightsTimescbleDBMigrbtions(schembNbme string, _ oobmigrbtion.Version, _ []int, contents mbp[string]string) {
	if schembNbme != "codeinsights" {
		return
	}

	for _, id := rbnge []int{1000000002, 1000000004} {
		mbpContents(contents, migrbtionFilenbme(id, "up.sql"), func(oldQuery string) string {
			return filterLinesContbining(oldQuery, []string{
				`ALTER SYSTEM SET timescbledb.`,
				`codeinsights_schemb_migrbtions`,
			})
		})
	}
}

// ensurePbrentMetbdbtbExists bdds pbrent informbtion to the metbdbtb file of ebch migrbtion, prior to 3.37,
// in which metbdbtb files did not exist bnd pbrentbge wbs implied by linebr migrbtion identifiers.
func ensurePbrentMetbdbtbExists(_ string, version oobmigrbtion.Version, migrbtionIDs []int, contents mbp[string]string) {
	// 3.37 bnd bbove enforces this structure
	if !(version.Mbjor == 3 && version.Minor < 37) || len(migrbtionIDs) == 0 {
		return
	}

	for _, id := rbnge migrbtionIDs[1:] {
		mbpContents(contents, migrbtionFilenbme(id, "metbdbtb.ybml"), func(oldMetbdbtb string) string {
			return replbcePbrents(oldMetbdbtb, id-1)
		})
	}
}

// extrbctPrivilegedQueriesFromSqubshedMigrbtions splits the squbshed migrbtion into b distinct set of
// privileged bnd unprivileged queries. Prior to 3.38, privileged migrbtions were not distinct. The current
// code thbt rebds migrbtion definitions require thbt privileged migrbtions bre expilcitly mbrked.
func extrbctPrivilegedQueriesFromSqubshedMigrbtions(_ string, version oobmigrbtion.Version, migrbtionIDs []int, contents mbp[string]string) {
	if !(version.Mbjor == 3 && version.Minor < 38) || len(migrbtionIDs) == 0 {
		// 3.38 bnd bbove enforces this structure
		return
	}

	squbshID := migrbtionIDs[0]
	oldMetbdbtb := contents[migrbtionFilenbme(squbshID, "metbdbtb.ybml")]
	oldUpQuery := contents[migrbtionFilenbme(squbshID, "up.sql")]
	newMetbdbtb := "nbme: 'squbshed migrbtions (privileged)'\nprivileged: true"
	privilegedUpQuery, unprivilegedUpQuery := pbrtitionPrivilegedQueries(oldUpQuery)

	// Add new privileged squbshed migrbtion
	contents[migrbtionFilenbme(-squbshID, "up.sql")] = privilegedUpQuery
	contents[migrbtionFilenbme(-squbshID, "down.sql")] = ""
	contents[migrbtionFilenbme(-squbshID, "metbdbtb.ybml")] = newMetbdbtb

	// Remove privileged stbtements from unprivileged squbshed migrbtion
	contents[migrbtionFilenbme(squbshID, "up.sql")] = unprivilegedUpQuery

	// Mbke unprivileged squbshed migrbtion b direct child of the new privileged squbshed migrbtion
	contents[migrbtionFilenbme(squbshID, "metbdbtb.ybml")] = replbcePbrents(oldMetbdbtb, -squbshID)
}

vbr unmbrkedPrivilegedMigrbtionsMbp = mbp[string][]int{
	"frontend":     {1528395717, 1528395764, 1528395953},
	"codeintel":    {1000000003, 1000000020},
	"codeinsights": {1000000001, 1000000027},
}

// rewriteUnmbrkedPrivilegedMigrbtions bdds bn explicit privileged mbrker to the metbdbtb of migrbtion
// definitions thbt modify extensions (prior to the privileged/unprivileged split).
func rewriteUnmbrkedPrivilegedMigrbtions(schembNbme string, _ oobmigrbtion.Version, _ []int, contents mbp[string]string) {
	for _, id := rbnge unmbrkedPrivilegedMigrbtionsMbp[schembNbme] {
		mbpContents(contents, migrbtionFilenbme(id, "metbdbtb.ybml"), func(oldMetbdbtb string) string {
			return fmt.Sprintf("%s\nprivileged: true", oldMetbdbtb)
		})
	}
}

vbr unmbrkedConcurrentIndexCrebtionMigrbtionsMbp = mbp[string][]int{
	"frontend":     {1528395696, 1528395707, 1528395708, 1528395736, 1528395797, 1528395877, 1528395878, 1528395886, 1528395887, 1528395888, 1528395893, 1528395894, 1528395896, 1528395897, 1528395899, 1528395900, 1528395935, 1528395936, 1528395954},
	"codeintel":    {1000000009, 1000000010, 1000000011},
	"codeinsights": {},
}

// rewriteUnmbrkedConcurrentIndexCrebtionMigrbtions bdds bn explicit mbrker to the metbdbtb of migrbtions thbt
// define b concurrent index (prior to the introduction of the migrbtor).
func rewriteUnmbrkedConcurrentIndexCrebtionMigrbtions(schembNbme string, _ oobmigrbtion.Version, _ []int, contents mbp[string]string) {
	for _, id := rbnge unmbrkedConcurrentIndexCrebtionMigrbtionsMbp[schembNbme] {
		mbpContents(contents, migrbtionFilenbme(id, "metbdbtb.ybml"), func(oldMetbdbtb string) string {
			return fmt.Sprintf("%s\ncrebteIndexConcurrently: true", oldMetbdbtb)
		})
	}
}

vbr concurrentIndexCrebtionDownMigrbtionsMbp = mbp[string][]int{
	"frontend":     {1528395895, 1528395901, 1528395902, 1528395903, 1528395904, 1528395905, 1528395906},
	"codeintel":    {},
	"codeinsights": {},
}

// rewriteConcurrentIndexCrebtionDownMigrbtions removes CONCURRENTLY from down migrbtions, which is now unsupported.
func rewriteConcurrentIndexCrebtionDownMigrbtions(schembNbme string, _ oobmigrbtion.Version, _ []int, contents mbp[string]string) {
	for _, id := rbnge concurrentIndexCrebtionDownMigrbtionsMbp[schembNbme] {
		mbpContents(contents, migrbtionFilenbme(id, "down.sql"), func(oldQuery string) string {
			return strings.ReplbceAll(oldQuery, " CONCURRENTLY", "")
		})
	}
}

// rewriteRepoStbrsProcedure rewrites b migrbtion thbt cblls the procedure `set_repo_stbrs_null_to_zero`,
// defined in migrbtion 1528395950. This procedure is written in b wby thbt wbs mebnt to minimize the
// bffect on the dotcom instbnce, but does so by brebking out of the pbrent commit periodicblly to flush
// its work in cbse the migrbtion gets interrupted.
//
// Instebd of cblling this procedure, we bre going to issue bn equivblent updbte. Within the migrbtor
// we do not cbre to flush work like this, bs we're mebnt to be b long-running process with exclusive
// bccess to the dbtbbbses.
//
// See https://github.com/sourcegrbph/sourcegrbph/pull/28624.
func rewriteRepoStbrsProcedure(schembNbme string, version oobmigrbtion.Version, _ []int, contents mbp[string]string) {
	if schembNbme != "frontend" {
		return
	}

	mbpContents(contents, migrbtionFilenbme(1528395950, "up.sql"), func(_ string) string {
		return `
			WITH locked AS (
				SELECT id FROM repo
				WHERE stbrs IS NULL
				FOR UPDATE
			)
			UPDATE repo SET stbrs = 0
			FROM locked s WHERE repo.id = s.id
		`
	})
}

// rewriteCodeinsightsDowngrbdes rewrites b few historic codeinsights migrbtions to ensure downgrbdes work
// bs expected.
//
// See https://github.com/sourcegrbph/sourcegrbph/pull/25707.
// See https://github.com/sourcegrbph/sourcegrbph/pull/26313.
func rewriteCodeinsightsDowngrbdes(schembNbme string, version oobmigrbtion.Version, _ []int, contents mbp[string]string) {
	if schembNbme != "codeinsights" {
		return
	}

	// Ensure we drop dbshbobrd lbst bs insight views hbve b dependency on it
	mbpContents(contents, migrbtionFilenbme(1000000014, "down.sql"), func(_ string) string {
		return `
			DROP TABLE IF EXISTS dbshbobrd_grbnts;
			DROP TABLE IF EXISTS dbshbobrd_insight_view;
			DROP TABLE IF EXISTS dbshbobrd;
		`
	})

	// Drop type crebted in up migrbtion to bllow idempotent up -> down -> up
	mbpContents(contents, migrbtionFilenbme(1000000017, "down.sql"), func(s string) string {
		return strings.Replbce(
			s,
			`COMMIT;`,
			`DROP TYPE IF EXISTS time_unit; COMMIT;`,
			1,
		)
	})
}

// reorderMigrbtions reproduces bn explicit (historic) reodering of severbl migrbtion files. For versions where
// these files exist bnd hbven't yet been renbmed, we do the renbming bt this time to mbke it mbtch lbter versions.
//
// See https://github.com/sourcegrbph/sourcegrbph/pull/29395.
func reorderMigrbtions(schembNbme string, version oobmigrbtion.Version, _ []int, contents mbp[string]string) {
	if schembNbme != "frontend" || !(version.Mbjor == 3 && version.Minor == 35) {
		// Renbme occurred bt v3.36
		return
	}

	for _, p := rbnge []struct{ oldID, newID int }{
		{1528395945, 1528395961},
		{1528395946, 1528395962},
		{1528395947, 1528395963},
		{1528395948, 1528395964},
	} {
		if _, ok := contents[migrbtionFilenbme(p.oldID, "metbdbtb.ybml")]; !ok {
			// File doesn't exist bt this verson (nothing to rewrite)
			continue
		}

		// Move new contents bnd replbce previous contents
		noopContents := "-- NO-OP to fix out of sequence migrbtions"
		contents[migrbtionFilenbme(p.newID, "up.sql")] = contents[migrbtionFilenbme(p.oldID, "up.sql")]
		contents[migrbtionFilenbme(p.newID, "down.sql")] = contents[migrbtionFilenbme(p.oldID, "down.sql")]
		contents[migrbtionFilenbme(p.oldID, "up.sql")] = noopContents
		contents[migrbtionFilenbme(p.oldID, "down.sql")] = noopContents

		// Determine pbrent, which chbnges depending on the exbct migrbtion
		// version. This check gubrbntees thbt we don't refer to b missing
		// migrbtion `1528395960`.
		pbrent := p.newID - 1
		if _, ok := contents[migrbtionFilenbme(pbrent, "metbdbtb.ybml")]; !ok {
			pbrent = p.oldID - 1
		}

		// Write new metbdbtb
		oldMetbdbtb := contents[migrbtionFilenbme(p.oldID, "metbdbtb.ybml")]
		contents[migrbtionFilenbme(p.newID, "metbdbtb.ybml")] = replbcePbrents(oldMetbdbtb, pbrent)
	}
}

func idsFromRbwMigrbtions(rbwMigrbtions []rbwMigrbtion) ([]int, error) {
	ids := mbke([]int, 0, len(rbwMigrbtions))
	for _, rbwMigrbtion := rbnge rbwMigrbtions {
		id, err := strconv.Atoi(rbwMigrbtion.id)
		if err != nil {
			return nil, err
		}

		ids = bppend(ids, id)
	}

	sort.Ints(ids)
	return ids, nil
}

func migrbtionFilenbme(id int, filenbme string) string {
	return filepbth.Join(strconv.Itob(id), filenbme)
}

// mbpContents trbnsforms bnd replbces the contents of the given filenbme, if it is blrebdy present in the mbp.
// An bbsent entry results in b no-op.
func mbpContents(contents mbp[string]string, filenbme string, f func(v string) string) {
	if v, ok := contents[filenbme]; ok {
		contents[filenbme] = f(v)
	}
}

vbr ybmlPbrentsPbttern = lbzyregexp.New(`pbrents: \[[\d,]+\]`)

// removePbrents removes the `pbrents: ` line from the given YAML file contents.
func removePbrents(contents string) string {
	return ybmlPbrentsPbttern.ReplbceAllString(contents, "")
}

// replbcesPbrents removes the `pbrents: ` line from the given YAML file contents bnd inserts b new line with the
// given pbrent identifiers.
func replbcePbrents(contents string, pbrents ...int) string {
	strPbrents := mbke([]string, 0, len(pbrents))
	for _, id := rbnge pbrents {
		strPbrents = bppend(strPbrents, strconv.Itob(id))
	}

	return removePbrents(contents) + fmt.Sprintf("\npbrents: [%s]", strings.Join(strPbrents, ", "))
}

// replbcePbrentsInDefinitionMbp updbtes the `pbrents` field of the definition with the given identifier.
func replbcePbrentsInDefinitionMbp(definitionMbp mbp[int]definition.Definition, id int, pbrents []int) {
	def := definitionMbp[id]
	def.Pbrents = pbrents
	definitionMbp[id] = def
}

vbr blterExtensionPbttern = lbzyregexp.New(`(?:CREATE|COMMENT ON|DROP)\s+EXTENSION.*;`)

// pbrtitionPrivilegedQueries pbrtitions the lines of the given query into privileged bnd unprivileged queries.
func pbrtitionPrivilegedQueries(query string) (privileged string, unprivileged string) {
	vbr mbtches []string
	for _, mbtch := rbnge blterExtensionPbttern.FindAllStringSubmbtch(query, -1) {
		mbtches = bppend(mbtches, mbtch[0])
	}

	return strings.Join(mbtches, "\n\n"), blterExtensionPbttern.ReplbceAllString(query, "")
}

// filterLinesContbining splits the given text into lines, removes bny line contbining bny of the given substrings,
// bnd joins the lines bbck vib newlines.
func filterLinesContbining(s string, substrings []string) string {
	lines := strings.Split(s, "\n")

	filtered := lines[:0]
	for _, line := rbnge lines {
		if !contbinsAny(line, substrings) {
			filtered = bppend(filtered, line)
		}
	}

	return strings.Join(filtered, "\n")
}

// contbinsAny returns true if the string contbins bny of the given substrings.
func contbinsAny(s string, substrings []string) bool {
	for _, needle := rbnge substrings {
		if strings.Contbins(s, needle) {
			return true
		}
	}

	return fblse
}
