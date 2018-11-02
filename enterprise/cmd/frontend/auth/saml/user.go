package saml

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	saml2 "github.com/russellhaering/gosaml2"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/external/auth"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc"
)

type authnResponseInfo struct {
	spec                 extsvc.ExternalAccountSpec
	email, displayName   string
	unnormalizedUsername string
	accountData          interface{}
}

func readAuthnResponse(p *provider, encodedResp string) (*authnResponseInfo, error) {
	{
		if raw, err := base64.StdEncoding.DecodeString(encodedResp); err == nil {
			traceLog(fmt.Sprintf("AuthnResponse: %s", p.ConfigID().ID), string(raw))
		}
	}

	assertions, err := p.samlSP.RetrieveAssertionInfo(encodedResp)
	if err != nil {
		return nil, errors.WithMessage(err, "reading AuthnResponse assertions")
	}
	if wi := assertions.WarningInfo; wi.InvalidTime || wi.NotInAudience {
		return nil, fmt.Errorf("invalid SAML AuthnResponse: %+v", wi)
	}

	pi, err := p.getCachedInfoAndError()
	if err != nil {
		return nil, err
	}

	firstNonempty := func(ss ...string) string {
		for _, s := range ss {
			if s := strings.TrimSpace(s); s != "" {
				return s
			}
		}
		return ""
	}
	attr := samlAssertionValues(assertions.Values)
	email := firstNonempty(attr.Get("email"), attr.Get("emailaddress"))
	if email == "" && mightBeEmail(assertions.NameID) {
		email = assertions.NameID
	}
	if pn := attr.Get("eduPersonPrincipalName"); email == "" && mightBeEmail(pn) {
		email = pn
	}
	info := authnResponseInfo{
		spec: extsvc.ExternalAccountSpec{
			ServiceType: providerType,
			ServiceID:   pi.ServiceID,
			ClientID:    pi.ClientID,
			AccountID:   assertions.NameID,
		},
		email:                email,
		unnormalizedUsername: firstNonempty(attr.Get("login"), attr.Get("uid"), email),
		displayName:          firstNonempty(attr.Get("displayName"), attr.Get("givenName")+" "+attr.Get("surname")),
		accountData:          assertions,
	}
	if assertions.NameID == "" {
		return nil, errors.New("the SAML response did not contain a valid NameID")
	}
	if info.email == "" {
		return nil, errors.New("the SAML response did not contain an email attribute")
	}
	if info.unnormalizedUsername == "" {
		return nil, errors.New("the SAML response did not contain a username attribute")
	}
	return &info, nil
}

// getOrCreateUser gets or creates a user account based on the SAML claims. It returns the
// authenticated actor if successful; otherwise it returns an friendly error message (safeErrMsg)
// that is safe to display to users, and a non-nil err with lower-level error details.
func getOrCreateUser(ctx context.Context, info *authnResponseInfo) (_ *actor.Actor, safeErrMsg string, err error) {
	var data extsvc.ExternalAccountData
	auth.SetExternalAccountData(&data.AccountData, info.accountData)

	username, err := auth.NormalizeUsername(info.unnormalizedUsername)
	if err != nil {
		return nil, fmt.Sprintf("Error normalizing the username %q. See https://about.sourcegraph.com/docs/config/authentication#username-normalization.", info.unnormalizedUsername), err
	}

	userID, safeErrMsg, err := auth.CreateOrUpdateUser(ctx, db.NewUser{
		Username:        username,
		Email:           info.email,
		EmailIsVerified: info.email != "", // TODO(sqs): https://github.com/sourcegraph/sourcegraph/issues/10118
		DisplayName:     info.displayName,
		// SAML has no standard way of providing an avatar URL.
	},
		info.spec,
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
