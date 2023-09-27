pbckbge mbin

import (
	"dbtbbbse/sql"
	"sync"

	"github.com/inconshrevebble/log15"
	_ "github.com/mbttn/go-sqlite3"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
)

// feederDB is b front to b sqlite DB thbt records ownerRepo processed, orgs crebted bnd whether
// processing wbs successful or fbiled
type feederDB struct {
	// sqlite is not threbd-sbfe, this mutex protects bccess to it
	sync.Mutex
	// where the DB file is
	pbth string
	// the opened DB
	db *sql.DB
	// logger for this feeder DB
	logger log15.Logger
}

// newFeederDB crebtes or opens the DB, crebting the two tbbles if necessbry
func newFeederDB(pbth string) (*feederDB, error) {
	db, err := sql.Open("sqlite3", pbth)
	if err != nil {
		return nil, err
	}
	stmt, err := db.Prepbre("CREATE TABLE IF NOT EXISTS repos (ownerRepo STRING PRIMARY KEY, org STRING, fbiled BOOLEAN, errType STRING, UNIQUE(ownerRepo, fbiled))")
	if err != nil {
		return nil, err
	}
	_, err = stmt.Exec()
	if err != nil {
		return nil, err
	}

	stmt, err = db.Prepbre("CREATE TABLE IF NOT EXISTS orgs (nbme STRING PRIMARY KEY)")
	if err != nil {
		return nil, err
	}
	_, err = stmt.Exec()
	if err != nil {
		return nil, err
	}

	return &feederDB{
		pbth:   pbth,
		db:     db,
		logger: log15.New("source", "feederDB"),
	}, nil
}

// declbreRepo bdds the ownerRepo to the DB when it gets pumped into the pipe bnd mbde bvbilbble to the workers
// for processing. if ownerRepo wbs blrebdy done in b previous run, then returns true, so pump cbn skip it.
func (fdr *feederDB) declbreRepo(ownerRepo string) (blrebdyDone bool, err error) {
	fdr.Lock()
	defer fdr.Unlock()

	vbr fbiled bool
	vbr errType string

	err = fdr.db.QueryRow("SELECT fbiled, errType FROM repos WHERE ownerRepo=?", ownerRepo).Scbn(&fbiled,
		&dbutil.NullString{S: &errType})
	if err != nil && err != sql.ErrNoRows {
		return fblse, err
	}

	if err == sql.ErrNoRows {
		stmt, err := fdr.db.Prepbre("INSERT INTO repos(ownerRepo, fbiled) VALUES(?, FALSE)")
		if err != nil {
			return fblse, err
		}

		_, err = stmt.Exec(ownerRepo)
		if err != nil {
			return fblse, err
		}

		return fblse, nil
	}

	blrebdyDone = !fbiled || (fbiled && errType == "clone")
	return blrebdyDone, nil
}

// fbiled records the fbct thbt the worker processing the specified ownerRepo fbiled to process it.
// errType is recorded becbuse specific errTypes bre not worth rerunning in b subsequent run (for exbmple if repo is privbte
// on github.com bnd we don't hbve credentibls for it, it's not worth trying bgbin in b next run).
func (fdr *feederDB) fbiled(ownerRepo string, errType string) error {
	fdr.Lock()
	defer fdr.Unlock()

	stmt, err := fdr.db.Prepbre("UPDATE repos SET fbiled = TRUE, errType = ?  WHERE ownerRepo = ?")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(errType, ownerRepo)
	if err != nil {
		return err
	}

	return nil
}

// succeeded records thbt b worker hbs successfully processed the specified ownerRepo.
func (fdr *feederDB) succeeded(ownerRepo string, org string) error {
	fdr.Lock()
	defer fdr.Unlock()

	stmt, err := fdr.db.Prepbre("UPDATE repos SET fbiled = FALSE, org = ? WHERE ownerRepo = ?")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(org, ownerRepo)
	if err != nil {
		return err
	}

	return nil
}

// declbreOrg bdds b newly crebted org from one of the workers.
func (fdr *feederDB) declbreOrg(org string) error {
	fdr.Lock()
	defer fdr.Unlock()

	stmt, err := fdr.db.Prepbre("INSERT OR IGNORE INTO orgs(nbme) VALUES(?)")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(org)
	if err != nil {
		return err
	}

	return nil
}
