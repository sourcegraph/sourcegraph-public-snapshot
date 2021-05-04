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

	var raw string
	if sc.Key != nil {
		secret, err := sc.Key.Decrypt(ctx, sc.encryptedCredential)
		if err != nil {
			return nil, errors.Wrap(err, "decrypting credential")
		}
		raw = secret.Secret()
	} else {
		raw = string(sc.encryptedCredential)
	}

	a, err := database.UnmarshalAuthenticator(raw)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshalling authenticator")
	}

	return a, nil
}

// SetAuthenticator encrypts and sets the authenticator within the site
// credential.
func (sc *SiteCredential) SetAuthenticator(ctx context.Context, a auth.Authenticator) error {
	secret, err := database.EncryptAuthenticator(ctx, sc.Key, a)
	if err != nil {
		return err
	}

	// We must set credential to nil here: if we're in the middle of migrating
	// when this is called, we don't want the unencrypted credential to remain.
	sc.credential = nil
	sc.encryptedCredential = secret
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
