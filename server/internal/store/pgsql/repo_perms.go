package pgsql

import (
	"fmt"
	"strings"
	"time"

	"github.com/sqs/modl"

	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/store"
)

type repoPermsRow struct {
	UID       int32      `db:"uid"`
	Repo      string     `db:"repo"`
	GrantedAt *time.Time `db:"granted_at"`
}

func init() {
	Schema.Map.AddTableWithName(repoPermsRow{}, "repo_perms").SetKeys(false, "UID", "Repo")
	Schema.CreateSQL = append(Schema.CreateSQL,
		`ALTER TABLE user_waitlist ALTER COLUMN granted_at TYPE timestamp with time zone USING granted_at::timestamp with time zone;`,
	)
}

// repoPerms is a DB-backed implementation of the RepoPerms store.
type repoPerms struct{}

func (r *repoPerms) Add(ctx context.Context, uid int32, repo string) error {
	if uid == 0 || repo == "" {
		return nil
	}

	currTime := time.Now()
	dbPerms := &repoPermsRow{
		UID:       uid,
		Repo:      repo,
		GrantedAt: &currTime,
	}

	err := dbh(ctx).Insert(dbPerms)
	if err != nil && strings.Contains(err.Error(), `duplicate key value violates unique constraint`) {
		return store.ErrRepoPermissionExists
	}
	if err != nil {
		return err
	}
	return nil
}

func (r *repoPerms) Update(ctx context.Context, uid int32, repos []string) error {
	if uid == 0 {
		return nil
	}

	// Insert all repo permissions
	for _, repo := range repos {
		err := r.Add(ctx, uid, repo)
		if err != nil && err != store.ErrRepoPermissionExists {
			return err
		}
	}

	var args []interface{}
	arg := func(a interface{}) string {
		v := modl.PostgresDialect{}.BindVar(len(args))
		args = append(args, a)
		return v
	}

	uidSQL := "uid=" + arg(uid)
	repoSQL := "true"
	if repos != nil && len(repos) > 0 {
		repoURIs := make([]string, len(repos))
		for i, r := range repos {
			repoURIs[i] = arg(r)
		}
		repoSQL = "repo NOT IN (" + strings.Join(repoURIs, ",") + ")"
	}

	// Remove extra permissions
	sql := fmt.Sprintf(`DELETE FROM repo_perms WHERE %s AND %s`, uidSQL, repoSQL)
	res, err := dbh(ctx).Exec(sql, uid)
	if err != nil {
		return err
	}
	if _, err := res.RowsAffected(); err != nil {
		return err
	}
	return nil
}

func (r *repoPerms) Delete(ctx context.Context, uid int32, repo string) error {
	if uid == 0 || repo == "" {
		return nil
	}

	res, err := dbh(ctx).Exec(`DELETE FROM reg_clients WHERE uid=$1 AND repo=$2;`, uid, repo)
	if err != nil {
		return err
	}
	if _, err := res.RowsAffected(); err != nil {
		return err
	}
	return nil
}

func (r *repoPerms) ListUserRepos(ctx context.Context, uid int32) ([]string, error) {
	if uid == 0 {
		return make([]string, 0), nil
	}

	var repoPermsRows []*repoPermsRow
	sql := `SELECT * FROM repo_perms WHERE uid=$1`
	if err := dbh(ctx).Select(&repoPermsRows, sql, uid); err != nil {
		return nil, err
	}

	repos := make([]string, len(repoPermsRows))
	for i, row := range repoPermsRows {
		repos[i] = row.Repo
	}
	return repos, nil
}

func (r *repoPerms) ListRepoUsers(ctx context.Context, repo string) ([]int32, error) {
	if repo == "" {
		return make([]int32, 0), nil
	}

	var repoPermsRows []*repoPermsRow
	sql := `SELECT * FROM repo_perms WHERE repo=$1`
	if err := dbh(ctx).Select(&repoPermsRows, sql, repo); err != nil {
		return nil, err
	}

	users := make([]int32, len(repoPermsRows))
	for i, row := range repoPermsRows {
		users[i] = row.UID
	}
	return users, nil
}

func (r *repoPerms) DeleteUser(ctx context.Context, uid int32) error {
	if uid == 0 {
		return nil
	}

	res, err := dbh(ctx).Exec(`DELETE FROM repo_perms WHERE uid=$1;`, uid)
	if err != nil {
		return err
	}
	if _, err := res.RowsAffected(); err != nil {
		return err
	}
	return nil
}

func (r *repoPerms) DeleteRepo(ctx context.Context, repo string) error {
	if repo == "" {
		return nil
	}

	res, err := dbh(ctx).Exec(`DELETE FROM repo_perms WHERE repo=$1;`, repo)
	if err != nil {
		return err
	}
	if _, err := res.RowsAffected(); err != nil {
		return err
	}
	return nil
}
