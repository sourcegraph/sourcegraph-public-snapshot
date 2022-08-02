package repos

import (
	"context"
	"strconv"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/oauthutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type RefreshTokenConfig struct {
	DB                database.DB
	ExternalServiceID int64
	TokenOauthRefresh string
}

func (r *RefreshTokenConfig) RefreshToken(ctx context.Context, doer httpcli.Doer, oauthCtx oauthutil.OauthContext) (string, error) {
	refreshedToken, err := oauthutil.RetrieveToken(ctx, doer, oauthCtx, r.TokenOauthRefresh, oauthutil.AuthStyleInParams)
	if err != nil {
		return "", errors.Wrap(err, "error retrieving token")
	}

	svc, err := r.DB.ExternalServices().GetByID(ctx, r.ExternalServiceID)
	if err != nil {
		return "", errors.Wrap(err, "getting external service")
	}

	defer func() {
		success := err == nil
		gitlab.TokenRefreshCounter.WithLabelValues("codehost", strconv.FormatBool(success)).Inc()
	}()

	svc.Config, err = jsonc.Edit(svc.Config, refreshedToken.AccessToken, "token")
	if err != nil {
		return "", errors.Wrap(err, "updating OAuth token")
	}
	svc.Config, err = jsonc.Edit(svc.Config, refreshedToken.RefreshToken, "token.oauth.refresh")
	if err != nil {
		return "", errors.Wrap(err, "updating OAuth refresh token")
	}
	svc.Config, err = jsonc.Edit(svc.Config, refreshedToken.Expiry.Unix(), "token.oauth.expiry")
	if err != nil {
		return "", errors.Wrap(err, "updating OAuth token expiry")
	}
	svc.UpdatedAt = time.Now()
	if err := r.DB.ExternalServices().Upsert(ctx, svc); err != nil {
		return "", errors.Wrap(err, "upserting external service")
	}

	return "", nil
}
