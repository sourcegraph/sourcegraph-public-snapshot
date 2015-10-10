package local

import (
	"fmt"
	"log"

	"code.google.com/p/rog-go/parallel"
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sourcegraph/srclib/graph"
	srcstore "sourcegraph.com/sourcegraph/srclib/store"
	"src.sourcegraph.com/sourcegraph/store"
)

func (s *defs) ListRefs(ctx context.Context, op *sourcegraph.DefsListRefsOp) (*sourcegraph.RefList, error) {
	defSpec := op.Def
	opt := op.Opt
	if opt == nil {
		opt = &sourcegraph.DefListRefsOptions{}
	}

	var repoFilters []srcstore.RefFilter
	if opt.Repo != "" {
		repoFilters = []srcstore.RefFilter{
			srcstore.ByRepos(opt.Repo),
		}
	} else {
		if defSpec.CommitID == "" {
			return nil, &sourcegraph.InvalidSpecError{Reason: "ListRefs: CommitID is empty"}
		}
		repoFilters = []srcstore.RefFilter{
			// TODO(sqs): don't restrict to same-commit
			srcstore.ByRepos(defSpec.Repo),
			srcstore.ByCommitIDs(defSpec.CommitID),
		}
	}
	refFilters := []srcstore.RefFilter{
		srcstore.ByRefDef(graph.RefDefKey{
			DefRepo:     defSpec.Repo,
			DefUnitType: defSpec.UnitType,
			DefUnit:     defSpec.Unit,
			DefPath:     defSpec.Path,
		}),
		srcstore.RefFilterFunc(func(ref *graph.Ref) bool { return !ref.Def }),
		srcstore.Limit(opt.Offset()+opt.Limit()+1, 0),
	}
	filters := append(repoFilters, refFilters...)
	bareRefs, err := store.GraphFromContext(ctx).Refs(filters...)
	if err != nil {
		return nil, err
	}

	// Convert to sourcegraph.Ref and file bareRefs.
	refs := make([]*sourcegraph.Ref, 0, opt.Limit())
	for i, bareRef := range bareRefs {
		if i >= opt.Offset() && i < (opt.Offset()+opt.Limit()) {
			refs = append(refs, &sourcegraph.Ref{Ref: *bareRef})
		}
	}
	hasMore := len(bareRefs) > opt.Offset()+opt.Limit()

	// Get authorship info, if requested.
	if opt.Authorship {
		// TODO(perf): optimize this to hit the cache more, assuming
		// we're blaming lots of small refs in the same file
		par := parallel.NewRun(8)
		for _, ref0 := range refs {
			ref := ref0
			par.Do(func() error {
				vcsrepo, err := store.RepoVCSFromContext(ctx).Open(ctx, ref.Repo)
				if err != nil {
					return err
				}
				br, ok := vcsrepo.(vcs.Blamer)
				if !ok {
					return &sourcegraph.NotImplementedError{What: fmt.Sprintf("repository %T does not support blaming files", vcsrepo)}
				}
				hunks, err := blameFileByteRange(br, ref.File, &vcs.BlameOptions{NewestCommit: vcs.CommitID(ref.CommitID)}, int(ref.Start), int(ref.End))
				if err != nil {
					return err
				}
				if len(hunks) != 1 {
					log.Printf("Warning: blaming ref %v: blame output has %d hunks, expected only one. Using first (or skipping if none).", ref, len(hunks))
				}
				if len(hunks) > 0 {
					h := hunks[0]
					ref.Authorship = &sourcegraph.AuthorshipInfo{
						AuthorEmail:    h.Author.Email, // TODO(privacy): leaks email addrs
						LastCommitDate: h.Author.Date,
						LastCommitID:   string(h.CommitID),
					}
				}
				return nil
			})
		}
		if err := par.Wait(); err != nil {
			log.Printf("Warning: error fetching ref authorship info for def %+v: %s. Continuing.", defSpec, err)
		}
	}

	return &sourcegraph.RefList{
		Refs:           refs,
		StreamResponse: sourcegraph.StreamResponse{HasMore: hasMore},
	}, nil
}
