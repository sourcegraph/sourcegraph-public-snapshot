package backend

import (
	"context"
	"net/url"
	"strconv"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"github.com/sourcegraph/sourcegraph/pkg/randstring"
)

func MakeRandomHardToGuessPassword() string {
	return randstring.NewLen(36)
}

func MakePasswordResetURL(ctx context.Context, userID int32) (*url.URL, error) {
	resetCode, err := db.Users.RenewPasswordResetCode(ctx, userID)
	if err != nil {
		return nil, err
	}
	query := url.Values{}
	query.Set("userID", strconv.Itoa(int(userID)))
	query.Set("code", resetCode)
	return &url.URL{Path: "/password-reset", RawQuery: query.Encode()}, nil
}
