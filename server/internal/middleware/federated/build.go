package federated

import (
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/svc"
)

// The Builds.List function has a custom federated implementation because it is currently
// not generated automatically. This happens because sourcegraph.BuildListOptions does not
// use a RepoSpec type field to store the repo path, but instead stores it as a string.
// Due to this, gen.RepoURIExpr does not identify Builds.List as necessary for federation.
// A temporary fix is to have this custom federated implementation. The correct way to fix
// this is to change sourcegraph.BuildListOptions (and all dependent code) to use a RepoSpec field.
//
// TODO(pararth): Fix this. This will be tackled along with the overall repo URI renaming task,
// which will modify the RepoSpec type.
func CustomBuildsList(ctx context.Context, v1 *sourcegraph.BuildListOptions, s sourcegraph.BuildsServer) (*sourcegraph.BuildList, error) {
	if v1.Repo == "" {
		// if the request is not for a particular repo but for aggregate
		// builds, then route it to the local builds store.
		return s.List(ctx, v1)
	}
	ctx2, err := RepoContext(ctx, &v1.Repo)
	if err != nil {
		return nil, err
	}
	if ctx2 == nil {
		return s.List(ctx, v1)
	}
	ctx = ctx2
	return svc.Builds(ctx).List(ctx, v1)
}
