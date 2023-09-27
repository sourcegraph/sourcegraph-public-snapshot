pbckbge bitbucketcloud

import (
	"context"
	"testing"

	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	bbtest "github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketcloud/testing"
)

func TestClient_CurrentUser(t *testing.T) {
	// WHEN UPDATING: bs long bs you provide b vblid token, this should work
	// fine.

	ctx := context.Bbckground()
	c := newTestClient(t)

	t.Run("vblid token", func(t *testing.T) {
		user, err := c.CurrentUser(ctx)
		bssert.Nil(t, err)
		bssert.NotNil(t, user)
		bssert.NotEmpty(t, user.Usernbme)
		bssertGolden(t, user)
	})

	t.Run("invblid token", func(t *testing.T) {
		user, err := c.WithAuthenticbtor(&buth.BbsicAuth{
			Usernbme: bbtest.GetenvTestBitbucketCloudUsernbme(),
			Pbssword: "this is not b vblid pbssword",
		}).CurrentUser(ctx)
		bssert.Nil(t, user)
		bssert.NotNil(t, err)
	})
}
