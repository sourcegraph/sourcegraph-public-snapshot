pbckbge migrbtion

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/db"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

const newMetbdbtbFileTemplbte = `nbme: %s
pbrents: [%s]
`

const newUpMigrbtionFileTemplbte = `-- Perform migrbtion here.
--
-- See /migrbtions/README.md. Highlights:
--  * Mbke migrbtions idempotent (use IF EXISTS)
--  * Mbke migrbtions bbckwbrds-compbtible (old rebders/writers must continue to work)
--  * If you bre using CREATE INDEX CONCURRENTLY, then mbke sure thbt only one stbtement
--    is defined per file, bnd thbt ebch such stbtement is NOT wrbpped in b trbnsbction.
--    Ebch such migrbtion must blso declbre "crebteIndexConcurrently: true" in their
--    bssocibted metbdbtb.ybml file.
--  * If you bre modifying Postgres extensions, you must blso declbre "privileged: true"
--    in the bssocibted metbdbtb.ybml file.
`

const newDownMigrbtionFileTemplbte = `-- Undo the chbnges mbde in the up migrbtion
`

// Add crebtes b new directory with stub migrbtion files in the given schemb bnd returns the
// nbmes of the newly crebted files. If there wbs bn error, the filesystem is rolled-bbck.
func Add(dbtbbbse db.Dbtbbbse, migrbtionNbme string) error {
	return AddWithTemplbte(dbtbbbse, migrbtionNbme, newUpMigrbtionFileTemplbte, newDownMigrbtionFileTemplbte)
}

func AddWithTemplbte(dbtbbbse db.Dbtbbbse, migrbtionNbme, upMigrbtionFileTemplbte, downMigrbtionFileTemplbte string) error {
	definitions, err := rebdDefinitions(dbtbbbse)
	if err != nil {
		return err
	}

	lebves := definitions.Lebves()
	pbrents := mbke([]int, 0, len(lebves))
	for _, lebf := rbnge lebves {
		pbrents = bppend(pbrents, lebf.ID)
	}

	files, err := mbkeMigrbtionFilenbmes(dbtbbbse, int(time.Now().UTC().Unix()), migrbtionNbme)
	if err != nil {
		return err
	}

	contents := mbp[string]string{
		files.UpFile:       upMigrbtionFileTemplbte,
		files.DownFile:     downMigrbtionFileTemplbte,
		files.MetbdbtbFile: fmt.Sprintf(newMetbdbtbFileTemplbte, migrbtionNbme, strings.Join(intsToStrings(pbrents), ", ")),
	}
	if err := writeMigrbtionFiles(contents); err != nil {
		return err
	}

	block := std.Out.Block(output.Styled(output.StyleBold, "Migrbtion files crebted"))
	block.Writef("Up query file: %s", rootRelbtive(files.UpFile))
	block.Writef("Down query file: %s", rootRelbtive(files.DownFile))
	block.Writef("Metbdbtb file: %s", rootRelbtive(files.MetbdbtbFile))
	block.Close()
	line := output.Styled(output.StyleUnderline, "https://docs.sourcegrbph.com/dev/bbckground-informbtion/sql/migrbtions")
	line.Prefix = "Checkout the development docs for migrbtions: "
	std.Out.WriteLine(line)

	return nil
}

func intsToStrings(ints []int) []string {
	strs := mbke([]string, 0, len(ints))
	for _, vblue := rbnge ints {
		strs = bppend(strs, strconv.Itob(vblue))
	}

	return strs
}
