package main

import (
	"database/sql"
	"fmt"
	"log"
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

var createUserTableStmt = `CREATE TABLE IF NOT EXISTS users (
  login STRING PRIMARY KEY,
  email STRING,
  failed STRING DEFAULT "",
  created BOOLEAN DEFAULT FALSE
)`

var createOrgTableStmt = `CREATE TABLE IF NOT EXISTS orgs (
  login STRING PRIMARY KEY,
  adminLogin STRING,
  failed STRING DEFAULT "",
  created BOOLEAN DEFAULT FALSE
)`

var createTeamTableStmt = `CREATE TABLE IF NOT EXISTS teams (
  name STRING PRIMARY KEY,
  org STRING,
  failed STRING DEFAULT "",
  created BOOLEAN DEFAULT FALSE,
  hasMembers BOOLEAN DEFAULT FALSE
)`

func newState(path string) (*state, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	for _, statement := range []string{createUserTableStmt, createOrgTableStmt, createTeamTableStmt} {
		stmt, err := db.Prepare(statement)
		if err != nil {
			return nil, err
		}
		_, err = stmt.Exec()
		if err != nil {
			return nil, err
		}
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

type team struct {
	Name       string
	Org        string
	Failed     string
	Created    bool
	HasMembers bool
}

type org struct {
	Login   string
	Admin   string
	Failed  string
	Created bool
}

func (s *state) loadUsers() ([]*user, error) {
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
		err = rows.Scan(&u.Login, &u.Email, &u.Failed, &u.Created)
		if err != nil {
			return nil, err
		}
		users = append(users, &u)
	}
	return users, nil
}

func (s *state) getRandomUsers(limit int) ([]string, error) {
	s.Lock()
	defer s.Unlock()
	rows, err := s.db.Query(fmt.Sprintf("SELECT login FROM users ORDER BY RANDOM() LIMIT %d", limit))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userLogins []string
	for rows.Next() {
		var uLogin string
		err = rows.Scan(&uLogin)
		if err != nil {
			return nil, err
		}
		userLogins = append(userLogins, uLogin)
	}
	return userLogins, nil
}

func (s *state) loadTeams() ([]*team, error) {
	s.Lock()
	defer s.Unlock()
	rows, err := s.db.Query(`SELECT name, org, failed, created, hasMembers FROM teams`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []*team
	for rows.Next() {
		var t team
		err = rows.Scan(&t.Name, &t.Org, &t.Failed, &t.Created, &t.HasMembers)
		if err != nil {
			return nil, err
		}
		teams = append(teams, &t)
	}
	return teams, nil
}

func (s *state) loadOrgs() ([]*org, error) {
	s.Lock()
	defer s.Unlock()
	rows, err := s.db.Query(`SELECT login, adminLogin, failed, created FROM orgs`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []*org
	for rows.Next() {
		var o org
		err = rows.Scan(&o.Login, &o.Admin, &o.Failed, &o.Created)
		if err != nil {
			return nil, err
		}
		orgs = append(orgs, &o)
	}
	return orgs, nil
}

func generateNames(prefix string, count int) []string {
	names := make([]string, count)
	for i := 0; i < count; i++ {
		names[i] = fmt.Sprintf("%s-%09d", prefix, i)
	}
	return names
}

func (s *state) generateUsers(cfg config) ([]*user, error) {
	names := generateNames("user", cfg.userCount)
	if err := s.insertUsers(names); err != nil {
		return nil, err
	}
	return s.loadUsers()
}

func (s *state) generateTeams(cfg config) ([]*team, error) {
	names := generateNames("team", cfg.teamCount)
	orgs, err := s.loadOrgs()
	if err != nil {
		return nil, err
	}
	if len(orgs) == 0 {
		log.Fatal("Organisations must be generated before teams")
	}
	if err = s.insertTeams(names, orgs); err != nil {
		return nil, err
	}
	return s.loadTeams()
}

func (s *state) generateOrgs(cfg config) ([]*org, error) {
	names := generateNames("org", cfg.orgCount)
	if err := s.insertOrgs(names, cfg.orgAdmin); err != nil {
		return nil, err
	}
	return s.loadOrgs()
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

var saveTeamStmt = `UPDATE teams SET
failed = ?,
created = ?,
hasMembers = ?,
WHERE name = ?`

func (s *state) saveTeam(t *team) error {
	s.Lock()
	defer s.Unlock()

	_, err := s.db.Exec(
		saveTeamStmt,
		t.Failed,
		t.Created,
		t.HasMembers,
		t.Name)

	if err != nil {
		return err
	}
	return nil
}

var saveOrgStmt = `UPDATE orgs SET
failed = ?,
created = ?
WHERE login = ?`

func (s *state) saveOrg(o *org) error {
	s.Lock()
	defer s.Unlock()

	_, err := s.db.Exec(
		saveOrgStmt,
		o.Failed,
		o.Created,
		o.Login)

	if err != nil {
		return err
	}
	return nil
}

func (s *state) insertUsers(logins []string) error {
	s.Lock()
	defer s.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	for _, login := range logins {
		if _, err = tx.Exec(`INSERT INTO users(login, email) VALUES (?, ?)`, login, fmt.Sprintf("%s@%s", login, emailDomain)); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *state) insertTeams(names []string, orgs []*org) error {
	s.Lock()
	defer s.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	for i, name := range names {
		// distribute teams evenly over orgs
		org := orgs[i%len(orgs)]
		if _, err = tx.Exec(`INSERT INTO teams(name, org) VALUES (?, ?)`, name, org.Login); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *state) insertOrgs(logins []string, admin string) error {
	s.Lock()
	defer s.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	for _, login := range logins {
		if _, err = tx.Exec(`INSERT INTO orgs(login, adminLogin) VALUES (?, ?)`, login, admin); err != nil {
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

func (s *state) countCompletedTeams() (int, error) {
	s.Lock()
	defer s.Unlock()

	row := s.db.QueryRow(`SELECT COUNT(name) FROM teams WHERE created = TRUE AND hasMembers = true AND failed == ""`)
	var count int
	err := row.Scan(&count)
	return count, err
}

func (s *state) countAllTeams() (int, error) {
	s.Lock()
	defer s.Unlock()

	row := s.db.QueryRow(`SELECT COUNT(name) FROM teams`)
	var count int
	err := row.Scan(&count)
	return count, err
}

func (s *state) countCompletedOrgs() (int, error) {
	s.Lock()
	defer s.Unlock()

	row := s.db.QueryRow(`SELECT COUNT(login) FROM orgs WHERE created = TRUE AND failed == ""`)
	var count int
	err := row.Scan(&count)
	return count, err
}

func (s *state) countAllOrgs() (int, error) {
	s.Lock()
	defer s.Unlock()

	row := s.db.QueryRow(`SELECT COUNT(login) FROM orgs`)
	var count int
	err := row.Scan(&count)
	return count, err
}
