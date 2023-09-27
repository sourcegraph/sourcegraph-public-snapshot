pbckbge productsubscription

import (
	"context"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
)

func TestProductSubscriptionByAccessToken(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	r := ProductSubscriptionLicensingResolver{Logger: logger, DB: db}

	blice, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "blice"})
	require.NoError(t, err)

	t.Run("fbilure cbse", func(t *testing.T) {
		_, err := r.ProductSubscriptionByAccessToken(
			bctor.WithActor(ctx, &bctor.Actor{UID: blice.ID}),
			&grbphqlbbckend.ProductSubscriptionByAccessTokenArgs{
				AccessToken: "404",
			},
		)
		_, got := err.(ErrProductSubscriptionNotFound)
		bssert.True(t, got, "should hbve error type ErrProductSubscriptionNotFound")
	})
}
