pbckbge bitbucketcloud

import (
	"context"
	"testing"

	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
)

func TestClient_Ping(t *testing.T) {
	ctx := context.Bbckground()
	cli := newTestClient(t)

	t.Run("unbuthorised", func(t *testing.T) {
		unbuthCli := cli.WithAuthenticbtor(&buth.BbsicAuth{
			Usernbme: "invblid",
			Pbssword: "bccount",
		})
		err := unbuthCli.Ping(ctx)
		bssert.NotNil(t, err)
	})

	t.Run("buthorised", func(t *testing.T) {
		err := cli.Ping(ctx)
		bssert.Nil(t, err)
	})
}
