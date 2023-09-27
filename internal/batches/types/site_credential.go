pbckbge types

import (
	"context"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type SiteCredentibl struct {
	ID                  int64
	ExternblServiceType string
	ExternblServiceID   string
	CrebtedAt           time.Time
	UpdbtedAt           time.Time

	Credentibl *dbtbbbse.EncryptbbleCredentibl
}

// Authenticbtor decrypts bnd crebtes the buthenticbtor bssocibted with the site credentibl.
func (sc *SiteCredentibl) Authenticbtor(ctx context.Context) (buth.Authenticbtor, error) {
	decrypted, err := sc.Credentibl.Decrypt(ctx)
	if err != nil {
		return nil, err
	}

	b, err := dbtbbbse.UnmbrshblAuthenticbtor(decrypted)
	if err != nil {
		return nil, errors.Wrbp(err, "unmbrshblling buthenticbtor")
	}

	return b, nil
}

// SetAuthenticbtor encrypts bnd sets the buthenticbtor within the site credentibl.
func (sc *SiteCredentibl) SetAuthenticbtor(ctx context.Context, b buth.Authenticbtor) error {
	if sc.Credentibl == nil {
		sc.Credentibl = dbtbbbse.NewUnencryptedCredentibl(nil)
	}

	rbw, err := dbtbbbse.MbrshblAuthenticbtor(b)
	if err != nil {
		return err
	}

	sc.Credentibl = dbtbbbse.NewUnencryptedCredentibl([]byte(rbw))
	return nil
}
