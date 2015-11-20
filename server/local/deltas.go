package local

import (
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"strings"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-diff/diff"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"src.sourcegraph.com/sourcegraph/emailaddrs"
	"src.sourcegraph.com/sourcegraph/errcode"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/svc"
)

var Deltas sourcegraph.DeltasServer = &deltas{}

type deltas struct {
	// mockDiffFunc, if set, is called by (deltas).diff instead of the
	// main method body. It allows mocking (deltas).diff in tests.
	mockDiffFunc func(context.Context, sourcegraph.DeltaSpec) ([]*diff.FileDiff, *sourcegraph.Delta, error)
}

var _ sourcegraph.DeltasServer = (*deltas)(nil)

func (s *deltas) Get(ctx context.Context, ds *sourcegraph.DeltaSpec) (*sourcegraph.Delta, error) {
	d := &sourcegraph.Delta{
		Base: ds.Base,
		Head: ds.Head,
	}

	get := func(repo **sourcegraph.Repo, repoRevSpec *sourcegraph.RepoRevSpec, commit **vcs.Commit, build **sourcegraph.Build) error {
		var err error
		*repo, err = svc.Repos(ctx).Get(ctx, &ds.Base.RepoSpec)
		if err != nil {
			return err
		}
		*commit, err = svc.Repos(ctx).GetCommit(ctx, repoRevSpec)
		if err != nil {
			return err
		}
		repoRevSpec.CommitID = string((*commit).ID)

		// Get build.
		buildInfo, err := svc.Builds(ctx).GetRepoBuildInfo(ctx, &sourcegraph.BuildsGetRepoBuildInfoOp{
			Repo: *repoRevSpec,
			Opt:  &sourcegraph.BuildsGetRepoBuildInfoOptions{Exact: true},
		})
		if err != nil && grpc.Code(err) != codes.NotFound {
			return err
		}
		if buildInfo != nil {
			*build = buildInfo.Exact
		}

		return nil
	}

	if err := get(&d.BaseRepo, &d.Base, &d.BaseCommit, &d.BaseBuild); err != nil {
		return d, err
	}
	if err := get(&d.HeadRepo, &d.Head, &d.HeadCommit, &d.HeadBuild); err != nil {
		return d, err
	}

	// Try to compute merge-base.
	vcsrepo, err := store.RepoVCSFromContext(ctx).Open(ctx, d.BaseRepo.URI)
	if err != nil {
		return d, err
	}
	var (
		ok bool
		id vcs.CommitID
	)
	if d.BaseRepo.URI == d.HeadRepo.URI {
		type MergeBaser interface {
			MergeBase(a, b vcs.CommitID) (vcs.CommitID, error)
		}
		var mBaser MergeBaser
		mBaser, ok = vcsrepo.(MergeBaser)
		if ok {
			id, err = mBaser.MergeBase(vcs.CommitID(d.BaseCommit.ID), vcs.CommitID(d.HeadCommit.ID))
			if err != nil {
				return d, err
			}
		}
	} else {
		type CrossRepoMergeBaser interface {
			CrossRepoMergeBase(a vcs.CommitID, headRepo vcs.Repository, b vcs.CommitID) (vcs.CommitID, error)
		}
		var crmBaser CrossRepoMergeBaser
		crmBaser, ok = vcsrepo.(CrossRepoMergeBaser)
		if ok {
			hrp, err := store.RepoVCSFromContext(ctx).Open(ctx, d.HeadRepo.URI)
			if err != nil {
				return d, err
			}
			id, err = crmBaser.CrossRepoMergeBase(vcs.CommitID(d.BaseCommit.ID), hrp, vcs.CommitID(d.HeadCommit.ID))
			if err != nil {
				return d, err
			}
		}
	}
	if ok && d.BaseCommit.ID != id {
		ds2 := *ds
		// There is most likely a merge conflict here, so we update the
		// delta to contain the actual merge base used in this diff A...B
		ds2.Base.CommitID = string(id)
		if strings.HasPrefix(ds.Base.CommitID, ds.Base.Rev) {
			// If the Revision is not a branch, but the commit ID, clear it.
			ds2.Base.Rev = ""
		}
		return s.Get(ctx, &ds2)
	}

	return d, nil
}

// listDefsOpt is the default options and pagination to use when fetching defs to list affected authors/dependents/etc.
//
// TODO(sqs): Because we only fetch the first 100 to compute the
// affected/impact results, it means we miss out on impact that is
// caused by defs after the 100th one. Make the ListAffected* methods
// fetch the entire set, not just the first 100, or show some kind of
// warning on the app.
var listDefsOpt = &sourcegraph.DeltaListDefsOptions{ListOptions: sourcegraph.ListOptions{PerPage: 100}}

func (s *deltas) ListAffectedAuthors(ctx context.Context, op *sourcegraph.DeltasListAffectedAuthorsOp) (*sourcegraph.DeltaAffectedPersonList, error) {
	ds := op.Ds
	opt := op.Opt

	listDefsOpt := *listDefsOpt
	if opt != nil {
		listDefsOpt.DeltaFilter = opt.DeltaFilter
	}
	defs, err := s.ListDefs(ctx, &sourcegraph.DeltasListDefsOp{Ds: ds, Opt: &listDefsOpt})
	if err != nil {
		return nil, err
	}

	defsChangedRemoved := baseDefsChangedAndRemoved(defs)
	authorsMap := map[sourcegraph.PersonSpec]*sourcegraph.DeltaAffectedPerson{}
	for _, def := range defsChangedRemoved {
		authors, err := svc.Defs(ctx).ListAuthors(ctx, &sourcegraph.DefsListAuthorsOp{
			Def: sourcegraph.NewDefSpecFromDefKey(def.DefKey),
			Opt: &sourcegraph.DefListAuthorsOptions{ListOptions: sourcegraph.ListOptions{PerPage: 100}},
		})
		if err != nil {
			if errcode.GRPC(err) == codes.NotFound {
				// This occurs when a def's file doesn't refer to an
				// existing file (e.g., a package def's File points to
				// a dir).
				log.Printf("Warning: ListAffectedAuthors: couldn't ListAuthors for def %v: %s.", def.DefKey, err)
				continue
			}
			return nil, err
		}
		for _, a := range authors.DefAuthors {
			var key sourcegraph.PersonSpec
			if a.Email != "" {
				email, err := emailaddrs.Deobfuscate(a.Email)
				if err != nil {
					return nil, err
				}
				key.Email = email
			} else {
				key.UID = a.UID
			}

			if _, present := authorsMap[key]; present {
				authorsMap[key].Defs = append(authorsMap[key].Defs, def)
			} else {
				person, err := svc.People(ctx).Get(ctx, &key)
				if err != nil {
					return nil, err
				}

				authorsMap[key] = &sourcegraph.DeltaAffectedPerson{
					Person: *person,
					Defs:   []*sourcegraph.Def{def},
				}
			}
		}
	}

	allAuthors := make([]*sourcegraph.DeltaAffectedPerson, len(authorsMap))
	i := 0
	for _, a := range authorsMap {
		allAuthors[i] = a
		i++
	}

	return &sourcegraph.DeltaAffectedPersonList{DeltaAffectedPersons: allAuthors}, nil
}

func (s *deltas) ListAffectedClients(ctx context.Context, op *sourcegraph.DeltasListAffectedClientsOp) (*sourcegraph.DeltaAffectedPersonList, error) {
	ds := op.Ds
	opt := op.Opt

	listDefsOpt := *listDefsOpt
	if opt != nil {
		listDefsOpt.DeltaFilter = opt.DeltaFilter
	}
	defs, err := s.ListDefs(ctx, &sourcegraph.DeltasListDefsOp{Ds: ds, Opt: &listDefsOpt})
	if err != nil {
		return nil, err
	}

	defsChangedRemoved := baseDefsChangedAndRemoved(defs)
	clientsMap := map[sourcegraph.PersonSpec]*sourcegraph.DeltaAffectedPerson{}
	for _, def := range defsChangedRemoved {
		clients, err := svc.Defs(ctx).ListClients(ctx, &sourcegraph.DefsListClientsOp{
			Def: sourcegraph.NewDefSpecFromDefKey(def.DefKey),
			Opt: &sourcegraph.DefListClientsOptions{ListOptions: sourcegraph.ListOptions{PerPage: 100}},
		})
		if err != nil {
			return nil, err
		}

		for _, a := range clients.DefClients {
			var key sourcegraph.PersonSpec
			if a.Email != "" {
				email, err := emailaddrs.Deobfuscate(a.Email)
				if err != nil {
					return nil, err
				}
				key.Email = email
			} else {
				key.UID = a.UID
			}

			if _, present := clientsMap[key]; present {
				clientsMap[key].Defs = append(clientsMap[key].Defs, def)
			} else {
				person, err := svc.People(ctx).Get(ctx, &key)
				if err != nil {
					return nil, err
				}

				clientsMap[key] = &sourcegraph.DeltaAffectedPerson{
					Person: *person,
					Defs:   []*sourcegraph.Def{def},
				}
			}
		}
	}

	allClients := make([]*sourcegraph.DeltaAffectedPerson, len(clientsMap))
	i := 0
	for _, a := range clientsMap {
		allClients[i] = a
		i++
	}

	return &sourcegraph.DeltaAffectedPersonList{DeltaAffectedPersons: allClients}, nil
}
