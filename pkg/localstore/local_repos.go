package localstore

import (
	"context"
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
		return fmt.Errorf("not a valid remote URI: %s", repo.RemoteURI)
	}
	if repo.AccessToken == "" {
		return fmt.Errorf("error creating local repo %s: AccessToken required", repo.RemoteURI)
	}
	return nil
}

type localRepos struct{}

func (l *localRepos) Get(ctx context.Context, remoteURI, accessToken string) (*sourcegraph.LocalRepo, error) {
	if Mocks.LocalRepos.Get != nil {
		return Mocks.LocalRepos.Get(ctx, remoteURI, accessToken)
	}

	// 🚨 SECURITY: must include access_token field in query as a permissions 🚨
	// check. Note other functions may rely on this method to verify repo permissions
	repos, err := l.getBySQL(ctx, "WHERE (remote_uri=$1 AND access_token=$2 AND deleted_at IS NULL) LIMIT 1", remoteURI, accessToken)
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
	var id int64
	err = appDBH(ctx).QueryRow(
		"INSERT INTO local_repos(remote_uri, access_token, created_at, updated_at) VALUES($1, $2, $3, $4) RETURNING id",
		newRepo.RemoteURI, newRepo.AccessToken, newRepo.CreatedAt, newRepo.UpdatedAt).Scan(&id)
	if err != nil {
		return nil, err
	}

	return newRepo, nil
}

// getBySQL returns localRepos matching the SQL query, if any exist.
func (*localRepos) getBySQL(ctx context.Context, query string, args ...interface{}) ([]*sourcegraph.LocalRepo, error) {
	rows, err := appDBH(ctx).Query("SELECT id, remote_uri, access_token, created_at, updated_at FROM local_repos "+query, args...)
	if err != nil {
		return nil, err
	}

	localRepos := []*sourcegraph.LocalRepo{}
	defer rows.Close()
	for rows.Next() {
		var l sourcegraph.LocalRepo
		err := rows.Scan(&l.ID, &l.RemoteURI, &l.AccessToken, &l.CreatedAt, &l.UpdatedAt)
		if err != nil {
			return nil, err
		}
		localRepos = append(localRepos, &l)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return localRepos, nil
}
