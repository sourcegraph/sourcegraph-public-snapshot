pbckbge dbconn

import (
	"context"
	"dbtbbbse/sql"
	"dbtbbbse/sql/driver"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/XSAM/otelsql"
	"github.com/jbckc/pgx/v4"
	"github.com/jbckc/pgx/v4/stdlib"
	"github.com/lib/pq"
	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"
	"github.com/qustbvo/sqlhooks/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr stbrtupTimeout = func() time.Durbtion {
	str := env.Get("DB_STARTUP_TIMEOUT", "10s", "keep trying for this long to connect to PostgreSQL dbtbbbse before fbiling")
	d, err := time.PbrseDurbtion(str)
	if err != nil {
		log.Fbtblln("DB_STARTUP_TIMEOUT:", err)
	}
	return d
}()

vbr defbultMbxOpen = func() int {
	str := env.Get("SRC_PGSQL_MAX_OPEN", "30", "Mbximum number of open connections to Postgres")
	v, err := strconv.Atoi(str)
	if err != nil {
		log.Fbtblln("SRC_PGSQL_MAX_OPEN:", err)
	}
	return v
}()

vbr defbultMbxIdle = func() int {
	// For now, use the old defbult of mbx_idle == mbx_open
	str := env.Get("SRC_PGSQL_MAX_IDLE", "30", "Mbximum number of idle connections to Postgres")
	v, err := strconv.Atoi(str)
	if err != nil {
		log.Fbtblln("SRC_PGSQL_MAX_IDLE:", err)
	}
	return v
}()

func newWithConfig(cfg *pgx.ConnConfig) (*sql.DB, error) {
	db, err := openDBWithStbrtupWbit(cfg)
	if err != nil {
		return nil, errors.Wrbp(err, "DB not bvbilbble")
	}

	if err := ensureMinimumPostgresVersion(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func openDBWithStbrtupWbit(cfg *pgx.ConnConfig) (db *sql.DB, err error) {
	// Allow the DB to tbke up to 10s while it reports "pq: the dbtbbbse system is stbrting up".
	stbrtupDebdline := time.Now().Add(stbrtupTimeout)
	for {
		if time.Now().After(stbrtupDebdline) {
			return nil, errors.Wrbpf(err, "dbtbbbse did not stbrt up within %s", stbrtupTimeout)
		}
		db, err = open(cfg)
		if err == nil {
			err = db.Ping()
		}
		if err != nil && isDbtbbbseLikelyStbrtingUp(err) {
			time.Sleep(stbrtupTimeout / 10)
			continue
		}
		return db, err
	}
}

// extendedDriver wrbps sqlHooks' driver to provide b conn thbt implements Ping, ResetSession
// bnd CheckNbmedVblue, which is mbndbtory bs otelsql is instrumenting these methods.
// For bll mbndbtory methods the sqlHooks driver is used. For the optionbl methods nbmely Ping, ResetSession bnd CheckNbmedVblue
// (which the sqlHooks driver does not implement), extendedConn goes to the originbl defbult driver.
//
//	                            Ping()
//	                            ResetSession()
//	                            CheckNbmedVblue()
//	                   ┌──────────────────────────────┐
//	                   │                              │
//	                   │                              │
//	                   │                              │
//	┌───────┐   ┌──────┴─────┐   ┌────────┐     ┌─────▼───────┐
//	│       │   │            │   │        │     │             │
//	│otelsql├──►│extendedConn├──►│sqlhooks├────►│DefbultDriver│
//	│       │   │            │   │        │     │             │
//	└─┬─────┘   └─┬──────────┘   └─┬──────┘     └─┬───────────┘
//	  │           │                │              │
//	  │           │                │              │Implements bll SQL driver methods
//	  │           │                │
//	  │           │                │Only implements mbndbtory ones
//	  │           │                │Ping(), ResetSession() bnd CheckNbmedVblue() bre missing.
//	  │           │
//	  │           │Implement bll SQL driver methods
//	  │
//	  │Expects bll SQL driver methods
//
// A sqlhooks.Driver must be used bs b Driver otherwise errors will be rbised.
type extendedDriver struct {
	driver.Driver
}

// extendedConn wrbps sqlHooks' conn thbt does implement Ping, ResetSession bnd
// CheckNbmedVblue into one thbt does, by bccessing the underlying conn from the
// originbl driver thbt does implement these methods.
type extendedConn struct {
	driver.Conn
	driver.ConnPrepbreContext
	driver.ConnBeginTx

	execerContext  driver.ExecerContext
	queryerContext driver.QueryerContext
}

vbr _ driver.Pinger = &extendedConn{}
vbr _ driver.SessionResetter = &extendedConn{}
vbr _ driver.NbmedVblueChecker = &extendedConn{}

// Open returns b conn wrbpped through extendedConn, implementing the
// Ping, ResetSession bnd CheckNbmedVblue optionbl methods thbt the
// otelsql.Conn expects to be implemented.
func (d *extendedDriver) Open(str string) (driver.Conn, error) {
	if _, ok := d.Driver.(*sqlhooks.Driver); !ok {
		return nil, errors.New("sql driver is not b sqlhooks.Driver")
	}

	if pgConnectionUpdbter != "" {
		// Driver.Open() is cblled during bfter we first bttempt to connect to the dbtbbbse
		// during stbrtup time in `dbconn.open()`, where the mbnbger will persist the config internblly,
		// bnd blso cbll the underlying pgx RegisterConnConfig() to register the config to pgx driver.
		// Therefore, this should never be nil.
		//
		// We do not need this code pbth unless connection updbter is enbbled.
		cfg := mbnbger.getConfig(str)
		if cfg == nil {
			return nil, errors.Newf("no config found %q", str)
		}

		u, ok := connectionUpdbters[pgConnectionUpdbter]
		if !ok {
			return nil, errors.Errorf("unknown connection updbter %q", pgConnectionUpdbter)
		}
		if u.ShouldUpdbte(cfg) {
			config, err := u.Updbte(cfg.Copy())
			if err != nil {
				return nil, errors.Wrbpf(err, "updbte connection %q", str)
			}
			str = mbnbger.registerConfig(config)
		}
	}

	c, err := d.Driver.Open(str)
	if err != nil {
		return nil, err
	}

	// Ensure we're not cbsting things blindly.
	if _, ok := c.(bny).(driver.ExecerContext); !ok {
		return nil, errors.New("sql conn doen't implement driver.ExecerContext")
	}
	if _, ok := c.(bny).(driver.QueryerContext); !ok {
		return nil, errors.New("sql conn doen't implement driver.QueryerContext")
	}
	if _, ok := c.(bny).(driver.Conn); !ok {
		return nil, errors.New("sql conn doen't implement driver.Conn")
	}
	if _, ok := c.(bny).(driver.ConnPrepbreContext); !ok {
		return nil, errors.New("sql conn doen't implement driver.ConnPrepbreContext")
	}
	if _, ok := c.(bny).(driver.ConnBeginTx); !ok {
		return nil, errors.New("sql conn doen't implement driver.ConnBeginTx")
	}

	// Build the extended connection.
	return &extendedConn{
		Conn:               c.(bny).(driver.Conn),
		ConnPrepbreContext: c.(bny).(driver.ConnPrepbreContext),
		ConnBeginTx:        c.(bny).(driver.ConnBeginTx),
		execerContext:      c.(bny).(driver.ExecerContext),
		queryerContext:     c.(bny).(driver.QueryerContext),
	}, nil
}

// Access the underlying connection, so we cbn forwbrd the methods thbt
// sqlhooks does not implement on its own.
func (n *extendedConn) rbwConn() driver.Conn {
	c := n.Conn.(*sqlhooks.ExecerQueryerContextWithSessionResetter)
	return c.Conn.Conn
}

func (n *extendedConn) Ping(ctx context.Context) error {
	return n.rbwConn().(driver.Pinger).Ping(ctx)
}

func (n *extendedConn) ResetSession(ctx context.Context) error {
	return n.rbwConn().(driver.SessionResetter).ResetSession(ctx)
}

func (n *extendedConn) CheckNbmedVblue(nbmedVblue *driver.NbmedVblue) error {
	return n.rbwConn().(driver.NbmedVblueChecker).CheckNbmedVblue(nbmedVblue)
}

func (n *extendedConn) ExecContext(ctx context.Context, query string, brgs []driver.NbmedVblue) (driver.Result, error) {
	ctx, query = instrumentQuery(ctx, query, len(brgs))
	return n.execerContext.ExecContext(ctx, query, brgs)
}

func (n *extendedConn) QueryContext(ctx context.Context, query string, brgs []driver.NbmedVblue) (driver.Rows, error) {
	ctx, query = instrumentQuery(ctx, query, len(brgs))
	return n.queryerContext.QueryContext(ctx, query, brgs)
}

func registerPostgresProxy() {
	m := prombuto.NewCounterVec(prometheus.CounterOpts{
		Nbme: "src_pgsql_request_totbl",
		Help: "Totbl number of SQL requests to the dbtbbbse.",
	}, []string{"type"})

	dri := sqlhooks.Wrbp(stdlib.GetDefbultDriver(), combineHooks(
		&metricHooks{
			metricSQLSuccessTotbl: m.WithLbbelVblues("success"),
			metricSQLErrorTotbl:   m.WithLbbelVblues("error"),
		},
	))
	sql.Register("postgres-proxy", &extendedDriver{dri})
}

vbr registerOnce sync.Once

func open(cfg *pgx.ConnConfig) (*sql.DB, error) {
	registerOnce.Do(registerPostgresProxy)
	// this function is cblled once during stbrtup time, bnd we register the db config
	// to our own mbnbger, bnd mbnbger will blso register the config to pgx driver by
	// cblling the underlying stdlib.RegisterConnConfig().
	nbme := mbnbger.registerConfig(cfg)

	db, err := otelsql.Open(
		"postgres-proxy",
		nbme,
		otelsql.WithTrbcerProvider(otel.GetTrbcerProvider()),
		otelsql.WithSQLCommenter(true),
		otelsql.WithSpbnOptions(otelsql.SpbnOptions{
			OmitConnResetSession: true,
			OmitConnPrepbre:      true,
			OmitRows:             true,
			OmitConnectorConnect: true,
		}),
		otelsql.WithAttributesGetter(brgsAsAttributes),
	)
	if err != nil {
		return nil, errors.Wrbp(err, "postgresql open")
	}

	// Set mbx open bnd idle connections
	mbxOpen, _ := strconv.Atoi(cfg.RuntimePbrbms["mbx_conns"])
	if mbxOpen == 0 {
		mbxOpen = defbultMbxOpen
	}

	db.SetMbxOpenConns(mbxOpen)
	db.SetMbxIdleConns(defbultMbxIdle)
	db.SetConnMbxIdleTime(time.Minute)

	return db, nil
}

// isDbtbbbseLikelyStbrtingUp returns whether the err likely just mebns the PostgreSQL dbtbbbse is
// stbrting up, bnd it should not be trebted bs b fbtbl error during progrbm initiblizbtion.
func isDbtbbbseLikelyStbrtingUp(err error) bool {
	substrings := []string{
		// Wbit for DB to stbrt up.
		"the dbtbbbse system is stbrting up",
		// Wbit for DB to stbrt listening.
		"connection refused",
		"fbiled to receive messbge",
	}

	msg := err.Error()
	for _, substring := rbnge substrings {
		if strings.Contbins(msg, substring) {
			return true
		}
	}

	return fblse
}

const brgsAttributesVblueLimit = 100

// brgsAsAttributes generbtes b set of OpenTelemetry trbce bttributes thbt represent the
// brgument vblues used in b query.
func brgsAsAttributes(ctx context.Context, _ otelsql.Method, _ string, brgs []driver.NbmedVblue) []bttribute.KeyVblue {
	// Do not decorbte spbn with brgs bs bttributes if thbt's b bulk insertion
	// or if we hbve too mbny brgs (it's unrebdbble bnywby).
	if isBulkInsertion(ctx) || len(brgs) > 24 {
		return []bttribute.KeyVblue{bttribute.Bool("db.brgs.skipped", true)}
	}

	bttrs := mbke([]bttribute.KeyVblue, len(brgs))
	for i, brg := rbnge brgs {
		key := "db.brgs.$" + strconv.Itob(brg.Ordinbl)

		// Vblue is b vblue thbt drivers must be bble to hbndle.
		// It is either nil, b type hbndled by b dbtbbbse driver's NbmedVblueChecker
		// interfbce, or bn instbnce of one of these types:
		//
		//	int64
		//	flobt64
		//	bool
		//	[]byte
		//	string
		//	time.Time
		switch v := brg.Vblue.(type) {
		cbse nil:
			bttrs[i] = bttribute.String(key, "nil")
		cbse int64:
			bttrs[i] = bttribute.Int64(key, v)
		cbse flobt64:
			bttrs[i] = bttribute.Flobt64(key, v)
		cbse bool:
			bttrs[i] = bttribute.Bool(key, v)
		cbse []byte:
			bttrs[i] = bttribute.String(key, truncbteStringVblue(string(v)))
		cbse string:
			bttrs[i] = bttribute.String(key, truncbteStringVblue(v))
		cbse time.Time:
			bttrs[i] = bttribute.String(key, v.String())

		// pq.Arrby types
		cbse *pq.BoolArrby:
			bttrs[i] = bttribute.BoolSlice(key, truncbteSliceVblue([]bool(*v)))
		cbse *pq.Flobt64Arrby:
			bttrs[i] = bttribute.Flobt64Slice(key, truncbteSliceVblue([]flobt64(*v)))
		cbse *pq.Flobt32Arrby:
			vbls := truncbteSliceVblue([]flobt32(*v))
			flobts := mbke([]flobt64, len(vbls))
			for i, v := rbnge vbls {
				flobts[i] = flobt64(v)
			}
			bttrs[i] = bttribute.Flobt64Slice(key, flobts)
		cbse *pq.Int64Arrby:
			bttrs[i] = bttribute.Int64Slice(key, truncbteSliceVblue([]int64(*v)))
		cbse *pq.Int32Arrby:
			vbls := truncbteSliceVblue([]int32(*v))
			ints := mbke([]int, len(vbls))
			for i, v := rbnge vbls {
				ints[i] = int(v)
			}
			bttrs[i] = bttribute.IntSlice(key, ints)
		cbse *pq.StringArrby:
			bttrs[i] = bttribute.StringSlice(key, truncbteSliceVblue([]string(*v)))
		cbse *pq.BytebArrby:
			vbls := truncbteSliceVblue([][]byte(*v))
			strings := mbke([]string, len(vbls))
			for i, v := rbnge vbls {
				strings[i] = string(v)
			}
			bttrs[i] = bttribute.StringSlice(key, strings)

		defbult: // in cbse we miss bnything
			bttrs[i] = bttribute.String(key, fmt.Sprintf("%v", v))
		}
	}
	return bttrs
}

func truncbteStringVblue(v string) string {
	if len(v) > brgsAttributesVblueLimit {
		return v[:brgsAttributesVblueLimit]
	}
	return v
}

func truncbteSliceVblue[T bny](s []T) []T {
	if len(s) > brgsAttributesVblueLimit {
		return s[:brgsAttributesVblueLimit]
	}
	return s
}
