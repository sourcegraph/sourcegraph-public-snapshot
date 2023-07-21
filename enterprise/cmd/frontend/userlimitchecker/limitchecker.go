package userlimitchecker

import (
	"context"
	"fmt"
	"time"

	ps "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/dotcom/productsubscription"
	"github.com/sourcegraph/sourcegraph/internal/api/internalapi"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/txemail"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// send email to site admins if approaching user limit on active license
func SendApproachingUserLimitAlert(ctx context.Context, db database.DB) error {
	licenseDb := ps.NewDbLicense(db)
	licenses, err := licenseDb.List(ctx, ps.DbLicencesListNoOpt())
	if err != nil {
		return errors.Wrap(err, "could not get list of db licenses")
	}

	var licenseID string
	for _, license := range licenses {
		if license.RevokedAt == nil {
			licenseID = license.ID
			break
		}
	}

	lastAlertSentAt, err := licenseDb.GetUserCountAlertSentAt(ctx, licenseID)
	if err != nil {
		return errors.Wrap(err, "could not get last time user account alert was sent")
	}

	// check if alert was recently sent
	if !time.Now().After(lastAlertSentAt.Add(7 * 24 * time.Hour)) {
		fmt.Println("email recently sent")
		return nil
	}

	currentUserCount, err := getUserCount(ctx, db)
	if err != nil {
		return errors.Wrap(err, "could not get user count")
	}

	currentUserLimit, err := getLicenseUserLimit(ctx, db)
	if err != nil {
		return errors.Wrap(err, "could not get license user limit")
	}

	percentOfLimitUsed := getPercentOfLimit(currentUserCount, currentUserLimit)
	if percentOfLimitUsed < 90 && currentUserCount < currentUserLimit-2 {
		fmt.Println("user count on license within limit")
		return nil
	}

	siteAdminEmails, err := getVerifiedSiteAdminEmails(ctx, db)
	if err != nil {
		return errors.Wrap(err, "could not get site admins")
	}

	messageId := "approaching_user_limit"
	replyTo := "support@sourcegraph.com"

	if err := internalapi.Client.SendEmail(ctx, "approaching_user_limit", txemail.Message{
		To:        siteAdminEmails,
		Template:  approachingUserLimitEmailTemplate,
		MessageID: &messageId,
		ReplyTo:   &replyTo,
		Data: struct {
			RemainingUsers int
			Percent        int
		}{
			RemainingUsers: currentUserLimit - currentUserCount,
			Percent:        percentOfLimitUsed,
		},
	}); err != nil {
		return errors.Wrap(err, "could not send email")
	}

	return nil
}

func getPercentOfLimit(userCount, userLimit int) int {
	if userCount == 0 {
		return 0
	}
	if userLimit == 0 {
		return userCount + 100
	}
	return (userCount * 100) / userLimit
}

func getUserCount(ctx context.Context, db database.DB) (int, error) {
	userStore := db.Users()
	userCount, err := userStore.Count(ctx, &database.UsersListOptions{})
	if err != nil {
		return 0, err
	}
	return userCount, nil
}

func getLicenseUserLimit(ctx context.Context, db database.DB) (int, error) {
	items, err := ps.NewDbLicense(db).List(ctx, ps.DbLicencesListNoOpt())
	if err != nil {
		return 0, err
	}

	for _, item := range items {
		if item.LicenseExpiresAt.After(time.Now()) {
			if item.LicenseUserCount != nil {
				return *item.LicenseUserCount, nil
			} else {
				return 0, nil
			}
		}
	}
	return 0, nil
}

func getVerifiedSiteAdminEmails(ctx context.Context, db database.DB) ([]string, error) {
	var siteAdminEmails []string
	users, err := db.Users().List(ctx, &database.UsersListOptions{})
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if user.SiteAdmin {
			email, verified, err := getUserEmail(ctx, db, user)
			if err != nil {
				return nil, err
			}
			if verified {
				siteAdminEmails = append(siteAdminEmails, email)
			}
		}
	}
	return siteAdminEmails, nil
}

func getUserEmail(ctx context.Context, db database.DB, u *types.User) (string, bool, error) {
	return database.UserEmailsWith(db).GetPrimaryEmail(ctx, u.ID)
}
