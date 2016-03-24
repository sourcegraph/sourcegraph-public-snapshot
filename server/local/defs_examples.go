package local

import (
	"log"

	"golang.org/x/net/context"

	"github.com/rogpeppe/rog-go/parallel"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/svc"
)

func (s *defs) ListExamples(ctx context.Context, op *sourcegraph.DefsListExamplesOp) (*sourcegraph.ExampleList, error) {
	defSpec := op.Def
	opt := op.Opt
	if opt == nil {
		opt = &sourcegraph.DefListExamplesOptions{}
	}
	if opt.PerPage > 1000 {
		opt.PerPage = 1000
	}

	refs, err := svc.Defs(ctx).ListRefs(ctx, &sourcegraph.DefsListRefsOp{
		Def: defSpec,
		Opt: &sourcegraph.DefListRefsOptions{
			Repo:        opt.Repo,
			Files:       opt.Files,
			ListOptions: opt.ListOptions,
		},
	})
	if err != nil {
		return nil, err
	}

	examples := make([]*sourcegraph.Example, len(refs.Refs))
	par := parallel.NewRun(8)
	for loopI, loopRef := range refs.Refs {
		i, ref := loopI, loopRef

		par.Do(func() error {
			examples[i] = &sourcegraph.Example{Ref: *ref}

			entrySpec := sourcegraph.TreeEntrySpec{
				RepoRev: sourcegraph.RepoRevSpec{
					RepoSpec: sourcegraph.RepoSpec{URI: ref.Repo},
					Rev:      op.Rev,
					CommitID: ref.CommitID,
				},
				Path: ref.File,
			}
			opt := &sourcegraph.RepoTreeGetOptions{
				GetFileOptions: sourcegraph.GetFileOptions{
					FileRange: sourcegraph.FileRange{
						StartByte: int64(ref.Start), EndByte: int64(ref.End),
					},
					FullLines:          true,
					ExpandContextLines: 2,
				},
			}
			e, err := svc.RepoTree(ctx).Get(ctx, &sourcegraph.RepoTreeGetOp{Entry: entrySpec, Opt: opt})
			if err != nil {
				log.Printf("Error fetching VCS file %v in def examples query for def %+v: %s. Proceeding with other examples.", entrySpec, defSpec, err)
				examples[i].Error = true
				return nil
			}
			if e.Type != sourcegraph.FileEntry {
				examples[i].Error = true
				return nil
			}

			examples[i].Contents = string(e.Contents)
			examples[i].FileRange = *e.FileRange
			if op.Rev != "" {
				examples[i].Rev = op.Rev
			}
			return nil
		})
	}
	if err := par.Wait(); err != nil {
		return nil, err
	}

	return &sourcegraph.ExampleList{Examples: examples, StreamResponse: refs.StreamResponse}, nil
}
