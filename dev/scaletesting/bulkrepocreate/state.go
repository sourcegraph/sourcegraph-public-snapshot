package main

import (
	"database/sql"
	"fmt"
	"sync"
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
  git_url STRING DEFAULT ""
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

type repo struct {
	Name    string
	Failed  string
	Created bool
	Pushed  bool
	GitURL  string

	blank *blankRepo
}

func (s *state) load() ([]*repo, error) {
	s.Lock()
	defer s.Unlock()
	rows, err := s.db.Query(`SELECT name, failed, created, pushed, git_url FROM repos`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var repos []*repo
	for rows.Next() {
		var r repo
		err := rows.Scan(&r.Name, &r.Failed, &r.Created, &r.Pushed, &r.GitURL)
		if err != nil {
			return nil, err
		}
		repos = append(repos, &r)
	}
	return repos, nil
}

func generateNames(prefix string, count int) []string {
	names := make([]string, count)
	for i := 0; i < count; i++ {
		names[i] = fmt.Sprintf("%s%09d", prefix, i)
	}
	return names
}

func (s *state) generate(cfg config) ([]*repo, error) {
	names := generateNames(cfg.prefix, cfg.count)
	if err := s.insert(names); err != nil {
		return nil, err
	}
	return s.load()
}

var saveRepoStmt = `UPDATE repos SET 
failed = ?, 
created = ?,
pushed = ?,
git_url = ?
WHERE name = ?`

func (s *state) saveRepo(r *repo) error {
	s.Lock()
	defer s.Unlock()

	_, err := s.db.Exec(
		saveRepoStmt,
		r.Failed,
		r.Created,
		r.Pushed,
		r.GitURL,
		r.Name)

	if err != nil {
		return err
	}
	return nil
}

func (s *state) insert(names []string) error {
	s.Lock()
	defer s.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	for _, name := range names {
		if _, err := tx.Exec(`INSERT INTO repos(name) VALUES (?)`, name); err != nil {
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
