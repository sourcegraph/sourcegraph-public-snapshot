package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

type externalAccountDataResolver struct {
	data *extsvc.PublicAccountData
}

func NewExternalAccountDataResolver(ctx context.Context, encryptedData *encryption.JSONEncryptable[any], serviceType string) (*externalAccountDataResolver, error) {
	data, err := publicAccountDataFromJSON(ctx, encryptedData, serviceType)
	if err != nil {
		return nil, err
	}
	return &externalAccountDataResolver{
		data: data,
	}, nil
}

type GitHubAccountData struct {
	Username *string `json:"name,omitempty"`
	Login    *string `json:"login,omitempty"`
	URL      *string `json:"html_url,omitempty"`
}

type GitLabAccountData struct {
	Username *string `json:"name,omitempty"`
	Login    *string `json:"username,omitempty"`
	URL      *string `json:"web_url,omitempty"`
}

type SAMLValues struct {
	Values *SAMLAccountData `json:"Values,omitempty"`
}

// SAML actually uses `http://schemas.xmlsoap.org...` as JSON keys for attributes
// And we need to check multiple different attributes to find a probable username
type SAMLAccountData struct {
	Nickname  *SAMLAttribute `json:"nickname,omitempty"`
	Login     *SAMLAttribute `json:"login,omitempty"`
	Username1 *SAMLAttribute `json:"username,omitempty"`
	Username2 *SAMLAttribute `json:"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name,omitempty"`
	Email1    *SAMLAttribute `json:"emailaddress,omitempty"`
	Email2    *SAMLAttribute `json:"http://schemas.xmlsoap.org/claims/EmailAddress,omitempty"`
	Email3    *SAMLAttribute `json:"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress,omitempty"`
}

type SAMLAttribute struct {
	Values []*SAMLValue `json:"Values"`
}

type SAMLValue struct {
	Value string
}

type OpenIDConnectAccountData struct {
	UserClaims *OpenIDUserClaims `json:"userClaims,omitempty"`
	UserInfo   *OpenIDUserInfo   `json:"userInfo,omitempty"`
}

type OpenIDUserClaims struct {
	Username  *string `json:"preferred_username,omitempty"`
	GivenName *string `json:"given_name,omitempty"`
	Name      *string `json:"name,omitempty"`
}

type OpenIDUserInfo struct {
	Email *string `json:"email,omitempty"`
}

func publicAccountDataFromJSON(ctx context.Context, val *encryption.JSONEncryptable[any], serviceType string) (*extsvc.PublicAccountData, error) {
	// should we support Gerrit or other auth types in here?
	switch serviceType {
	case extsvc.TypeGitHub:
		data, err := encryption.DecryptJSON[GitHubAccountData](ctx, val)
		if err != nil {
			return nil, err
		}
		casted := extsvc.PublicAccountData(*data)
		return &casted, nil
	case extsvc.TypeGitLab:
		data, err := encryption.DecryptJSON[GitLabAccountData](ctx, val)
		if err != nil {
			return nil, err
		}
		casted := extsvc.PublicAccountData(*data)
		return &casted, nil
	case "saml": //TODO: define constants for SAML and OpenID Connect service types
		data, err := encryption.DecryptJSON[SAMLValues](ctx, val)
		if err != nil {
			return nil, err
		}
		v := data.Values
		if v == nil {
			return nil, nil
		}

		var username string
		candidates := []*SAMLAttribute{v.Nickname, v.Login, v.Username1, v.Username2, v.Email1, v.Email2, v.Email3}
		for _, c := range candidates {
			if c != nil && len(c.Values) > 0 && c.Values[0] != nil && c.Values[0].Value != "" {
				username = c.Values[0].Value
				break
			}
		}
		if username == "" {
			return nil, nil
		}
		return &extsvc.PublicAccountData{
			Username: &username,
		}, nil
	case "openidconnect":
		data, err := encryption.DecryptJSON[OpenIDConnectAccountData](ctx, val)
		if err != nil {
			return nil, err
		}

		var username string
		candidates := []*string{data.UserClaims.Username, data.UserClaims.GivenName, data.UserClaims.Name, data.UserInfo.Email}
		for _, value := range candidates {
			if value != nil && *value != "" {
				username = *value
				break
			}
		}
		if username == "" {
			return nil, nil
		}
		return &extsvc.PublicAccountData{
			Username: &username,
		}, nil
	}

	return nil, nil
}

func (r *externalAccountDataResolver) Username() *string {
	return r.data.Username
}

func (r *externalAccountDataResolver) Login() *string {
	return r.data.Login
}

func (r *externalAccountDataResolver) URL() *string {
	return r.data.URL
}
