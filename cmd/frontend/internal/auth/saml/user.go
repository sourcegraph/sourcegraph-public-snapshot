package saml

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	saml2 "github.com/russellhaering/gosaml2"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
)

// getOrCreateUser gets or creates a user account based on the SAML claims. It returns the
// authenticated actor if successful; otherwise it returns an friendly error message (safeErrMsg)
// that is safe to display to users, and a non-nil err with lower-level error details.
func getOrCreateUser(ctx context.Context, p *provider, info *saml2.AssertionInfo) (_ *actor.Actor, safeErrMsg string, err error) {
	var data db.ExternalAccountData
	auth.SetExternalAccountData(&data.AccountData, info)

	attr := samlAssertionValues(info.Values)
	email := attr.Get("email")
	if email == "" && mightBeEmail(info.NameID) {
		email = info.NameID
	}
	login := attr.Get("login")
	if login == "" {
		login = attr.Get("uid")
	}
	displayName := attr.Get("displayName")
	if displayName == "" {
		displayName = attr.Get("givenName")
	}
	if displayName == "" {
		displayName = login
	}
	if displayName == "" {
		displayName = email
	}
	if displayName == "" {
		displayName = info.NameID
	}
	if login == "" {
		login = email
	}
	if login == "" {
		return nil, "The SAML authentication provider did not contain an email attribute.", errors.New("SAML response did not contain email")
	}
	login, err = auth.NormalizeUsername(login)
	if err != nil {
		return nil, fmt.Sprintf("Error normalizing the username %q. See https://about.sourcegraph.com/docs/config/authentication#username-normalization.", login), err
	}

	userID, safeErrMsg, err := auth.CreateOrUpdateUser(ctx, db.NewUser{
		Username:        login,
		Email:           email,
		EmailIsVerified: email != "", // TODO(sqs): https://github.com/sourcegraph/sourcegraph/issues/10118
		DisplayName:     displayName,
		// SAML has no standard way of providing an avatar URL.
	},
		db.ExternalAccountSpec{
			ServiceType: providerType,
			ServiceID:   p.ID().ID,
			AccountID:   info.NameID,
		},
		data,
	)
	if err != nil {
		return nil, safeErrMsg, err
	}
	return actor.FromUser(userID), "", nil
}

func mightBeEmail(s string) bool {
	return strings.Count(s, "@") == 1
}

type samlAssertionValues saml2.Values

func (v samlAssertionValues) Get(key string) string {
	for _, a := range v {
		if a.Name == key || a.FriendlyName == key {
			return a.Values[0].Value
		}
	}
	return ""
}
