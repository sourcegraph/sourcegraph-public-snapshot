package main

import (
	"database/sql"
	"sync"

	_ "github.com/mattn/go-sqlite3"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

// feederDB is a front to a sqlite DB that records ownerRepo processed, orgs created and whether
// processing was successful or failed
type feederDB struct {
	// sqlite is not thread-safe, this mutex protects access to it
	sync.Mutex
	// where the DB file is
	path string
	// the opened DB
	db *sql.DB
}

// newFeederDB creates or opens the DB, creating the two tables if necessary
func newFeederDB(path string) (*feederDB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	stmt, err := db.Prepare("CREATE TABLE IF NOT EXISTS repos (ownerRepo STRING PRIMARY KEY, org STRING, failed BOOLEAN, errType STRING, UNIQUE(ownerRepo, failed))")
	if err != nil {
		return nil, err
	}
	_, err = stmt.Exec()
	if err != nil {
		return nil, err
	}

	stmt, err = db.Prepare("CREATE TABLE IF NOT EXISTS orgs (name STRING PRIMARY KEY)")
	if err != nil {
		return nil, err
	}
	_, err = stmt.Exec()
	if err != nil {
		return nil, err
	}

	return &feederDB{
		path: path,
		db:   db,
	}, nil
}

// declareRepo adds the ownerRepo to the DB when it gets pumped into the pipe and made available to the workers
// for processing. if ownerRepo was already done in a previous run, then returns true, so pump can skip it.
func (fdr *feederDB) declareRepo(ownerRepo string) (alreadyDone bool, err error) {
	fdr.Lock()
	defer fdr.Unlock()

	var failed bool
	var errType string

	err = fdr.db.QueryRow("SELECT failed, errType FROM repos WHERE ownerRepo=?", ownerRepo).Scan(&failed,
		&dbutil.NullString{S: &errType})
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}

	if err == sql.ErrNoRows {
		stmt, err := fdr.db.Prepare("INSERT INTO repos(ownerRepo, failed) VALUES(?, FALSE)")
		if err != nil {
			return false, err
		}

		_, err = stmt.Exec(ownerRepo)
		if err != nil {
			return false, err
		}

		return false, nil
	}

	alreadyDone = !failed || (failed && errType == "clone")
	return alreadyDone, nil
}

// failed records the fact that the worker processing the specified ownerRepo failed to process it.
// errType is recorded because specific errTypes are not worth rerunning in a subsequent run (for example if repo is private
// on github.com and we don't have credentials for it, it's not worth trying again in a next run).
func (fdr *feederDB) failed(ownerRepo string, errType string) error {
	fdr.Lock()
	defer fdr.Unlock()

	stmt, err := fdr.db.Prepare("UPDATE repos SET failed = TRUE, errType = ?  WHERE ownerRepo = ?")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(errType, ownerRepo)
	if err != nil {
		return err
	}

	return nil
}

// succeeded records that a worker has successfully processed the specified ownerRepo.
func (fdr *feederDB) succeeded(ownerRepo string, org string) error {
	fdr.Lock()
	defer fdr.Unlock()

	stmt, err := fdr.db.Prepare("UPDATE repos SET failed = FALSE, org = ? WHERE ownerRepo = ?")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(org, ownerRepo)
	if err != nil {
		return err
	}

	return nil
}

// declareOrg adds a newly created org from one of the workers.
func (fdr *feederDB) declareOrg(org string) error {
	fdr.Lock()
	defer fdr.Unlock()

	stmt, err := fdr.db.Prepare("INSERT OR IGNORE INTO orgs(name) VALUES(?)")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(org)
	if err != nil {
		return err
	}

	return nil
}
