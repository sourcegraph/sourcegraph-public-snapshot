package graphqlbackend

import (
	"context"
	"encoding/json"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

type externalAccountResolver struct {
	db      database.DB
	account extsvc.Account
}

func externalAccountByID(ctx context.Context, db database.DB, id graphql.ID) (*externalAccountResolver, error) {
	externalAccountID, err := unmarshalExternalAccountID(id)
	if err != nil {
		return nil, err
	}
	account, err := db.UserExternalAccounts().Get(ctx, externalAccountID)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Only the user and site admins should be able to see a user's external accounts.
	if err := auth.CheckSiteAdminOrSameUser(ctx, db, account.UserID); err != nil {
		return nil, err
	}

	return &externalAccountResolver{db: db, account: *account}, nil
}

func marshalExternalAccountID(repo int32) graphql.ID { return relay.MarshalID("ExternalAccount", repo) }

func unmarshalExternalAccountID(id graphql.ID) (externalAccountID int32, err error) {
	err = relay.UnmarshalSpec(id, &externalAccountID)
	return
}

func (r *externalAccountResolver) ID() graphql.ID { return marshalExternalAccountID(r.account.ID) }
func (r *externalAccountResolver) User(ctx context.Context) (*UserResolver, error) {
	return UserByIDInt32(ctx, r.db, r.account.UserID)
}
func (r *externalAccountResolver) ServiceType() string { return r.account.ServiceType }
func (r *externalAccountResolver) ServiceID() string   { return r.account.ServiceID }
func (r *externalAccountResolver) ClientID() string    { return r.account.ClientID }
func (r *externalAccountResolver) AccountID() string   { return r.account.AccountID }
func (r *externalAccountResolver) CreatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.account.CreatedAt}
}
func (r *externalAccountResolver) UpdatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.account.UpdatedAt}
}

func (r *externalAccountResolver) RefreshURL() *string {
	// TODO(sqs): Not supported.
	return nil
}

func (r *externalAccountResolver) AccountData(ctx context.Context) (*JSONValue, error) {
	// 🚨 SECURITY: It is only safe to assume account data of GitHub and GitLab do
	// not contain sensitive information that is not known to the user (which is
	// accessible via APIs by users themselves). We cannot take the same assumption
	// for other types of external accounts.
	//
	// Therefore, the site admins and the user can view account data of GitHub and
	// GitLab, but only site admins can view account data for all other types.
	var err error
	if r.account.ServiceType == extsvc.TypeGitHub || r.account.ServiceType == extsvc.TypeGitLab {
		err = auth.CheckSiteAdminOrSameUser(ctx, r.db, actor.FromContext(ctx).UID)
	} else {
		err = auth.CheckUserIsSiteAdmin(ctx, r.db, actor.FromContext(ctx).UID)
	}
	if err != nil {
		return nil, err
	}

	if r.account.Data != nil {
		raw, err := r.account.Data.Decrypt(ctx)
		if err != nil {
			return nil, err
		}

		return &JSONValue{raw}, nil
	}
	return nil, nil
}

func (r *externalAccountResolver) PublicAccountData(ctx context.Context) (*publicExternalAccountDataResolver, error) {
	// 🚨 SECURITY: We only return this data to site admin or user who is linked to the external account
	// This method differs from the one above - here we only return specific attributes
	// from the account that are public info, e.g. username, email, etc.
	err := auth.CheckSiteAdminOrSameUser(ctx, r.db, actor.FromContext(ctx).UID)
	if err != nil {
		return nil, err
	}

	if r.account.Data != nil {
		raw, err := r.account.Data.Decrypt(ctx)
		if err != nil {
			return nil, err
		}
		jsonValue := &JSONValue{raw}
		// I could not find other way to convert the raw decrypted map[string]interface{} back to JSON string data
		bytes, err := jsonValue.MarshalJSON()
		if err != nil {
			return nil, err
		}
		data, err := publicAccountDataFromJSON(bytes, r.account.ServiceType)
		if err != nil {
			return nil, err
		}

		return &publicExternalAccountDataResolver{
			data: data,
		}, nil
	}

	return nil, nil
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

// this looks ugly, because SAML actually uses `http://schemas.xmlsoap.org...` as JSON
// keys for attributes. And we need to check multiple different attributes
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

func publicAccountDataFromJSON(val []byte, serviceType string) (*extsvc.PublicAccountData, error) {
	// TODO: do we need Gerrit or other things in here?
	switch serviceType {
	case extsvc.TypeGitHub:
		m := &GitHubAccountData{}
		err := json.Unmarshal(val, &m)
		if err != nil {
			return nil, err
		}
		return &extsvc.PublicAccountData{
			UserName: m.Username,
			Login:    m.Login,
			URL:      m.URL,
		}, nil
	case extsvc.TypeGitLab:
		m := &GitLabAccountData{}
		err := json.Unmarshal(val, &m)
		if err != nil {
			return nil, err
		}
		return &extsvc.PublicAccountData{
			UserName: m.Username,
			Login:    m.Login,
			URL:      m.URL,
		}, nil
	case "saml": //TODO: define constants for SAML and OpenID Connect service types
		m := &SAMLValues{}
		err := json.Unmarshal(val, &m)
		if err != nil {
			return nil, err
		}
		var username *string
		v := m.Values
		if v == nil {
			return nil, nil
		}
		assignIfDefined := func(data []*SAMLAttribute) bool {
			assigned := false
			for _, d := range data {
				if d != nil && len(d.Values) > 0 && d.Values[0] != nil && d.Values[0].Value != "" {
					*username = d.Values[0].Value
					assigned = true
					break
				}
			}
			return assigned
		}
		if !assignIfDefined([]*SAMLAttribute{v.Nickname, v.Login, v.Username1, v.Username2, v.Email1, v.Email2, v.Email3}) {
			return nil, nil
		}
		return &extsvc.PublicAccountData{
			UserName: username,
		}, nil
	case "openidconnect":
		m := &OpenIDConnectAccountData{}
		err := json.Unmarshal(val, &m)
		if err != nil {
			return nil, err
		}

		var username *string
		assignIfDefined := func(data []*string) bool {
			assigned := false
			for _, value := range data {
				if value != nil && *value != "" {
					username = value
					assigned = true
					break
				}
			}
			return assigned
		}
		if !assignIfDefined([]*string{m.UserClaims.Username, m.UserClaims.GivenName, m.UserClaims.Name, m.UserInfo.Email}) {
			return nil, nil
		}
		return &extsvc.PublicAccountData{
			UserName: username,
		}, nil
	}

	return nil, nil
}

type publicExternalAccountDataResolver struct {
	data *extsvc.PublicAccountData
}

func (r *publicExternalAccountDataResolver) Username() *string {
	return r.data.UserName
}

func (r *publicExternalAccountDataResolver) Login() *string {
	return r.data.Login
}

func (r *publicExternalAccountDataResolver) URL() *string {
	return r.data.URL
}
