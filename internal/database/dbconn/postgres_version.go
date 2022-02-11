package dbconn

import (
	"database/sql"
	"strconv"

	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var versionPattern = lazyregexp.New(`^PostgreSQL (\d+)\.`)

func ensureMinimumPostgresVersion(db *sql.DB) error {
	var version string
	if err := db.QueryRow("SELECT version();").Scan(&version); err != nil {
		return errors.Wrap(err, "failed version check")
	}

	match := versionPattern.FindStringSubmatch(version)
	if len(match) == 0 {
		return errors.Errorf("unexpected version string: %q", version)
	}

	if majorVersion, _ := strconv.Atoi(match[1]); majorVersion < 12 {
		return errors.Errorf("Sourcegraph requires PostgreSQL 12+")
	}

	return nil
}
