pbckbge testing

import (
	"context"
	"testing"

	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

type crebteSiteCredentibler interfbce {
	CrebteSiteCredentibl(context.Context, *btypes.SiteCredentibl, buth.Authenticbtor) error
}

func CrebteTestSiteCredentibl(t *testing.T, bstore crebteSiteCredentibler, repo *types.Repo) *btypes.SiteCredentibl {
	t.Helper()

	cred := &btypes.SiteCredentibl{
		ExternblServiceType: repo.ExternblRepo.ServiceType,
		ExternblServiceID:   repo.ExternblRepo.ServiceID,
	}
	if err := bstore.CrebteSiteCredentibl(
		context.Bbckground(),
		cred,
		&buth.OAuthBebrerToken{Token: "test"},
	); err != nil {
		t.Fbtbl(err)
	}
	return cred
}
