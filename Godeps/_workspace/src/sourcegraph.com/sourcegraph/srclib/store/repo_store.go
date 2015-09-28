package store

import (
	"sync"

	"code.google.com/p/rog-go/parallel"
	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

// A RepoStore stores and accesses srclib build data for a repository
// (consisting of any number of commits, each of which have any number
// of source units).
type RepoStore interface {
	// Versions returns all commits that match the VersionFilter.
	Versions(...VersionFilter) ([]*Version, error)

	// TreeStore's methods call the corresponding methods on the
	// TreeStore of each version contained within this repository. The
	// combined results are returned (in undefined order).
	TreeStore
}

// A RepoImporter imports srclib build data for a source unit at a
// specific version into a RepoStore.
type RepoImporter interface {
	// Import imports srclib build data for a source unit at a
	// specific version into the store.
	Import(commitID string, unit *unit.SourceUnit, data graph.Output) error
}

type RepoIndexer interface {
	// Index builds indexes for the store.
	Index(commitID string) error
}

// A RepoStoreImporter implements both RepoStore and RepoImporter.
type RepoStoreImporter interface {
	RepoStore
	RepoImporter
}

// A VersionKey is a unique identifier for a version across all
// repositories.
type VersionKey struct {
	// Repo is the URI of the commit's repository.
	Repo string

	// CommitID is the commit ID of the commit.
	CommitID string
}

// A Version represents a revision (i.e., commit) of a repository.
type Version struct {
	// Repo is the URI of the repository that contains this commit.
	Repo string

	// CommitID is the commit ID of the VCS revision that this version
	// represents. If blank, then this version refers to the current
	// workspace.
	CommitID string

	// TODO(sqs): add build metadata fields (build logs, timings, what
	// was actually built, incremental build tracking, diff/pack
	// compression helper info, etc.)
}

// IsCurrentWorkspace returns a boolean indicating whether this
// version represents the current workspace, as opposed to a specific
// VCS commit.
func (v Version) IsCurrentWorkspace() bool { return v.CommitID == "" }

// A repoStores is a RepoStore whose methods call the
// corresponding method on each of the repo stores returned by the
// repoStores func.
type repoStores struct {
	opener repoStoreOpener
}

var _ RepoStore = (*repoStores)(nil)

func (s repoStores) Versions(f ...VersionFilter) ([]*Version, error) {
	rss, err := openRepoStores(s.opener, f)
	if err != nil {
		return nil, err
	}

	var allVersions []*Version
	for repo, rs := range rss {
		if rs == nil {
			continue
		}

		versions, err := rs.Versions(filtersForRepo(repo, f).([]VersionFilter)...)
		if err != nil && !isStoreNotExist(err) {
			return nil, err
		}
		for _, version := range versions {
			version.Repo = repo
		}
		allVersions = append(allVersions, versions...)
	}
	return allVersions, nil
}

func (s repoStores) Units(f ...UnitFilter) ([]*unit.SourceUnit, error) {
	rss, err := openRepoStores(s.opener, f)
	if err != nil {
		return nil, err
	}

	var (
		allUnits   []*unit.SourceUnit
		allUnitsMu sync.Mutex
	)
	par := parallel.NewRun(storeFetchPar)
	for repo_, rs_ := range rss {
		repo, rs := repo_, rs_
		if rs == nil {
			continue
		}

		par.Do(func() error {
			units, err := rs.Units(filtersForRepo(repo, f).([]UnitFilter)...)
			if err != nil && !isStoreNotExist(err) {
				return err
			}
			for _, unit := range units {
				unit.Repo = repo
			}
			allUnitsMu.Lock()
			allUnits = append(allUnits, units...)
			allUnitsMu.Unlock()
			return nil
		})
	}
	err = par.Wait()
	return allUnits, err
}

func (s repoStores) Defs(f ...DefFilter) ([]*graph.Def, error) {
	rss, err := openRepoStores(s.opener, f)
	if err != nil {
		return nil, err
	}

	var (
		allDefs   []*graph.Def
		allDefsMu sync.Mutex
	)
	par := parallel.NewRun(storeFetchPar)
	for repo_, rs_ := range rss {
		repo, rs := repo_, rs_
		if rs == nil {
			continue
		}

		par.Do(func() error {
			defs, err := rs.Defs(filtersForRepo(repo, f).([]DefFilter)...)
			if err != nil && !isStoreNotExist(err) {
				return err
			}
			for _, def := range defs {
				def.Repo = repo
			}
			allDefsMu.Lock()
			allDefs = append(allDefs, defs...)
			allDefsMu.Unlock()
			return nil
		})
	}
	err = par.Wait()
	return allDefs, err
}

func (s repoStores) Refs(f ...RefFilter) ([]*graph.Ref, error) {
	rss, err := openRepoStores(s.opener, f)
	if err != nil {
		return nil, err
	}

	var allRefs []*graph.Ref
	for repo, rs := range rss {
		if rs == nil {
			continue
		}

		setImpliedRepo(f, repo)
		refs, err := rs.Refs(filtersForRepo(repo, f).([]RefFilter)...)
		if err != nil && !isStoreNotExist(err) {
			return nil, err
		}
		for _, ref := range refs {
			ref.Repo = repo
			if ref.DefRepo == "" {
				ref.DefRepo = repo
			}
		}
		allRefs = append(allRefs, refs...)
	}
	return allRefs, nil
}
