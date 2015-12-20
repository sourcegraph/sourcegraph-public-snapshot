package pgsql

import (
	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
)

func init() {
	Schema.Map.AddTableWithName(dbRepoConfig{}, "repo_config").SetKeys(false, "Repo")
}

// dbRepoConfig DB-maps a sourcegraph.RepoConfig object.
type dbRepoConfig struct {
	// Repo is the URI of the repository that this config is for.
	Repo string
}

func (c *dbRepoConfig) toRepoConfig() *sourcegraph.RepoConfig {
	return &sourcegraph.RepoConfig{}
}

func (c *dbRepoConfig) fromRepoConfig(repo string, c2 *sourcegraph.RepoConfig) {
	c.Repo = repo
}

// RepoConfigs is a DB-backed implementation of the RepoConfigs store.
type RepoConfigs struct{}

var _ store.RepoConfigs = (*RepoConfigs)(nil)

func (s *RepoConfigs) Get(ctx context.Context, repo string) (*sourcegraph.RepoConfig, error) {
	var confRows []*dbRepoConfig
	sql := `SELECT * FROM repo_config WHERE repo=$1;`
	if err := dbh(ctx).Select(&confRows, sql, repo); err != nil {
		return nil, err
	}
	if len(confRows) == 0 {
		return nil, nil
	}
	return confRows[0].toRepoConfig(), nil
}

func (s *RepoConfigs) Update(ctx context.Context, repo string, conf sourcegraph.RepoConfig) error {
	var dbConf dbRepoConfig
	dbConf.fromRepoConfig(repo, &conf)
	n, err := dbh(ctx).Update(&dbConf)
	if err != nil {
		return err
	}
	if n == 0 {
		// No config row yet exists, so we must insert it.
		return dbh(ctx).Insert(&dbConf)
	}
	return nil
}
