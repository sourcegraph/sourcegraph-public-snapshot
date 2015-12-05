package tracer

import (
	"fmt"
	"time"

	"golang.org/x/tools/godoc/vfs"
	"sourcegraph.com/sourcegraph/appdash"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
)

// repository implements the vcs.Repository interface.
type repository struct {
	r   vcs.Repository
	rec *appdash.Recorder
}

// ResolveRevision implements the vcs.Repository interface.
func (r repository) ResolveRevision(spec string) (vcs.CommitID, error) {
	start := time.Now()
	rev, err := r.r.ResolveRevision(spec)
	r.rec.Child().Event(GoVCS{
		Name:      "vcs.Repository.ResolveRevision",
		Args:      fmt.Sprintf("%#v", spec),
		StartTime: start,
		EndTime:   time.Now(),
	})
	return rev, err
}

// ResolveTag implements the vcs.Repository interface.
func (r repository) ResolveTag(name string) (vcs.CommitID, error) {
	start := time.Now()
	tag, err := r.r.ResolveTag(name)
	r.rec.Child().Event(GoVCS{
		Name:      "vcs.Repository.ResolveTag",
		Args:      fmt.Sprintf("%#v", name),
		StartTime: start,
		EndTime:   time.Now(),
	})
	return tag, err
}

// ResolveBranch implements the vcs.Repository interface.
func (r repository) ResolveBranch(name string) (vcs.CommitID, error) {
	start := time.Now()
	branch, err := r.r.ResolveBranch(name)
	r.rec.Child().Event(GoVCS{
		Name:      "vcs.Repository.ResolveBranch",
		Args:      fmt.Sprintf("%#v", name),
		StartTime: start,
		EndTime:   time.Now(),
	})
	return branch, err
}

// Branches implements the vcs.Repository interface.
func (r repository) Branches(opt vcs.BranchesOptions) ([]*vcs.Branch, error) {
	start := time.Now()
	branches, err := r.r.Branches(opt)
	r.rec.Child().Event(GoVCS{
		Name:      "vcs.Repository.Branches",
		Args:      fmt.Sprintf("%#v", opt),
		StartTime: start,
		EndTime:   time.Now(),
	})
	return branches, err
}

// Tags implements the vcs.Repository interface.
func (r repository) Tags() ([]*vcs.Tag, error) {
	start := time.Now()
	tags, err := r.r.Tags()
	r.rec.Child().Event(GoVCS{
		Name:      "vcs.Repository.Tags",
		StartTime: start,
		EndTime:   time.Now(),
	})
	return tags, err
}

// GetCommit implements the vcs.Repository interface.
func (r repository) GetCommit(commitID vcs.CommitID) (*vcs.Commit, error) {
	start := time.Now()
	commit, err := r.r.GetCommit(commitID)
	r.rec.Child().Event(GoVCS{
		Name:      "vcs.Repository.GetCommit",
		Args:      fmt.Sprintf("%#v", commitID),
		StartTime: start,
		EndTime:   time.Now(),
	})
	return commit, err
}

// Commits implements the vcs.Repository interface.
func (r repository) Commits(opt vcs.CommitsOptions) (commits []*vcs.Commit, total uint, err error) {
	start := time.Now()
	commits, total, err = r.r.Commits(opt)
	r.rec.Child().Event(GoVCS{
		Name:      "vcs.Repository.Commits",
		Args:      fmt.Sprintf("%#v", opt),
		StartTime: start,
		EndTime:   time.Now(),
	})
	return commits, total, err
}

// Committers implements the vcs.Repository interface.
func (r repository) Committers(opt vcs.CommittersOptions) ([]*vcs.Committer, error) {
	start := time.Now()
	committers, err := r.r.Committers(opt)
	r.rec.Child().Event(GoVCS{
		Name:      "vcs.Repository.Committers",
		Args:      fmt.Sprintf("%#v", opt),
		StartTime: start,
		EndTime:   time.Now(),
	})
	return committers, err
}

// FileSystem implements the vcs.Repository interface.
func (r repository) FileSystem(at vcs.CommitID) (vfs.FileSystem, error) {
	start := time.Now()
	fs, err := r.r.FileSystem(at)
	r.rec.Child().Event(GoVCS{
		Name:      "vcs.Repository.FileSystem",
		Args:      fmt.Sprintf("%#v", at),
		StartTime: start,
		EndTime:   time.Now(),
	})
	if err != nil {
		return nil, err
	}
	return fileSystem{fs: fs, rec: r.rec}, nil
}
