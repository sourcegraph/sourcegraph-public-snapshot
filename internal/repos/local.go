package repos

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// LocalSources connects to a local code host.
type LocalSource struct {
	svc    *types.ExternalService
	config *schema.LocalExternalService
	logger log.Logger
}

func NewLocalSource(ctx context.Context, svc *types.ExternalService, logger log.Logger) (*LocalSource, error) {
	rawConfig, err := svc.Config.Decrypt(ctx)
	if err != nil {
		return nil, errors.Errorf("external service id=%d config error: %s", svc.ID, err)
	}
	var config schema.LocalExternalService
	if err := jsonc.Unmarshal(rawConfig, &config); err != nil {
		return nil, errors.Wrapf(err, "external service id=%d config error", svc.ID)
	}

	return &LocalSource{
		svc:    svc,
		config: &config,
		logger: logger,
	}, nil
}

// CheckConnection at this point assumes availability and relies on errors returned
// from the subsequent calls. This is going to be expanded as part of issue #44683
// to actually only return true if the source can serve requests.
func (s *LocalSource) CheckConnection(ctx context.Context) error {
	return nil
}

func (s *LocalSource) ExternalServices() types.ExternalServices {
	return types.ExternalServices{s.svc}
}

func (s *LocalSource) ListRepos(ctx context.Context, results chan SourceResult) {
	urn := s.svc.URN()
	repoPaths := getRepoPaths(s.config)
	for _, r := range repoPaths {
		uri := "file:///" + r.Path
		s.logger.Info("found repo " + uri)
		results <- SourceResult{
			Source: s,
			Repo: &types.Repo{
				Name: r.fullName(),
				URI:  uri,
				ExternalRepo: api.ExternalRepoSpec{
					ID:          uri,
					ServiceType: extsvc.VariantLocalGit.AsType(),
					ServiceID:   uri,
				},
				Sources: map[string]*types.SourceInfo{
					urn: {
						ID:       urn,
						CloneURL: uri,
					},
				},
				Metadata: nil,
			},
		}
	}
}

// Checks if git thinks the given path is a valid .git folder for a repository
func isBareRepo(path string) bool {
	c := exec.Command("git", "-C", path, "rev-parse", "--is-bare-repository")
	out, err := c.CombinedOutput()

	if err != nil {
		return false
	}

	return string(out) != "false\n"
}

// Check if git thinks the given path is a proper git checkout
func isGitRepo(path string) bool {
	// Executing git rev-parse --git-dir in the root of a worktree returns .git
	c := exec.Command("git", "-C", path, "rev-parse")
	err := c.Run()
	return err == nil
}

type repoConfig struct {
	Path  string
	Group string
}

func (c repoConfig) fullName() api.RepoName {
	name := ""
	if c.Group != "" {
		name = c.Group + "/"
	}
	name += strings.TrimSuffix(filepath.Base(c.Path), ".git")
	return api.RepoName(name)
}

func getRepoPaths(config *schema.LocalExternalService) []repoConfig {
	paths := []repoConfig{}
	for _, pathConfig := range config.Repos {
		pattern, err := homedir.Expand(pathConfig.Pattern)
		if err != nil {
			continue
		}
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}

		for _, match := range matches {
			if isGitRepo(match) {
				paths = append(paths, repoConfig{Path: match, Group: pathConfig.Group})
			}
		}
	}

	return paths
}
