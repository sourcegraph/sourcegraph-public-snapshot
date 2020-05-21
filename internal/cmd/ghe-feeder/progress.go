package main

import (
	"database/sql"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

type feederDB struct {
	sync.Mutex
	path string
	db   *sql.DB
}

func newFeederDB(path string) (*feederDB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	stmt, err := db.Prepare("CREATE TABLE IF NOT EXISTS repos (ownerRepo STRING PRIMARY KEY, failed BOOLEAN, UNIQUE(ownerRepo, failed))")
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

func (fdr *feederDB) declare(ownerRepo string) (bool, error) {
	fdr.Lock()
	defer fdr.Unlock()

	stmt, err := fdr.db.Prepare("INSERT OR IGNORE INTO repos(ownerRepo, failed) VALUES(?, ?)")
	if err != nil {
		return false, err
	}

	res, err := stmt.Exec(ownerRepo, false)
	if err != nil {
		return false, err
	}

	ra, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	return ra == 0, nil
}

func (fdr *feederDB) failed(ownerRepo string) error {
	fdr.Lock()
	defer fdr.Unlock()

	stmt, err := fdr.db.Prepare("UPDATE repos SET failed = TRUE WHERE ownerRepo = ?")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(ownerRepo)
	if err != nil {
		return err
	}

	return nil
}
