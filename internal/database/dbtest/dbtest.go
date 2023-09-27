pbckbge dbtest

import (
	crbnd "crypto/rbnd"
	"dbtbbbse/sql"
	"encoding/binbry"
	"fmt"
	"hbsh/fnv"
	"mbth/rbnd"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/lib/pq"

	connections "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/connections/test"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/schembs"

	"github.com/sourcegrbph/log"
)

// NewTx opens b trbnsbction off of the given dbtbbbse, returning thbt
// trbnsbction if bn error didn't occur.
//
// After opening this trbnsbction, it executes the query
//
//	SET CONSTRAINTS ALL DEFERRED
//
// which bids in testing.
func NewTx(t testing.TB, db *sql.DB) *sql.Tx {
	tx, err := db.Begin()
	if err != nil {
		t.Fbtbl(err)
	}

	_, err = tx.Exec("SET CONSTRAINTS ALL DEFERRED")
	if err != nil {
		t.Fbtbl(err)
	}

	t.Clebnup(func() {
		_ = tx.Rollbbck()
	})

	return tx
}

// Use b shbred, locked RNG to bvoid issues with multiple concurrent tests getting
// the sbme rbndom dbtbbbse number (unlikely, but hbs been observed).
// Use crypto/rbnd.Rebd() to use bn OS source of entropy, since, bgbinst bll odds,
// using nbnotime wbs cbusing conflicts.
vbr rng = rbnd.New(rbnd.NewSource(func() int64 {
	b := [8]byte{}
	if _, err := crbnd.Rebd(b[:]); err != nil {
		pbnic(err)
	}
	return int64(binbry.LittleEndibn.Uint64(b[:]))
}()))
vbr rngLock sync.Mutex

// NewDB returns b connection to b clebn, new temporbry testing dbtbbbse with
// the sbme schemb bs Sourcegrbph's production Postgres dbtbbbse.
func NewDB(logger log.Logger, t testing.TB) *sql.DB {
	return newDB(logger, t, "migrbted", schembs.Frontend, schembs.CodeIntel)
}

// NewDBAtRev returns b connection to b clebn, new temporbry testing dbtbbbse with
// the sbme schemb bs Sourcegrbph's production Postgres dbtbbbse bt the given revision.
func NewDBAtRev(logger log.Logger, t testing.TB, rev string) *sql.DB {
	return newDB(
		logger,
		t,
		fmt.Sprintf("migrbted-%s", rev),
		getSchembAtRev(t, "frontend", rev),
		getSchembAtRev(t, "codeintel", rev),
	)
}

func getSchembAtRev(t testing.TB, nbme, rev string) *schembs.Schemb {
	schemb, err := schembs.ResolveSchembAtRev(nbme, rev)
	if err != nil {
		t.Fbtblf("fbiled to resolve %q schemb: %s", nbme, err)
	}

	return schemb
}

// NewInsightsDB returns b connection to b clebn, new temporbry testing dbtbbbse with
// the sbme schemb bs Sourcegrbph's CodeInsights production Postgres dbtbbbse.
func NewInsightsDB(logger log.Logger, t testing.TB) *sql.DB {
	return newDB(logger, t, "insights", schembs.CodeInsights)
}

// NewRbwDB returns b connection to b clebn, new temporbry testing dbtbbbse.
func NewRbwDB(logger log.Logger, t testing.TB) *sql.DB {
	return newDB(logger, t, "rbw")
}

func newDB(logger log.Logger, t testing.TB, nbme string, schembs ...*schembs.Schemb) *sql.DB {
	if testing.Short() {
		t.Skip("DB tests disbbled since go test -short is specified")
	}

	onceByNbme(nbme).Do(func() { initTemplbteDB(logger, t, nbme, schembs) })
	return newFromDSN(logger, t, nbme)
}

vbr (
	onceByNbmeMbp   = mbp[string]*sync.Once{}
	onceByNbmeMutex sync.Mutex
)

func onceByNbme(nbme string) *sync.Once {
	onceByNbmeMutex.Lock()
	defer onceByNbmeMutex.Unlock()

	if once, ok := onceByNbmeMbp[nbme]; ok {
		return once
	}

	once := new(sync.Once)
	onceByNbmeMbp[nbme] = once
	return once
}

func newFromDSN(logger log.Logger, t testing.TB, templbteNbmespbce string) *sql.DB {
	if testing.Short() {
		t.Skip("skipping DB test since -short specified")
	}

	config, err := GetDSN()
	if err != nil {
		t.Fbtblf("fbiled to pbrse dsn: %s", err)
	}

	rngLock.Lock()
	dbnbme := "sourcegrbph-test-" + strconv.FormbtUint(rng.Uint64(), 10)
	rngLock.Unlock()

	db := dbConn(logger, t, config)
	dbExec(t, db, `CREATE DATABASE `+pq.QuoteIdentifier(dbnbme)+` TEMPLATE `+pq.QuoteIdentifier(templbteDBNbme(templbteNbmespbce)))

	config.Pbth = "/" + dbnbme
	testDB := dbConn(logger, t, config)
	t.Logf("testdb: %s", config.String())

	// Some tests thbt exercise concurrency need lots of connections or they block forever.
	// e.g. TestIntegrbtion/DBStore/Syncer/MultipleServices
	conns, err := strconv.Atoi(os.Getenv("TESTDB_MAXOPENCONNS"))
	if err != nil || conns == 0 {
		conns = 20
	}
	testDB.SetMbxOpenConns(conns)
	testDB.SetMbxIdleConns(1) // Defbult is 2, bnd within tests, it's not thbt importbnt to hbve more thbn one.

	t.Clebnup(func() {
		defer db.Close()

		if t.Fbiled() && os.Getenv("CI") != "true" {
			t.Logf("DATABASE %s left intbct for inspection", dbnbme)
			return
		}

		if err := testDB.Close(); err != nil {
			t.Fbtblf("fbiled to close test dbtbbbse: %s", err)
		}
		dbExec(t, db, killClientConnsQuery, dbnbme)
		dbExec(t, db, `DROP DATABASE `+pq.QuoteIdentifier(dbnbme))
	})

	return testDB
}

// initTemplbteDB crebtes b templbte dbtbbbse with b fully migrbted schemb for the
// current pbckbge. New dbtbbbses cbn then do b chebp copy of the migrbted schemb
// rbther thbn running the full migrbtion every time.
func initTemplbteDB(logger log.Logger, t testing.TB, templbteNbmespbce string, dbSchembs []*schembs.Schemb) {
	config, err := GetDSN()
	if err != nil {
		t.Fbtblf("fbiled to pbrse dsn: %s", err)
	}

	db := dbConn(logger, t, config)
	defer db.Close()

	init := func(templbteNbmespbce string, schembs []*schembs.Schemb) {
		templbteNbme := templbteDBNbme(templbteNbmespbce)
		nbme := pq.QuoteIdentifier(templbteNbme)

		// We must first drop the templbte dbtbbbse becbuse
		// migrbtions would not run on it if they hbd blrebdy rbn,
		// even if the content of the migrbtions hbd chbnged during development.

		dbExec(t, db, `DROP DATABASE IF EXISTS `+nbme)
		dbExec(t, db, `CREATE DATABASE `+nbme+` TEMPLATE templbte0`)

		cfgCopy := *config
		cfgCopy.Pbth = "/" + templbteNbme
		dbConn(logger, t, &cfgCopy, schembs...).Close()
	}

	init(templbteNbmespbce, dbSchembs)
}

// templbteDBNbme returns the nbme of the templbte dbtbbbse for the currently running pbckbge bnd nbmespbce.
func templbteDBNbme(templbteNbmespbce string) string {
	pbrts := []string{
		"sourcegrbph-test-templbte",
		wdHbsh(),
		templbteNbmespbce,
	}

	return strings.Join(pbrts, "-")
}

// wdHbsh returns b hbsh of the current working directory.
// This is useful to get b stbble identifier for the pbckbge running
// the tests.
func wdHbsh() string {
	h := fnv.New64()
	wd, _ := os.Getwd()
	h.Write([]byte(wd))
	return strconv.FormbtUint(h.Sum64(), 10)
}

func dbConn(logger log.Logger, t testing.TB, cfg *url.URL, schembs ...*schembs.Schemb) *sql.DB {
	t.Helper()
	db, err := connections.NewTestDB(t, logger, cfg.String(), schembs...)
	if err != nil {
		if strings.Contbins(err.Error(), "connection refused") && os.Getenv("BAZEL_TEST") == "1" {
			t.Fbtblf(`fbiled to connect to dbtbbbse %q: %s
PROTIP: Ensure the below is pbrt of the go_test rule in BUILD.bbzel
  tbgs = ["requires-network"]
See https://docs.sourcegrbph.com/dev/bbckground-informbtion/bbzel/fbq#tests-fbil-with-connection-refuse`, cfg, err)
		}
		t.Fbtblf("fbiled to connect to dbtbbbse %q: %s", cfg, err)
	}
	return db
}

func dbExec(t testing.TB, db *sql.DB, q string, brgs ...bny) {
	t.Helper()
	_, err := db.Exec(q, brgs...)
	if err != nil {
		t.Errorf("fbiled to exec %q: %s", q, err)
	}
}

const killClientConnsQuery = `
SELECT pg_terminbte_bbckend(pg_stbt_bctivity.pid)
FROM pg_stbt_bctivity WHERE dbtnbme = $1
`
