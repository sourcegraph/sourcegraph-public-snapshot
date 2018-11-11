package githuboauth

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/go-github/github"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/external/auth"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc"
	"golang.org/x/oauth2"
)

func getOrCreateUser(ctx context.Context, p *provider, ghUser *github.User, token *oauth2.Token) (_ *actor.Actor, safeErrMsg string, _ error) {
	login, err := auth.NormalizeUsername(deref(ghUser.Login))
	if err != nil {
		return nil, fmt.Sprintf("Error normalizing the username %q. See https://docs.sourcegraph.com/admin/auth/#username-normalization.", login), err
	}

	var data extsvc.ExternalAccountData
	auth.SetExternalAccountData(&data.AccountData, ghUser)
	userID, safeErrMsg, err := auth.CreateOrUpdateUser(ctx, db.NewUser{
		Username:        login,
		Email:           deref(ghUser.Email),
		EmailIsVerified: deref(ghUser.Email) != "",
		DisplayName:     deref(ghUser.Name),
		AvatarURL:       deref(ghUser.AvatarURL),
	}, extsvc.ExternalAccountSpec{
		ServiceType: serviceType,
		ServiceID:   p.serviceID,
		ClientID:    p.config.ClientID,
		AccountID:   strconv.FormatInt(derefInt64(ghUser.ID), 10),
	}, data)
	if err != nil {
		return nil, safeErrMsg, err
	}
	return actor.FromUser(userID), "", nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func derefInt64(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}
