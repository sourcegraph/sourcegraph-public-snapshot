package pgsql

import (
	"log"
	"time"

	"github.com/sqs/modl"
	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/util/dbutil"
	"src.sourcegraph.com/sourcegraph/util/randstring"
)

func init() {
	Schema.Map.AddTableWithName(dbAuthCode{}, "oauth2_auth_code").SetKeys(false, "Code", "client_id", "redirect_uri")
	Schema.CreateSQL = append(Schema.CreateSQL,
		`ALTER TABLE oauth2_auth_code ALTER COLUMN expires_at TYPE timestamp with time zone USING expires_at::timestamp with time zone;`,
		"ALTER TABLE oauth2_auth_code ALTER COLUMN scope TYPE text[] USING array[scope]::text[];",
	)
}

// dbAuthCode DB-maps an OAuth2 authorization code grant and related
// metadata.
type dbAuthCode struct {
	Code        string
	ClientID    string `db:"client_id"`
	RedirectURI string `db:"redirect_uri"`
	Scope       *dbutil.StringSlice
	UID         int32
	ExpiresAt   time.Time `db:"expires_at"`
	Exchanged   bool
}

// Authorizations is a FS-backed implementation of the Authorizations store.
type Authorizations struct{}

var _ store.Authorizations = (*Authorizations)(nil)

func (s *Authorizations) CreateAuthCode(ctx context.Context, req *sourcegraph.AuthorizationCodeRequest, expires time.Duration) (string, error) {
	code := &dbAuthCode{
		Code:        randstring.NewLen(40),
		ClientID:    req.ClientID,
		RedirectURI: req.RedirectURI,
		Scope:       &dbutil.StringSlice{Slice: req.Scope},
		UID:         req.UID,
		ExpiresAt:   time.Now().Add(expires),
	}

	if err := dbh(ctx).Insert(code); err != nil {
		return "", err
	}

	// Clean up.
	if err := s.removeExpiredAuthCodes(ctx); err != nil {
		return "", err
	}

	return code.Code, nil
}

func (s *Authorizations) MarkExchanged(ctx context.Context, code *sourcegraph.AuthorizationCode, clientID string) (*sourcegraph.AuthorizationCodeRequest, error) {
	var args []interface{}
	arg := func(a interface{}) string {
		v := modl.PostgresDialect{}.BindVar(len(args))
		args = append(args, a)
		return v
	}

	sql := `
SELECT * FROM oauth2_auth_code c
WHERE c.code=` + arg(code.Code) + ` AND c.redirect_uri=` + arg(code.RedirectURI) + ` AND
      c.client_id=` + arg(clientID) + ` AND c.expires_at > current_timestamp;`

	var dbCodes []*dbAuthCode
	if err := dbh(ctx).Select(&dbCodes, sql, args...); err != nil {
		return nil, err
	}
	if len(dbCodes) == 0 {
		return nil, store.ErrAuthCodeNotFound
	}
	dbCode := dbCodes[0]

	// Don't allow it to be exchanged twice!
	if dbCode.Exchanged {
		log.Printf("Warning: auth code %q (UID %d, scope %v) exchanged twice! Possible attack in progress.", dbCode.Code, dbCode.UID, dbCode.Scope)
		return nil, store.ErrAuthCodeAlreadyExchanged
	}

	dbCode.Exchanged = true
	if _, err := dbh(ctx).Update(dbCode); err != nil {
		return nil, err
	}

	// Clean up.
	if err := s.removeExpiredAuthCodes(ctx); err != nil {
		return nil, err
	}

	return &sourcegraph.AuthorizationCodeRequest{
		ClientID:    dbCode.ClientID,
		RedirectURI: dbCode.RedirectURI,
		Scope:       dbCode.Scope.Slice,
		UID:         dbCode.UID,
	}, nil
}

// removeExpiredAuthCodes is run when we write to the auth code DB, to
// occasionally purge the DB of expired grants.
func (s *Authorizations) removeExpiredAuthCodes(ctx context.Context) error {
	_, err := dbh(ctx).Exec(`DELETE FROM oauth2_auth_code WHERE expires_at <= current_timestamp;`)
	return err
}
