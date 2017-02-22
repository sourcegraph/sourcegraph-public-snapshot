package zap

import (
	"fmt"

	"github.com/sourcegraph/zap/pkg/config"
)

type sortableRefIdentifiers []RefIdentifier

func (v sortableRefIdentifiers) Len() int      { return len(v) }
func (v sortableRefIdentifiers) Swap(i, j int) { v[i], v[j] = v[j], v[i] }
func (v sortableRefIdentifiers) Less(i, j int) bool {
	if v[i].Repo != v[j].Repo {
		return v[i].Repo < v[j].Repo
	}
	return v[i].Ref < v[j].Ref
}

type sortableRefInfos []RefInfo

func (v sortableRefInfos) Len() int      { return len(v) }
func (v sortableRefInfos) Swap(i, j int) { v[i], v[j] = v[j], v[i] }
func (v sortableRefInfos) Less(i, j int) bool {
	if v[i].Repo != v[j].Repo {
		return v[i].Repo < v[j].Repo
	}
	return v[i].Ref < v[j].Ref
}

// RemoteOrDefault returns r.Remotes[remoteName] if it exists. Otherwise, the
// global configuration file is consulted for a global default remote. If a
// global default remote exists, it is returned for the given repository name
// which should come from gitutil.DefaultRepoName.
//
// If repoName is an empty string, i.e. gitutil.DefaultRepoName could not derive
// one and returned an error, then the global configuration is not consulted.
func (r RepoConfiguration) RemoteOrDefault(remoteName, repoName string) (RepoRemoteConfiguration, error) {
	v, ok := r.Remotes[remoteName]
	if ok {
		return v, nil
	}
	if repoName == "" {
		return RepoRemoteConfiguration{}, fmt.Errorf("remote does not exist: %s", remoteName)
	}

	// At this point, we don't have the remote. But maybe we do have a global
	// default remote. Let's check.
	//
	// Note: err == nil if the file does not exist, i.e. cfg is just zero
	cfg, err := config.ReadGlobalFile()
	if err != nil {
		return RepoRemoteConfiguration{}, err
	}
	defaultEndpoint := cfg.Section("default").Option("remote")
	if defaultEndpoint == "" {
		return RepoRemoteConfiguration{}, fmt.Errorf("remote does not exist: %s", remoteName)
	}
	return RepoRemoteConfiguration{
		Endpoint: defaultEndpoint,
		Repo:     repoName,
		Refspecs: []string{"*"},
	}, nil
}
