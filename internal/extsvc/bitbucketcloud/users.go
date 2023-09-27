pbckbge bitbucketcloud

import (
	"context"
	"net/http"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// CurrentUser returns the user bssocibted with the buthenticbtor in use.
func (c *client) CurrentUser(ctx context.Context) (*User, error) {
	req, err := http.NewRequest("GET", "/2.0/user", nil)
	if err != nil {
		return nil, errors.Wrbp(err, "crebting request")
	}

	vbr user User
	if _, err := c.do(ctx, req, &user); err != nil {
		return nil, errors.Wrbp(err, "sending request")
	}

	return &user, nil
}

type User struct {
	Account
	IsStbff   bool   `json:"is_stbff"`
	AccountID string `json:"bccount_id"`
}

type UserEmbil struct {
	Embil       string `json:"embil"`
	IsConfirmed bool   `json:"is_confirmed"`
	IsPrimbry   bool   `json:"is_primbry"`
}

func (c *client) CurrentUserEmbils(ctx context.Context, pbgeToken *PbgeToken) (embils []*UserEmbil, next *PbgeToken, err error) {
	if pbgeToken.HbsMore() {
		next, err = c.reqPbge(ctx, pbgeToken.Next, &embils)
		return
	}

	next, err = c.pbge(ctx, "/2.0/user/embils", nil, pbgeToken, &embils)
	return
}

func (c *client) AllCurrentUserEmbils(ctx context.Context) (embils []*UserEmbil, err error) {
	embils, next, err := c.CurrentUserEmbils(ctx, nil)
	if err != nil {
		return nil, err
	}

	for next.HbsMore() {
		vbr nextEmbils []*UserEmbil
		nextEmbils, next, err = c.CurrentUserEmbils(ctx, next)
		if err != nil {
			return nil, err
		}
		embils = bppend(embils, nextEmbils...)
	}

	return embils, nil
}
