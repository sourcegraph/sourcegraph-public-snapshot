package graphqlbackend

import (
	"context"

	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
)

var mockAllEmailsForOrg func(ctx context.Context, orgID int32, excludeByUserID []int32) ([]string, error)

func allEmailsForOrg(ctx context.Context, orgID int32, excludeByUserID []int32) ([]string, error) {
	if mockAllEmailsForOrg != nil {
		return mockAllEmailsForOrg(ctx, orgID, excludeByUserID)
	}

	members, err := db.OrgMembers.GetByOrgID(ctx, orgID)
	if err != nil {
		return nil, err
	}
	exclude := make(map[int32]interface{})
	for _, id := range excludeByUserID {
		exclude[id] = struct{}{}
	}
	emails := []string{}
	for _, m := range members {
		if _, ok := exclude[m.UserID]; ok {
			continue
		}
		email, _, err := db.UserEmails.GetEmail(ctx, m.UserID)
		if err != nil {
			// This shouldn't happen, but we don't want to prevent the notification,
			// so swallow the error.
			log15.Error("get user", "uid", m.UserID, "error", err)
			continue
		}
		emails = append(emails, email)
	}
	return emails, nil
}
