package types

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
)

type SiteCredential struct {
	ID                  int64
	ExternalServiceType string
	ExternalServiceID   string
	EncryptionKeyID     string
	CreatedAt           time.Time
	UpdatedAt           time.Time

	// TODO(batch-changes-site-credential-encryption): once we're through the
	// migration period, this should be simplified back to a single, exported
	// EncryptedCredential []byte field.
	credential          auth.Authenticator
	encryptedCredential []byte

	Key encryption.Key
}

// Authenticator decrypts and creates the authenticator associated with the site
// credential.
func (sc *SiteCredential) Authenticator(ctx context.Context) (auth.Authenticator, error) {
	if sc.credential != nil {
		return sc.credential, nil
	}

	if sc.encryptedCredential == nil {
		return nil, errors.New("no unencrypted or encrypted credential found")
	}

	// The record includes a field indicating the encryption key ID. We don't
	// really have a way to look up a key by ID right now, so this is used as a
	// marker of whether we should expect a key or not.
	if sc.EncryptionKeyID == "" {
		return database.UnmarshalAuthenticator(string(sc.encryptedCredential))
	}
	if sc.Key == nil {
		return nil, errors.New("user credential is encrypted, but no key is available to decrypt it")
	}

	secret, err := sc.Key.Decrypt(ctx, sc.encryptedCredential)
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

	// We must set credential to nil here: if we're in the middle of migrating
	// when this is called, we don't want the unencrypted credential to remain.
	sc.credential = nil
	sc.encryptedCredential = secret
	sc.EncryptionKeyID = id

	return nil
}

// GetRawCredential gets the raw fields within the SiteCredential instance. This
// must only be called from the batches store when updating a site credential;
// all other users must use Authenticator() and SetAuthenticator() to access and
// mutate the credential state.
func (sc *SiteCredential) GetRawCredential() (auth.Authenticator, []byte) {
	return sc.credential, sc.encryptedCredential
}

// SetRawCredential sets the raw fields within the SiteCredential instance. This
// must only be called from the batches store when scanning a site credential;
// all other users must use Authenticator() and SetAuthenticator() to access and
// mutate the credential state.
func (sc *SiteCredential) SetRawCredential(credential auth.Authenticator, encryptedCredential []byte) {
	sc.credential = credential
	sc.encryptedCredential = encryptedCredential
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
