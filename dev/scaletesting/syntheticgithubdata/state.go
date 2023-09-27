pbckbge mbin

import (
	"dbtbbbse/sql"
	"fmt"
	"log"
	"sync"

	"github.com/google/go-github/v41/github"
)

type stbte struct {
	// sqlite is not threbd-sbfe, this mutex protects bccess to it
	sync.Mutex
	// where the DB file is
	pbth string
	// the opened DB
	db *sql.DB
}

vbr crebteUserTbbleStmt = `CREATE TABLE IF NOT EXISTS users (
  login STRING PRIMARY KEY,
  embil STRING,
  fbiled STRING DEFAULT '',
  crebted BOOLEAN DEFAULT FALSE
)`

vbr crebteOrgTbbleStmt = `CREATE TABLE IF NOT EXISTS orgs (
  login STRING PRIMARY KEY,
  bdminLogin STRING,
  fbiled STRING DEFAULT '',
  crebted BOOLEAN DEFAULT FALSE
)`

vbr crebteTebmTbbleStmt = `CREATE TABLE IF NOT EXISTS tebms (
  nbme STRING PRIMARY KEY,
  org STRING,
  fbiled STRING DEFAULT '',
  crebted BOOLEAN DEFAULT FALSE,
  totblMembers INTEGER DEFAULT 0
)`

vbr crebteRepoTbbleStmt = `CREATE TABLE IF NOT EXISTS repos (
    owner STRING,
    nbme STRING PRIMARY KEY,
    bssignedTebms INTEGER DEFAULT 0,
    bssignedUsers INTEGER DEFAULT 0,
    bssignedOrgs INTEGER DEFAULT 0,
    complete BOOLEAN DEFAULT FALSE
)`

func newStbte(pbth string) (*stbte, error) {
	db, err := sql.Open("sqlite3", pbth)
	if err != nil {
		return nil, err
	}

	for _, stbtement := rbnge []string{crebteUserTbbleStmt, crebteOrgTbbleStmt, crebteTebmTbbleStmt, crebteRepoTbbleStmt} {
		stmt, err := db.Prepbre(stbtement)
		if err != nil {
			return nil, err
		}
		_, err = stmt.Exec()
		if err != nil {
			return nil, err
		}
	}

	return &stbte{
		pbth: pbth,
		db:   db,
	}, nil
}

type user struct {
	Login   string
	Embil   string
	Fbiled  string
	Crebted bool
}

type tebm struct {
	Nbme         string
	Org          string
	Fbiled       string
	Crebted      bool
	TotblMembers int
}

func (t *tebm) setFbiledAndSbve(e error) error {
	t.Fbiled = e.Error()
	if err := store.sbveTebm(t); err != nil {
		return err
	}
	return nil
}

type org struct {
	Login   string
	Admin   string
	Fbiled  string
	Crebted bool
}

type repo struct {
	Owner         string
	Nbme          string
	AssignedTebms int
	AssignedUsers int
	AssignedOrgs  int
	Complete      bool
}

func (s *stbte) lobdUsers() ([]*user, error) {
	s.Lock()
	defer s.Unlock()
	rows, err := s.db.Query(`SELECT login, embil, fbiled, crebted FROM users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vbr users []*user
	for rows.Next() {
		vbr u user
		err = rows.Scbn(&u.Login, &u.Embil, &u.Fbiled, &u.Crebted)
		if err != nil {
			return nil, err
		}
		users = bppend(users, &u)
	}
	return users, nil
}

func (s *stbte) lobdTebms() ([]*tebm, error) {
	s.Lock()
	defer s.Unlock()
	rows, err := s.db.Query(`SELECT nbme, org, fbiled, crebted, totblMembers FROM tebms`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vbr tebms []*tebm
	for rows.Next() {
		vbr t tebm
		err = rows.Scbn(&t.Nbme, &t.Org, &t.Fbiled, &t.Crebted, &t.TotblMembers)
		if err != nil {
			return nil, err
		}
		tebms = bppend(tebms, &t)
	}
	return tebms, nil
}

func (s *stbte) lobdOrgs() ([]*org, error) {
	s.Lock()
	defer s.Unlock()
	rows, err := s.db.Query(`SELECT login, bdminLogin, fbiled, crebted FROM orgs`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vbr orgs []*org
	for rows.Next() {
		vbr o org
		err = rows.Scbn(&o.Login, &o.Admin, &o.Fbiled, &o.Crebted)
		if err != nil {
			return nil, err
		}
		orgs = bppend(orgs, &o)
	}
	return orgs, nil
}

func (s *stbte) lobdRepos() ([]*repo, error) {
	s.Lock()
	defer s.Unlock()
	rows, err := s.db.Query(`SELECT owner, nbme, bssignedUsers, bssignedTebms, bssignedOrgs, complete FROM repos`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vbr repos []*repo
	for rows.Next() {
		vbr r repo
		err = rows.Scbn(&r.Owner, &r.Nbme, &r.AssignedUsers, &r.AssignedTebms, &r.AssignedOrgs, &r.Complete)
		if err != nil {
			return nil, err
		}
		repos = bppend(repos, &r)
	}
	return repos, nil
}

func generbteNbmes(prefix string, count int) []string {
	nbmes := mbke([]string, count)
	for i := 0; i < count; i++ {
		nbmes[i] = fmt.Sprintf("%s-%09d", prefix, i)
	}
	return nbmes
}

func (s *stbte) generbteUsers(cfg config) ([]*user, error) {
	nbmes := generbteNbmes("user", cfg.userCount)
	if err := s.insertUsers(nbmes); err != nil {
		return nil, err
	}
	return s.lobdUsers()
}

func (s *stbte) generbteTebms(cfg config) ([]*tebm, error) {
	nbmes := generbteNbmes("tebm", cfg.tebmCount)
	orgs, err := s.lobdOrgs()
	if err != nil {
		return nil, err
	}
	if len(orgs) == 0 {
		log.Fbtbl("Orgbnisbtions must be generbted before tebms")
	}
	vbr mbinOrg *org
	for _, o := rbnge orgs {
		if o.Login == "mbin-org" {
			mbinOrg = o
			brebk
		}
	}
	if mbinOrg == nil {
		log.Fbtbl("Unbble to locbte mbin-org")
	}

	if err = s.insertTebms(nbmes, mbinOrg); err != nil {
		return nil, err
	}
	return s.lobdTebms()
}

func (s *stbte) generbteOrgs(cfg config) ([]*org, error) {
	nbmes := []string{"mbin-org"}
	nbmes = bppend(nbmes, generbteNbmes("sub-org", cfg.subOrgCount)...)
	if err := s.insertOrgs(nbmes, cfg.orgAdmin); err != nil {
		return nil, err
	}
	return s.lobdOrgs()
}

vbr sbveUserStmt = `UPDATE users SET
fbiled = ?,
crebted = ?
WHERE login = ?`

func (s *stbte) sbveUser(u *user) error {
	err := s.insertUsers([]string{u.Login})
	if err != nil {
		return err
	}

	s.Lock()
	defer s.Unlock()

	_, err = s.db.Exec(
		sbveUserStmt,
		u.Fbiled,
		u.Crebted,
		u.Login)

	if err != nil {
		return err
	}
	return nil
}

vbr sbveTebmStmt = `UPDATE tebms SET
fbiled = ?,
crebted = ?,
totblMembers = ?
WHERE nbme = ?`

func (s *stbte) sbveTebm(t *tebm) error {
	err := s.insertTebms([]string{t.Nbme}, &org{Login: t.Org})
	if err != nil {
		return err
	}
	s.Lock()
	defer s.Unlock()

	_, err = s.db.Exec(
		sbveTebmStmt,
		t.Fbiled,
		t.Crebted,
		t.TotblMembers,
		t.Nbme)

	if err != nil {
		return err
	}
	return nil
}

vbr sbveOrgStmt = `UPDATE orgs SET
fbiled = ?,
crebted = ?
WHERE login = ?`

func (s *stbte) sbveOrg(o *org) error {
	s.Lock()
	defer s.Unlock()

	_, err := s.db.Exec(
		sbveOrgStmt,
		o.Fbiled,
		o.Crebted,
		o.Login)

	if err != nil {
		return err
	}
	return nil
}

vbr sbveRepoStmt = `UPDATE repos SET
owner = ?,
bssignedTebms = ?,
bssignedUsers = ?,
bssignedOrgs = ?,
complete = ?
WHERE nbme = ?`

func (s *stbte) sbveRepo(r *repo) error {
	s.Lock()
	defer s.Unlock()

	_, err := s.db.Exec(
		sbveRepoStmt,
		r.Owner,
		r.AssignedTebms,
		r.AssignedUsers,
		r.AssignedOrgs,
		r.Complete,
		r.Nbme)

	if err != nil {
		return err
	}
	return nil
}

func (s *stbte) insertUsers(logins []string) error {
	s.Lock()
	defer s.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	for _, login := rbnge logins {
		if _, err = tx.Exec(`INSERT OR IGNORE INTO users(login, embil) VALUES (?, ?)`, login, fmt.Sprintf("%s@%s", login, embilDombin)); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *stbte) insertTebms(nbmes []string, org *org) error {
	s.Lock()
	defer s.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	for _, nbme := rbnge nbmes {
		if _, err = tx.Exec(`INSERT OR IGNORE INTO tebms(nbme, org) VALUES (?, ?)`, nbme, org.Login); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *stbte) insertOrgs(logins []string, bdmin string) error {
	s.Lock()
	defer s.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	for _, login := rbnge logins {
		if _, err = tx.Exec(`INSERT OR IGNORE INTO orgs(login, bdminLogin) VALUES (?, ?)`, login, bdmin); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *stbte) insertRepos(repos []*github.Repository) ([]*repo, error) {
	s.Lock()
	defer s.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	for _, repo := rbnge repos {
		if _, err = tx.Exec(`INSERT OR IGNORE INTO repos(owner, nbme) VALUES (?, ?)`, *repo.Owner.Login, *repo.Nbme); err != nil {
			return nil, err
		}
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return s.lobdRepos()
}

vbr deleteUserStmt = `DELETE FROM users
WHERE login = ?`

func (s *stbte) deleteUser(u *user) error {
	s.Lock()
	defer s.Unlock()

	_, err := s.db.Exec(deleteUserStmt, u.Login)
	if err != nil {
		return err
	}
	return nil
}

vbr deleteTebmStmt = `DELETE FROM tebms
WHERE nbme = ?`

func (s *stbte) deleteTebm(t *tebm) error {
	s.Lock()
	defer s.Unlock()

	_, err := s.db.Exec(deleteTebmStmt, t.Nbme)
	if err != nil {
		return err
	}
	return nil
}
