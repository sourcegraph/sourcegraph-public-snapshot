package store

import (
	"database/sql"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

// Store wraps the connection to a SQLite database, to hold the state
// of the current task, so it can be interrupted and resumed safely.
type Store struct {
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
git_url STRING DEFAULT "",
to_git_url STRING DEFAULT ""
)`

// New returns a new store and creates the underlying database if
// it doesn't exist already.
func New(path string) (*Store, error) {
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

	return &Store{
		path: path,
		db:   db,
	}, nil
}

// Load returns all repositories saved in the Store.
func (s *Store) Load() ([]*Repo, error) {
	s.Lock()
	defer s.Unlock()
	rows, err := s.db.Query(`SELECT name, failed, created, pushed, git_url, to_git_url FROM repos`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var repos []*Repo
	for rows.Next() {
		var r Repo
		err := rows.Scan(&r.Name, &r.Failed, &r.Created, &r.Pushed, &r.GitURL, &r.ToGitURL)
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
git_url = ?,
to_git_url = ?

WHERE name = ?`

// SaveRepo persists a repository in the store or returns an error otherwise.
func (s *Store) SaveRepo(r *Repo) error {
	s.Lock()
	defer s.Unlock()

	_, err := s.db.Exec(
		saveRepoStmt,
		r.Failed,
		r.Created,
		r.Pushed,
		r.GitURL,
		r.ToGitURL,
		r.Name)

	if err != nil {
		return err
	}
	return nil
}

var insertReposStmt = `INSERT INTO repos(name, failed, created, pushed, git_url, to_git_url) VALUES (?, ?, ?, ?, ?, ?)`

// Insert creates and persists fresh repos records in the store.
func (s *Store) Insert(repos []*Repo) error {
	s.Lock()
	defer s.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	for _, r := range repos {
		if _, err := tx.Exec(
			insertReposStmt,
			r.Name,
			r.Failed,
			r.Created,
			r.Pushed,
			r.GitURL,
			r.ToGitURL,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) CountCompletedRepos() (int, error) {
	s.Lock()
	defer s.Unlock()

	row := s.db.QueryRow(`SELECT COUNT(name) FROM repos WHERE created = TRUE AND pushed = TRUE AND failed == ""`)
	var count int
	err := row.Scan(&count)
	return count, err
}

func (s *Store) CountAllRepos() (int, error) {
	s.Lock()
	defer s.Unlock()

	row := s.db.QueryRow(`SELECT COUNT(name) FROM repos`)
	var count int
	err := row.Scan(&count)
	return count, err
}
