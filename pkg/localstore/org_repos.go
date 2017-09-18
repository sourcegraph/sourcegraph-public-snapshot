package localstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"time"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

// validRemoteURI matches a relative path, with the first component being a
// valid domain, e.g. "github.com/gorilla/mux".
var validRemoteURI = regexp.MustCompile(`^[^/@]+(\.[^/]+)+(:\d+)?(/[^/]+)+$`)

func validateRepo(repo *sourcegraph.OrgRepo) error {
	if repo.RemoteURI == "" {
		return errors.New("error creating local repo: RemoteURI required")
	}
	matched := validRemoteURI.MatchString(repo.RemoteURI)
	if !matched {
		return fmt.Errorf("error creating local repo %s: not a valid remote uri", repo.RemoteURI)
	}
	if repo.AccessToken == "" && repo.OrgID == 0 {
		return fmt.Errorf("error creating local repo %s: AccessToken or OrgID required", repo.RemoteURI)
	}
	return nil
}

type orgRepos struct{}

func (r *orgRepos) GetByID(ctx context.Context, id int32) (*sourcegraph.OrgRepo, error) {
	return r.getOneBySQL(ctx, "WHERE id=$1 AND deleted_at IS NULL LIMIT 1", id)
}

func (r *orgRepos) GetByOrg(ctx context.Context, orgID int32) ([]*sourcegraph.OrgRepo, error) {
	return r.getBySQL(ctx, "WHERE org_id=$1 AND access_token IS NULL AND deleted_at IS NULL", orgID)
}

// deprecated
func (r *orgRepos) Get(ctx context.Context, remoteURI, accessToken string) (*sourcegraph.OrgRepo, error) {
	if Mocks.OrgRepos.Get != nil {
		return Mocks.OrgRepos.Get(ctx, remoteURI, accessToken)
	}
	return r.getOneBySQL(ctx, "WHERE (remote_uri=$1 AND access_token=$2 AND org_id IS NULL AND deleted_at IS NULL) LIMIT 1", remoteURI, accessToken)
}

func (r *orgRepos) getOneBySQL(ctx context.Context, query string, args ...interface{}) (*sourcegraph.OrgRepo, error) {
	repos, err := r.getBySQL(ctx, query, args...)
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
	var orgID *int32
	var accessToken *string
	if newRepo.OrgID > 0 {
		orgID = &newRepo.OrgID
	} else {
		accessToken = &newRepo.AccessToken
	}
	err = globalDB.QueryRow(
		"INSERT INTO org_repos(remote_uri, org_id, access_token, created_at, updated_at) VALUES($1, $2, $3, $4, $5) RETURNING id",
		newRepo.RemoteURI, orgID, accessToken, newRepo.CreatedAt, newRepo.UpdatedAt).Scan(&newRepo.ID)
	if err != nil {
		return nil, err
	}

	return newRepo, nil
}

// getBySQL returns org repos matching the SQL query, if any exist.
func (*orgRepos) getBySQL(ctx context.Context, query string, args ...interface{}) ([]*sourcegraph.OrgRepo, error) {
	rows, err := globalDB.Query("SELECT id, remote_uri, org_id, access_token, created_at, updated_at FROM org_repos "+query, args...)
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
		var accessToken sql.NullString
		err := rows.Scan(&repo.ID, &repo.RemoteURI, &orgID, &accessToken, &repo.CreatedAt, &repo.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if orgID.Valid {
			repo.OrgID = int32(orgID.Int64)
		} else if accessToken.Valid {
			repo.AccessToken = accessToken.String
		}
		repos = append(repos, &repo)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return repos, nil
}
