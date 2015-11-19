package fs

import (
	"crypto/subtle"
	"encoding/json"
	"log"
	"os"
	"time"

	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/rwvfs"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/util/randstring"
)

const authCodesDBFilename = "authorization_codes.json"

// authCode is an OAuth2 authorization code grant and related
// metadata.
type authCode struct {
	Code string
	*sourcegraph.AuthorizationCodeRequest
	ExpiresAt time.Time
	Exchanged bool
}

func (c *authCode) expired() bool {
	return time.Now().After(c.ExpiresAt)
}

// readAuthCodesDB reads the regClient/account database from disk. If no such
// file exists, an empty slice is returned (and no error).
func readAuthCodesDB(ctx context.Context) ([]*authCode, error) {
	f, err := dbVFS(ctx).Open(authCodesDBFilename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var authCodes []*authCode
	if err := json.NewDecoder(f).Decode(&authCodes); err != nil {
		return nil, err
	}
	return authCodes, nil
}

// writeAuthCodesDB writes the regClient/account database to disk.
func writeAuthCodesDB(ctx context.Context, authCodes []*authCode) (err error) {
	data, err := json.MarshalIndent(authCodes, "", "  ")
	if err != nil {
		return err
	}

	if err := rwvfs.MkdirAll(dbVFS(ctx), "."); err != nil {
		return err
	}
	f, err := dbVFS(ctx).Create(authCodesDBFilename)
	if err != nil {
		return err
	}
	defer func() {
		if err2 := f.Close(); err2 != nil {
			if err == nil {
				err = err2
			} else {
				log.Printf("Warning: closing auth codes DB after error (%s) failed: %s.", err, err2)
			}
		}
	}()

	_, err = f.Write(data)
	return err
}

// Authorizations is a FS-backed implementation of the Authorizations store.
type Authorizations struct{}

var _ store.Authorizations = (*Authorizations)(nil)

func (s *Authorizations) CreateAuthCode(ctx context.Context, req *sourcegraph.AuthorizationCodeRequest, expires time.Duration) (string, error) {
	codes, err := readAuthCodesDB(ctx)
	if err != nil {
		return "", err
	}

	// Append auth code.
	code := &authCode{
		Code:                     randstring.NewLen(40),
		ExpiresAt:                time.Now().Add(expires),
		AuthorizationCodeRequest: req,
	}
	codes = append(codes, code)

	// Save to disk.
	if err := writeAuthCodesDB(ctx, removeExpiredAuthCodes(codes)); err != nil {
		return "", err
	}

	return code.Code, nil
}

func (s *Authorizations) MarkExchanged(ctx context.Context, code *sourcegraph.AuthorizationCode, clientID string) (*sourcegraph.AuthorizationCodeRequest, error) {
	codes, err := readAuthCodesDB(ctx)
	if err != nil {
		return nil, err
	}

	// Find the code.
	var dbCode *authCode
	for _, c := range codes {
		if subtle.ConstantTimeCompare([]byte(c.Code), []byte(code.Code)) == 1 && c.RedirectURI == code.RedirectURI && c.ClientID == clientID && !c.expired() {
			dbCode = c
			break
		}
	}
	if dbCode == nil {
		return nil, store.ErrAuthCodeNotFound
	}

	// Don't allow it to be exchanged twice!
	if dbCode.Exchanged {
		log.Printf("Warning: auth code %q (UID %d, scope %v) exchanged twice! Possible attack in progress.", dbCode.Code, dbCode.UID, dbCode.Scope)
		return nil, store.ErrAuthCodeAlreadyExchanged
	}
	dbCode.Exchanged = true

	// Save to disk.
	if err := writeAuthCodesDB(ctx, removeExpiredAuthCodes(codes)); err != nil {
		return nil, err
	}

	return dbCode.AuthorizationCodeRequest, nil
}

// removeExpiredAuthCodes is run when we write to the auth code DB, to
// occasionally purge the DB of expired grants.
func removeExpiredAuthCodes(codes []*authCode) []*authCode {
	unexpired := make([]*authCode, 0, len(codes))
	for _, c := range codes {
		if !c.expired() {
			unexpired = append(unexpired, c)
		}
	}
	return unexpired
}
