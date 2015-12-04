package tracer

import (
	"fmt"
	"time"

	"sourcegraph.com/sourcegraph/appdash"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
)

func init() { appdash.RegisterEvent(GoVCS{}) }

// GoVCS records a go-vcs method invocation.
type GoVCS struct {
	Name, Args string

	StartTime time.Time
	EndTime   time.Time
}

// Schema returns the constant "GoVCS".
func (GoVCS) Schema() string { return "GoVCS" }

func (e GoVCS) Start() time.Time { return e.StartTime }
func (e GoVCS) End() time.Time   { return e.EndTime }

// Wrap wraps the given VCS repository, returning a repository which emits
// tracing events.
func Wrap(r vcs.Repository, rec *appdash.Recorder) vcs.Repository {
	// Wrap the repository.
	t := repository{r: r, rec: rec}

	// Also wrap optional interfaces. Yes, this is ugly. Yes, this works. No,
	// there isn't an easier way to do this. Because vcs.Repository has these
	// optional interfaces that we need to wrap, our only choice is to form an
	// anonymous struct which is the union of all optional interfaces that the
	// input repository implements.

	// Detect which optional interfaces the input vcs.Repository implements.
	realBlamer, isBlamer := r.(vcs.Blamer)
	realDiffer, isDiffer := r.(vcs.Differ)
	realCrossRepoDiffer, isCrossRepoDiffer := r.(vcs.CrossRepoDiffer)
	realFileLister, isFileLister := r.(vcs.FileLister)
	realMerger, isMerger := r.(vcs.Merger)
	realCrossRepoMerger, isCrossRepoMerger := r.(vcs.CrossRepoMerger)
	realRemoteUpdater, isRemoteUpdater := r.(vcs.RemoteUpdater)
	realSearcher, isSearcher := r.(vcs.Searcher)

	// Wrap the optional interfaces.
	blamer := blamer{repository: t, b: realBlamer, rec: rec}
	differ := differ{repository: t, d: realDiffer, rec: rec}
	crossRepoDiffer := crossRepoDiffer{repository: t, c: realCrossRepoDiffer, rec: rec}
	fileLister := fileLister{repository: t, f: realFileLister, rec: rec}
	merger := merger{repository: t, m: realMerger, rec: rec}
	crossRepoMerger := crossRepoMerger{repository: t, m: realCrossRepoMerger, rec: rec}
	remoteUpdater := remoteUpdater{repository: t, r: realRemoteUpdater, rec: rec}
	searcher := searcher{repository: t, s: realSearcher, rec: rec}

	// Return a union of all optional interfaces that the input vcs.Repository
	// implements.
	switch {
	case isBlamer && isDiffer && isCrossRepoDiffer && isFileLister && isMerger && isCrossRepoMerger && isRemoteUpdater && isSearcher:
		return struct {
			vcs.Repository
			vcs.Blamer
			vcs.Differ
			vcs.CrossRepoDiffer
			vcs.FileLister
			vcs.Merger
			vcs.CrossRepoMerger
			vcs.RemoteUpdater
			vcs.Searcher
		}{t, blamer, differ, crossRepoDiffer, fileLister, merger, crossRepoMerger, remoteUpdater, searcher}

	case isBlamer && isDiffer && isCrossRepoDiffer && isFileLister && isMerger && isCrossRepoMerger && isRemoteUpdater:
		return struct {
			vcs.Repository
			vcs.Blamer
			vcs.Differ
			vcs.CrossRepoDiffer
			vcs.FileLister
			vcs.Merger
			vcs.CrossRepoMerger
			vcs.RemoteUpdater
		}{t, blamer, differ, crossRepoDiffer, fileLister, merger, crossRepoMerger, remoteUpdater}

	case isBlamer && isDiffer && isCrossRepoDiffer && isFileLister && isMerger && isCrossRepoMerger:
		return struct {
			vcs.Repository
			vcs.Blamer
			vcs.Differ
			vcs.CrossRepoDiffer
			vcs.FileLister
			vcs.Merger
			vcs.CrossRepoMerger
		}{t, blamer, differ, crossRepoDiffer, fileLister, merger, crossRepoMerger}

	case isBlamer && isDiffer && isCrossRepoDiffer && isFileLister && isMerger:
		return struct {
			vcs.Repository
			vcs.Blamer
			vcs.Differ
			vcs.CrossRepoDiffer
			vcs.FileLister
			vcs.Merger
		}{t, blamer, differ, crossRepoDiffer, fileLister, merger}

	case isBlamer && isDiffer && isCrossRepoDiffer && isFileLister:
		return struct {
			vcs.Repository
			vcs.Blamer
			vcs.Differ
			vcs.CrossRepoDiffer
			vcs.FileLister
		}{t, blamer, differ, crossRepoDiffer, fileLister}

	case isBlamer && isDiffer && isCrossRepoDiffer:
		return struct {
			vcs.Repository
			vcs.Blamer
			vcs.Differ
			vcs.CrossRepoDiffer
		}{t, blamer, differ, crossRepoDiffer}

	case isBlamer && isDiffer:
		return struct {
			vcs.Repository
			vcs.Blamer
			vcs.Differ
		}{t, blamer, differ}

	case isBlamer:
		return struct {
			vcs.Repository
			vcs.Blamer
		}{t, blamer}

	default:
		return t
	}
}

// blamer implements the vcs.Repository interface, adding a wrapped vcs.Blamer
// implementation.
type blamer struct {
	repository
	b   vcs.Blamer
	rec *appdash.Recorder
}

// BlameFile implements the vcs.Blamer interface.
func (b blamer) BlameFile(path string, opt *vcs.BlameOptions) ([]*vcs.Hunk, error) {
	start := time.Now()
	hunks, err := b.b.BlameFile(path, opt)
	b.rec.Child().Event(GoVCS{
		Name:      "vcs.Blamer.BlameFile",
		Args:      fmt.Sprintf("%#v, %#v", path, opt),
		StartTime: start,
		EndTime:   time.Now(),
	})
	return hunks, err
}

// differ implements the vcs.Repository interface, adding a wrapped vcs.Differ
// implementation.
type differ struct {
	repository
	d   vcs.Differ
	rec *appdash.Recorder
}

// Diff implements the vcs.Differ interface.
func (d differ) Diff(base, head vcs.CommitID, opt *vcs.DiffOptions) (*vcs.Diff, error) {
	start := time.Now()
	diff, err := d.d.Diff(base, head, opt)
	d.rec.Child().Event(GoVCS{
		Name:      "vcs.Differ.Diff",
		Args:      fmt.Sprintf("%#v, %#v, %#v", base, head, opt),
		StartTime: start,
		EndTime:   time.Now(),
	})
	return diff, err
}

// crossRepoDiffer implements the vcs.Repository interface, adding a wrapped
// vcs.CrossRepoDiffer implementation.
type crossRepoDiffer struct {
	repository
	c   vcs.CrossRepoDiffer
	rec *appdash.Recorder
}

// CrossRepoDiff implements the vcs.CrossRepoDiffer interface.
func (c crossRepoDiffer) CrossRepoDiff(base vcs.CommitID, headRepo vcs.Repository, head vcs.CommitID, opt *vcs.DiffOptions) (*vcs.Diff, error) {
	start := time.Now()
	diff, err := c.c.CrossRepoDiff(base, headRepo, head, opt)
	c.rec.Child().Event(GoVCS{
		Name:      "vcs.CrossRepoDiffer.CrossRepoDiff",
		Args:      fmt.Sprintf("%#v, %#v, %#v, %#v", base, headRepo, head, opt),
		StartTime: start,
		EndTime:   time.Now(),
	})
	return diff, err
}

// fileLister implements the vcs.Repository interface, adding a wrapped
// vcs.FileListener implementation.
type fileLister struct {
	repository
	f   vcs.FileLister
	rec *appdash.Recorder
}

// ListFiles implements the vcs.FileLister interface.
func (f fileLister) ListFiles(commit vcs.CommitID) ([]string, error) {
	start := time.Now()
	files, err := f.f.ListFiles(commit)
	f.rec.Child().Event(GoVCS{
		Name:      "vcs.FileLister.ListFiles",
		Args:      fmt.Sprintf("%#v", commit),
		StartTime: start,
		EndTime:   time.Now(),
	})
	return files, err
}

// merger implements the vcs.Repository interface, adding a wrapped vcs.Merger
// implementation.
type merger struct {
	repository
	m   vcs.Merger
	rec *appdash.Recorder
}

// MergeBase implements the vcs.Merger interface.
func (m merger) MergeBase(a vcs.CommitID, b vcs.CommitID) (vcs.CommitID, error) {
	start := time.Now()
	commit, err := m.m.MergeBase(a, b)
	m.rec.Child().Event(GoVCS{
		Name:      "vcs.Merger.MergeBase",
		Args:      fmt.Sprintf("%#v, %#v", a, b),
		StartTime: start,
		EndTime:   time.Now(),
	})
	return commit, err
}

// crossRepoMerger implements the vcs.Repository interface, adding a wrapped
// vcs.CrossRepoMerger implementation.
type crossRepoMerger struct {
	repository
	m   vcs.CrossRepoMerger
	rec *appdash.Recorder
}

// CrossRepoMergeBase implements the vcs.CrossRepoMerger interface.
func (m crossRepoMerger) CrossRepoMergeBase(a vcs.CommitID, repoB vcs.Repository, b vcs.CommitID) (vcs.CommitID, error) {
	start := time.Now()
	commit, err := m.m.CrossRepoMergeBase(a, repoB, b)
	m.rec.Child().Event(GoVCS{
		Name:      "vcs.CrossRepoMerger.CrossRepoMergeBase",
		Args:      fmt.Sprintf("%#v, %#v, %#v", a, repoB, b),
		StartTime: start,
		EndTime:   time.Now(),
	})
	return commit, err
}

// remoteUpdater implements the vcs.Repository interface, adding a wrapped
// vcs.RemoteUpdater implementation.
type remoteUpdater struct {
	repository
	r   vcs.RemoteUpdater
	rec *appdash.Recorder
}

// UpdateEverything implements the vcs.RemoteUpdater interface.
func (r remoteUpdater) UpdateEverything(opts vcs.RemoteOpts) (*vcs.UpdateResult, error) {
	start := time.Now()
	result, err := r.r.UpdateEverything(opts)
	r.rec.Child().Event(GoVCS{
		Name:      "vcs.RemoteUpdater.UpdateEverything",
		Args:      fmt.Sprintf("%#v", opts),
		StartTime: start,
		EndTime:   time.Now(),
	})
	return result, err
}

// searcher implements the vcs.Repository interface, adding a wrapped
// vcs.Searcher implementation.
type searcher struct {
	repository
	s   vcs.Searcher
	rec *appdash.Recorder
}

// Search implements the vcs.Searcher interface.
func (s searcher) Search(commit vcs.CommitID, opts vcs.SearchOptions) ([]*vcs.SearchResult, error) {
	start := time.Now()
	results, err := s.s.Search(commit, opts)
	s.rec.Child().Event(GoVCS{
		Name:      "vcs.Searcher.Search",
		Args:      fmt.Sprintf("%#v, %#v", commit, opts),
		StartTime: start,
		EndTime:   time.Now(),
	})
	return results, err
}
