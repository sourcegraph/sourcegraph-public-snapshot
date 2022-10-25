package main

import (
	"database/sql"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

type state struct {
	// sqlite is not thread-safe, this mutex protects access to it
	sync.Mutex
	// where the DB file is
	path string
	// the opened DB
	db *sql.DB
}

var createTableStmt = `CREATE TABLE IF NOT EXISTS repos (
  name STRING PRIMARY KEY,
  failed STRING DEFAULT "",
  created BOOLEAN DEFAULT FALSE,
  pushed BOOLEAN DEFAULT FALSE,
  from_git_url STRING DEFAULT "",
  to_git_url STRING DEFAULT ""
)`

func newState(path string) (*state, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	stmt, err := db.Prepare(createTableStmt)
	if err != nil {
		return nil, err
	}
	_, err = stmt.Exec()
	if err != nil {
		return nil, err
	}

	return &state{
		path: path,
		db:   db,
	}, nil
}

func (s *state) load() ([]*Repo, error) {
	s.Lock()
	defer s.Unlock()
	rows, err := s.db.Query(`SELECT name, failed, created, pushed, from_git_url, to_git_url FROM repos`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var repos []*Repo
	for rows.Next() {
		var r Repo
		err := rows.Scan(&r.Name, &r.Failed, &r.Created, &r.Pushed, &r.FromGitURL, &r.ToGitURL)
		if err != nil {
			return nil, err
		}
		repos = append(repos, &r)
	}
	return repos, nil
}

var saveRepoStmt = `UPDATE repos SET 
failed = ?, 
created = ?,
pushed = ?,
from_git_url = ?,
to_git_url = ?

WHERE name = ?`

func (s *state) saveRepo(r *Repo) error {
	s.Lock()
	defer s.Unlock()

	_, err := s.db.Exec(
		saveRepoStmt,
		r.Failed,
		r.Created,
		r.Pushed,
		r.FromGitURL,
		r.ToGitURL,
		r.Name)

	if err != nil {
		return err
	}
	return nil
}

func (s *state) insert(repos []*Repo) error {
	s.Lock()
	defer s.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	for _, r := range repos {
		if _, err := tx.Exec(`INSERT INTO repos(name, from_git_url) VALUES (?, ?)`, r.Name, r.FromGitURL); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *state) countCompletedRepos() (int, error) {
	s.Lock()
	defer s.Unlock()

	row := s.db.QueryRow(`SELECT COUNT(name) FROM repos WHERE created = TRUE AND pushed = TRUE AND failed == ""`)
	var count int
	err := row.Scan(&count)
	return count, err
}

func (s *state) countAllRepos() (int, error) {
	s.Lock()
	defer s.Unlock()

	row := s.db.QueryRow(`SELECT COUNT(name) FROM repos`)
	var count int
	err := row.Scan(&count)
	return count, err
}
