package server

import (
	"bufio"
	"bytes"
	"context"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestVcsDependenciesSyncer_Fetch(t *testing.T) {
	ctx := context.Background()
	placeholder, _ := parseFakeDependency("sourcegraph/placeholder@0.0.0")

	depsSource := &fakeDepsSource{
		deps:     map[string]reposource.PackageDependency{},
		get:      map[string]error{},
		download: map[string]error{},
	}
	depsService := &fakeDepsService{deps: map[string][]dependencies.Repo{}}

	s := vcsDependenciesSyncer{
		typ:         "fake",
		scheme:      "fake",
		placeholder: placeholder,
		source:      depsSource,
		svc:         depsService,
	}

	remoteURL := &vcs.URL{URL: url.URL{Path: "fake/foo"}}

	dir := GitDir(t.TempDir())
	_, err := s.CloneCommand(ctx, remoteURL, string(dir))
	require.NoError(t, err)

	depsService.Add("foo@0.0.1")
	depsSource.Add("foo@0.0.1")

	t.Run("one version from service", func(t *testing.T) {
		err := s.Fetch(ctx, remoteURL, dir)
		require.NoError(t, err)

		s.assertRefs(t, dir, map[string]string{
			"refs/heads/latest":   "759dab7e4a7fc384522cb75519660cb0d6f6e49d",
			"refs/tags/v0.0.1":    "b47eb15deed08abc9d437c81f42c1635febaa218",
			"refs/tags/v0.0.1^{}": "759dab7e4a7fc384522cb75519660cb0d6f6e49d",
		})
	})

	s.configDeps = []string{"foo@0.0.2"}
	depsSource.Add("foo@0.0.2")

	t.Run("two versions, service and config", func(t *testing.T) {
		err := s.Fetch(ctx, remoteURL, dir)
		require.NoError(t, err)

		s.assertRefs(t, dir, map[string]string{
			"refs/heads/latest":   "6cff53ec57702e8eec10569a3d981dacbaee4ed3",
			"refs/tags/v0.0.1":    "b47eb15deed08abc9d437c81f42c1635febaa218",
			"refs/tags/v0.0.1^{}": "759dab7e4a7fc384522cb75519660cb0d6f6e49d",
			"refs/tags/v0.0.2":    "7e2e4506ef1f5cd97187917a67bfb7a310f78687",
			"refs/tags/v0.0.2^{}": "6cff53ec57702e8eec10569a3d981dacbaee4ed3",
		})
	})

	depsSource.Delete("foo@0.0.2")

	t.Run("one version missing in source", func(t *testing.T) {
		err := s.Fetch(ctx, remoteURL, dir)
		require.NoError(t, err)

		s.assertRefs(t, dir, map[string]string{
			"refs/heads/latest":   "759dab7e4a7fc384522cb75519660cb0d6f6e49d",
			"refs/tags/v0.0.1":    "b47eb15deed08abc9d437c81f42c1635febaa218",
			"refs/tags/v0.0.1^{}": "759dab7e4a7fc384522cb75519660cb0d6f6e49d",
		})
	})

	depsSource.Add("foo@0.0.2")
	depsSource.get["foo@0.0.1"] = errors.New("401 unauthorized")

	t.Run("error tolerance", func(t *testing.T) {
		err := s.Fetch(ctx, remoteURL, dir)
		require.ErrorContains(t, err, "401 unauthorized")
		// When any fatal error is returned by source.Get, we add other new versions
		// that didn't return an error and delete no versions since we can't know
		// they have really been deleted in the presence of fatal errors.
		s.assertRefs(t, dir, map[string]string{
			"refs/heads/latest":   "6cff53ec57702e8eec10569a3d981dacbaee4ed3",
			"refs/tags/v0.0.1":    "b47eb15deed08abc9d437c81f42c1635febaa218",
			"refs/tags/v0.0.1^{}": "759dab7e4a7fc384522cb75519660cb0d6f6e49d",
			"refs/tags/v0.0.2":    "7e2e4506ef1f5cd97187917a67bfb7a310f78687",
			"refs/tags/v0.0.2^{}": "6cff53ec57702e8eec10569a3d981dacbaee4ed3",
		})
	})

	depsSource.get = map[string]error{}
	depsService.Delete("foo@0.0.1")

	t.Run("service version deleted", func(t *testing.T) {
		err := s.Fetch(ctx, remoteURL, dir)
		require.NoError(t, err)

		s.assertRefs(t, dir, map[string]string{
			"refs/heads/latest":   "6cff53ec57702e8eec10569a3d981dacbaee4ed3",
			"refs/tags/v0.0.2":    "7e2e4506ef1f5cd97187917a67bfb7a310f78687",
			"refs/tags/v0.0.2^{}": "6cff53ec57702e8eec10569a3d981dacbaee4ed3",
		})
	})

	s.configDeps = []string{}

	t.Run("all versions deleted", func(t *testing.T) {
		err := s.Fetch(ctx, remoteURL, dir)
		require.NoError(t, err)

		s.assertRefs(t, dir, map[string]string{})
	})
}

type fakeDepsService struct {
	deps map[string][]dependencies.Repo
}

func (s *fakeDepsService) ListDependencyRepos(ctx context.Context, opts dependencies.ListDependencyReposOpts) ([]dependencies.Repo, error) {
	return s.deps[opts.Name], nil
}

func (s *fakeDepsService) Add(deps ...string) {
	for _, d := range deps {
		dep, _ := parseFakeDependency(d)
		name := dep.PackageSyntax()
		s.deps[name] = append(s.deps[name], dependencies.Repo{
			Scheme:  dep.Scheme(),
			Name:    name,
			Version: dep.PackageVersion(),
		})
	}
}

func (s *fakeDepsService) Delete(deps ...string) {
	for _, d := range deps {
		dep, _ := parseFakeDependency(d)
		name := dep.PackageSyntax()
		version := dep.PackageVersion()
		filtered := s.deps[name][:0]
		for _, r := range s.deps[name] {
			if r.Version != version {
				filtered = append(filtered, r)
			}
		}
		s.deps[name] = filtered
	}
}

type fakeDepsSource struct {
	deps          map[string]reposource.PackageDependency
	get, download map[string]error
}

func (s *fakeDepsSource) Add(deps ...string) {
	for _, d := range deps {
		dep, _ := parseFakeDependency(d)
		s.deps[d] = dep
	}
}

func (s *fakeDepsSource) Delete(deps ...string) {
	for _, d := range deps {
		delete(s.deps, d)
	}
}

func (s *fakeDepsSource) Get(ctx context.Context, name, version string) (reposource.PackageDependency, error) {
	d := name + "@" + version

	err := s.get[d]
	if err != nil {
		return nil, err
	}

	dep, ok := s.deps[d]
	if !ok {
		return nil, notFoundError{errors.Errorf("%s@%s not found", name, version)}
	}

	return dep, nil
}

type notFoundError struct{ error }

func (e notFoundError) NotFound() bool { return true }

func (s *fakeDepsSource) Download(ctx context.Context, dir string, dep reposource.PackageDependency) error {
	err := s.download[dep.PackageManagerSyntax()]
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "README.md"), []byte("README for "+dep.PackageManagerSyntax()), 0666)
}

func (fakeDepsSource) ParseDependency(dep string) (reposource.PackageDependency, error) {
	return parseFakeDependency(dep)
}

func (fakeDepsSource) ParseDependencyFromRepoName(repoName string) (reposource.PackageDependency, error) {
	return parseFakeDependency(strings.TrimPrefix(repoName, "fake/"))
}

type fakeDep struct {
	name    string
	version string
}

func parseFakeDependency(dep string) (reposource.PackageDependency, error) {
	i := strings.LastIndex(dep, "@")
	if i == -1 {
		return fakeDep{name: dep}, nil
	}
	return fakeDep{name: dep[:i], version: dep[i+1:]}, nil
}

func (f fakeDep) Scheme() string               { return "fake" }
func (f fakeDep) PackageSyntax() string        { return f.name }
func (f fakeDep) PackageManagerSyntax() string { return f.name + "@" + f.version }
func (f fakeDep) PackageVersion() string       { return f.version }
func (f fakeDep) Description() string          { return f.name + "@" + f.version }
func (f fakeDep) RepoName() api.RepoName       { return api.RepoName("fake/" + f.name) }
func (f fakeDep) GitTagFromVersion() string    { return "v" + f.version }
func (f fakeDep) Less(other reposource.PackageDependency) bool {
	return f.PackageManagerSyntax() > other.PackageManagerSyntax()
}

func (s vcsDependenciesSyncer) runCloneCommand(t *testing.T, examplePackageURL, bareGitDirectory string, dependencies []string) {
	u := vcs.URL{
		URL: url.URL{Path: examplePackageURL},
	}
	s.configDeps = dependencies
	cmd, err := s.CloneCommand(context.Background(), &u, bareGitDirectory)
	assert.Nil(t, err)
	assert.Nil(t, cmd.Run())
}

func (s vcsDependenciesSyncer) assertRefs(t *testing.T, dir GitDir, want map[string]string) {
	t.Helper()

	cmd := exec.Command("git", "show-ref", "--head", "--dereference")
	cmd.Dir = string(dir)

	out, _ := cmd.CombinedOutput()

	sc := bufio.NewScanner(bytes.NewReader(out))
	have := map[string]string{}
	for sc.Scan() {
		fs := strings.Fields(sc.Text())
		have[fs[1]] = fs[0]
	}

	require.NoError(t, sc.Err())
	require.Equal(t, want, have)
}
