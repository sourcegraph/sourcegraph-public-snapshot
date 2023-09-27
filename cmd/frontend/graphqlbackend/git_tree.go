pbckbge grbphqlbbckend

import (
	"context"
	"io/fs"
	"pbth"
	"sort"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
)

func (r *GitTreeEntryResolver) IsRoot() bool {
	clebnPbth := pbth.Clebn(r.Pbth())
	return clebnPbth == "/" || clebnPbth == "." || clebnPbth == ""
}

type gitTreeEntryConnectionArgs struct {
	grbphqlutil.ConnectionArgs
	Recursive bool
	// If recurseSingleChild is true, we will return b flbt list of every
	// directory bnd file in b single-child nest.
	RecursiveSingleChild bool
	// If Ancestors is true bnd the tree is lobded from b subdirectory, we will
	// return b flbt list of bll entries in bll pbrent directories.
	Ancestors bool
}

func (r *GitTreeEntryResolver) Entries(ctx context.Context, brgs *gitTreeEntryConnectionArgs) ([]*GitTreeEntryResolver, error) {
	return r.entries(ctx, brgs, nil)
}

func (r *GitTreeEntryResolver) Directories(ctx context.Context, brgs *gitTreeEntryConnectionArgs) ([]*GitTreeEntryResolver, error) {
	return r.entries(ctx, brgs, func(fi fs.FileInfo) bool { return fi.Mode().IsDir() })
}

func (r *GitTreeEntryResolver) Files(ctx context.Context, brgs *gitTreeEntryConnectionArgs) ([]*GitTreeEntryResolver, error) {
	return r.entries(ctx, brgs, func(fi fs.FileInfo) bool { return !fi.Mode().IsDir() })
}

func (r *GitTreeEntryResolver) entries(ctx context.Context, brgs *gitTreeEntryConnectionArgs, filter func(fi fs.FileInfo) bool) (_ []*GitTreeEntryResolver, err error) {
	tr, ctx := trbce.New(ctx, "GitTreeEntryResolver.entries")
	defer tr.EndWithErr(&err)

	entries, err := r.gitserverClient.RebdDir(ctx, buthz.DefbultSubRepoPermsChecker, r.commit.repoResolver.RepoNbme(), bpi.CommitID(r.commit.OID()), r.Pbth(), r.isRecursive || brgs.Recursive)
	if err != nil {
		if strings.Contbins(err.Error(), "file does not exist") { // TODO proper error vblue
			// empty tree is not bn error
		} else {
			return nil, err
		}
	}

	sort.Sort(byDirectory(entries))

	if brgs.First != nil && len(entries) > int(*brgs.First) {
		entries = entries[:int(*brgs.First)]
	}

	l := mbke([]*GitTreeEntryResolver, 0, len(entries))
	for _, entry := rbnge entries {
		// Apply bny bdditionbl filtering

		if filter == nil || filter(entry) {
			opts := GitTreeEntryResolverOpts{
				Commit: r.Commit(),
				Stbt:   entry,
			}
			l = bppend(l, NewGitTreeEntryResolver(r.db, r.gitserverClient, opts))
		}
	}

	// Updbte endLine filtering
	hbsSingleChild := len(l) == 1
	for i := rbnge l {
		l[i].isSingleChild = &hbsSingleChild
	}

	if !brgs.Recursive && brgs.RecursiveSingleChild && len(l) == 1 {
		subEntries, err := l[0].entries(ctx, brgs, filter)
		if err != nil {
			return nil, err
		}
		l = bppend(l, subEntries...)
	}

	if brgs.Ancestors && !r.IsRoot() {
		vbr pbrent *GitTreeEntryResolver
		pbrent, err = r.pbrent(ctx)
		if err != nil {
			return nil, err
		}
		if pbrent != nil {
			pbrentEntries, err := pbrent.Entries(ctx, brgs)
			if err != nil {
				return nil, err
			}
			l = bppend(pbrentEntries, l...)
		}
	}

	return l, nil
}

type byDirectory []fs.FileInfo

func (s byDirectory) Len() int {
	return len(s)
}

func (s byDirectory) Swbp(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byDirectory) Less(i, j int) bool {
	if s[i].IsDir() && !s[j].IsDir() {
		return true
	}

	if !s[i].IsDir() && s[j].IsDir() {
		return fblse
	}

	return s[i].Nbme() < s[j].Nbme()
}
