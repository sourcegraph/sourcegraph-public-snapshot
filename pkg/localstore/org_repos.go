package localstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/keegancsmith/sqlf"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

// validCanonicalRemoteID matches a relative path, with the first component being a
// valid domain, e.g. "github.com/gorilla/mux".
var validCanonicalRemoteID = regexp.MustCompile(`^[^/@]+(\.[^/]+)+(:\d+)?(/[^/]+)+$`)

func validateRepo(repo *sourcegraph.OrgRepo) error {
	if repo.CanonicalRemoteID == "" {
		return errors.New("error creating local repo: CanonicalRemoteID required")
	}
	matched := validCanonicalRemoteID.MatchString(repo.CanonicalRemoteID)
	if !matched {
		return fmt.Errorf("error creating local repo %s: not a valid remote uri", repo.CanonicalRemoteID)
	}
	if repo.OrgID == 0 {
		return fmt.Errorf("error creating local repo %s: OrgID required", repo.CanonicalRemoteID)
	}
	return nil
}

type orgRepos struct{}

func (r *orgRepos) GetByID(ctx context.Context, id int32) (*sourcegraph.OrgRepo, error) {
	if Mocks.OrgRepos.GetByID != nil {
		return Mocks.OrgRepos.GetByID(ctx, id)
	}
	return expectOne(r.getBySQL(ctx, "WHERE id=$1 AND deleted_at IS NULL", id))
}

func (r *orgRepos) GetByOrg(ctx context.Context, orgID int32) ([]*sourcegraph.OrgRepo, error) {
	return r.getBySQL(ctx, "WHERE org_id=$1 AND deleted_at IS NULL", orgID)
}

func (r *orgRepos) GetByCanonicalRemoteID(ctx context.Context, orgID int32, canonicalRemoteID string) (*sourcegraph.OrgRepo, error) {
	if Mocks.OrgRepos.GetByCanonicalRemoteID != nil {
		return Mocks.OrgRepos.GetByCanonicalRemoteID(ctx, orgID, canonicalRemoteID)
	}
	q := r.listQuery(orgID, []string{canonicalRemoteID})
	return expectOne(r.getBySQL(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...))
}

func (r *orgRepos) GetByCanonicalRemoteIDs(ctx context.Context, orgID int32, canonicalRemoteIDs []string) ([]*sourcegraph.OrgRepo, error) {
	q := r.listQuery(orgID, canonicalRemoteIDs)
	return r.getBySQL(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
}

func (r *orgRepos) listQuery(orgID int32, canonicalRemoteIDs []string) *sqlf.Query {
	conds := []*sqlf.Query{}
	conds = append(conds, sqlf.Sprintf("org_id=%d", orgID))
	if len(canonicalRemoteIDs) > 0 {
		ids := []*sqlf.Query{}
		for _, id := range canonicalRemoteIDs {
			ids = append(ids, sqlf.Sprintf("%s", id))
		}
		conds = append(conds, sqlf.Sprintf("canonical_remote_id IN (%s)", sqlf.Join(ids, ",")))
	}
	conds = append(conds, sqlf.Sprintf("deleted_at IS NULL"))
	return sqlf.Sprintf("WHERE %s", sqlf.Join(conds, "AND"))
}

func expectOne(repos []*sourcegraph.OrgRepo, err error) (*sourcegraph.OrgRepo, error) {
	if err != nil {
		return nil, err
	}
	if len(repos) != 1 {
		return nil, ErrRepoNotFound
	}
	// repos should have exactly one entry after passing the error check above.
	return repos[0], nil
}

func (*orgRepos) Create(ctx context.Context, newRepo *sourcegraph.OrgRepo) (*sourcegraph.OrgRepo, error) {
	if Mocks.OrgRepos.Create != nil {
		return Mocks.OrgRepos.Create(ctx, newRepo)
	}

	newRepo.CreatedAt = time.Now()
	newRepo.UpdatedAt = newRepo.CreatedAt
	err := validateRepo(newRepo)
	if err != nil {
		return nil, err
	}

	// orgID is temporarily nullable while we support both orgs and access tokens.
	// TODO(nick): make org_id non-null when dropping support for access tokens.
	err = globalDB.QueryRowContext(
		ctx,
		"INSERT INTO org_repos(canonical_remote_id, org_id, clone_url, created_at, updated_at) VALUES($1, $2, $3, $4, $5) RETURNING id",
		newRepo.CanonicalRemoteID, newRepo.OrgID, newRepo.CloneURL, newRepo.CreatedAt, newRepo.UpdatedAt).Scan(&newRepo.ID)
	if err != nil {
		return nil, err
	}

	return newRepo, nil
}

// getBySQL returns org repos matching the SQL query, if any exist.
func (*orgRepos) getBySQL(ctx context.Context, query string, args ...interface{}) ([]*sourcegraph.OrgRepo, error) {
	rows, err := globalDB.QueryContext(ctx, "SELECT id, canonical_remote_id, clone_url, org_id, created_at, updated_at FROM org_repos "+query, args...)
	if err != nil {
		return nil, err
	}

	repos := []*sourcegraph.OrgRepo{}
	defer rows.Close()
	for rows.Next() {
		var repo sourcegraph.OrgRepo
		// orgID is temporarily nullable while we support both orgs and access tokens.
		// TODO(nick): make org_id non-null when dropping support for access tokens.
		var orgID sql.NullInt64
		err := rows.Scan(&repo.ID, &repo.CanonicalRemoteID, &repo.CloneURL, &orgID, &repo.CreatedAt, &repo.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if orgID.Valid {
			repo.OrgID = int32(orgID.Int64)
		}
		repos = append(repos, &repo)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return repos, nil
}
