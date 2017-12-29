package backend

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/url"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/randstring"
)

var Users = &users{}

type users struct{}

func (s *users) List(ctx context.Context) (res *sourcegraph.UserList, err error) {
	if Mocks.Users.List != nil {
		return Mocks.Users.List(ctx)
	}

	// 🚨 SECURITY:  only admins are allowed to use this endpoint
	if err := CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}
	users, err := localstore.Users.List(ctx)
	if err != nil {
		return nil, err
	}
	return &sourcegraph.UserList{Users: users}, nil
}

func NativeAuthUserAuthID(email string) string {
	return fmt.Sprintf("%s:%s", sourcegraph.UserProviderNative, email)
}

func MakeEmailVerificationCode() string {
	emailCodeBytes := make([]byte, 20)
	if _, err := rand.Read(emailCodeBytes); err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(emailCodeBytes)
}

func MakeRandomHardToGuessPassword() string {
	return randstring.NewLen(36)
}

func MakePasswordResetURL(ctx context.Context, userID int32, userEmail string) (*url.URL, error) {
	resetCode, err := localstore.Users.RenewPasswordResetCode(ctx, userID)
	if err != nil {
		return nil, err
	}
	query := url.Values{}
	query.Set("email", userEmail)
	query.Set("code", resetCode)
	return &url.URL{Path: "/password-reset", RawQuery: query.Encode()}, nil
}
