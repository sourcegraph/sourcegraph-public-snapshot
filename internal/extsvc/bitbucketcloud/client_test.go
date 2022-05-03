package bitbucketcloud

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
)

func TestClient_Ping(t *testing.T) {
	ctx := context.Background()
	cli := newTestClient(t)

	t.Run("unauthorised", func(t *testing.T) {
		unauthCli := cli.WithAuthenticator(&auth.BasicAuth{
			Username: "invalid",
			Password: "account",
		})
		err := unauthCli.Ping(ctx)
		assert.NotNil(t, err)
	})

	t.Run("authorised", func(t *testing.T) {
		err := cli.Ping(ctx)
		assert.Nil(t, err)
	})
}
