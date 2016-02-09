package tracer

import (
	"fmt"
	"time"

	"sourcegraph.com/sourcegraph/appdash"
	"src.sourcegraph.com/sourcegraph/pkg/vcs"
	"src.sourcegraph.com/sourcegraph/pkg/vcs/gitcmd"
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
func Wrap(r *gitcmd.Repository, rec *appdash.Recorder) vcs.Repository {
	return repository{r: r, rec: rec}
}

func (r repository) BlameFile(path string, opt *vcs.BlameOptions) ([]*vcs.Hunk, error) {
	start := time.Now()
	hunks, err := r.r.BlameFile(path, opt)
	r.rec.Child().Event(GoVCS{
		Name:      "vcs.Blamer.BlameFile",
		Args:      fmt.Sprintf("%#v, %#v", path, opt),
		StartTime: start,
		EndTime:   time.Now(),
	})
	return hunks, err
}

func (r repository) Diff(base, head vcs.CommitID, opt *vcs.DiffOptions) (*vcs.Diff, error) {
	start := time.Now()
	diff, err := r.r.Diff(base, head, opt)
	r.rec.Child().Event(GoVCS{
		Name:      "vcs.Differ.Diff",
		Args:      fmt.Sprintf("%#v, %#v, %#v", base, head, opt),
		StartTime: start,
		EndTime:   time.Now(),
	})
	return diff, err
}

func (r repository) CrossRepoDiff(base vcs.CommitID, headRepo vcs.Repository, head vcs.CommitID, opt *vcs.DiffOptions) (*vcs.Diff, error) {
	start := time.Now()
	diff, err := r.r.CrossRepoDiff(base, headRepo, head, opt)
	r.rec.Child().Event(GoVCS{
		Name:      "vcs.CrossRepoDiffer.CrossRepoDiff",
		Args:      fmt.Sprintf("%#v, %#v, %#v, %#v", base, headRepo, head, opt),
		StartTime: start,
		EndTime:   time.Now(),
	})
	return diff, err
}

func (r repository) ListFiles(commit vcs.CommitID) ([]string, error) {
	start := time.Now()
	files, err := r.r.ListFiles(commit)
	r.rec.Child().Event(GoVCS{
		Name:      "vcs.FileLister.ListFiles",
		Args:      fmt.Sprintf("%#v", commit),
		StartTime: start,
		EndTime:   time.Now(),
	})
	return files, err
}

func (r repository) MergeBase(a vcs.CommitID, b vcs.CommitID) (vcs.CommitID, error) {
	start := time.Now()
	commit, err := r.r.MergeBase(a, b)
	r.rec.Child().Event(GoVCS{
		Name:      "vcs.Merger.MergeBase",
		Args:      fmt.Sprintf("%#v, %#v", a, b),
		StartTime: start,
		EndTime:   time.Now(),
	})
	return commit, err
}

func (r repository) CrossRepoMergeBase(a vcs.CommitID, repoB vcs.Repository, b vcs.CommitID) (vcs.CommitID, error) {
	start := time.Now()
	commit, err := r.r.CrossRepoMergeBase(a, repoB, b)
	r.rec.Child().Event(GoVCS{
		Name:      "vcs.CrossRepoMerger.CrossRepoMergeBase",
		Args:      fmt.Sprintf("%#v, %#v, %#v", a, repoB, b),
		StartTime: start,
		EndTime:   time.Now(),
	})
	return commit, err
}

func (r repository) UpdateEverything(opts vcs.RemoteOpts) (*vcs.UpdateResult, error) {
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

func (r repository) Search(commit vcs.CommitID, opts vcs.SearchOptions) ([]*vcs.SearchResult, error) {
	start := time.Now()
	results, err := r.r.Search(commit, opts)
	r.rec.Child().Event(GoVCS{
		Name:      "vcs.Searcher.Search",
		Args:      fmt.Sprintf("%#v, %#v", commit, opts),
		StartTime: start,
		EndTime:   time.Now(),
	})
	return results, err
}
