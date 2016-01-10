package fs

import (
	"os"

	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	platformstorage "src.sourcegraph.com/sourcegraph/platform/storage"
	"src.sourcegraph.com/sourcegraph/store"
)

// repoConfigs is a FS-backed implementation of the RepoConfigs store.
//
// TODO(slimsag): consider replacing RepoConfigs entirely with just platform
// storage.
type repoConfigs struct{}

var _ store.RepoConfigs = (*repoConfigs)(nil)

const (
	repoConfigsAppName = "core.repo-configs"
	repoConfigsBucket  = "config"
	repoConfigsKey     = "config.json"
)

func (s *repoConfigs) Get(ctx context.Context, repo string) (*sourcegraph.RepoConfig, error) {
	sys := platformstorage.Namespace(ctx, repoConfigsAppName, repo)
	var conf sourcegraph.RepoConfig
	if err := platformstorage.GetJSON(sys, repoConfigsBucket, repoConfigsKey, &conf); err != nil {
		if os.IsNotExist(err) {
			// By default, all repos are enabled on the fs backend.
			return &conf, nil
		}
		return nil, err
	}
	return &conf, nil
}

func (s *repoConfigs) Update(ctx context.Context, repo string, conf sourcegraph.RepoConfig) error {
	sys := platformstorage.Namespace(ctx, repoConfigsAppName, repo)
	return platformstorage.PutJSON(sys, repoConfigsBucket, repoConfigsKey, &conf)
}
