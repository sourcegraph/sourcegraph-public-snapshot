package localstore

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

type orgs struct{}

// GetByUserID returns a list of all organizations for the user. An empty slice is
// returned if the user is not authenticated or is not a member of any org.
func (*orgs) GetByUserID(userID string) ([]*sourcegraph.Org, error) {
	rows, err := globalDB.Query("SELECT orgs.id, orgs.name, orgs.created_at, orgs.updated_at FROM org_members LEFT OUTER JOIN orgs ON org_members.org_id = orgs.id WHERE user_id=$1", userID)
	if err != nil {
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

func validateOrg(org sourcegraph.Org) error {
	if org.Name == "" {
		return errors.New("error creating org: name required")
	}
	return nil
}

func (o *orgs) GetByID(ctx context.Context, orgID int32) (*sourcegraph.Org, error) {
	if Mocks.Orgs.GetByID != nil {
		return Mocks.Orgs.GetByID(ctx, orgID)
	}
	orgs, err := o.getBySQL(ctx, "WHERE id=$1 LIMIT 1", orgID)
	if err != nil {
		return nil, err
	}
	if len(orgs) == 0 {
		return nil, fmt.Errorf("org %d not found", orgID)
	}
	return orgs[0], nil
}

func (*orgs) getBySQL(ctx context.Context, query string, args ...interface{}) ([]*sourcegraph.Org, error) {
	rows, err := globalDB.Query("SELECT id, name, created_at, updated_at FROM orgs "+query, args...)
	if err != nil {
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
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Constraint == "org_name_valid_chars" {
				return nil, errors.New(`org name invalid`)
			}
			if pqErr.Constraint == "org_name_unique" {
				return nil, errors.New(`org name already exists`)
			}
		}

		return nil, err
	}

	return &newOrg, nil
}
