package backend

import (
	"sort"
	"time"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/services/svc"
)

var RepoTree sourcegraph.RepoTreeServer = &repoTree{}

type repoTree struct{}

var _ sourcegraph.RepoTreeServer = (*repoTree)(nil)

func (s *repoTree) Get(ctx context.Context, op *sourcegraph.RepoTreeGetOp) (*sourcegraph.TreeEntry, error) {
	// Cap Get operation to some reasonable time.
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	entrySpec := op.Entry
	opt := op.Opt
	if opt == nil {
		opt = &sourcegraph.RepoTreeGetOptions{}
	}

	// It's OK if entrySpec is a dir. GetFileOptions will be ignored
	// by the vcsstore server in that case.
	entry0, err := s.getFromVCS(ctx, entrySpec, &opt.GetFileOptions)
	if err != nil {
		return nil, err
	}

	entry := &sourcegraph.TreeEntry{
		BasicTreeEntry: entry0.BasicTreeEntry,
	}
	if entry0.Type == sourcegraph.FileEntry {
		entry.FileRange = &entry0.FileRange
	}

	if opt.ContentsAsString {
		entry.ContentsString = string(entry.Contents)
		entry.Contents = nil
	}

	return entry, nil
}

// getFromVCS gets a tree entry from the vcsstore. Even though the
// return type is FileWithRange, it can return dirs too (the FileRange
// embedded struct will just be zeroed).
func (s *repoTree) getFromVCS(ctx context.Context, entrySpec sourcegraph.TreeEntrySpec, opt *sourcegraph.GetFileOptions) (*sourcegraph.FileWithRange, error) {
	if opt == nil {
		opt = &sourcegraph.GetFileOptions{}
	}

	if !isAbsCommitID(entrySpec.RepoRev.CommitID) {
		return nil, errNotAbsCommitID
	}

	vcsrepo, err := store.RepoVCSFromContext(ctx).Open(ctx, entrySpec.RepoRev.Repo)
	if err != nil {
		return nil, err
	}

	commit := vcs.CommitID(entrySpec.RepoRev.CommitID)

	fi, err := vcsrepo.Lstat(commit, entrySpec.Path)
	if err != nil {
		return nil, err
	}

	e := newTreeEntry(fi)
	e.CommitID = string(commit)
	fwr := sourcegraph.FileWithRange{BasicTreeEntry: e}

	if fi.Mode().IsDir() {
		ee, err := readDir(vcsrepo, commit, entrySpec.Path, int(opt.RecurseSingleSubfolderLimit), true)
		if err != nil {
			return nil, err
		}
		sort.Sort(TreeEntriesByTypeByName(ee))
		e.Entries = ee
	} else if fi.Mode().IsRegular() {
		contents, err := vcsrepo.ReadFile(commit, entrySpec.Path)
		if err != nil {
			return nil, err
		}

		e.Contents = contents

		if empty := (sourcegraph.GetFileOptions{}); *opt != empty {
			fr, _, err := computeFileRange(contents, *opt)
			if err != nil {
				return nil, err
			}

			// Trim to only requested range.
			e.Contents = e.Contents[fr.StartByte:fr.EndByte]
			fwr.FileRange = *fr
		}
	}

	return &fwr, nil
}

func (s *repoTree) List(ctx context.Context, op *sourcegraph.RepoTreeListOp) (*sourcegraph.RepoTreeListResult, error) {
	// Cap List operation to some reasonable time.
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	repoRevSpec := op.Rev

	if !isAbsCommitID(repoRevSpec.CommitID) {
		return nil, errNotAbsCommitID
	}

	repo, err := svc.Repos(ctx).Get(ctx, &sourcegraph.RepoSpec{ID: repoRevSpec.Repo})
	if err != nil {
		return nil, err
	}

	vcsrepo, err := store.RepoVCSFromContext(ctx).Open(ctx, repo.ID)
	if err != nil {
		return nil, err
	}

	infos, err := vcsrepo.ReadDir(vcs.CommitID(repoRevSpec.CommitID), ".", true)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, info := range infos {
		if !info.IsDir() {
			files = append(files, info.Name())
		}
	}

	return &sourcegraph.RepoTreeListResult{Files: files}, nil
}
