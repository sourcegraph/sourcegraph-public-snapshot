package fs

import (
	"os"

	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/platform/storage"
	"src.sourcegraph.com/sourcegraph/store"
)

// RepoConfigs is a FS-backed implementation of the RepoConfigs store.
//
// TODO(slimsag): consider replacing RepoConfigs entirely with just platform
// storage.
type RepoConfigs struct{}

var _ store.RepoConfigs = (*RepoConfigs)(nil)

const (
	repoConfigsAppName = "core.repo-configs"
)

func (s *RepoConfigs) Get(ctx context.Context, repo string) (*sourcegraph.RepoConfig, error) {
	sys := storage.Namespace(ctx, repoConfigsAppName, repo)
	var conf sourcegraph.RepoConfig
	if err := storage.GetJSON(sys, "config", "config.json", &conf); err != nil {
		if os.IsNotExist(err) {
			return &conf, nil
		}
		return nil, err
	}
	return &conf, nil
}

func (s *RepoConfigs) Update(ctx context.Context, repo string, conf sourcegraph.RepoConfig) error {
	sys := storage.Namespace(ctx, repoConfigsAppName, repo)
	return storage.PutJSON(sys, "configs", "config.json", &conf)
}
