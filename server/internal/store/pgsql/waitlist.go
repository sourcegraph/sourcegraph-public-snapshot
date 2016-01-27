package pgsql

import (
	"strings"
	"time"

	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
)

type userWaitlistRow struct {
	UID       int32      `db:"uid"`
	AddedAt   *time.Time `db:"added_at"`
	GrantedAt *time.Time `db:"granted_at"`
}

type orgWaitlistRow struct {
	Name      string     `db:"name"`
	AddedAt   *time.Time `db:"added_at"`
	GrantedAt *time.Time `db:"granted_at"`
}

func init() {
	Schema.Map.AddTableWithName(userWaitlistRow{}, "user_waitlist").SetKeys(false, "UID")
	Schema.CreateSQL = append(Schema.CreateSQL,
		`ALTER TABLE user_waitlist ALTER COLUMN added_at TYPE timestamp with time zone USING added_at::timestamp with time zone;`,
		`ALTER TABLE user_waitlist ALTER COLUMN granted_at TYPE timestamp with time zone USING granted_at::timestamp with time zone;`,
	)

	Schema.Map.AddTableWithName(userWaitlistRow{}, "org_waitlist").SetKeys(false, "Name")
	Schema.CreateSQL = append(Schema.CreateSQL,
		`ALTER TABLE org_waitlist ALTER COLUMN added_at TYPE timestamp with time zone USING added_at::timestamp with time zone;`,
		`ALTER TABLE org_waitlist ALTER COLUMN granted_at TYPE timestamp with time zone USING granted_at::timestamp with time zone;`,
	)
}

// waitlist is a DB-backed implementation of the Waitlist store.
type waitlist struct{}

func (w *waitlist) AddUser(ctx context.Context, uid int32) error {
	currTime := time.Now()
	dbUser := &userWaitlistRow{
		UID:     uid,
		AddedAt: &currTime,
	}

	err := dbh(ctx).Insert(dbUser)
	if err != nil && strings.Contains(err.Error(), `duplicate key value violates unique constraint`) {
		return store.ErrWaitlistedUserExists
	}
	if err != nil {
		return err
	}
	return nil
}

func (w *waitlist) getUser(ctx context.Context, uid int32) (*userWaitlistRow, error) {
	var users []*userWaitlistRow
	err := dbh(ctx).Select(&users, "SELECT * FROM user_waitlist WHERE uid=$1 LIMIT 1", uid)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, &store.WaitlistedUserNotFoundError{}
	}
	return users[0], err
}

func (w *waitlist) GetUser(ctx context.Context, uid int32) (*sourcegraph.WaitlistedUser, error) {
	dbUser, err := w.getUser(ctx, uid)
	if err != nil {
		return nil, err
	}
	return &sourcegraph.WaitlistedUser{
		UID:       dbUser.UID,
		AddedAt:   ts(dbUser.AddedAt),
		GrantedAt: ts(dbUser.GrantedAt),
	}, err
}

func (w *waitlist) GrantUser(ctx context.Context, uid int32) error {
	dbUser, err := w.getUser(ctx, uid)
	if err != nil {
		return err
	}

	if dbUser.GrantedAt != nil {
		return nil
	}

	currTime := time.Now()
	dbUser.GrantedAt = &currTime
	_, err = dbh(ctx).Update(dbUser)
	return err
}

func (w *waitlist) ListUsers(ctx context.Context, onlyWaitlisted bool) ([]*sourcegraph.WaitlistedUser, error) {
	var userWaitlistRows []*userWaitlistRow
	sql := `SELECT * FROM user_waitlist`
	if onlyWaitlisted {
		sql += ` WHERE granted_at is null`
	}
	if err := dbh(ctx).Select(&userWaitlistRows, sql); err != nil {
		return nil, err
	}

	waitlistedUsers := make([]*sourcegraph.WaitlistedUser, len(userWaitlistRows))
	for i, row := range userWaitlistRows {
		waitlistedUsers[i] = &sourcegraph.WaitlistedUser{
			UID:       row.UID,
			AddedAt:   ts(row.AddedAt),
			GrantedAt: ts(row.GrantedAt),
		}
	}
	return waitlistedUsers, nil
}

func (w *waitlist) AddOrg(ctx context.Context, orgName string) error {
	currTime := time.Now()
	dbOrg := &orgWaitlistRow{
		Name:    orgName,
		AddedAt: &currTime,
	}

	err := dbh(ctx).Insert(dbOrg)
	if err != nil && strings.Contains(err.Error(), `duplicate key value violates unique constraint`) {
		return store.ErrWaitlistedOrgExists
	}
	if err != nil {
		return err
	}
	return nil
}

func (w *waitlist) getOrg(ctx context.Context, orgName string) (*orgWaitlistRow, error) {
	var orgs []*orgWaitlistRow
	err := dbh(ctx).Select(&orgs, "SELECT * FROM org_waitlist WHERE name=$1 LIMIT 1", orgName)
	if err != nil {
		return nil, err
	}
	if len(orgs) == 0 {
		return nil, &store.WaitlistedOrgNotFoundError{}
	}
	return orgs[0], err
}

func (w *waitlist) GetOrg(ctx context.Context, orgName string) (*sourcegraph.WaitlistedOrg, error) {
	dbOrg, err := w.getOrg(ctx, orgName)
	if err != nil {
		return nil, err
	}
	return &sourcegraph.WaitlistedOrg{
		Name:      dbOrg.Name,
		AddedAt:   ts(dbOrg.AddedAt),
		GrantedAt: ts(dbOrg.GrantedAt),
	}, err
}

func (w *waitlist) GrantOrg(ctx context.Context, orgName string) error {
	dbOrg, err := w.getOrg(ctx, orgName)
	if err != nil {
		return err
	}

	if dbOrg.GrantedAt != nil {
		return nil
	}

	currTime := time.Now()
	dbOrg.GrantedAt = &currTime
	_, err = dbh(ctx).Update(dbOrg)
	return err
}

func (w *waitlist) ListOrgs(ctx context.Context, onlyWaitlisted bool) ([]*sourcegraph.WaitlistedOrg, error) {
	var orgWaitlistRows []*orgWaitlistRow
	sql := `SELECT * FROM org_waitlist`
	if onlyWaitlisted {
		sql += ` WHERE granted_at is null`
	}
	if err := dbh(ctx).Select(&orgWaitlistRows, sql); err != nil {
		return nil, err
	}

	waitlistedOrgs := make([]*sourcegraph.WaitlistedOrg, len(orgWaitlistRows))
	for i, row := range orgWaitlistRows {
		waitlistedOrgs[i] = &sourcegraph.WaitlistedOrg{
			Name:      row.Name,
			AddedAt:   ts(row.AddedAt),
			GrantedAt: ts(row.GrantedAt),
		}
	}
	return waitlistedOrgs, nil
}
