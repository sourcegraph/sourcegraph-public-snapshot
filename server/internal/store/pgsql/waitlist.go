package pgsql

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"gopkg.in/gorp.v1"

	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/server/accesscontrol"
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

	Schema.Map.AddTableWithName(orgWaitlistRow{}, "org_waitlist").SetKeys(false, "Name")
	Schema.CreateSQL = append(Schema.CreateSQL,
		`ALTER TABLE org_waitlist ALTER COLUMN added_at TYPE timestamp with time zone USING added_at::timestamp with time zone;`,
		`ALTER TABLE org_waitlist ALTER COLUMN granted_at TYPE timestamp with time zone USING granted_at::timestamp with time zone;`,
	)
}

// waitlist is a DB-backed implementation of the Waitlist store.
type waitlist struct{}

func (w *waitlist) AddUser(ctx context.Context, uid int32) error {
	if err := accesscontrol.VerifyUserSelfOrAdmin(ctx, "Waitlist.AddUser", uid); err != nil {
		return err
	}
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
	var user userWaitlistRow
	if err := dbh(ctx).SelectOne(&user, "SELECT * FROM user_waitlist WHERE uid=$1 LIMIT 1", uid); err == sql.ErrNoRows {
			return nil, &store.WaitlistedUserNotFoundError{UID: uid}
	} else if err != nil {
		return nil, err
	}
	return &user, nil
}

func (w *waitlist) GetUser(ctx context.Context, uid int32) (*sourcegraph.WaitlistedUser, error) {
	if err := accesscontrol.VerifyUserSelfOrAdmin(ctx, "Waitlist.GetUser", uid); err != nil {
		return nil, err
	}
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
	if err := accesscontrol.VerifyUserHasAdminAccess(ctx, "Waitlist.GrantUser"); err != nil {
		return err
	}
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
	if err := accesscontrol.VerifyUserHasAdminAccess(ctx, "Waitlist.ListUsers"); err != nil {
		return nil, err
	}
	var userWaitlistRows []*userWaitlistRow
	sql := `SELECT * FROM user_waitlist`
	if onlyWaitlisted {
		sql += ` WHERE granted_at is null`
	}
	if _, err := dbh(ctx).Select(&userWaitlistRows, sql); err != nil {
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
	if err := accesscontrol.VerifyUserHasAdminAccess(ctx, "Waitlist.AddOrg"); err != nil {
		return err
	}
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
	var org orgWaitlistRow
	if err := dbh(ctx).SelectOne(&org, "SELECT * FROM org_waitlist WHERE name=$1 LIMIT 1", orgName); err == sql.ErrNoRows {
		return nil, &store.WaitlistedOrgNotFoundError{OrgName: orgName}
	} else if err != nil {
		return nil, err
	}
	return &org, nil
}

func (w *waitlist) GetOrg(ctx context.Context, orgName string) (*sourcegraph.WaitlistedOrg, error) {
	if err := accesscontrol.VerifyUserHasAdminAccess(ctx, "Waitlist.GetOrg"); err != nil {
		return nil, err
	}
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
	if err := accesscontrol.VerifyUserHasAdminAccess(ctx, "Waitlist.GrantOrg"); err != nil {
		return err
	}
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

func (w *waitlist) ListOrgs(ctx context.Context, onlyWaitlisted, onlyGranted bool, filterNames []string) ([]*sourcegraph.WaitlistedOrg, error) {
	if err := accesscontrol.VerifyUserHasAdminAccess(ctx, "Waitlist.ListOrgs"); err != nil {
		return nil, err
	}
	var orgWaitlistRows []*orgWaitlistRow

	var args []interface{}
	arg := func(a interface{}) string {
		v := gorp.PostgresDialect{}.BindVar(len(args))
		args = append(args, a)
		return v
	}

	whereSQL := "true"
	var conds []string
	if onlyWaitlisted {
		conds = append(conds, "granted_at is null")
	}
	if onlyGranted {
		conds = append(conds, "granted_at is not null")
	}
	if filterNames != nil && len(filterNames) > 0 {
		orgNames := make([]string, len(filterNames))
		for i, name := range filterNames {
			orgNames[i] = arg(name)
		}
		conds = append(conds, "name IN ("+strings.Join(orgNames, ",")+")")
	}
	if conds != nil && len(conds) > 0 {
		whereSQL = "(" + strings.Join(conds, ") AND (") + ")"
	}
	sql := fmt.Sprintf(`SELECT * FROM org_waitlist WHERE %s`, whereSQL)
	if _, err := dbh(ctx).Select(&orgWaitlistRows, sql, args...); err != nil {
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
