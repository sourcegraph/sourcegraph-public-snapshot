package types

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type SiteCredential struct {
	ID                  int64
	ExternalServiceType string
	ExternalServiceID   string
	CreatedAt           time.Time
	UpdatedAt           time.Time

	Credential *database.EncryptableCredential
}

// Authenticator decrypts and creates the authenticator associated with the site credential.
func (sc *SiteCredential) Authenticator(ctx context.Context) (auth.Authenticator, error) {
	decrypted, err := sc.Credential.Decrypt(ctx)
	if err != nil {
		return nil, err
	}

	a, err := database.UnmarshalAuthenticator(decrypted)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshalling authenticator")
	}

	return a, nil
}

// SetAuthenticator encrypts and sets the authenticator within the site credential.
func (sc *SiteCredential) SetAuthenticator(ctx context.Context, a auth.Authenticator) error {
	if sc.Credential == nil {
		sc.Credential = database.NewUnencryptedCredential(nil)
	}

	raw, err := database.MarshalAuthenticator(a)
	if err != nil {
		return err
	}

	sc.Credential = database.NewUnencryptedCredential([]byte(raw))
	return nil
}
