package local

import (
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/server/accesscontrol"
	"sourcegraph.com/sourcegraph/sourcegraph/store"
	"sourcegraph.com/sourcegraph/srclib/store/pb"
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
	if _, err := pb.Server(store.GraphFromContext(ctx)).Import(ctx, op); err != nil {
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
		// Currently the xref store holds data for only the HEAD commit of the default
		// branch of the repo. We keep the commitID field empty to signify that
		// the refs are always pointing to the HEAD commit of the default branch (which
		// is the default behavior on our app for empty repoRevSpecs).
		op.CommitID = ""
		if err := store.GlobalRefsFromContext(ctx).Update(ctx, op); err != nil {
			return nil, err
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
