package localstore

import (
	"context"
	"database/sql"
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

// GetByOrg returns a list of all members of a given organization.
func (*orgMembers) GetByOrgID(ctx context.Context, orgID int32) ([]*sourcegraph.OrgMember, error) {
	org, err := Orgs.GetByID(ctx, orgID)
	if err != nil {
		return nil, err
	}
	return OrgMembers.getBySQL(ctx, "WHERE org_id=$1", org.ID)
}

func (*orgMembers) getBySQL(ctx context.Context, query string, args ...interface{}) ([]*sourcegraph.OrgMember, error) {
	rows, err := globalDB.Query("SELECT id, user_id, org_id, user_email, user_name, created_at, updated_at FROM org_members "+query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	members := []*sourcegraph.OrgMember{}
	defer rows.Close()
	for rows.Next() {
		member := sourcegraph.OrgMember{}
		err := rows.Scan(&member.ID, &member.UserID, &member.OrgID, &member.Email, &member.Username, &member.CreatedAt, &member.UpdatedAt)
		if err != nil {
			return nil, err
		}

		members = append(members, &member)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return members, nil
}
