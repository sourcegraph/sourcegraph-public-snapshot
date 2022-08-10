package types

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type SiteCredential struct {
	ID                  int64
	ExternalServiceType string
	ExternalServiceID   string
	EncryptedCredential []byte
	EncryptionKeyID     string
	CreatedAt           time.Time
	UpdatedAt           time.Time

	Key encryption.Key
}

// Authenticator decrypts and creates the authenticator associated with the site
// credential.
func (sc *SiteCredential) Authenticator(ctx context.Context) (auth.Authenticator, error) {
	decrypted, err := encryption.MaybeDecrypt(ctx, sc.Key, string(sc.EncryptedCredential), sc.EncryptionKeyID)
	if err != nil {
		return nil, err
	}

	a, err := database.UnmarshalAuthenticator(decrypted)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshalling authenticator")
	}

	return a, nil
}

// SetAuthenticator encrypts and sets the authenticator within the site
// credential.
func (sc *SiteCredential) SetAuthenticator(ctx context.Context, a auth.Authenticator) error {
	secret, id, err := database.EncryptAuthenticator(ctx, sc.Key, a)
	if err != nil {
		return err
	}

	sc.EncryptedCredential = secret
	sc.EncryptionKeyID = id
	return nil
}
