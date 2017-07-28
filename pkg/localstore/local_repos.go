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

func init() {
	AppSchema.Map.AddTableWithName(dbLocalRepo{}, "local_repos").SetKeys(true, "ID").SetUniqueTogether("remote_uri", "access_token")
	AppSchema.CreateSQL = append(AppSchema.CreateSQL,
		"CREATE INDEX ON local_repos(remote_uri);",
		"ALTER TABLE local_repos ALTER COLUMN remote_uri TYPE citext;",
	)
}

// dbLocalRepo DB-maps a sourcegraph.LocalRepo object.
type dbLocalRepo struct {
	ID          int64
	RemoteURI   string     `db:"remote_uri"`
	AccessToken string     `db:"access_token"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at"`
}

func (r *dbLocalRepo) fromLocalRepo(r2 *sourcegraph.LocalRepo) {
	r.ID = int64(r2.ID)
	r.RemoteURI = r2.RemoteURI
	r.AccessToken = r2.AccessToken
}

func (r *dbLocalRepo) toLocalRepo() *sourcegraph.LocalRepo {
	r2 := &sourcegraph.LocalRepo{}
	r2.ID = int32(r.ID)
	r2.RemoteURI = r.RemoteURI
	r2.AccessToken = r.AccessToken
	r2.CreatedAt = r.CreatedAt
	r2.UpdatedAt = r.UpdatedAt
	return r2
}

func (r *dbLocalRepo) validate() error {
	matched := validRemoteURI.MatchString(r.RemoteURI)
	if !matched {
		return fmt.Errorf("not a valid remote URI: %s", r.RemoteURI)
	}
	return nil
}

type localRepos struct{}

func (*localRepos) Get(ctx context.Context, remoteURI, accessToken string) (*sourcegraph.LocalRepo, error) {
	if Mocks.LocalRepos.Get != nil {
		return Mocks.LocalRepos.Get(ctx, remoteURI, accessToken)
	}

	var r dbLocalRepo
	// 🚨 SECURITY: must include access_token field in query as a permissions 🚨
	// check. Note other functions may rely on this method to verify repo permissions
	err := appDBH(ctx).SelectOne(&r, "SELECT * FROM local_repos WHERE (remote_uri=$1 AND access_token=$2 AND deleted_at IS NULL)", remoteURI, accessToken)
	if err == sql.ErrNoRows {
		return nil, ErrRepoNotFound
	} else if err != nil {
		return nil, err
	}

	return r.toLocalRepo(), nil
}

func (*localRepos) Create(ctx context.Context, newRepo *sourcegraph.LocalRepo) (*sourcegraph.LocalRepo, error) {
	if Mocks.LocalRepos.Create != nil {
		return Mocks.LocalRepos.Create(ctx, newRepo)
	}

	if newRepo.RemoteURI == "" {
		return nil, errors.New("error creating local repo: RemoteURI required")
	}
	if newRepo.AccessToken == "" {
		return nil, fmt.Errorf("error creating local repo %s: AccessToken required", newRepo.RemoteURI)
	}

	var r dbLocalRepo
	r.fromLocalRepo(newRepo)
	r.CreatedAt = time.Now()
	r.UpdatedAt = r.CreatedAt
	err := r.validate()
	if err != nil {
		return nil, err
	}
	err = appDBH(ctx).Insert(&r)
	if err != nil {
		return nil, err
	}
	return r.toLocalRepo(), nil
}
