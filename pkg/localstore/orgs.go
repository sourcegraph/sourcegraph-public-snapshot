package localstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

type orgs struct{}

type OrgID int32

const NoOrg OrgID = 0

// AllOrgsFromUID returns a list of all organizations for the user represented by a UID. An empty slice is
// returned if the user is not authenticated or is not a member of any org.
func (*orgs) AllOrgsFromUID(UID string) ([]*sourcegraph.Org, error) {
	rows, err := globalDB.Query("SELECT orgs.id, orgs.name, orgs.created_at, orgs.updated_at FROM org_members LEFT OUTER JOIN orgs ON org_members.org_id = orgs.id WHERE user_id=$1", UID)
	if err != nil {
		if err == sql.ErrNoRows {
			return []*sourcegraph.Org{}, nil
		}
		return []*sourcegraph.Org{}, err
	}

	orgs := []*sourcegraph.Org{}
	defer rows.Close()
	for rows.Next() {
		org := sourcegraph.Org{}
		err := rows.Scan(&org.ID, &org.Name, &org.CreatedAt, &org.UpdatedAt)
		if err != nil {
			return nil, err
		}

		orgs = append(orgs, &org)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orgs, nil
}

// CurrentUserIsMember returns a boolean indicating if the current user is member of the given
// organization.
func (*orgs) CurrentUserIsMember(ctx context.Context, org OrgID) (bool, error) {
	a := actor.FromContext(ctx)
	if !a.IsAuthenticated() {
		return false, nil
	}

	if err := globalDB.QueryRow("SELECT FROM org_members WHERE user_id=$1 AND org_id=$2 LIMIT 1", a.UID, org).Scan(); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func validateOrg(org sourcegraph.Org) error {
	if org.Name == "" {
		return errors.New("error creating org: name required")
	}
	return nil
}

func (o *orgs) GetByID(ctx context.Context, OrgID int32) (*sourcegraph.Org, error) {
	orgs, err := o.getBySQL(ctx, "WHERE id=$1 LIMIT 1", OrgID)
	if err != nil {
		return nil, err
	}
	if len(orgs) == 0 {
		return nil, fmt.Errorf("org %d not found", OrgID)
	}
	return orgs[0], nil
}

func (*orgs) getBySQL(ctx context.Context, query string, args ...interface{}) ([]*sourcegraph.Org, error) {
	rows, err := globalDB.Query("SELECT id, name, created_at, updated_at FROM orgs "+query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	orgs := []*sourcegraph.Org{}
	defer rows.Close()
	for rows.Next() {
		org := sourcegraph.Org{}
		err := rows.Scan(&org.ID, &org.Name, &org.CreatedAt, &org.UpdatedAt)
		if err != nil {
			return nil, err
		}

		orgs = append(orgs, &org)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return orgs, nil
}

func (*orgs) Create(ctx context.Context, name string) (*sourcegraph.Org, error) {
	newOrg := sourcegraph.Org{
		Name: name,
	}
	newOrg.CreatedAt = time.Now()
	newOrg.UpdatedAt = newOrg.CreatedAt
	err := validateOrg(newOrg)
	if err != nil {
		return nil, err
	}
	err = globalDB.QueryRow(
		"INSERT INTO orgs(name, created_at, updated_at) VALUES($1, $2, $3) RETURNING id",
		newOrg.Name, newOrg.CreatedAt, newOrg.UpdatedAt).Scan(&newOrg.ID)
	if err != nil {
		if strings.Contains(err.Error(), "org_name_valid_chars") {
			return nil, errors.New(`error creating org: name must only contain alphanumeric characters, "_", and "-"`)
		}
		return nil, err
	}

	return &newOrg, nil
}
