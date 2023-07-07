package userlimitchecker

import (
	"context"

	ps "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/dotcom/productsubscription"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/txemail"
	"github.com/sourcegraph/sourcegraph/internal/txemail/txtypes"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var approachingUserLimitEmailTemplate = txemail.MustValidate(txtypes.Templates{
	Subject: `Your user count is approaching your license's limit`,
	Text: `
Hi there! You're approaching the user limit allowed by your Sourcegraph License. You currently have {{.RemainingUsers}} left.

Reach out to your rep at Sourcegraph if you'd like to increase the limit.
`,
	HTML: `
<p>
Hi there! You're approaching the user limit allowed by your Sourcegraph License. You currently have {{.RemainingUsers}} left.
</p>
<p>Reach out to your rep at Sourcegraph if you'd like to increase the limit.</p>
`,
})

// function to send email if approaching
func sendApproachingUserLimitAlert(ctx context.Context, db database.DB) error {
	atOrOverUserLimit, err := atOrOverUserLimit(ctx, db)
	if err != nil {
		return errors.Wrap(err, "could not check user limit")
	}

	siteAdminEmails, err := getSiteAdmins(ctx, db)
	if err != nil {
		return errors.Wrap(err, "could not get site admins")
	}

	if atOrOverUserLimit {
		if err := txemail.Send(ctx, "approaching_user_limit", txemail.Message{
			To:       siteAdminEmails,
			Template: approachingUserLimitEmailTemplate,
			Data: struct {
				FromName string
				Message  string
			}{
				FromName: "Jason Hawk Harris",
				Message:  "This is a test email",
			},
		}); err != nil {
			return errors.Wrap(err, "could not send email")
		}
	}

	return nil
}

// function to check whether the user count is approaching the license user limit
func atOrOverUserLimit(ctx context.Context, db database.DB) (bool, error) {
	userCount, err := getUserCount(ctx, db)
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
func getUserCount(ctx context.Context, db database.DB) (int, error) {
	userStore := db.Users()
	userCount, err := userStore.Count(ctx, &database.UsersListOptions{})
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

func getSiteAdmins(ctx context.Context, db database.DB) ([]string, error) {
	var siteAdminEmails []string

	userStore := db.Users()
	users, err := userStore.List(ctx, &database.UsersListOptions{})
	if err != nil {
		return siteAdminEmails, err
	}

	for _, user := range users {
		if user.SiteAdmin {
			email, _, err := getUserEmail(ctx, db, user)
			if err != nil {
				return siteAdminEmails, err
			}

			siteAdminEmails = append(siteAdminEmails, email)
		}
	}
	return siteAdminEmails, nil
}

func getUserEmail(ctx context.Context, db database.DB, u *types.User) (string, bool, error) {
	return database.UserEmailsWith(db).GetPrimaryEmail(ctx, u.ID)
}
