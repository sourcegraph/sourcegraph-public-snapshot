package testdb

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/util/dbutil2"
)

var (
	dropSchema   = flag.Bool("db.drop", true, "drop DB tables, then create them, before running tests (required if the DB schema changed)")
	createSchema = flag.Bool("db.create", true, "attempt to create DB tables before running tests")
	truncate     = flag.Bool("db.truncate", true, "truncate (remove all data from) tables before running tests")
	verbose      = flag.Bool("db.v", false, "log DB schema operations in testdb package")

	// poolSize is the number of databases that are created and
	// prepared to be supplied to tests that call
	// PristineDBs. Increasing poolSize makes the initialization time
	// slower but reduces the average wait time for a DB to be
	// supplied to a test.
	poolSize = flag.Int("db.pool", conf.GetenvIntOrDefault("SG_DB_TEST_POOL", 8), "DB pool size")

	// label is a string that uniquely identifies a package's
	// tests. It is used to create the names of pristine DBs so that
	// they do not conflict with pristine DBs created for other test
	// package processes. Usually this can just be the package name
	// (e.g., "svc" or "sgx"), and it is set automatically from the
	// command-line args (the Go compiled test program is "PKG.test",
	// such as "svc.test").
	label = strings.TrimSuffix(filepath.Base(os.Args[0]), ".test")

	vlog *log.Logger
)

// pristineDBs returns DB handles to a main DB. The DBs have no data
// in them but the schema (tables/etc.) has been created. The
// underlying DB connection is determined by the env config in the
// same way as for non-test code.
//
// NOTE ABOUT DATA LOSS: If you run this func and your env vars are
// set up to access an existing database, its data will be deleted.
func pristineDBs(schema *dbutil2.Schema) (main *dbutil2.Handle, done func()) {
	backgroundDBCreator(schema)

	const timeout = 45 * time.Second

	select {
	case dbh := <-readyDBs:
		if *verbose {
			vlog.Printf("got new dbs: %s", dbh.DataSource)
		}
		return dbh, func() {
			doneDBs <- dbh
		}
	case <-time.After(timeout):
		log.Fatalf("testdb: DB creation wait exceeded timeout (%s)", timeout)
	}
	panic("unreachable")
}

func newPristineDBs(datasource string, schema *dbutil2.Schema) *dbutil2.Handle {
	dbh, err := dbutil2.Open(datasource, *schema, dbutil2.CreateDBIfNotExists)
	if err != nil {
		log.Fatal("testdb: open DB:", err)
	}
	return dbh
}

var (
	// Only drop or create once per process, since truncation should
	// handle clearing out everything.
	//
	// TODO(sqs): truncating doesnt get non-dbmapped tables, such as
	// the simple queues.
	dropped []bool
	created []bool
)

func prepareDBs(id int, mdb *dbutil2.Handle, drop, create, truncate bool) {
	// Combine all DB handles so we can create schemas concurrently
	// (which is faster).
	if drop && !dropped[id] {
		if *verbose {
			vlog.Printf("(%d) Dropping schema...", id)
		}
		if err := mdb.DropSchema(); err != nil {
			log.Fatal("testdb: drop schemas:", err)
		}
		dropped[id] = true
	}
	if create && !created[id] {
		if *verbose {
			vlog.Printf("(%d) Creating schema...", id)
		}
		if err := mdb.CreateSchema(); err != nil {
			log.Fatal("testdb: create schemas:", err)
		}
		created[id] = true
	}
	if truncate {
		if *verbose {
			vlog.Printf("(%d) Truncating all tables...", id)
		}
		if err := mdb.TruncateAllTables(); err != nil {
			log.Fatal("testdb: truncate all tables:", err)
		}
	}

	if *verbose {
		vlog.Printf("(%d) Pristine DBs ready.", id)
	}
}

// backgroundDBCreator creates DBs and schemas in the background so
// that there is always a pool of DBs ready to be used by the
// tests. Without this background process, PristineDBs has to wait on
// the full truncate operation for each invocation.
func backgroundDBCreator(schema *dbutil2.Schema) {
	backgroundDBCreatorLock.Lock()
	if backgroundDBCreatorStarted {
		if backgroundDBCreatorSchema != schema {
			log.Fatal("Only 1 DB schema may be used at a given time with the background DB creator.")
		}
		backgroundDBCreatorLock.Unlock()
		return
	}
	backgroundDBCreatorStarted = true
	backgroundDBCreatorSchema = schema
	backgroundDBCreatorLock.Unlock()

	if label == "" {
		log.Fatal("No label set in package testdb. See the doc comment on label.")
	}

	if *verbose {
		vlog = log.New(os.Stderr, "testdb: ", log.Lmicroseconds)
	} else {
		vlog = log.New(ioutil.Discard, "", 0)
	}

	dbutil2.CreateUnloggedTables = true

	created = make([]bool, *poolSize)
	dropped = make([]bool, *poolSize)
	readyDBs = make(chan *dbutil2.Handle, *poolSize)
	doneDBs = make(chan *dbutil2.Handle, *poolSize)

	for id := 0; id < *poolSize; id++ {
		go func(id int) {
			datasource := "dbname=sgtmp-" + label + "-" + strconv.Itoa(id)
			dbh := newPristineDBs(datasource, schema)
			prepareDBs(id, dbh, *dropSchema, *createSchema, *truncate)
			if *verbose {
				vlog.Printf("opened new DB (%s)", datasource)
			}
			readyDBs <- dbh
		}(id)
	}

	for i := 0; i < *poolSize; i++ {
		go func() {
			for dbh := range doneDBs {
				if *verbose {
					vlog.Println("(background) done with DB; truncating it and prepping for reuse")
				}
				if *truncate {
					start := time.Now()
					if *verbose {
						vlog.Println("(background) Truncating all tables...")
					}
					if err := dbh.TruncateAllTables(); err != nil {
						log.Fatal("testdb: truncate all tables:", err)
					}
					if *verbose {
						vlog.Println("(background) Truncated all tables in ", time.Since(start))
					}
				}
				readyDBs <- dbh
			}
		}()
	}
}

var (
	backgroundDBCreatorStarted bool
	backgroundDBCreatorSchema  *dbutil2.Schema // only 1 schema may be used
	backgroundDBCreatorLock    sync.Mutex      // protects backgroundDBCreatorStarted

	readyDBs chan *dbutil2.Handle
	doneDBs  chan *dbutil2.Handle
)
