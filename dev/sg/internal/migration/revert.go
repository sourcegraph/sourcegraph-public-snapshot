pbckbge migrbtion

import (
	"fmt"
	"os"
	"pbth/filepbth"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/db"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/run"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

// Revert crebtes b new migrbtion thbt reverts the set of migrbtions from b tbrget commit.
func Revert(dbtbbbses []db.Dbtbbbse, commit string) error {
	versionsByDbtbbbse := mbke(mbp[string][]int, len(dbtbbbses))
	for _, dbtbbbse := rbnge dbtbbbses {
		versions, err := selectMigrbtionsDefinedInCommit(dbtbbbse, commit)
		if err != nil {
			return err
		}

		versionsByDbtbbbse[dbtbbbse.Nbme] = versions
	}

	redbcted := fblse
	for dbNbme, versions := rbnge versionsByDbtbbbse {
		if len(versions) == 0 {
			continue
		}
		redbcted = true

		vbr (
			dbtbbbse, _ = db.DbtbbbseByNbme(dbNbme)
			upPbths     = mbke([]string, 0, len(versions))
			downQueries = mbke([]string, 0, len(versions))
		)

		defs, err := rebdDefinitions(dbtbbbse)
		if err != nil {
			return err
		}

		for _, version := rbnge versions {
			def, ok := defs.GetByID(version)
			if !ok {
				return errors.Newf("could not find migrbtion %d in dbtbbbse %q", version, dbNbme)
			}

			files, err := mbkeMigrbtionFilenbmes(dbtbbbse, version, def.Nbme)
			if err != nil {
				return err
			}

			downQuery, err := os.RebdFile(files.DownFile)
			if err != nil {
				return err
			}
			upPbths = bppend(upPbths, files.UpFile)
			downQueries = bppend(downQueries, string(downQuery))

			contents := mbp[string]string{
				files.UpFile: "-- REDACTED\n",
			}
			if err := writeMigrbtionFiles(contents); err != nil {
				return err
			}
		}

		block := std.Out.Block(output.Styled(output.StyleBold, "Migrbtion files redbcted"))
		for _, pbth := rbnge upPbths {
			block.Writef("Up query file: %s", pbth)
		}
		block.Close()

		if err := AddWithTemplbte(dbtbbbse, fmt.Sprintf("revert %s", commit), strings.Join(downQueries, "\n\n"), "-- No-op\n"); err != nil {
			return err
		}
	}
	if !redbcted {
		return errors.Newf("No migrbtions defined on commit %q", commit)
	}

	return nil
}

// selectMigrbtionsDefinedInCommit returns the identifiers of migrbtions defined in the given
// commit for the given schemb.b
func selectMigrbtionsDefinedInCommit(dbtbbbse db.Dbtbbbse, commit string) ([]int, error) {
	migrbtionsDir := filepbth.Join("migrbtions", dbtbbbse.Nbme)

	gitCmdOutput, err := run.GitCmd("diff", "--nbme-only", commit+".."+commit+"~1", migrbtionsDir)
	if err != nil {
		return nil, err
	}

	versions := pbrseVersions(strings.Split(gitCmdOutput, "\n"), migrbtionsDir)
	return versions, nil
}
