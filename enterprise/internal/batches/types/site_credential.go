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

const (
	SiteCredentialPlaceholderEncryptionKeyID = "previously-migrated"
	SiteCredentialUnmigratedEncryptionKeyID  = "unmigrated"
)

// Authenticator decrypts and creates the authenticator associated with the site
// credential.
func (sc *SiteCredential) Authenticator(ctx context.Context) (auth.Authenticator, error) {
	// The record includes a field indicating the encryption key ID. We don't
	// really have a way to look up a key by ID right now, so this is used as a
	// marker of whether we should expect a key or not.
	if sc.EncryptionKeyID == "" || sc.EncryptionKeyID == SiteCredentialUnmigratedEncryptionKeyID {
		return database.UnmarshalAuthenticator(string(sc.EncryptedCredential))
	}
	if sc.Key == nil {
		return nil, errors.New("user credential is encrypted, but no key is available to decrypt it")
	}

	secret, err := sc.Key.Decrypt(ctx, sc.EncryptedCredential)
	if err != nil {
		return nil, errors.Wrap(err, "decrypting credential")
	}

	a, err := database.UnmarshalAuthenticator(secret.Secret())
	if err != nil {
		return nil, errors.Wrap(err, "unmarshalling authenticator")
	}

	return a, nil
}

// SetAuthenticator encrypts and sets the authenticator within the site
// credential.
func (sc *SiteCredential) SetAuthenticator(ctx context.Context, a auth.Authenticator) error {
	// Set the key ID. This is cargo culted from external_accounts.go, and the
	// key ID doesn't appear to be actually useful as anything other than a
	// marker of whether the data is expected to be encrypted or not.
	id, err := keyID(ctx, sc.Key)
	if err != nil {
		return errors.Wrap(err, "getting key version")
	}

	secret, err := database.EncryptAuthenticator(ctx, sc.Key, a)
	if err != nil {
		return err
	}

	sc.EncryptedCredential = secret
	sc.EncryptionKeyID = id

	return nil
}

func keyID(ctx context.Context, key encryption.Key) (string, error) {
	if key != nil {
		version, err := key.Version(ctx)
		if err != nil {
			return "", errors.Wrap(err, "getting key version")
		}
		return version.JSON(), nil
	}

	return "", nil
}
