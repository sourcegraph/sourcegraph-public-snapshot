package pgsql

import (
	"crypto/subtle"
	"errors"
	"strings"
	"time"

	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/util/randstring"
)

func init() {
	Schema.Map.AddTableWithName(dbInvites{}, "invites").SetKeys(false, "email")
	Schema.CreateSQL = append(Schema.CreateSQL,
		`ALTER TABLE invites ALTER COLUMN created_at TYPE timestamp with time zone USING created_at::timestamp with time zone;`,
		"CREATE UNIQUE INDEX invites_token ON invites(token)",
	)
}

// dbInvites DB-maps an account invite and related metadata.
type dbInvites struct {
	Email     string
	Token     string
	Write     bool
	Admin     bool
	InUse     bool      `db:"in_use"`
	CreatedAt time.Time `db:"created_at"`
}

func toInvite(d *dbInvites) *sourcegraph.AccountInvite {
	return &sourcegraph.AccountInvite{
		Email: d.Email,
		Write: d.Write,
		Admin: d.Admin,
	}
}

// Authorizations is a FS-backed implementation of the Authorizations store.
type Invites struct{}

var _ store.Invites = (*Invites)(nil)

func (s *Invites) CreateOrUpdate(ctx context.Context, invite *sourcegraph.AccountInvite) (string, error) {
	dbInvite := &dbInvites{
		Email:     invite.Email,
		Token:     randstring.NewLen(20),
		Write:     invite.Write,
		Admin:     invite.Admin,
		CreatedAt: time.Now(),
	}

	err := dbh(ctx).Insert(dbInvite)
	if err != nil && strings.Contains(err.Error(), `duplicate key value violates unique constraint`) {
		_, err = dbh(ctx).Update(dbInvite)
	}
	if err != nil {
		return "", err
	}
	return dbInvite.Token, nil
}

func (s *Invites) Retrieve(ctx context.Context, token string) (*sourcegraph.AccountInvite, error) {
	dbInvite, err := s.get(ctx, token)
	if err != nil {
		return nil, err
	}
	if dbInvite.InUse {
		return nil, errors.New("already used")
	}

	dbInvite.InUse = true
	if _, err = dbh(ctx).Update(dbInvite); err != nil {
		return nil, err
	}

	return toInvite(dbInvite), nil
}

func (s *Invites) MarkUnused(ctx context.Context, token string) error {
	dbInvite, err := s.get(ctx, token)
	if err != nil {
		return err
	}

	if !dbInvite.InUse {
		return nil
	}

	dbInvite.InUse = false

	if _, err := dbh(ctx).Update(dbInvite); err != nil {
		return err
	}
	return nil
}

func (s *Invites) get(ctx context.Context, token string) (*dbInvites, error) {
	var invites []*dbInvites
	err := dbh(ctx).Select(&invites, `SELECT * FROM invites;`)
	if err != nil {
		return nil, err
	}
	// Constant time comparison to prevent timing attacks.
	for i := range invites {
		if subtle.ConstantTimeCompare([]byte(token), []byte(invites[i].Token)) == 1 {
			return invites[i], nil
		}
	}
	return nil, errors.New("not found")
}

func (s *Invites) Delete(ctx context.Context, token string) error {
	_, err := dbh(ctx).Exec(`DELETE FROM invites WHERE "token" = $1;`, token)
	return err
}

func (s *Invites) DeleteByEmail(ctx context.Context, email string) error {
	res, err := dbh(ctx).Exec(`DELETE FROM invites WHERE "email" = $1;`, email)
	if n, err := res.RowsAffected(); err != nil {
		return err
	} else if n == 0 {
		return errors.New("not found")
	}
	return err
}

func (s *Invites) List(ctx context.Context) ([]*sourcegraph.AccountInvite, error) {
	var invites []*dbInvites
	err := dbh(ctx).Select(&invites, `SELECT * FROM invites;`)
	if err != nil {
		return nil, err
	}
	accountInvites := make([]*sourcegraph.AccountInvite, 0)
	for _, invite := range invites {
		accountInvites = append(accountInvites, toInvite(invite))
	}
	return accountInvites, nil
}
