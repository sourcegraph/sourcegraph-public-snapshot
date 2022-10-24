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

var createTableStmt = `CREATE TABLE IF NOT EXISTS users (
  login STRING PRIMARY KEY,
  email STRING,
  failed STRING DEFAULT "",
  created BOOLEAN DEFAULT FALSE
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

type user struct {
	Login   string
	Email   string
	Failed  string
	Created bool
}

func (s *state) load() ([]*user, error) {
	s.Lock()
	defer s.Unlock()
	rows, err := s.db.Query(`SELECT login, email, failed, created FROM users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*user
	for rows.Next() {
		var u user
		err := rows.Scan(&u.Login, &u.Email, &u.Failed, &u.Created)
		if err != nil {
			return nil, err
		}
		users = append(users, &u)
	}
	return users, nil
}

func generateNames(count int) []string {
	names := make([]string, count)
	for i := 0; i < count; i++ {
		names[i] = fmt.Sprintf("user-%09d", i)
	}
	return names
}

func (s *state) generate(cfg config) ([]*user, error) {
	names := generateNames(cfg.count)
	if err := s.insert(names); err != nil {
		return nil, err
	}
	return s.load()
}

var saveUserStmt = `UPDATE users SET
failed = ?,
created = ?
WHERE login = ?`

func (s *state) saveUser(u *user) error {
	s.Lock()
	defer s.Unlock()

	_, err := s.db.Exec(
		saveUserStmt,
		u.Failed,
		u.Created,
		u.Login)

	if err != nil {
		return err
	}
	return nil
}

func (s *state) insert(logins []string) error {
	s.Lock()
	defer s.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	for _, login := range logins {
		if _, err := tx.Exec(`INSERT INTO users(login, email) VALUES (?, ?)`, login, fmt.Sprintf("%s@%s", login, emailDomain)); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *state) countCompletedUsers() (int, error) {
	s.Lock()
	defer s.Unlock()

	row := s.db.QueryRow(`SELECT COUNT(login) FROM users WHERE created = TRUE AND failed == ""`)
	var count int
	err := row.Scan(&count)
	return count, err
}

func (s *state) countAllUsers() (int, error) {
	s.Lock()
	defer s.Unlock()

	row := s.db.QueryRow(`SELECT COUNT(login) FROM users`)
	var count int
	err := row.Scan(&count)
	return count, err
}
