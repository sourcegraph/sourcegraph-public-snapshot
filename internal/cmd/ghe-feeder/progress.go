package main

import (
	"database/sql"
	"sync"

	"github.com/inconshreveable/log15"
	_ "github.com/mattn/go-sqlite3"
)

type feederDB struct {
	sync.Mutex
	path   string
	db     *sql.DB
	logger log15.Logger
}

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
		path:   path,
		db:     db,
		logger: log15.New("source", "feederDB"),
	}, nil
}

func (fdr *feederDB) declareRepo(ownerRepo string) (alreadyDone bool, err error) {
	fdr.Lock()
	defer fdr.Unlock()

	var failed bool
	var errType string

	err = fdr.db.QueryRow("SELECT failed, errType FROM repos WHERE ownerRepo=?", ownerRepo).Scan(&failed, &errType)
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
	return
}

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
