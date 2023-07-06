package userlimitchecker

import (
	"context"
	"fmt"

	"github.com/sourcegraph/log"
	ps "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/dotcom/productsubscription"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// function to send email if approaching
func sendApproachingUserLimitAlert(ctx context.Context, db database.DB) error {
	atOrOverUserLimit, err := atOrOverUserLimit(ctx, db)
	if err != nil {
		return errors.Wrap(err, "could not check user limit")
	}

	if atOrOverUserLimit {
		// TODO replace with logic for sending email
		fmt.Println("Send email!")
	}
	return nil
}

// function to check whether the user count is approaching the license user limit
func atOrOverUserLimit(ctx context.Context, db database.DB) (bool, error) {
	userCount, err := getUserCount(ctx, log.NoOp())
	if err != nil {
		return false, err
	}

	userLimit, err := getUserLimit(ctx, db)
	if err != nil {
		return false, err
	}

	if userCount >= userLimit-5 {
		return true, nil
	}
	return false, nil
}

// function to get the user count
func getUserCount(ctx context.Context, logger log.Logger) (int, error) {
	userCount, err := database.Users(logger).Count(ctx, &database.UsersListOptions{})
	if err != nil {
		return 0, err
	}
	return userCount, nil
}

// function to get the license user count
func getUserLimit(ctx context.Context, db database.DB) (int, error) {
	items, err := ps.NewDbLicense(db).List(ctx, ps.DbLicencesListNoOpt())
	if err != nil {
		return 0, err
	}
	if len(items) == 0 {
		return 0, nil
	}

	// loop over all licenses and look for a RevokeReason == nil
	// we know this license is active, so return the user count
	// if loop finishes, we know there are no active licenses, return 0
	for _, item := range items {
		if item.RevokeReason != nil {
			return *item.LicenseUserCount, nil
		}
	}

	return 0, nil
}
