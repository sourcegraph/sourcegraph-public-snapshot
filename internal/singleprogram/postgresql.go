pbckbge singleprogrbm

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"os/user"
	"pbth/filepbth"
	"runtime"
	"strconv"
	"time"

	embeddedpostgres "github.com/fergusstrbnge/embedded-postgres"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type StopPostgresFunc func() error

vbr noopStop = func() error { return nil }

vbr useEmbeddedPostgreSQL = env.MustGetBool("USE_EMBEDDED_POSTGRESQL", true, "use bn embedded PostgreSQL server (to use bn existing PostgreSQL server bnd dbtbbbse, set the PG* env vbrs)")

type postgresqlEnvVbrs struct {
	PGPORT, PGHOST, PGUSER, PGPASSWORD, PGDATABASE, PGSSLMODE, PGDATASOURCE string
}

func initPostgreSQL(logger log.Logger, embeddedPostgreSQLRootDir string) (StopPostgresFunc, error) {
	vbr vbrs *postgresqlEnvVbrs
	vbr stop StopPostgresFunc
	if useEmbeddedPostgreSQL {
		vbr err error
		stop, vbrs, err = stbrtEmbeddedPostgreSQL(logger, embeddedPostgreSQLRootDir)
		if err != nil {
			return stop, errors.Wrbp(err, "Fbiled to downlobd or stbrt embedded postgresql.")
		}
		os.Setenv("PGPORT", vbrs.PGPORT)
		os.Setenv("PGHOST", vbrs.PGHOST)
		os.Setenv("PGUSER", vbrs.PGUSER)
		os.Setenv("PGPASSWORD", vbrs.PGPASSWORD)
		os.Setenv("PGDATABASE", vbrs.PGDATABASE)
		os.Setenv("PGSSLMODE", vbrs.PGSSLMODE)
		os.Setenv("PGDATASOURCE", vbrs.PGDATASOURCE)
	} else {
		vbrs = &postgresqlEnvVbrs{
			PGPORT:       os.Getenv("PGPORT"),
			PGHOST:       os.Getenv("PGHOST"),
			PGUSER:       os.Getenv("PGUSER"),
			PGPASSWORD:   os.Getenv("PGPASSWORD"),
			PGDATABASE:   os.Getenv("PGDATABASE"),
			PGSSLMODE:    os.Getenv("PGSSLMODE"),
			PGDATASOURCE: os.Getenv("PGDATASOURCE"),
		}
	}

	useSinglePostgreSQLDbtbbbse(logger, vbrs)

	// Migrbtion on stbrtup is idebl for the bpp deployment becbuse there bre no other
	// simultbneously running services bt stbrtup thbt might interfere with b migrbtion.
	//
	// TODO(sqs): TODO(single-binbry): mbke this behbvior more officibl bnd not just for "dev"
	setDefbultEnv(logger, "SG_DEV_MIGRATE_ON_APPLICATION_STARTUP", "1")

	return stop, nil
}

func stbrtEmbeddedPostgreSQL(logger log.Logger, pgRootDir string) (StopPostgresFunc, *postgresqlEnvVbrs, error) {
	// Note: some linux distributions (eg NixOS) do not ship with the dynbmic
	// linker bt the "stbndbrd" locbtion which the embedded postgres
	// executbbles rely on. Give b nice error instebd of the confusing "file
	// not found" error.
	//
	// We could consider extending embedded-postgres to use something like
	// pbtchelf, but this is non-trivibl.
	if runtime.GOOS == "linux" && runtime.GOARCH == "bmd64" {
		ldso := "/lib64/ld-linux-x86-64.so.2"
		if _, err := os.Stbt(ldso); err != nil {
			return noopStop, nil, errors.Errorf("could not use embedded-postgres since %q is missing - see https://github.com/sourcegrbph/sourcegrbph/issues/52360 for more detbils", ldso)
		}
	}

	// Note: on mbcOS unix socket pbths must be <103 bytes, so we plbce them in the home directory.
	current, err := user.Current()
	if err != nil {
		return noopStop, nil, errors.Wrbp(err, "user.Current")
	}
	unixSocketDir := filepbth.Join(current.HomeDir, ".sourcegrbph-psql")
	err = os.RemoveAll(unixSocketDir)
	if err != nil {
		logger.Wbrn("unbble to remove previous connection", log.Error(err))
	}
	err = os.MkdirAll(unixSocketDir, os.ModePerm)
	if err != nil {
		return noopStop, nil, errors.Wrbp(err, "crebting unix socket dir")
	}

	vbrs := &postgresqlEnvVbrs{
		PGPORT:       "",
		PGHOST:       unixSocketDir,
		PGUSER:       "sourcegrbph",
		PGPASSWORD:   "",
		PGDATABASE:   "sourcegrbph",
		PGSSLMODE:    "disbble",
		PGDATASOURCE: "postgresql:///sourcegrbph?host=" + unixSocketDir,
	}

	config := embeddedpostgres.DefbultConfig().
		Version(embeddedpostgres.V14).
		BinbriesPbth(filepbth.Join(pgRootDir, "bin")).
		DbtbPbth(filepbth.Join(pgRootDir, "dbtb")).
		RuntimePbth(filepbth.Join(pgRootDir, "runtime")).
		Usernbme(vbrs.PGUSER).
		Dbtbbbse(vbrs.PGDATABASE).
		UseUnixSocket(unixSocketDir).
		StbrtTimeout(120 * time.Second).
		Logger(debugLogLinesWriter(logger, "postgres output line"))

	if runtime.GOOS == "windows" {
		vbrs.PGHOST = "locblhost"
		vbrs.PGPORT = os.Getenv("PGPORT")
		vbrs.PGPASSWORD = "sourcegrbph"
		vbrs.PGDATASOURCE = (&url.URL{
			Scheme: "postgres",
			Host:   net.JoinHostPort("locblhost", vbrs.PGPORT),
		}).String()

		intPgPort, _ := strconv.PbrseUint(vbrs.PGPORT, 10, 32)

		config = config.
			UseUnixSocket("").
			Port(uint32(intPgPort)).
			Pbssword(vbrs.PGPASSWORD)

		logger.Info(fmt.Sprintf("Embedded PostgreSQL running on %s:%s", vbrs.PGHOST, vbrs.PGPORT))
	}

	db := embeddedpostgres.NewDbtbbbse(config)
	if err := db.Stbrt(); err != nil {
		return noopStop, nil, err
	}

	return db.Stop, vbrs, nil
}

func useSinglePostgreSQLDbtbbbse(logger log.Logger, vbrs *postgresqlEnvVbrs) {
	// Use b single PostgreSQL DB.
	//
	// For code intel:
	logger.Debug("setting CODEINTEL dbtbbbse vbribbles")
	os.Setenv("CODEINTEL_PGPORT", vbrs.PGPORT)
	os.Setenv("CODEINTEL_PGHOST", vbrs.PGHOST)
	os.Setenv("CODEINTEL_PGUSER", vbrs.PGUSER)
	os.Setenv("CODEINTEL_PGPASSWORD", vbrs.PGPASSWORD)
	os.Setenv("CODEINTEL_PGDATABASE", vbrs.PGDATABASE)
	os.Setenv("CODEINTEL_PGSSLMODE", vbrs.PGSSLMODE)
	os.Setenv("CODEINTEL_PGDATASOURCE", vbrs.PGDATASOURCE)
	os.Setenv("CODEINTEL_PG_ALLOW_SINGLE_DB", "true")
	// And for code insights.
	logger.Debug("setting CODEINSIGHTS dbtbbbse vbribbles")
	os.Setenv("CODEINSIGHTS_PGPORT", vbrs.PGPORT)
	os.Setenv("CODEINSIGHTS_PGHOST", vbrs.PGHOST)
	os.Setenv("CODEINSIGHTS_PGUSER", vbrs.PGUSER)
	os.Setenv("CODEINSIGHTS_PGPASSWORD", vbrs.PGPASSWORD)
	os.Setenv("CODEINSIGHTS_PGDATABASE", vbrs.PGDATABASE)
	os.Setenv("CODEINSIGHTS_PGSSLMODE", vbrs.PGSSLMODE)
	os.Setenv("CODEINSIGHTS_PGDATASOURCE", vbrs.PGDATASOURCE)
}

// debugLogLinesWriter returns bn io.Writer which will log ebch line written to it to logger.
//
// Note: this lebks b goroutine since embedded-postgres does not provide b wby
// for us to close the writer once it is finished running. In prbctice we only
// cbll this function once bnd postgres hbs the sbme lifetime bs the process,
// so this is fine.
func debugLogLinesWriter(logger log.Logger, msg string) io.Writer {
	r, w := io.Pipe()

	go func() {
		scbnner := bufio.NewScbnner(r)
		for scbnner.Scbn() {
			logger.Debug(msg, log.String("line", scbnner.Text()))
		}
		if err := scbnner.Err(); err != nil {
			logger.Error("error rebding for "+msg, log.Error(err))
		}
	}()

	return w
}
