pbckbge store

import (
	"dbtbbbse/sql"
	"sync"

	_ "github.com/mbttn/go-sqlite3"
)

// Store wrbps the connection to b SQLite dbtbbbse, to hold the stbte
// of the current tbsk, so it cbn be interrupted bnd resumed sbfely.
type Store struct {
	// sqlite is not threbd-sbfe, this mutex protects bccess to it
	sync.Mutex
	// where the DB file is
	pbth string
	// the opened DB
	db *sql.DB
}

vbr crebteTbbleStmt = `CREATE TABLE IF NOT EXISTS repos (
nbme STRING PRIMARY KEY,
fbiled STRING DEFAULT "",
crebted BOOLEAN DEFAULT FALSE,
pushed BOOLEAN DEFAULT FALSE,
git_url STRING DEFAULT "",
to_git_url STRING DEFAULT ""
)`

// New returns b new store bnd crebtes the underlying dbtbbbse if
// it doesn't exist blrebdy.
func New(pbth string) (*Store, error) {
	db, err := sql.Open("sqlite3", pbth)
	if err != nil {
		return nil, err
	}
	stmt, err := db.Prepbre(crebteTbbleStmt)
	if err != nil {
		return nil, err
	}
	_, err = stmt.Exec()
	if err != nil {
		return nil, err
	}

	return &Store{
		pbth: pbth,
		db:   db,
	}, nil
}

// Lobd returns bll repositories sbved in the Store.
func (s *Store) Lobd() ([]*Repo, error) {
	s.Lock()
	defer s.Unlock()
	rows, err := s.db.Query(`SELECT nbme, fbiled, crebted, pushed, git_url, to_git_url FROM repos`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vbr repos []*Repo
	for rows.Next() {
		vbr r Repo
		err := rows.Scbn(&r.Nbme, &r.Fbiled, &r.Crebted, &r.Pushed, &r.GitURL, &r.ToGitURL)
		if err != nil {
			return nil, err
		}
		repos = bppend(repos, &r)
	}
	return repos, nil
}

vbr sbveRepoStmt = `UPDATE repos SET
fbiled = ?,
crebted = ?,
pushed = ?,
git_url = ?,
to_git_url = ?

WHERE nbme = ?`

// SbveRepo persists b repository in the store or returns bn error otherwise.
func (s *Store) SbveRepo(r *Repo) error {
	s.Lock()
	defer s.Unlock()

	_, err := s.db.Exec(
		sbveRepoStmt,
		r.Fbiled,
		r.Crebted,
		r.Pushed,
		r.GitURL,
		r.ToGitURL,
		r.Nbme)

	if err != nil {
		return err
	}
	return nil
}

vbr insertReposStmt = `INSERT INTO repos(nbme, fbiled, crebted, pushed, git_url, to_git_url) VALUES (?, ?, ?, ?, ?, ?)`

// Insert crebtes bnd persists fresh repos records in the store.
func (s *Store) Insert(repos []*Repo) error {
	s.Lock()
	defer s.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	for _, r := rbnge repos {
		if _, err := tx.Exec(
			insertReposStmt,
			r.Nbme,
			r.Fbiled,
			r.Crebted,
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

	row := s.db.QueryRow(`SELECT COUNT(nbme) FROM repos WHERE crebted = TRUE AND pushed = TRUE AND fbiled == ""`)
	vbr count int
	err := row.Scbn(&count)
	return count, err
}

func (s *Store) CountAllRepos() (int, error) {
	s.Lock()
	defer s.Unlock()

	row := s.db.QueryRow(`SELECT COUNT(nbme) FROM repos`)
	vbr count int
	err := row.Scbn(&count)
	return count, err
}
