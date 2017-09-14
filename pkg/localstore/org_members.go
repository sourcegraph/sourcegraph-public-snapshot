package localstore

import (
	"context"
	"time"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

type orgMembers struct{}

func (*orgMembers) Create(ctx context.Context, orgID int, userID, username, email string) (*sourcegraph.OrgMember, error) {
	newMember := sourcegraph.OrgMember{
		OrgID:    int32(orgID),
		UserID:   userID,
		Username: username,
		Email:    email,
	}
	newMember.CreatedAt = time.Now()
	newMember.UpdatedAt = newMember.CreatedAt
	err := globalDB.QueryRow(
		"INSERT INTO org_members(org_id, user_id, user_email, user_name, created_at, updated_at) VALUES($1, $2, $3, $4, $5, $6) RETURNING id",
		newMember.OrgID, newMember.UserID, newMember.Email, newMember.Username, newMember.CreatedAt, newMember.UpdatedAt).Scan(&newMember.ID)
	if err != nil {
		return nil, err
	}

	return &newMember, nil
}

func (*orgMembers) Remove(ctx context.Context, orgID int, userID string) error {
	_, err := globalDB.Exec("DELETE FROM org_members WHERE (org_id=$1 AND user_id=$2)", orgID, userID)
	return err
}
