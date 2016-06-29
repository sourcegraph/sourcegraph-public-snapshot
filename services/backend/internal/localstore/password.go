package localstore

import (
	"database/sql"
	"math"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/clock"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
)

// password is a pgsql backed implementation of the passwords store.
// password uses the clock interface for testing purposes. Use the
// newPassword() constructor in order to make a normal password instance
// for production.
type password struct {
	clock clock.Clock
}

func newPassword() *password {
	return &password{clock: clock.NewProductionClock()}
}

var _ store.Password = (*password)(nil)

type dbPassword struct {
	UID              int32
	HashedPassword   []byte
	ConsecutiveFails int32     `db:"consecutive_fails"`
	LastFail         time.Time `db:"last_fail"`
}

var tableName = "passwords"

func init() {
	AppSchema.Map.AddTableWithName(dbPassword{}, tableName).SetKeys(false, "UID")
}

func unmarshalPassword(ctx context.Context, UID int32) (dbPassword, error) {
	var pass dbPassword
	err := appDBH(ctx).SelectOne(&pass, `SELECT * FROM passwords WHERE uid=$1`, UID)
	return pass, err
}

// calcWaitPeriod calculates how long someone should wait in between login attempts given that
// that they have failed 'fails' time in a row previously. If "fails" is zero, the duration is
// guaranteed to be 0 as well. Fails must be >= 0.
// Based off of recommendations from: https://www.owasp.org/index.php/Guide_to_Authentication#Thresholds_Governor
func calcWaitPeriod(fails int) time.Duration {
	if fails < 0 {
		panic("fails must be non-negative")
	}
	// The formula currently is: -5 + 5*3^(n/3) -> 0s, ~2s, ~5.4s, 10s, ~16.6s, ~26s, 40s, ...
	return -5*time.Second + 5*time.Duration(math.Pow(3, float64(fails)/3))*time.Second
}

// As recommeded by: https://www.owasp.org/index.php/Authentication_Cheat_Sheet#Prevent_Brute-Force_Attacks
const maxWaitPeriod = 20 * time.Minute

// CheckUIDPassword returns an error if the password argument is not correct for
// the user, or if the waiting period before the user can try logging in again
// has expired.
func (p password) CheckUIDPassword(ctx context.Context, UID int32, password string) error {
	if err := accesscontrol.VerifyUserHasAdminAccess(ctx, "Password.CheckUIDPassword"); err != nil {
		return err
	}
	genericErr := grpc.Errorf(codes.PermissionDenied, "password login not allowed for uid %d", UID)

	pass, err := unmarshalPassword(ctx, UID)
	if err == sql.ErrNoRows {
		return genericErr // User doesn't exist
	} else if err != nil {
		return err
	}

	waitPeriod := calcWaitPeriod(int(pass.ConsecutiveFails))

	if waitPeriod > maxWaitPeriod {
		waitPeriod = maxWaitPeriod
		log15.Warn("SECURITY: Maximum wait period for password attempt reached", "uid", UID, "waitPeriod", maxWaitPeriod)
	}

	// Has the user waited long enough (waitPeriod) since their last failure?
	// We round to the microsecond because that is the highest resolution that Postgres offers
	// for the timestamp data type.
	now := p.clock.Now().Round(1 * time.Microsecond)
	if !pass.LastFail.Add(waitPeriod).Round(1 * time.Microsecond).Before(now) {
		return grpc.Errorf(codes.PermissionDenied, "must wait for %s before trying again", waitPeriod.String())
	}

	if len(pass.HashedPassword) == 0 {
		// User has no password (and can only log in via
		// GitHub OAuth2, etc.)
		return genericErr
	}

	cmpRes := bcrypt.CompareHashAndPassword([]byte(pass.HashedPassword), []byte(password))
	if cmpRes == bcrypt.ErrMismatchedHashAndPassword {
		// Wrong password - update the failure information
		pass.ConsecutiveFails++
		pass.LastFail = now
	} else if cmpRes == nil {
		// Correct password - reset the failure count
		pass.ConsecutiveFails = 0
	}
	_, err = appDBH(ctx).Update(&pass)
	if err != nil {
		log15.Warn("Error when trying to update password failure information", "uid", UID, "error", err)
	}
	return cmpRes
}

func (p password) SetPassword(ctx context.Context, uid int32, password string) error {
	if err := accesscontrol.VerifyUserSelfOrAdmin(ctx, "Password.SetPassword", uid); err != nil {
		return err
	}

	if password == "" {
		// Clear password (user can only log in via GitHub OAuth2, for
		// example).
		_, err := appDBH(ctx).Exec(`DELETE FROM passwords WHERE uid=$1;`, uid)
		return err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 11)
	if err != nil {
		return err
	}

	query := `
WITH upsert AS (
  UPDATE passwords SET hashedpassword=$2, consecutive_fails=$3, last_fail=$4 WHERE uid=$1 RETURNING *
)
INSERT INTO passwords(uid, hashedpassword, consecutive_fails, last_fail) SELECT $1, $2, $3, $4 WHERE NOT EXISTS (SELECT * FROM upsert);`
	_, err = appDBH(ctx).Exec(query, uid, hashed, 0, p.clock.Now())
	return err
}
