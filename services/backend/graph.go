package backend

import (
	"golang.org/x/net/context"
	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
	"sourcegraph.com/sourcegraph/srclib/store/pb"
	"sourcegraph.com/sourcegraph/srclib/unit"
	"sourcegraph.com/sqs/pbtypes"
)

// NOTE: Only the importer part of the graph service is available over
// gRPC for now, because it is more necessary and has a simpler
// interface to convert to protobufs.

var Graph pb.MultiRepoImporterServer = &graph_{}

// The "_" differentiates it from package graph, which other files in
// this package import.
type graph_ struct{}

func (s *graph_) Import(ctx context.Context, op *pb.ImportOp) (*pbtypes.Void, error) {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "Graph.Import", op.Repo); err != nil {
		return nil, err
	}

	// Update global deps
	dstore := store.GlobalDepsFromContext(ctx)
	resolution := &unit.Resolution{
		Resolved: unit.Key{
			Repo:     op.Repo,
			CommitID: op.CommitID,
			Name:     op.Unit.Name,
			Type:     op.Unit.Type,
		},
		Raw: unit.Key{
			Repo:    unit.UnitRepoUnresolved,
			Version: op.Unit.Version,
			Name:    op.Unit.Name,
			Type:    op.Unit.Type,
		},
	}
	if err := dstore.Upsert(ctx, []*unit.Resolution{resolution}); err != nil {
		return nil, err
	}

	// Resolve ref dependency info
	resolveCache := make(map[unit.Key]*unit.Key)
	for _, ref := range op.Data.Refs {
		if ref.DefRepo == unit.UnitRepoUnresolved {
			raw := unit.Key{Type: ref.DefUnitType, Name: ref.DefUnit}
			if _, isResolved := resolveCache[raw]; !isResolved {
				resolved, err := dstore.Resolve(ctx, &raw)
				if err != nil {
					return nil, err
				}
				if len(resolved) >= 1 {
					r := resolved[0]
					resolveCache[raw] = &r
				} else {
					resolveCache[raw] = nil
				}
			}
			resolved := resolveCache[raw]
			if resolved != nil {
				ref.DefRepo, ref.DefUnit, ref.DefUnitType = resolved.Repo, resolved.Name, resolved.Type
			}
		}
	}

	gstore := store.GraphFromContext(ctx)
	if _, err := pb.Server(gstore).Import(ctx, op); err != nil {
		return nil, err
	}

	// If this build is for the head commit of the default branch of the repo,
	// update the global ref index.
	vcsrepo, err := store.RepoVCSFromContext(ctx).Open(ctx, op.Repo)
	if err != nil {
		return nil, err
	}
	commitID, err := vcsrepo.ResolveRevision("HEAD")
	if err != nil {
		return nil, err
	}
	if string(commitID) == op.CommitID {
		// Currently the global graph stores hold data for only the HEAD commit of the default
		// branch of the repo. We keep the commitID field empty to signify that
		// the data is always pointing to the HEAD commit of the default branch (which
		// is the default behavior on our app for empty repoRevSpecs).

		op.CommitID = ""
		if err := store.GlobalRefsFromContext(ctx).Update(ctx, op); err != nil {
			// Temporarily log and ignore error in updating the global ref store.
			// TODO: fail with error here once the rollout of global ref store is complete.
			log15.Error("error updating global ref store", "repo", op.Repo, "error", err)
		}
	}

	return &pbtypes.Void{}, nil
}

func (s *graph_) Index(ctx context.Context, op *pb.IndexOp) (*pbtypes.Void, error) {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "Graph.Index", op.Repo); err != nil {
		return nil, err
	}
	return pb.Server(store.GraphFromContext(ctx)).Index(ctx, op)
}

func (s *graph_) CreateVersion(ctx context.Context, op *pb.CreateVersionOp) (*pbtypes.Void, error) {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "Graph.CreateVersion", op.Repo); err != nil {
		return nil, err
	}
	return pb.Server(store.GraphFromContext(ctx)).CreateVersion(ctx, op)
}
