package saml

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	saml2 "github.com/russellhaering/gosaml2"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type authnResponseInfo struct {
	spec                 extsvc.AccountSpec
	email, displayName   string
	unnormalizedUsername string
	groups               map[string]bool
	accountData          any
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
		return nil, errors.Errorf("invalid SAML AuthnResponse: %+v", wi)
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
	email := firstNonempty(attr.Get("email"), attr.Get("emailaddress"), attr.Get("http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"), attr.Get("http://schemas.xmlsoap.org/claims/EmailAddress"))
	if email == "" && mightBeEmail(assertions.NameID) {
		email = assertions.NameID
	}
	if pn := attr.Get("eduPersonPrincipalName"); email == "" && mightBeEmail(pn) {
		email = pn
	}
	groupsAttr := "groups"
	if p.config.GroupsAttributeName != "" {
		groupsAttr = p.config.GroupsAttributeName
	}
	info := authnResponseInfo{
		spec: extsvc.AccountSpec{
			ServiceType: providerType,
			ServiceID:   pi.ServiceID,
			ClientID:    pi.ClientID,
			AccountID:   assertions.NameID,
		},
		email:                email,
		unnormalizedUsername: firstNonempty(attr.Get("login"), attr.Get("uid"), attr.Get("username"), attr.Get("http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name"), email),
		displayName:          firstNonempty(attr.Get("displayName"), attr.Get("givenName")+" "+attr.Get("surname"), attr.Get("http://schemas.xmlsoap.org/claims/CommonName"), attr.Get("http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname")),
		groups:               attr.GetMap(groupsAttr),
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
func getOrCreateUser(ctx context.Context, db database.DB, allowSignup bool, info *authnResponseInfo) (newUserCreated bool, _ *actor.Actor, safeErrMsg string, err error) {
	var data extsvc.AccountData
	if err := SetExternalAccountData(&data, info); err != nil {
		return false, nil, "", err
	}

	username, err := auth.NormalizeUsername(info.unnormalizedUsername)
	if err != nil {
		return false, nil, fmt.Sprintf("Error normalizing the username %q. See https://docs.sourcegraph.com/admin/auth/#username-normalization.", info.unnormalizedUsername), err
	}

	newUserCreated, userID, safeErrMsg, err := auth.GetAndSaveUser(ctx, db, auth.GetAndSaveUserOp{
		UserProps: database.NewUser{
			Username:        username,
			Email:           info.email,
			EmailIsVerified: info.email != "", // SAML emails are assumed to be verified
			DisplayName:     info.displayName,
			// SAML has no standard way of providing an avatar URL.
		},
		ExternalAccount:     info.spec,
		ExternalAccountData: data,
		CreateIfNotExist:    allowSignup,
	})
	if err != nil {
		return false, nil, safeErrMsg, err
	}
	return newUserCreated, actor.FromUser(userID), "", nil
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

func (v samlAssertionValues) GetMap(key string) map[string]bool {
	for _, a := range v {
		if a.Name == key || a.FriendlyName == key {
			output := make(map[string]bool)
			for _, v := range a.Values {
				output[v.Value] = true
			}
			return output
		}
	}
	return nil
}

type SAMLValues struct {
	Values map[string]SAMLAttribute `json:"Values,omitempty"`
}

type SAMLAttribute struct {
	Values []SAMLValue `json:"Values"`
}

type SAMLValue struct {
	Value string
}

// GetExternalAccountData returns the deserialized JSON blob from user external accounts table
func GetExternalAccountData(ctx context.Context, data *extsvc.AccountData) (val *SAMLValues, err error) {
	if data.Data != nil {
		val, err = encryption.DecryptJSON[SAMLValues](ctx, data.Data)
		if err != nil {
			return nil, err
		}
	}
	if val == nil {
		return nil, errors.New("could not find data for the external account")
	}

	return val, nil
}

func GetPublicExternalAccountData(ctx context.Context, accountData *extsvc.AccountData) (*extsvc.PublicAccountData, error) {
	data, err := GetExternalAccountData(ctx, accountData)
	if err != nil {
		return nil, err
	}

	values := data.Values
	if values == nil {
		return nil, errors.New("could not find data values for external account")
	}

	// convert keys to lower case for case insensitive matching of candidates
	lowerCaseValues := make(map[string]SAMLAttribute, len(values))
	for k, v := range values {
		lowerCaseValues[strings.ToLower(k)] = v
	}

	var displayName string
	// all candidates are lower case
	candidates := []string{
		"nickname",
		"login",
		"username",
		"name",
		"http://schemas.xmlsoap.org/claims/name",
		"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name",
		"email",
		"emailaddress",
		"http://schemas.xmlsoap.org/claims/emailaddress",
		"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress",
	}
	for _, key := range candidates {
		candidate, ok := lowerCaseValues[key]
		if ok && len(candidate.Values) > 0 && candidate.Values[0].Value != "" {
			displayName = candidate.Values[0].Value
			break
		}
	}
	if displayName == "" {
		return nil, nil
	}
	return &extsvc.PublicAccountData{
		DisplayName: displayName,
	}, nil
}

// SetExternalAccountData sets the user and token into the external account data blob.
func SetExternalAccountData(data *extsvc.AccountData, info *authnResponseInfo) error {
	// TODO: leverage the whole info object instead of just storing JSON blob without any structure
	serializedData, err := json.Marshal(info.accountData)
	if err != nil {
		return err
	}

	data.Data = extsvc.NewUnencryptedData(serializedData)
	return nil
}
