package backend

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/url"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/randstring"
)

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
	resetCode, err := db.Users.RenewPasswordResetCode(ctx, userID)
	if err != nil {
		return nil, err
	}
	query := url.Values{}
	query.Set("email", userEmail)
	query.Set("code", resetCode)
	return &url.URL{Path: "/password-reset", RawQuery: query.Encode()}, nil
}
