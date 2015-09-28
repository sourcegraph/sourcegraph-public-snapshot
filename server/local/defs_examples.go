package local

import (
	"log"

	"golang.org/x/net/context"

	"code.google.com/p/rog-go/parallel"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/svc"
	"sourcegraph.com/sourcegraph/vcsstore/vcsclient"
)

func (s *defs) ListExamples(ctx context.Context, op *sourcegraph.DefsListExamplesOp) (*sourcegraph.ExampleList, error) {
	defSpec := op.Def
	opt := op.Opt
	if opt == nil {
		opt = &sourcegraph.DefListExamplesOptions{}
	}

	refs, err := svc.Defs(ctx).ListRefs(ctx, &sourcegraph.DefsListRefsOp{
		Def: defSpec,
		Opt: &sourcegraph.DefListRefsOptions{
			Repo:        opt.Repo,
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
			examples[i] = &sourcegraph.Example{Ref: ref.Ref}

			if opt.Formatted || opt.TokenizedSource {
				entrySpec := sourcegraph.TreeEntrySpec{
					RepoRev: sourcegraph.RepoRevSpec{
						RepoSpec: sourcegraph.RepoSpec{URI: ref.Repo},
						Rev:      op.Rev,
						CommitID: ref.CommitID,
					},
					Path: ref.File,
				}
				opt := &sourcegraph.RepoTreeGetOptions{
					Formatted:       opt.Formatted,
					TokenizedSource: opt.TokenizedSource,
					GetFileOptions: vcsclient.GetFileOptions{
						FileRange: vcsclient.FileRange{
							StartByte: int64(ref.Start), EndByte: int64(ref.End),
						},
						FullLines:          true,
						ExpandContextLines: 4,
					},
				}
				e, err := svc.RepoTree(ctx).Get(ctx, &sourcegraph.RepoTreeGetOp{Entry: entrySpec, Opt: opt})
				if err != nil {
					log.Printf("Error fetching VCS file %v in def examples query for def %+v: %s. Proceeding with other examples.", entrySpec, defSpec, err)
					examples[i].Error = true
					return nil
				}
				if e.Type != vcsclient.FileEntry {
					examples[i].Error = true
					return nil
				}

				if e.SourceCode != nil {
					examples[i].SourceCode = e.SourceCode
				} else if opt.Formatted {
					examples[i].SrcHTML = string(e.Contents)
				}

				examples[i].StartLine = int32(e.StartLine)
				examples[i].EndLine = int32(e.EndLine)
				if op.Rev != "" {
					examples[i].Rev = op.Rev
				}
			}
			return nil
		})
	}
	if err := par.Wait(); err != nil {
		return nil, err
	}

	return &sourcegraph.ExampleList{Examples: examples, ListResponse: refs.ListResponse}, nil
}
