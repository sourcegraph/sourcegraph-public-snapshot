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

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"
	"golang.org/x/exp/maps"

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
		availableVersions: []string{"0.0.1", "0.0.2", "0.0.3"},
		download:          map[string]error{},
		downloadCount:     map[string]int{},
	}
	// contains the versions of packages synced to the "instance"
	depsService := &fakeDepsService{deps: map[reposource.PackageName][]dependencies.Repo{}}

	s := vcsPackagesSyncer{
		logger:      logtest.Scoped(t),
		typ:         "fake",
		scheme:      "fake",
		placeholder: placeholder,
		source:      depsSource,
		svc:         depsService,
	}

	remoteURL := &vcs.URL{URL: url.URL{Path: "fake/foo"}}

	dir := GitDir(t.TempDir())
	if _, err := s.CloneCommand(ctx, remoteURL, string(dir)); err != nil {
		t.Fatalf("unexpected error preparing package clone directory: %v", err)
	}

	versionRefs := map[string]string{
		"refs/tags/v0.0.1":    "b47eb15deed08abc9d437c81f42c1635febaa218",
		"refs/tags/v0.0.1^{}": "759dab7e4a7fc384522cb75519660cb0d6f6e49d",
		"refs/tags/v0.0.2":    "ba0ae2f9c0799884212519824ad2e38ae72dee85",
		"refs/tags/v0.0.2^{}": emptyTreeObject,
		"refs/tags/v0.0.3":    "ba94b95e16bf902e983ead70dc6ee0edd6b03a3b",
		"refs/tags/v0.0.3^{}": "c93e10f82d5d34341b2836202ebb6b0faa95fa71",
		"refs/heads/latest":   "c93e10f82d5d34341b2836202ebb6b0faa95fa71",
	}

	depsService.Add("foo@0.0.1")

	t.Run("0.0.{1,2,3} available, 0.0.{1,3} syncing", func(t *testing.T) {
		err := s.Fetch(ctx, remoteURL, dir, "")
		if err != nil {
			t.Fatalf("unexpected error fetching package: %v", err)
		}

		s.assertRefs(t, dir, versionRefs)
		s.assertDownloadCounts(t, depsSource, map[string]int{"foo@0.0.1": 1, "foo@0.0.3": 1})
	})

	s.configDeps = []string{"foo@0.0.2"}
	maps.Copy(versionRefs, map[string]string{
		"refs/tags/v0.0.2":    "7e2e4506ef1f5cd97187917a67bfb7a310f78687",
		"refs/tags/v0.0.2^{}": "6cff53ec57702e8eec10569a3d981dacbaee4ed3",
	})

	t.Run("0.0.{1,2,3} available, 0.0.2 syncing, 0.0.{1,3} cached", func(t *testing.T) {
		err := s.Fetch(ctx, remoteURL, dir, "")
		if err != nil {
			t.Fatalf("unexpected error fetching package: %v", err)
		}

		s.assertRefs(t, dir, versionRefs)
		s.assertDownloadCounts(t, depsSource, map[string]int{"foo@0.0.1": 1, "foo@0.0.2": 1, "foo@0.0.3": 1})
	})

	depsSource.download["foo@0.0.1"] = errors.New("401 unauthorized")

	t.Run("0.0.1 401s, 0.0.{1,2,3} cached", func(t *testing.T) {
		if err := s.Fetch(ctx, remoteURL, dir, ""); err != nil {
			t.Fatalf("unexpected error fetching package: %v", err)
		}
		// v0.0.1 is still present in the git repo because we didn't send a second download request. This may
		// be something to revisit, but accurately determining whether we should delete or not from response
		// may be hard to do given artifact host differences
		s.assertRefs(t, dir, versionRefs)
		s.assertDownloadCounts(t, depsSource, map[string]int{"foo@0.0.1": 1, "foo@0.0.2": 1, "foo@0.0.3": 1})
	})

	depsSource.availableVersions = append(depsSource.availableVersions, "0.0.4")
	maps.Copy(versionRefs, map[string]string{
		"refs/heads/latest":   "c1150351bcacca8b0b513192da7673702033a519",
		"refs/tags/v0.0.4":    "fbb95220111fb527f1c58c4ec4f3f9438540a916",
		"refs/tags/v0.0.4^{}": "c1150351bcacca8b0b513192da7673702033a519",
	})

	t.Run("lazy-sync 0.0.4 via revspec", func(t *testing.T) {
		// the v0.0.4 tag should be created on-demand through the revspec parameter
		// For context, see https://github.com/sourcegraph/sourcegraph/pull/38811
		err := s.Fetch(ctx, remoteURL, dir, "v0.0.4^0")
		if err != nil {
			t.Fatalf("unexpected error fetching package: %v", err)
		}

		if diff := cmp.Diff(s.svc.(*fakeDepsService).upsertedDeps, []dependencies.Repo{{
			ID:      0,
			Scheme:  fakeVersionedPackage{}.Scheme(),
			Name:    "foo",
			Version: "0.0.4",
		}}); diff != "" {
			t.Errorf("unexpected list of upserted dependencies (-want +got):\n%s", diff)
		}

		s.assertRefs(t, dir, versionRefs)
		// We triggered a single download for v0.0.3 since it was lazily requested.
		// We triggered a v0.0.1 download since it's still erroring.
		s.assertDownloadCounts(t, depsSource, map[string]int{"foo@0.0.1": 1, "foo@0.0.2": 1, "foo@0.0.3": 1, "foo@0.0.4": 1})
	})

	depsSource.availableVersions = append(depsSource.availableVersions, "0.0.5", "0.0.7")
	depsSource.download["foo@0.0.6"] = errors.New("0.0.6 not found")
	maps.Copy(versionRefs, map[string]string{
		"refs/heads/latest":   "b63e479c262c38061fcf8c86ae9284836109331e",
		"refs/tags/v0.0.5":    "3d8601d3b0f45f43a7366b8948fa9b938a345754",
		"refs/tags/v0.0.5^{}": emptyTreeObject,
		"refs/tags/v0.0.7":    "8b77a98c4d9979551b442b0287700442f62cd549",
		"refs/tags/v0.0.7^{}": "b63e479c262c38061fcf8c86ae9284836109331e",
	})

	t.Run("lazy-sync error 0.0.6 via revspec, 0.0.{1,2,3,4,5,7} available", func(t *testing.T) {
		numUpsertedDeps := len(s.svc.(*fakeDepsService).upsertedDeps)

		// the v0.0.6 tag cannot be created on-demand because it returns a "0.0.6 not found" error
		if err := s.Fetch(ctx, remoteURL, dir, "v0.0.6^0"); err != nil {
			t.Fatalf("unexpected error fetching package: %v", err)
		}
		// the 0.0.6 dependency was not stored in the database because the download failed.
		if numUpsertedDeps != len(s.svc.(*fakeDepsService).upsertedDeps) {
			t.Errorf("unexpected number of upserted dependencies: want=%d got=%d", numUpsertedDeps, len(s.svc.(*fakeDepsService).upsertedDeps))
		}
		s.assertRefs(t, dir, versionRefs)
		// We triggered downloads only for v0.0.4, no new downloads were triggered for cached or other errored versions.
		s.assertDownloadCounts(t, depsSource, map[string]int{"foo@0.0.1": 1, "foo@0.0.2": 1, "foo@0.0.3": 1, "foo@0.0.4": 1, "foo@0.0.6": 1, "foo@0.0.7": 1})
	})

	depsSource.download["org.springframework.boot:spring-boot:3.0"] = notFoundError{errors.New("Please contact Josh Long")}

	t.Run("trying to download non-existent Maven dependency", func(t *testing.T) {
		springBootDep, err := reposource.ParseMavenVersionedPackage("org.springframework.boot:spring-boot:3.0")
		if err != nil {
			t.Fatal("Cannot parse Maven dependency")
		}
		err = s.gitPushDependencyTag(ctx, string(dir), springBootDep)
		if err == nil {
			t.Fatalf("unexpected nil error for non-existent dependency")
		}
	})
}

type fakeDepsService struct {
	deps         map[reposource.PackageName][]dependencies.Repo
	upsertedDeps []dependencies.Repo
}

func (s *fakeDepsService) UpsertDependencyRepos(ctx context.Context, deps []dependencies.Repo) ([]dependencies.Repo, error) {
	s.upsertedDeps = append(s.upsertedDeps, deps...)
	for _, dep := range deps {
		alreadyExists := false
		for _, existingDep := range s.deps[dep.Name] {
			if existingDep.Version == dep.Version {
				alreadyExists = true
				break
			}
		}
		if !alreadyExists {
			s.deps[dep.Name] = append(s.deps[dep.Name], dep)
		}
	}
	return deps, nil
}

func (s *fakeDepsService) ListDependencyRepos(ctx context.Context, opts dependencies.ListDependencyReposOpts) ([]dependencies.Repo, error) {
	return s.deps[opts.Name], nil
}

// Add adds a version that should be synced if it hasnt been added already. While the source may have additional versions,
// they wont be synced unless 1) added via this method 2) they are/were the latest version 3) theyre listed in vcsPackageSyncer.configDeps
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
		filtered := s.deps[name][:0]
		for _, r := range s.deps[name] {
			if r.Version != dep.PackageVersion() {
				filtered = append(filtered, r)
			}
		}
		s.deps[name] = filtered
	}
}

type fakeDepsSource struct {
	// all versions available when listing versions on the package host
	availableVersions []string
	download          map[string]error
	// previously used to track re-downloads of deleted and re-added dependencies.
	downloadCount map[string]int
}

func (s *fakeDepsSource) ListVersions(ctx context.Context, dep reposource.Package) (tags []reposource.VersionedPackage, err error) {
	for _, version := range s.availableVersions {
		pkg, _ := parseFakeDependency(string(dep.PackageSyntax()) + "@" + version)
		tags = append(tags, pkg)
	}
	return
}

func (s *fakeDepsSource) Download(ctx context.Context, dir string, dep reposource.VersionedPackage) error {
	s.downloadCount[dep.VersionedPackageSyntax()] = 1 + s.downloadCount[dep.VersionedPackageSyntax()]

	err := s.download[dep.VersionedPackageSyntax()]
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "README.md"), []byte("README for "+dep.VersionedPackageSyntax()), 0o666)
}

func (fakeDepsSource) ParseVersionedPackageFromNameAndVersion(name reposource.PackageName, version string) (reposource.VersionedPackage, error) {
	return parseFakeDependency(string(name) + "@" + version)
}

func (fakeDepsSource) ParseVersionedPackageFromConfiguration(dep string) (reposource.VersionedPackage, error) {
	return parseFakeDependency(dep)
}

func (fakeDepsSource) ParsePackageFromName(name reposource.PackageName) (reposource.Package, error) {
	return parseFakeDependency(string(name))
}

func (s *fakeDepsSource) ParsePackageFromRepoName(repoName api.RepoName) (reposource.Package, error) {
	return s.ParsePackageFromName(reposource.PackageName(strings.TrimPrefix(string(repoName), "fake/")))
}

type fakeVersionedPackage struct {
	name    reposource.PackageName
	version string
}

func parseFakeDependency(dep string) (reposource.VersionedPackage, error) {
	i := strings.LastIndex(dep, "@")
	if i == -1 {
		return fakeVersionedPackage{name: reposource.PackageName(dep)}, nil
	}
	return fakeVersionedPackage{name: reposource.PackageName(dep[:i]), version: dep[i+1:]}, nil
}

func (f fakeVersionedPackage) Scheme() string                        { return "fake" }
func (f fakeVersionedPackage) PackageSyntax() reposource.PackageName { return f.name }
func (f fakeVersionedPackage) VersionedPackageSyntax() string {
	return string(f.name) + "@" + f.version
}
func (f fakeVersionedPackage) PackageVersion() string    { return f.version }
func (f fakeVersionedPackage) Description() string       { return string(f.name) + "@" + f.version }
func (f fakeVersionedPackage) RepoName() api.RepoName    { return api.RepoName("fake/" + f.name) }
func (f fakeVersionedPackage) GitTagFromVersion() string { return "v" + f.version }
func (f fakeVersionedPackage) Less(other reposource.VersionedPackage) bool {
	return f.VersionedPackageSyntax() > other.VersionedPackageSyntax()
}

func (s *vcsPackagesSyncer) runCloneCommand(t *testing.T, examplePackageURL, bareGitDirectory string, dependencies []string) {
	u := vcs.URL{
		URL: url.URL{Path: examplePackageURL},
	}
	s.configDeps = dependencies
	cmd, err := s.CloneCommand(context.Background(), &u, bareGitDirectory)
	if err != nil {
		t.Fatalf("unexpected error building clone command: %s", err)
	}
	if err := cmd.Run(); err != nil {
		t.Fatalf("unexpected error running clone command: %s", err)
	}
}

func (s *vcsPackagesSyncer) assertDownloadCounts(t *testing.T, depsSource *fakeDepsSource, want map[string]int) {
	t.Helper()

	if diff := cmp.Diff(want, depsSource.downloadCount); diff != "" {
		t.Fatalf("unexpected difference in download counts (-want +got):\n%s", diff)
	}
}

func (s *vcsPackagesSyncer) assertRefs(t *testing.T, dir GitDir, want map[string]string) {
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

	if sc.Err() != nil {
		t.Fatalf("unexpected error while scanning: %v", sc.Err())
	}

	if diff := cmp.Diff(want, have); diff != "" {
		t.Fatalf("unexpected difference in git refs (-want +got):\n%s", diff)
	}
}
