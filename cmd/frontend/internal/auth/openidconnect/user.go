package openidconnect

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/coreos/go-oidc"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/hubspot"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/hubspot/hubspotutil"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/telemetry"
	"github.com/sourcegraph/sourcegraph/internal/telemetry/telemetryrecorder"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

type ExternalAccountData struct {
	IDToken    oidc.IDToken  `json:"idToken"`
	UserInfo   oidc.UserInfo `json:"userInfo"`
	UserClaims userClaims    `json:"userClaims"`
}

// getOrCreateUser gets or creates a user account based on the OpenID Connect token. It returns the
// authenticated actor if successful; otherwise it returns a friendly error message (safeErrMsg)
// that is safe to display to users, and a non-nil err with lower-level error details.
func getOrCreateUser(
	ctx context.Context,
	logger log.Logger,
	db database.DB,
	p schema.OpenIDConnectAuthProvider,
	token *oauth2.Token,
	idToken *oidc.IDToken,
	userInfo *oidc.UserInfo,
	claims *userClaims,
	usernamePrefix string,
	userCreateEventProperties telemetry.EventMetadata,
	hubSpotProps *hubspot.ContactProperties,
) (newUserCreated bool, _ *actor.Actor, safeErrMsg string, err error) {
	if userInfo.Email == "" {
		return false, nil, "Only users with an email address may authenticate to Sourcegraph.", errors.New("no email address in claims")
	}
	if unverifiedEmail := claims.EmailVerified != nil && !*claims.EmailVerified; unverifiedEmail {
		// If the OP explicitly reports `"email_verified": false`, then reject the authentication
		// attempt. If undefined or true, then it will be allowed.
		return false, nil, fmt.Sprintf("Only users with verified email addresses may authenticate to Sourcegraph. The email address %q is not verified on the external authentication provider.", userInfo.Email), errors.Errorf("refusing unverified user email address %q", userInfo.Email)
	}

	login := getLogin(claims, userInfo)
	email := userInfo.Email
	displayName := getDisplayName(claims, login)

	if usernamePrefix != "" {
		login = usernamePrefix + login
	}
	login, err = auth.NormalizeUsername(login)
	if err != nil {
		return false, nil,
			fmt.Sprintf("Error normalizing the username %q. See https://sourcegraph.com/docs/admin/auth/#username-normalization.", login),
			errors.Wrap(err, "normalize username")
	}

	serializedToken, err := json.Marshal(token)
	if err != nil {
		return false, nil, "", err
	}
	serializedUser, err := json.Marshal(ExternalAccountData{
		IDToken:    *idToken,
		UserInfo:   *userInfo,
		UserClaims: *claims,
	})
	if err != nil {
		return false, nil, "", err
	}
	data := extsvc.AccountData{
		AuthData: extsvc.NewUnencryptedData(serializedToken),
		Data:     extsvc.NewUnencryptedData(serializedUser),
	}

	recorder := telemetryrecorder.New(db)
	newUserCreated, userID, safeErrMsg, err := auth.GetAndSaveUser(ctx, logger, db, recorder, auth.GetAndSaveUserOp{
		UserProps: database.NewUser{
			Username:        login,
			Email:           email,
			EmailIsVerified: email != "", // verified email check is at the top of the function
			DisplayName:     displayName,
			AvatarURL:       claims.Picture,
		},
		ExternalAccount: extsvc.AccountSpec{
			ServiceType: p.Type,
			ServiceID:   p.Issuer,
			ClientID:    p.ClientID,
			AccountID:   idToken.Subject,
		},
		UserCreateEventProperties: userCreateEventProperties,
		ExternalAccountData:       data,
		CreateIfNotExist:          p.AllowSignup == nil || *p.AllowSignup,
		SingleIdentityPerUser:     p.SingleIdentityPerUser,
	})
	if err != nil {
		return false, nil, safeErrMsg, err
	}
	go hubspotutil.SyncUser(email, hubspotutil.SignupEventID, hubSpotProps)
	return newUserCreated, actor.FromUser(userID), "", nil
}

// GetExternalAccountData returns the deserialized JSON blob from user external accounts table
func GetExternalAccountData(ctx context.Context, data *extsvc.AccountData) (val *ExternalAccountData, err error) {
	if data.Data != nil {
		val, err = encryption.DecryptJSON[ExternalAccountData](ctx, data.Data)
		if err != nil {
			return nil, err
		}
	}

	return val, nil
}

func GetPublicExternalAccountData(ctx context.Context, accountData *extsvc.AccountData) (*extsvc.PublicAccountData, error) {
	data, err := GetExternalAccountData(ctx, accountData)
	if err != nil {
		return nil, err
	}

	login := data.UserClaims.PreferredUsername
	if login == "" {
		login = data.UserInfo.Email
	}
	displayName := data.UserClaims.GivenName
	if displayName == "" {
		if data.UserClaims.Name == "" {
			displayName = data.UserClaims.Name
		} else {
			displayName = login
		}
	}

	return &extsvc.PublicAccountData{
		Login:       login,
		DisplayName: displayName,
		URL:         data.UserInfo.Profile,
	}, nil
}
