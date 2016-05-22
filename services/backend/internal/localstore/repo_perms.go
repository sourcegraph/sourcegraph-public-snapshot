package localstore

import (
	"time"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
)

type repoPermsRow struct {
	UID       int32      `db:"uid"`
	Repo      string     `db:"repo"`
	GrantedAt *time.Time `db:"granted_at"`
}

func init() {
	AppSchema.Map.AddTableWithName(repoPermsRow{}, "repo_perms").SetKeys(false, "UID", "Repo")
	AppSchema.CreateSQL = append(AppSchema.CreateSQL,
		`ALTER TABLE repo_perms ALTER COLUMN granted_at TYPE timestamp with time zone USING granted_at::timestamp with time zone;`,
	)
}

// repoPerms is a DB-backed implementation of the RepoPerms store.
type repoPerms struct{}

func (r *repoPerms) ListRepoUsers(ctx context.Context, repo string) ([]int32, error) {
	if err := accesscontrol.VerifyUserHasAdminAccess(ctx, "RepoPerms.ListRepoUsers"); err != nil {
		return nil, err
	}
	if repo == "" {
		return make([]int32, 0), nil
	}

	var repoPermsRows []*repoPermsRow
	sql := `SELECT * FROM repo_perms WHERE repo=$1`
	if _, err := appDBH(ctx).Select(&repoPermsRows, sql, repo); err != nil {
		return nil, err
	}

	users := make([]int32, len(repoPermsRows))
	for i, row := range repoPermsRows {
		users[i] = row.UID
	}
	return users, nil
}
