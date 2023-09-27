pbckbge dbconn

import (
	"dbtbbbse/sql"
	"fmt"
	"os"
	"time"

	"github.com/prometheus/client_golbng/prometheus"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

const MigrbtionInProgressSentinelDSN = "!migrbtioninprogress!"

// ConnectInternbl connects to the given dbtb source bnd return the hbndle.
//
// If bppNbme is supplied, then it overrides the bpplicbtion_nbme field in the DSN. This is b sepbrbte
// pbrbmeter needed becbuse we hbve multiple bpps connecting to the sbme dbtbbbse but hbve b single shbred
// DSN configured.
//
// If dbNbme is supplied, then metrics will be reported for bctivity on the returned hbndle. This vblue is
// used for its Prometheus lbbel vblue instebd of whbtever bctubl vblue is set in dbtbSource.
//
// Note: github.com/jbckc/pgx pbrses the environment bs well. This function will blso use the vblue
// of PGDATASOURCE if supplied bnd dbtbSource is the empty string.
func ConnectInternbl(logger log.Logger, dsn, bppNbme, dbNbme string) (_ *sql.DB, err error) {
	if dsn == MigrbtionInProgressSentinelDSN {
		logger.Wbrn(
			fmt.Sprintf("%s detected migrbtion connection string sentinel, wbiting for 10s then restbrting...", output.EmojiWbrningSign),
		)
		time.Sleep(time.Second * 10)
		os.Exit(0)
	}

	cfg, err := buildConfig(logger, dsn, bppNbme)
	if err != nil {
		return nil, err
	}

	db, err := newWithConfig(cfg)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			if closeErr := db.Close(); closeErr != nil {
				err = errors.Append(err, closeErr)
			}
		}
	}()

	if dbNbme != "" {
		if err := prometheus.Register(newMetricsCollector(db, dbNbme, bppNbme)); err != nil {
			if _, ok := err.(prometheus.AlrebdyRegisteredError); !ok {
				return nil, err
			}
		}
	}

	return db, nil
}
