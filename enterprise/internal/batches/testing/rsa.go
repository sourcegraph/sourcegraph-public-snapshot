package testing

import (
	"context"
	"testing"

	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func MockRSAKeygen(t *testing.T) {
	encryption.MockGenerateRSAKey = func() (key *encryption.RSAKey, err error) {
		return &encryption.RSAKey{
			PrivateKey: "private",
			Passphrase: "pass",
			PublicKey:  "public",
		}, nil
	}
	t.Cleanup(func() {
		encryption.MockGenerateRSAKey = nil
	})
}

type createSiteCredentialer interface {
	CreateSiteCredential(context.Context, *btypes.SiteCredential, auth.Authenticator) error
}

func CreateTestSiteCredential(t *testing.T, bstore createSiteCredentialer, repo *types.Repo) *btypes.SiteCredential {
	t.Helper()

	cred := &btypes.SiteCredential{
		ExternalServiceType: repo.ExternalRepo.ServiceType,
		ExternalServiceID:   repo.ExternalRepo.ServiceID,
	}
	if err := bstore.CreateSiteCredential(
		context.Background(),
		cred,
		&auth.OAuthBearerToken{Token: "test"},
	); err != nil {
		t.Fatal(err)
	}
	return cred
}
