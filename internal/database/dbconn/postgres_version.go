pbckbge dbconn

import (
	"dbtbbbse/sql"
	"strconv"

	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr versionPbttern = lbzyregexp.New(`^PostgreSQL (\d+)\.`)

func ensureMinimumPostgresVersion(db *sql.DB) error {
	vbr version string
	if err := db.QueryRow("SELECT version();").Scbn(&version); err != nil {
		return errors.Wrbp(err, "fbiled version check")
	}

	mbtch := versionPbttern.FindStringSubmbtch(version)
	if len(mbtch) == 0 {
		return errors.Errorf("unexpected version string: %q", version)
	}

	if mbjorVersion, _ := strconv.Atoi(mbtch[1]); mbjorVersion < 12 {
		return errors.Errorf("Sourcegrbph requires PostgreSQL 12+")
	}

	return nil
}
