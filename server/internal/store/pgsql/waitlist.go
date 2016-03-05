package pgsql

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/net/context"
	"gopkg.in/gorp.v1"

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

type userOrgRow struct {
	UID     int32      `db:"uid"`
	OrgName string     `db:"org"`
	AddedAt *time.Time `db:"added_at"`
}

type pendingReposRow struct {
	URI         string     `db:"uri"`
	CloneURL    string     `db:"clone_url"`
	Owner       string     `db:"owner"`
	IsOrg       bool       `db:"is_org"`
	Language    string     `db:"language"`
	Size        int32      `db:"size"`
	Forks       int32      `db:"forks"`
	Stars       int32      `db:"stars"`
	Watchers    int32      `db:"watchers"`
	Subscribers int32      `db:"subscribers"`
	Issues      int32      `db:"issues"`
	UpdatedAt   *time.Time `db:"updated_at"`
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

	Schema.Map.AddTableWithName(userOrgRow{}, "user_github_orgs").SetKeys(false, "UID", "OrgName")
	Schema.CreateSQL = append(Schema.CreateSQL,
		`ALTER TABLE user_github_orgs ALTER COLUMN added_at TYPE timestamp with time zone USING added_at::timestamp with time zone;`,
	)

	Schema.Map.AddTableWithName(pendingReposRow{}, "pending_repos").SetKeys(false, "URI")
	Schema.CreateSQL = append(Schema.CreateSQL,
		`ALTER TABLE pending_repos ALTER COLUMN updated_at TYPE timestamp with time zone USING updated_at::timestamp with time zone;`,
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

func (w *waitlist) UpdateUserOrgs(ctx context.Context, uid int32, orgNames []string) error {
	if err := accesscontrol.VerifyUserHasAdminAccess(ctx, "Waitlist.UpdateUserOrgs"); err != nil {
		return err
	}
	if uid == 0 {
		return nil
	}

	currTime := time.Now()
	// Insert all user orgs
	for _, org := range orgNames {
		dbUserOrg := &userOrgRow{
			UID:     uid,
			OrgName: org,
			AddedAt: &currTime,
		}

		err := dbh(ctx).Insert(dbUserOrg)
		if err != nil && !strings.Contains(err.Error(), `duplicate key value violates unique constraint`) {
			return err
		}
	}

	var args []interface{}
	arg := func(a interface{}) string {
		v := gorp.PostgresDialect{}.BindVar(len(args))
		args = append(args, a)
		return v
	}

	uidSQL := "uid=" + arg(uid)
	orgSQL := "true"
	if orgNames != nil && len(orgNames) > 0 {
		orgNameVars := make([]string, len(orgNames))
		for i, o := range orgNames {
			orgNameVars[i] = arg(o)
		}
		orgSQL = "org NOT IN (" + strings.Join(orgNameVars, ",") + ")"
	}

	// Remove extra orgs
	sql := fmt.Sprintf(`DELETE FROM user_github_orgs WHERE %s AND %s`, uidSQL, orgSQL)
	res, err := dbh(ctx).Exec(sql, args...)
	if err != nil {
		return err
	}
	if _, err := res.RowsAffected(); err != nil {
		return err
	}
	return nil
}

func (w *waitlist) RecordPendingRepo(ctx context.Context, repo *sourcegraph.RemoteRepo) error {
	if err := accesscontrol.VerifyUserHasAdminAccess(ctx, "Waitlist.RecordPendingRepo"); err != nil {
		return err
	}
	if repo == nil {
		return errors.New("invalid argument: nil repo")
	}
	currTime := time.Now()
	dbRepo := pendingReposRow{
		URI:       "github.com/" + repo.Owner + "/" + repo.Name,
		CloneURL:  repo.HTTPCloneURL,
		Owner:     repo.Owner,
		IsOrg:     repo.OwnerIsOrg,
		Language:  repo.Language,
		Stars:     repo.Stars,
		UpdatedAt: &currTime,
	}
	n, err := dbh(ctx).Update(&dbRepo)
	if err != nil {
		return err
	}
	if n == 0 {
		// No pending repo row yet exists, so we must insert it.
		return dbh(ctx).Insert(&dbRepo)
	}
	return nil
}
