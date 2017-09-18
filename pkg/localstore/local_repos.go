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

func validateLocalRepo(repo *sourcegraph.LocalRepo) error {
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

type localRepos struct{}

func (l *localRepos) GetByID(ctx context.Context, id int32) (*sourcegraph.LocalRepo, error) {
	return l.getOneBySQL(ctx, "WHERE id=$1 AND deleted_at IS NULL LIMIT 1", id)
}

func (l *localRepos) GetByOrg(ctx context.Context, orgID int32) ([]*sourcegraph.LocalRepo, error) {
	return l.getBySQL(ctx, "WHERE org_id=$1 AND access_token IS NULL AND deleted_at IS NULL", orgID)
}

// deprecated
func (l *localRepos) Get(ctx context.Context, remoteURI, accessToken string) (*sourcegraph.LocalRepo, error) {
	if Mocks.LocalRepos.Get != nil {
		return Mocks.LocalRepos.Get(ctx, remoteURI, accessToken)
	}
	return l.getOneBySQL(ctx, "WHERE (remote_uri=$1 AND access_token=$2 AND org_id IS NULL AND deleted_at IS NULL) LIMIT 1", remoteURI, accessToken)
}

func (l *localRepos) getOneBySQL(ctx context.Context, query string, args ...interface{}) (*sourcegraph.LocalRepo, error) {
	repos, err := l.getBySQL(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	if len(repos) != 1 {
		return nil, ErrRepoNotFound
	}
	// repos should have exactly one entry after passing the error check above.
	return repos[0], nil
}

func (*localRepos) Create(ctx context.Context, newRepo *sourcegraph.LocalRepo) (*sourcegraph.LocalRepo, error) {
	if Mocks.LocalRepos.Create != nil {
		return Mocks.LocalRepos.Create(ctx, newRepo)
	}

	newRepo.CreatedAt = time.Now()
	newRepo.UpdatedAt = newRepo.CreatedAt
	err := validateLocalRepo(newRepo)
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
		"INSERT INTO local_repos(remote_uri, org_id, access_token, created_at, updated_at) VALUES($1, $2, $3, $4, $5) RETURNING id",
		newRepo.RemoteURI, orgID, accessToken, newRepo.CreatedAt, newRepo.UpdatedAt).Scan(&newRepo.ID)
	if err != nil {
		return nil, err
	}

	return newRepo, nil
}

// getBySQL returns localRepos matching the SQL query, if any exist.
func (*localRepos) getBySQL(ctx context.Context, query string, args ...interface{}) ([]*sourcegraph.LocalRepo, error) {
	rows, err := globalDB.Query("SELECT id, remote_uri, org_id, access_token, created_at, updated_at FROM local_repos "+query, args...)
	if err != nil {
		return nil, err
	}

	localRepos := []*sourcegraph.LocalRepo{}
	defer rows.Close()
	for rows.Next() {
		var l sourcegraph.LocalRepo
		// orgID is temporarily nullable while we support both orgs and access tokens.
		// TODO(nick): make org_id non-null when dropping support for access tokens.
		var orgID sql.NullInt64
		var accessToken sql.NullString
		err := rows.Scan(&l.ID, &l.RemoteURI, &orgID, &accessToken, &l.CreatedAt, &l.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if orgID.Valid {
			l.OrgID = int32(orgID.Int64)
		} else if accessToken.Valid {
			l.AccessToken = accessToken.String
		}
		localRepos = append(localRepos, &l)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return localRepos, nil
}
