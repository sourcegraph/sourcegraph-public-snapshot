package globalstatedb

import (
	"context"
	"database/sql"

	cryptorand "crypto/rand"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbutil"
	"golang.org/x/crypto/bcrypt"
)

type State struct {
	SiteID      string
	Initialized bool // whether the initial site admin account has been created
}

func Get(ctx context.Context) (*State, error) {
	if Mock.Get != nil {
		return Mock.Get(ctx)
	}

	configuration, err := getConfiguration(ctx)
	if err == nil {
		return configuration, nil
	}
	err = tryInsertNew(ctx, dbconn.Global)
	if err != nil {
		return nil, err
	}
	return getConfiguration(ctx)
}

func SiteInitialized(ctx context.Context) (alreadyInitialized bool, err error) {
	if err := dbconn.Global.QueryRowContext(ctx, `SELECT initialized FROM global_state LIMIT 1`).Scan(&alreadyInitialized); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return alreadyInitialized, err
}

// EnsureInitialized ensures the site is marked as having been initialized. If the site was already
// initialized, it does nothing. It returns whether the site was already initialized prior to the
// call.
//
// 🚨 SECURITY: Initialization is an important security measure. If a new account is created on a
// site that is not initialized, and no other accounts exist, it is granted site admin
// privileges. If the site *has* been initialized, then a new account is not granted site admin
// privileges (even if all other users are deleted). This reduces the risk of (1) a site admin
// accidentally deleting all user accounts and opening up their site to any attacker becoming a site
// admin and (2) a bug in user account creation code letting attackers create site admin accounts.
func EnsureInitialized(ctx context.Context, dbh interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}) (alreadyInitialized bool, err error) {
	if err := tryInsertNew(ctx, dbh); err != nil {
		return false, err
	}

	// The "SELECT ... FOR UPDATE" prevents a race condition where two calls, each in their own transaction,
	// would see this initialized value as false and then set it to true below.
	if err := dbh.QueryRowContext(ctx, `SELECT initialized FROM global_state FOR UPDATE LIMIT 1`).Scan(&alreadyInitialized); err != nil {
		return false, err
	}

	if !alreadyInitialized {
		_, err = dbh.ExecContext(ctx, "UPDATE global_state SET initialized=true")
	}

	return alreadyInitialized, err
}

func getConfiguration(ctx context.Context) (*State, error) {
	configuration := &State{}
	err := dbconn.Global.QueryRowContext(ctx, "SELECT site_id, initialized FROM global_state LIMIT 1").Scan(
		&configuration.SiteID,
		&configuration.Initialized,
	)
	return configuration, err
}

func tryInsertNew(ctx context.Context, dbh interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}) error {
	siteID, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	// In the normal case (when no users exist yet because the instance is brand new), create the row
	// with initialized=false.
	//
	// If any users exist, then set the site as initialized so that the init screen doesn't show
	// up. (It would not let the visitor initialize the site anyway, because other users exist.) The
	// most likely reason the instance would get into this state (uninitialized but has users) is
	// because previously global state had a siteID and now we ignore that (or someone ran `DELETE
	// FROM global_state;` in the PostgreSQL database). In either case, it's safe to generate a new
	// site ID and set the site as initialized.
	_, err = dbh.ExecContext(ctx, "INSERT INTO global_state(site_id, initialized) values($1, EXISTS (SELECT 1 FROM users WHERE deleted_at IS NULL))", siteID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Constraint == "global_state_pkey" {
				// The row we were trying to insert already exists.
				// Don't treat this as an error.
				err = nil
			}

		}
	}
	return err
}

// ManagementConsoleState describes state regarding the management console.
type ManagementConsoleState struct {
	// PasswordPlaintext is the plaintext version of the management console
	// password. It is automatically generated if there is not an existing
	// management console password. However, the plaintext version here only
	// remains until the admin dismisses it. After that, only the bcrypt form
	// remains (see DismissManagementConsolePassword).
	PasswordPlaintext string

	// PasswordBcrypt is the bcrypt form of the management console password.
	PasswordBcrypt string
}

// ErrCannotUseManagementConsole is returned by AuthenticateManagementConsole
// and GetManagementConsolePlaintextPassword if the site is not yet
// initialized.
var ErrCannotUseManagementConsole = errors.New("cannot use management console until site is initialized")

// AuthenticateManagementConsole handles authentication for the management
// console. It returns nil on success and an error otherwise.
//
// Because the management console is considered both public on the internet AND
// without an authentication rate limit (this could lock out the real site
// admin), authentication uses a prohibitively costly bcrypt hash which takes
// 2.2s on e.g. a modern laptop. It is therefor also important for the password
// to be very long and random.
//
// 🚨 SECURITY: This method MUST be called before granting anyone access to the
// management console (it is the only form of authentication).
func AuthenticateManagementConsole(ctx context.Context, password string) error {
	mgmt, err := getManagementConsoleState(ctx)
	if err != nil {
		return err
	}
	return bcrypt.CompareHashAndPassword([]byte(mgmt.PasswordBcrypt), []byte(password))
}

// ClearManagementConsolePlaintextPassword clears the plaintext form of the
// management console password so that it is no longer stored insecurely in the
// DB. It is stored initially so the admin can retrieve it, and once the admin
// does it is cleared.
//
// 🚨 SECURITY: Only site admins or someone authenticated via AuthenticateManagementConsole
// should be allowed to invoke this method. Any other user could be a malicious
// actor attempting to lock the admin out.
func ClearManagementConsolePlaintextPassword(ctx context.Context) error {
	_, err := dbconn.Global.ExecContext(ctx, "UPDATE global_state SET mgmt_password_plaintext=''")
	if err != nil {
		return errors.Wrap(err, "UPDATE")
	}
	return nil
}

// GetManagementConsolePlaintextPassword fetches and returns the management console
// plaintext password form. After the admin is aware of this, ClearManagementConsolePlaintextPassword
// should be called to clear it.
//
// 🚨 SECURITY: The result of this function should ONLY ever be shown to site
// admins. Any other user could be a malicious actor.
func GetManagementConsolePlaintextPassword(ctx context.Context) (string, error) {
	mgmt, err := getManagementConsoleState(ctx)
	if err != nil {
		return "", err
	}
	return mgmt.PasswordPlaintext, nil
}

// getManagementConsoleState fetches and returns the global management console
// state containing the password of the management console.
func getManagementConsoleState(ctx context.Context) (*ManagementConsoleState, error) {
	// We first enforce that the site is initialized. This is purely to ensure
	// the table row is created for us already.
	initialized, err := SiteInitialized(ctx)
	if err != nil {
		return nil, err
	}
	if !initialized {
		return nil, ErrCannotUseManagementConsole
	}

	var finalState *ManagementConsoleState
	err = dbutil.Transaction(ctx, dbconn.Global, func(tx *sql.Tx) error {
		// Get the current management console state.
		mgmt, err := doGetManagementConsoleState(ctx, tx)
		if err != nil {
			return errors.Wrap(err, "doGetManagementConsoleState")
		}

		// If there is no password (e.g. new instance, old instances migrated
		// to new Sourcegraph version with management console, etc), then we
		// generate and store one.
		if mgmt.PasswordBcrypt != "" {
			// We have a password already.
			finalState = mgmt
			return nil
		}

		// Generate a new password and store both plaintext and bcrypt forms.
		passwordPlaintext, err := generateRandomPassword()
		if err != nil {
			return errors.Wrap(err, "generateRandomPassword")
		}
		passwordBcrypt, err := bcrypt.GenerateFromPassword([]byte(passwordPlaintext), 15)
		if err != nil {
			return errors.Wrap(err, "bcrypt")
		}
		_, err = tx.ExecContext(ctx,
			"UPDATE global_state SET mgmt_password_plaintext=$1, mgmt_password_bcrypt=$2",
			passwordPlaintext, passwordBcrypt,
		)
		if err != nil {
			return errors.Wrap(err, "UPDATE")
		}

		finalState = &ManagementConsoleState{
			PasswordPlaintext: passwordPlaintext,
			PasswordBcrypt:    string(passwordBcrypt),
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "transaction")
	}
	return finalState, nil
}

func doGetManagementConsoleState(ctx context.Context, tx *sql.Tx) (*ManagementConsoleState, error) {
	mgmt := &ManagementConsoleState{}
	err := dbconn.Global.QueryRowContext(ctx, "SELECT mgmt_password_plaintext, mgmt_password_bcrypt FROM global_state LIMIT 1").Scan(
		&mgmt.PasswordPlaintext,
		&mgmt.PasswordBcrypt,
	)
	return mgmt, err
}

var allowedPasswordCharacters []rune

func init() {
	allowedPasswordCharacters = append(allowedPasswordCharacters, []rune("abcdefghijklmnopqrstuvwxyz")...)
	allowedPasswordCharacters = append(allowedPasswordCharacters, []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")...)
	allowedPasswordCharacters = append(allowedPasswordCharacters, []rune("0123456789")...)
	allowedPasswordCharacters = append(allowedPasswordCharacters, []rune(`~!@#$%^&*_-+=<,>.?`)...)
}

// generateRandomPassword generates a random ASCII password of length 128 using
// crypto/rand as the source.
func generateRandomPassword() (string, error) {
	data := make([]byte, 128)
	_, err := cryptorand.Read(data)
	if err != nil {
		return "", errors.Wrap(err, "failed to generate crypto/rand data")
	}
	var generated []rune
	for _, n := range data {
		generated = append(generated, allowedPasswordCharacters[int(n)%len(allowedPasswordCharacters)])
	}
	return string(generated), nil
}
