package userlimitchecker

import (
	"context"
	"fmt"
	"time"

	ps "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/dotcom/productsubscription"
	lc "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/dotcom/userlimitchecker"

	"github.com/sourcegraph/sourcegraph/internal/api/internalapi"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/txemail"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// send email to site admins if approaching user limit on active license
func SendApproachingUserLimitAlert(ctx context.Context, db database.DB) error {
	licenseID, err := getActiveLicense(ctx, db)
	if err != nil {
		return errors.Wrap(err, ACTIVE_LICENSE_ERR)
	}

	checkerStore := lc.NewUserLimitChecker(db)
	c, err := checkerStore.GetByLicenseID(ctx, licenseID)
	if err != nil {
		return errors.Wrap(err, ACTIVE_CHECKER_ERR)
	}

	userCount, err := getUserCount(ctx, db)
	if err != nil {
		return errors.Wrap(err, USER_COUNT_ERR)
	}

	userLimit, err := getLicenseUserLimit(ctx, db)
	if err != nil {
		return errors.Wrap(err, USER_LIMIT_ERR)
	}

	if userCountWithinLimit(userCount, userLimit) {
		fmt.Println(WITHIN_LIMIT_MSG)
		return nil
	}

	if emailRecentlySent(c.UserCountAlertSentAt) && !userCountIncreased(userCount, c.UserCountWhenEmailLastSent) {
		fmt.Println(EMAIL_RECENTLY_SENT_MSG)
		return nil
	}

	siteAdminEmails, err := getVerifiedSiteAdminEmails(ctx, db)
	if err != nil {
		return errors.Wrap(err, ADMIN_EMAILS_ERR)
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
			RemainingUsers: userLimit - userCount,
			Percent:        getPercentage(userCount, userLimit),
		},
	}); err != nil {
		return errors.Wrap(err, EMAIL_SEND_ERR)
	}

	updateLicenseUserLimitCheckerFields(ctx, db, c.ID)
	return nil
}

func getActiveLicense(ctx context.Context, db database.DB) (string, error) {
	licenseStore := ps.NewDbLicense(db)
	licenses, err := licenseStore.List(ctx, ps.DbLicencesListNoOpt())
	if err != nil {
		return "", errors.Wrap(err, LICENSES_ERR)
	}

	for _, l := range licenses {
		if l.RevokedAt == nil {
			return l.ID, nil
		}
	}

	return "", errors.Wrap(err, NO_LICENSE_ERR)
}

func userCountWithinLimit(count int, limit int) bool {
	limitUsed := getPercentage(count, limit)
	if limitUsed < 90 && count <= limit-2 {
		return true
	}
	return false
}

func userCountIncreased(currentUserCount int, lastUserCount int) bool {
	return currentUserCount > lastUserCount
}

func emailRecentlySent(lastSent *time.Time) bool {
	if lastSent == nil {
		return false
	}
	now := time.Now()
	return now.Before(lastSent.Add(7 * 24 * time.Hour))
}

func updateLicenseUserLimitCheckerFields(ctx context.Context, db database.DB, checkerId string) error {
	currentUserCount, err := getUserCount(ctx, db)
	if err != nil {
		return errors.Wrap(err, USER_COUNT_ERR)
	}

	checkerStore := lc.NewUserLimitChecker(db)
	err = checkerStore.Update(ctx, checkerId, currentUserCount)
	if err != nil {
		return errors.Wrap(err, USER_LIMIT_ERR)
	}
	return nil
}

func getPercentage(userCount, userLimit int) int {
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
	licenses, err := ps.NewDbLicense(db).List(ctx, ps.DbLicencesListNoOpt())
	if err != nil {
		return 0, errors.Wrap(err, LICENSES_ERR)
	}

	for _, l := range licenses {
		if l.LicenseExpiresAt.After(time.Now()) {
			if l.LicenseUserCount != nil {
				return *l.LicenseUserCount, nil
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
		return nil, errors.Wrap(err, USER_LIST_ERR)
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
