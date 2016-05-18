package store

import (
	"fmt"
	"log"
	"path"
	"reflect"
	"strings"
	"sync"

	"sort"

	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

// A DefFilter filters a set of defs to only those for which SelectDef
// returns true.
type DefFilter interface {
	SelectDef(*graph.Def) bool
}

// DefFilters wraps a list of individual def filters and has a
// SelectDef method that returns true iff all individual def filters
// select the def.
type DefFilters []DefFilter

func (fs DefFilters) SelectDef(def *graph.Def) bool {
	for _, f := range fs {
		if !f.SelectDef(def) {
			return false
		}
	}
	return true
}

func (fs DefFilters) SelectDefs(defs ...*graph.Def) []*graph.Def {
	var sel []*graph.Def
	for _, def := range defs {
		if fs.SelectDef(def) {
			sel = append(sel, def)
		}
	}
	return sel
}

// A DefFilterFunc is a DefFilter that selects only those defs for
// which the func returns true.
type DefFilterFunc func(*graph.Def) bool

// SelectDef calls f(def).
func (f DefFilterFunc) SelectDef(def *graph.Def) bool { return f(def) }
func (f DefFilterFunc) String() string                { return "DefFilterFunc" }

// A RefFilter filters a set of refs to only those for which SelectRef
// returns true.
type RefFilter interface {
	SelectRef(*graph.Ref) bool
}

type refFilters []RefFilter

func (fs refFilters) SelectRef(ref *graph.Ref) bool {
	for _, f := range fs {
		if !f.SelectRef(ref) {
			return false
		}
	}
	return true
}

// A RefFilterFunc is a RefFilter that selects only those refs for
// which the func returns true.
type RefFilterFunc func(*graph.Ref) bool

// SelectRef calls f(ref).
func (f RefFilterFunc) SelectRef(ref *graph.Ref) bool { return f(ref) }
func (f RefFilterFunc) String() string                { return "RefFilterFunc" }

// A UnitFilter filters a set of units to only those for which Select
// returns true.
type UnitFilter interface {
	SelectUnit(*unit.SourceUnit) bool
}

type unitFilters []UnitFilter

func (fs unitFilters) SelectUnit(unit *unit.SourceUnit) bool {
	for _, f := range fs {
		if !f.SelectUnit(unit) {
			return false
		}
	}
	return true
}

// A UnitFilterFunc is a UnitFilter that selects only those units for
// which the func returns true.
type UnitFilterFunc func(*unit.SourceUnit) bool

// SelectUnit calls f(unit).
func (f UnitFilterFunc) SelectUnit(unit *unit.SourceUnit) bool { return f(unit) }
func (f UnitFilterFunc) String() string                        { return "UnitFilterFunc" }

// A VersionFilter filters a set of versions to only those for which SelectVersion
// returns true.
type VersionFilter interface {
	SelectVersion(*Version) bool
}

type versionFilters []VersionFilter

func (fs versionFilters) SelectVersion(version *Version) bool {
	for _, f := range fs {
		if !f.SelectVersion(version) {
			return false
		}
	}
	return true
}

// A VersionFilterFunc is a VersionFilter that selects only those
// versions for which the func returns true.
type VersionFilterFunc func(*Version) bool

// SelectVersion calls f(version).
func (f VersionFilterFunc) SelectVersion(version *Version) bool { return f(version) }
func (f VersionFilterFunc) String() string                      { return "VersionFilterFunc" }

// A RepoFilter filters a set of repos to only those for which SelectRepo
// returns true.
type RepoFilter interface {
	SelectRepo(string) bool
}

type repoFilters []RepoFilter

func (fs repoFilters) SelectRepo(repo string) bool {
	for _, f := range fs {
		if !f.SelectRepo(repo) {
			return false
		}
	}
	return true
}

// A RepoFilterFunc is a RepoFilter that selects only those repos for
// which the func returns true.
type RepoFilterFunc func(string) bool

// SelectRepo calls f(repo).
func (f RepoFilterFunc) SelectRepo(repo string) bool { return f(repo) }
func (f RepoFilterFunc) String() string              { return "RepoFilterFunc" }

// ByUnitsFilter is implemented by filters that restrict their
// selections to items that are in a set of source units. It allows
// the store to optimize calls by skipping data that it knows is not
// any of the the specified source units.
type ByUnitsFilter interface {
	ByUnits() []unit.ID2
}

// ByUnits creates a new filter that matches objects in any of the
// given source units. It panics if any of the unit IDs' names or
// types are empty.
func ByUnits(units ...unit.ID2) interface {
	DefFilter
	RefFilter
	UnitFilter
	ByUnitsFilter
} {
	for _, u := range units {
		if u.Type == "" {
			panic("unit.Type: empty")
		}
		if strings.Contains(u.Type, "/") {
			log.Printf("WARNING: srclib store.ByUnits was called with a source unit type of %q, which resembles a unit *name*. Did you mix up the order of ByUnits's arguments?", u.Type)
		}
	}
	return byUnitsFilter(units)
}

type byUnitsFilter []unit.ID2

func (f byUnitsFilter) contains(u unit.ID2) bool {
	for _, uu := range f {
		if uu == u {
			return true
		}
	}
	return false
}

func (f byUnitsFilter) String() string      { return fmt.Sprintf("ByUnits(%v)", ([]unit.ID2)(f)) }
func (f byUnitsFilter) ByUnits() []unit.ID2 { return f }
func (f byUnitsFilter) SelectDef(def *graph.Def) bool {
	return (def.Unit == "" && def.UnitType == "") || f.contains(unit.ID2{Type: def.UnitType, Name: def.Unit})
}
func (f byUnitsFilter) SelectRef(ref *graph.Ref) bool {
	return (ref.Unit == "" && ref.UnitType == "") || f.contains(unit.ID2{Type: ref.UnitType, Name: ref.Unit})
}
func (f byUnitsFilter) SelectUnit(unit *unit.SourceUnit) bool {
	return (unit.Type == "" && unit.Name == "") || f.contains(unit.ID2())
}

// ByCommitIDsFilter is implemented by filters that restrict their
// selection to items at specific commit IDs. It allows the store to
// optimize calls by skipping data that it knows is not at any of the
// specified commits.
type ByCommitIDsFilter interface {
	ByCommitIDs() []string
}

// ByCommitIDs creates a new filter by commit IDs. It panics if any
// commit ID is empty.
func ByCommitIDs(commitIDs ...string) interface {
	DefFilter
	RefFilter
	UnitFilter
	VersionFilter
	ByCommitIDsFilter
} {
	for _, c := range commitIDs {
		if c == "" {
			panic("empty commit ID")
		}
	}
	return byCommitIDsFilter(commitIDs)
}

type byCommitIDsFilter []string

func (f byCommitIDsFilter) String() string        { return fmt.Sprintf("ByCommitIDs(%s)", []string(f)) }
func (f byCommitIDsFilter) ByCommitIDs() []string { return []string(f) }
func (f byCommitIDsFilter) contains(commitID string) bool {
	for _, c := range f {
		if c == commitID {
			return true
		}
	}
	return false
}
func (f byCommitIDsFilter) SelectDef(def *graph.Def) bool {
	return def.CommitID == "" || f.contains(def.CommitID)
}
func (f byCommitIDsFilter) SelectRef(ref *graph.Ref) bool {
	return ref.CommitID == "" || f.contains(ref.CommitID)
}
func (f byCommitIDsFilter) SelectUnit(unit *unit.SourceUnit) bool {
	return unit.CommitID == "" || f.contains(unit.CommitID)
}
func (f byCommitIDsFilter) SelectVersion(version *Version) bool {
	return version.CommitID == "" || f.contains(version.CommitID)
}

// ByReposFilter is implemented by filters that restrict their
// selections to items in a set of repository. It allows the store to
// optimize calls by skipping data that it knows is not in any of the
// specified repositories.
type ByReposFilter interface {
	ByRepos() []string
}

// ByRepos creates a new filter by a set of repositories. It panics if
// repo is empty.
func ByRepos(repos ...string) interface {
	DefFilter
	RefFilter
	UnitFilter
	VersionFilter
	RepoFilter
	ByReposFilter
} {
	for _, repo := range repos {
		if repo == "" {
			panic("empty repo")
		}
	}
	return byReposFilter(repos)
}

type byReposFilter []string

func (f byReposFilter) String() string    { return fmt.Sprintf("ByRepos(%v)", []string(f)) }
func (f byReposFilter) ByRepos() []string { return []string(f) }
func (f byReposFilter) contains(repo string) bool {
	for _, r := range f {
		if r == repo {
			return true
		}
	}
	return false
}
func (f byReposFilter) SelectDef(def *graph.Def) bool {
	return def.Repo == "" || f.contains(def.Repo)
}
func (f byReposFilter) SelectRef(ref *graph.Ref) bool {
	return ref.Repo == "" || f.contains(ref.Repo)
}
func (f byReposFilter) SelectUnit(unit *unit.SourceUnit) bool {
	return unit.Repo == "" || f.contains(unit.Repo)
}
func (f byReposFilter) SelectVersion(version *Version) bool {
	return version.Repo == "" || f.contains(version.Repo)
}
func (f byReposFilter) SelectRepo(repo string) bool {
	return f.contains(repo)
}

// ByRepoCommitIDsFilter is implemented by filters that restrict their
// selections to items in a set of repositories (and in each
// repository, to a specific version). It allows the store to optimize
// calls by skipping data that it knows is not in any of the specified
// repository versions.
type ByRepoCommitIDsFilter interface {
	ByRepoCommitIDs() []Version
}

// ByRepoCommitIDs creates a new filter by a set of repository
// versions. It panics if repo is empty.
func ByRepoCommitIDs(versions ...Version) interface {
	DefFilter
	RefFilter
	UnitFilter
	VersionFilter
	RepoFilter
	ByReposFilter
	ByRepoCommitIDsFilter
} {
	for _, v := range versions {
		if v.Repo == "" {
			panic("empty version.Repo")
		}
		if v.CommitID == "" {
			panic("empty version.CommitID")
		}
	}
	return byRepoCommitIDsFilter(versions)
}

type byRepoCommitIDsFilter []Version

func (f byRepoCommitIDsFilter) String() string {
	return fmt.Sprintf("ByRepoCommitIDs(%v)", []Version(f))
}
func (f byRepoCommitIDsFilter) ByRepos() []string {
	repos := make([]string, len(f))
	for i, v := range f {
		repos[i] = v.Repo
	}
	return repos
}
func (f byRepoCommitIDsFilter) ByRepoCommitIDs() []Version { return []Version(f) }
func (f byRepoCommitIDsFilter) contains(repo, commitID string) bool {
	for _, v := range f {
		if v.Repo == repo && v.CommitID == commitID {
			return true
		}
	}
	return false
}
func (f byRepoCommitIDsFilter) SelectDef(def *graph.Def) bool {
	return (def.Repo == "" && def.CommitID == "") || f.contains(def.Repo, def.CommitID)
}
func (f byRepoCommitIDsFilter) SelectRef(ref *graph.Ref) bool {
	return (ref.Repo == "" && ref.CommitID == "") || f.contains(ref.Repo, ref.CommitID)
}
func (f byRepoCommitIDsFilter) SelectUnit(unit *unit.SourceUnit) bool {
	return (unit.Repo == "" && unit.CommitID == "") || f.contains(unit.Repo, unit.CommitID)
}
func (f byRepoCommitIDsFilter) SelectVersion(version *Version) bool {
	return (version.Repo == "" && version.CommitID == "") || f.contains(version.Repo, version.CommitID)
}
func (f byRepoCommitIDsFilter) SelectRepo(repo string) bool {
	for _, v := range f {
		if v.Repo == repo {
			return true
		}
	}
	return false
}

// ByUnitKey returns a filter by a source unit key. It panics if any
// fields on the unit key are not set. To filter by only source unit
// name and type, use ByUnits.
func ByUnitKey(key unit.Key) interface {
	DefFilter
	RefFilter
	UnitFilter
	ByReposFilter
	ByCommitIDsFilter
	ByUnitsFilter
} {
	if key.Repo == "" {
		panic("key.Repo: empty")
	}
	if key.CommitID == "" {
		panic("key.CommitID: empty")
	}
	if key.Type == "" {
		panic("key.Type: empty")
	}
	if key.Name == "" {
		panic("key.Name: empty")
	}
	return byUnitKeyFilter{key}
}

type byUnitKeyFilter struct{ key unit.Key }

func (f byUnitKeyFilter) String() string        { return fmt.Sprintf("ByUnitKey(%+v)", f.key) }
func (f byUnitKeyFilter) ByRepos() []string     { return []string{f.key.Repo} }
func (f byUnitKeyFilter) ByCommitIDs() []string { return []string{f.key.CommitID} }
func (f byUnitKeyFilter) ByUnits() []unit.ID2   { return []unit.ID2{f.key.ID2()} }
func (f byUnitKeyFilter) SelectDef(def *graph.Def) bool {
	return (def.Repo == "" || def.Repo == f.key.Repo) && (def.CommitID == "" || def.CommitID == f.key.CommitID) &&
		(def.UnitType == "" || def.UnitType == f.key.Type) && (def.Unit == "" || def.Unit == f.key.Name)
}
func (f byUnitKeyFilter) SelectRef(ref *graph.Ref) bool {
	return (ref.Repo == "" || ref.Repo == f.key.Repo) && (ref.CommitID == "" || ref.CommitID == f.key.CommitID) &&
		(ref.UnitType == "" || ref.UnitType == f.key.Type) && (ref.Unit == "" || ref.Unit == f.key.Name)
}
func (f byUnitKeyFilter) SelectUnit(unit *unit.SourceUnit) bool {
	return (unit.Repo == "" || unit.Repo == f.key.Repo) && (unit.CommitID == "" || unit.CommitID == f.key.CommitID) &&
		(unit.Type == "" || unit.Type == f.key.Type) && (unit.Name == "" || unit.Name == f.key.Name)
}

// ByDefKey returns a filter by a def key. It panics if the def path
// is not set. If you pass a ByDefKey filter to a store that's scoped
// to a specific repo/version/unit, then it will match all items in
// that repo/version/unit even if the
// key.Repo/key.CommitID/key.UnitType/key.Unit fields do not match
// (because stores do not "know" the repo/version/unit they store data
// for, and therefore they can't apply filter criteria for the level
// above them).
func ByDefKey(key graph.DefKey) interface {
	DefFilter
	ByReposFilter
	ByCommitIDsFilter
	ByUnitsFilter
} {
	if key.Path == "" {
		panic("key.Path: empty")
	}
	return byDefKeyFilter{key}
}

type byDefKeyFilter struct{ key graph.DefKey }

func (f byDefKeyFilter) String() string        { return fmt.Sprintf("ByDefKey(%+v)", f.key) }
func (f byDefKeyFilter) ByRepos() []string     { return []string{f.key.Repo} }
func (f byDefKeyFilter) ByCommitIDs() []string { return []string{f.key.CommitID} }
func (f byDefKeyFilter) ByUnits() []unit.ID2 {
	return []unit.ID2{{Type: f.key.UnitType, Name: f.key.Unit}}
}
func (f byDefKeyFilter) ByDefPath() string { return f.key.Path }
func (f byDefKeyFilter) SelectDef(def *graph.Def) bool {
	return (def.Repo == "" || def.Repo == f.key.Repo) && (def.CommitID == "" || def.CommitID == f.key.CommitID) &&
		(def.UnitType == "" || def.UnitType == f.key.UnitType) && (def.Unit == "" || def.Unit == f.key.Unit) &&
		def.Path == f.key.Path
}

// ByRefDefFilter is implemented by filters that restrict their
// selection to refs with a specific target definition.
type ByRefDefFilter interface {
	ByDefRepo() string
	ByDefUnitType() string
	ByDefUnit() string
	ByDefPath() string

	withEmptyImpliedValues() graph.RefDefKey // see docstring on impl method
}

// ByRefDef returns a filter by ref target def. It panics if
// def.DefPath is empty. If other fields are empty, they are assumed
// to match any value.
func ByRefDef(def graph.RefDefKey) RefFilter {
	if def.DefPath == "" {
		panic("def.DefPath: empty")
	}
	return &byRefDefFilter{def: def}
}

type byRefDefFilter struct {
	def graph.RefDefKey

	// These fields hold the repo and source unit that the filter is
	// being applied to. This workaround is necessary because
	// otherwise the filter would have no way to match the Ref.DefXyz
	// fields, since they are zeroed out if they equal the values for
	// the current repo and source unit.
	//
	// Unlike for other filters, a ByRefDef filter can't be used
	// (currently) to narrow the scope since a ref's def does not
	// determine where the ref is stored.

	impliedRepo string   // the implied DefRepo value when ref.DefRepo == ""
	impliedUnit unit.ID2 // the implied DefUnit{,Type} value when ref.DefUnit{,Type} == ""
}

func (f *byRefDefFilter) String() string {
	return fmt.Sprintf("ByRefDef(%+v, impliedRepo=%q, impliedUnit=%+v)", f.def, f.impliedRepo, f.impliedUnit)
}
func (f *byRefDefFilter) ByDefRepo() string          { return f.def.DefRepo }
func (f *byRefDefFilter) ByDefUnitType() string      { return f.def.DefUnitType }
func (f *byRefDefFilter) ByDefUnit() string          { return f.def.DefUnit }
func (f *byRefDefFilter) ByDefPath() string          { return f.def.DefPath }
func (f *byRefDefFilter) setImpliedRepo(repo string) { f.impliedRepo = repo }
func (f *byRefDefFilter) withImpliedUnit(u unit.ID2) RefFilter {
	newF := *f
	newF.impliedUnit = u
	return &newF
}
func (f *byRefDefFilter) SelectRef(ref *graph.Ref) bool {
	return ((ref.DefRepo == "" && f.impliedRepo == f.def.DefRepo) || ref.DefRepo == f.def.DefRepo) &&
		((ref.DefUnitType == "" && f.impliedUnit.Type == f.def.DefUnitType) || ref.DefUnitType == f.def.DefUnitType) &&
		((ref.DefUnit == "" && f.impliedUnit.Name == f.def.DefUnit) || ref.DefUnit == f.def.DefUnit) &&
		ref.DefPath == f.def.DefPath
}

// withEmptyImpliedValues returns the RefDefKey with empty field
// values for fields whose value in f.def matches the implied value.
//
// This is useful because in the index, the RefDefKey keys use the
// standard implicit values: a ref to a def in the same repo has an
// empty DefRepo, etc. But the ByRefDefFilter might have the def repo
// specified, since it was created at a higher level. Only set those
// values if they are not the same as the implicit ones (because the
// implicit ones should be blank).
func (f *byRefDefFilter) withEmptyImpliedValues() graph.RefDefKey {
	def := graph.RefDefKey{DefPath: f.ByDefPath()}
	if f.ByDefRepo() != f.impliedRepo {
		def.DefRepo = f.ByDefRepo()
	}
	if f.ByDefUnitType() != f.impliedUnit.Type {
		def.DefUnitType = f.ByDefUnitType()
	}
	if f.ByDefUnit() != f.impliedUnit.Name {
		def.DefUnit = f.ByDefUnit()
	}
	return def
}

var _ impliedRepoSetter = (*byRefDefFilter)(nil)
var _ impliedUnitSetter = (*byRefDefFilter)(nil)

// An AbsRefFilterFunc creates a RefFilter that selects only those
// refs for which the func returns true. Unlike RefFilterFunc, the
// ref's Def{Repo,UnitType,Unit,Path}, Repo, and CommitID fields are
// populated.
//
// AbsRefFilterFunc is less efficient than RefFilterFunc because it
// must make a copy of each ref before passing it to the func. If you
// don't need to access any of the fields it sets, use a
// RefFilterFunc.
func AbsRefFilterFunc(f RefFilterFunc) RefFilter {
	if f == nil {
		panic("AbsRefFilterFunc: f == nil")
	}
	return &absRefFilterFunc{f: f}
}

type absRefFilterFunc struct {
	f RefFilterFunc

	impliedRepo     string   // the implied DefRepo/Repo value when those are empty
	impliedCommitID string   // the CommitID currently being filtered
	impliedUnit     unit.ID2 // the implied DefUnitType/UnitType value when those are empty
}

func (f *absRefFilterFunc) String() string {
	return fmt.Sprintf("AbsRefFilterFunc(func %p, impliedRepo=%q, impliedCommitID=%q, impliedUnit=%+v)", f.f, f.impliedRepo, f.impliedCommitID, f.impliedUnit)
}
func (f *absRefFilterFunc) setImpliedRepo(repo string)         { f.impliedRepo = repo }
func (f *absRefFilterFunc) setImpliedCommitID(commitID string) { f.impliedCommitID = commitID }
func (f *absRefFilterFunc) withImpliedUnit(u unit.ID2) RefFilter {
	newF := *f
	newF.impliedUnit = u
	return &newF
}
func (f *absRefFilterFunc) SelectRef(ref *graph.Ref) bool {
	copy := *ref
	copy.Repo = f.impliedRepo
	copy.UnitType = f.impliedUnit.Type
	copy.Unit = f.impliedUnit.Name
	copy.CommitID = f.impliedCommitID
	if copy.DefRepo == "" {
		copy.DefRepo = f.impliedRepo
	}
	if copy.DefUnitType == "" {
		copy.DefUnitType = f.impliedUnit.Type
	}
	if copy.DefUnit == "" {
		copy.DefUnit = f.impliedUnit.Name
	}
	return f.f(&copy)
}

var _ impliedRepoSetter = (*absRefFilterFunc)(nil)
var _ impliedCommitIDSetter = (*absRefFilterFunc)(nil)
var _ impliedUnitSetter = (*absRefFilterFunc)(nil)

type impliedRepoSetter interface {
	setImpliedRepo(string)
}
type impliedCommitIDSetter interface {
	setImpliedCommitID(string)
}
type impliedUnitSetter interface {
	withImpliedUnit(unit.ID2) RefFilter
}

func setImpliedRepo(fs []RefFilter, repo string) {
	for _, f := range fs {
		if f, ok := f.(impliedRepoSetter); ok {
			f.setImpliedRepo(repo)
		}
	}
}

func setImpliedCommitID(fs []RefFilter, commitID string) {
	for _, f := range fs {
		if f, ok := f.(impliedCommitIDSetter); ok {
			f.setImpliedCommitID(commitID)
		}
	}
}

func withImpliedUnit(fs []RefFilter, u unit.ID2) []RefFilter {
	fCopy := make([]RefFilter, len(fs))
	for i, f := range fs {
		if fUnitSetter, ok := f.(impliedUnitSetter); ok {
			fCopy[i] = fUnitSetter.withImpliedUnit(u)
		} else {
			fCopy[i] = f
		}
	}
	return fCopy
}

// ByDefPathFilter is implemented by filters that restrict their
// selection to defs with a specific def path.
type ByDefPathFilter interface {
	ByDefPath() string
}

// ByDefPath returns a filter by def path. It panics if defPath is
// empty.
func ByDefPath(defPath string) interface {
	DefFilter
	ByDefPathFilter
} {
	if defPath == "" {
		panic("defPath: empty")
	}
	return byDefPathFilter(defPath)
}

type byDefPathFilter string

func (f byDefPathFilter) String() string    { return fmt.Sprintf("ByDefPath(%s)", string(f)) }
func (f byDefPathFilter) ByDefPath() string { return string(f) }
func (f byDefPathFilter) SelectDef(def *graph.Def) bool {
	return def.Path == string(f)
}

// ByDefQueryFilter is implemented by filters that restrict their
// selection to defs whose names match the query.
type ByDefQueryFilter interface {
	ByDefQuery() string
}

// ByDefQuery returns a filter by def query. It panics if q is empty.
func ByDefQuery(q string) interface {
	DefFilter
	ByDefQueryFilter
} {
	if q == "" {
		panic("ByDefQuery: empty")
	}
	return byDefQueryFilter(q)
}

type byDefQueryFilter string

func (f byDefQueryFilter) String() string     { return fmt.Sprintf("ByDefQuery(%q)", string(f)) }
func (f byDefQueryFilter) ByDefQuery() string { return string(f) }
func (f byDefQueryFilter) SelectDef(def *graph.Def) bool {
	// TODO(sqs): be smarter about the query matching semantics.
	return strings.HasPrefix(strings.ToLower(def.Name), strings.ToLower(string(f)))
}

// ByFilesFilter is implemented by filters that restrict their
// selection to defs, refs, etc., that exist in any file in a set, or
// source units that contain any of the files in the set.
type ByFilesFilter interface {
	ByFiles() []string
}

// ByFiles returns a filter that selects objects that are defined in
// or contain any of the listed files. It panics if any file path is
// empty, or if the file path has not been cleaned (i.e., if file !=
// path.Clean(file)).
//
// If exact == true, then only the exact files are accepted (i.e. a directory
// path will not include all files in that directory as it would normally).
func ByFiles(exact bool, files ...string) interface {
	DefFilter
	RefFilter
	UnitFilter
	ByFilesFilter
} {
	for _, f := range files {
		if f == "" {
			panic("file: empty")
		}
		if f != path.Clean(f) {
			panic("file: not cleaned (file != path.Clean(file))")
		}
	}
	return byFilesFilter{files: files, exact: exact}
}

type byFilesFilter struct {
	files []string
	exact bool
}

func (f byFilesFilter) String() string {
	return fmt.Sprintf("ByFiles(%v, exact=%t)", ([]string)(f.files), f.exact)
}
func (f byFilesFilter) ByFiles() []string { return f.files }
func (f byFilesFilter) SelectDef(def *graph.Def) bool {
	for _, ff := range f.files {
		if def.File == ff || (!f.exact && strings.HasPrefix(def.File, ff+"/")) {
			return true
		}
	}
	return false
}
func (f byFilesFilter) SelectRef(ref *graph.Ref) bool {
	for _, ff := range f.files {
		if ref.File == ff || (!f.exact && strings.HasPrefix(ref.File, ff+"/")) {
			return true
		}
	}
	return false
}
func (f byFilesFilter) SelectUnit(unit *unit.SourceUnit) bool {
	for _, unitFile := range unit.Files {
		for _, ff := range f.files {
			if ff == unitFile || (!f.exact && strings.HasPrefix(unitFile, ff+"/")) {
				return true
			}
		}
	}
	return false
}

// Limit is an EXPERIMENTAL filter for limiting the number of
// results. It is not correct because it assumes that if it is called
// on an object, it gets to decide whether that object appears in the
// final results. In reality, other filters could reject the object,
// and then those would count toward the limit for this filter but
// would never get returned. We could guarantee that Limit always runs
// last (after all other filters have accepted something).
func Limit(limit, offset int) interface {
	DefFilter
	RefFilter
} {
	return &limiter{n: limit, ofs: offset}
}

type limiter struct {
	n   int
	ofs int

	mu      sync.Mutex
	skipped map[interface{}]struct{}
	seen    map[interface{}]struct{}
}

func (l *limiter) String() string {
	return fmt.Sprintf("Limit(%d offset %d: %d remaining)", l.n, l.ofs, l.remainingOffsetPlusLimit())
}
func (l *limiter) remainingOffsetPlusLimit() int {
	l.mu.Lock()
	r := l.n + l.ofs - len(l.seen) - len(l.skipped)
	l.mu.Unlock()
	return r
}
func (l *limiter) SelectDef(def *graph.Def) bool { return l.selectObj(def) }
func (l *limiter) SelectRef(ref *graph.Ref) bool { return l.selectObj(ref) }
func (l *limiter) selectObj(obj interface{}) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.ofs > 0 && l.skipped == nil {
		l.skipped = make(map[interface{}]struct{}, l.ofs)
	}
	if len(l.skipped) < l.ofs {
		l.skipped[obj] = struct{}{}
		return false
	}
	if _, skipped := l.skipped[obj]; skipped {
		return false
	}
	if l.seen == nil {
		l.seen = make(map[interface{}]struct{}, l.n)
	}
	if _, seen := l.seen[obj]; seen {
		return true
	}
	if len(l.seen) < l.n {
		l.seen[obj] = struct{}{}
		return true
	}
	return false
}

// LimitRemaining returns how many more results may be added before
// the limit is exceeded (specified by the Limit filter, for
// example). If additional results may be added (either because the
// limit has not been reached, or there is no limit), moreOK is true.
func LimitRemaining(filters interface{}) (remaining int, moreOK bool) {
	for _, f := range storeFilters(filters) {
		switch f := f.(type) {
		case *limiter:
			m := f.remainingOffsetPlusLimit()
			return m, m > 0
		}
	}
	return 0, true
}

// storeFilters converts from slice-of-filter-type (e.g., []DefFilter,
// []UnitFilter) to []interface{}. It enables us to write generic
// functions that operate on any type of filter list without having
// duplicate verbose conversion code.
func storeFilters(anyFilters interface{}) []interface{} {
	// Optimized special cases for known filter types.
	switch o := anyFilters.(type) {
	case DefFilter:
		return []interface{}{o}
	case []DefFilter:
		fs := make([]interface{}, len(o))
		for i, f := range o {
			fs[i] = f
		}
		return fs
	}

	v := reflect.ValueOf(anyFilters)
	if !v.IsValid() {
		// no filters
		return nil
	}

	switch v.Kind() {
	case reflect.Slice:
		if v.Len() == 0 {
			return nil
		}
		filters := make([]interface{}, v.Len())
		for i := 0; i < v.Len(); i++ {
			filters[i] = v.Index(i).Interface()
		}
		return filters

	default:
		return []interface{}{anyFilters}
	}
}

// toTypedFilterSlice takes a typ like reflect.TypeOf([]DefFilter{})
// and a filters list like []interface{}{DefFilter1, DefFilter2} and
// returns the filters list as a []DefFilter{DefFilter1,
// DefFilter2}. It merely returns a modified type.
func toTypedFilterSlice(typ reflect.Type, filters []interface{}) interface{} {
	fs := reflect.MakeSlice(typ, len(filters), len(filters))
	for i, f := range filters {
		e := fs.Index(i)
		e.Set(reflect.ValueOf(f))
	}
	return fs.Interface()
}

type defsSortByName []*graph.Def

func (ds defsSortByName) Len() int           { return len(ds) }
func (ds defsSortByName) Swap(i, j int)      { ds[i], ds[j] = ds[j], ds[i] }
func (ds defsSortByName) Less(i, j int) bool { return ds[i].Name < ds[j].Name }

type DefsSortByName struct{}

func (ds DefsSortByName) String() string { return "DefsSortByName" }
func (ds DefsSortByName) DefsSort(defs []*graph.Def) {
	sort.Sort(defsSortByName(defs))
}
func (ds DefsSortByName) SelectDef(def *graph.Def) bool {
	return true
}

type DefsSortByKey struct{}

func (ds DefsSortByKey) String() string { return "DefsSortByKey" }
func (ds DefsSortByKey) DefsSort(defs []*graph.Def) {
	sort.Sort(graph.Defs(defs))
}
func (ds DefsSortByKey) SelectDef(def *graph.Def) bool {
	return true
}

type DefsSorter interface {
	DefsSort(defs []*graph.Def)
}
