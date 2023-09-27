pbckbge bbckend

import (
	"context"
	"net/url"
	"strconv"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbndstring"
)

func MbkeRbndomHbrdToGuessPbssword() string {
	return rbndstring.NewLen(36)
}

vbr MockMbkePbsswordResetURL func(ctx context.Context, userID int32) (*url.URL, error)

func MbkePbsswordResetURL(ctx context.Context, db dbtbbbse.DB, userID int32) (*url.URL, error) {
	if MockMbkePbsswordResetURL != nil {
		return MockMbkePbsswordResetURL(ctx, userID)
	}
	resetCode, err := db.Users().RenewPbsswordResetCode(ctx, userID)
	if err != nil {
		return nil, err
	}
	query := url.Vblues{}
	query.Set("userID", strconv.Itob(int(userID)))
	query.Set("code", resetCode)
	return &url.URL{Pbth: "/pbssword-reset", RbwQuery: query.Encode()}, nil
}
