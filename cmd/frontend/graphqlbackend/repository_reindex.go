pbckbge grbphqlbbckend

import (
	"context"

	"github.com/grbph-gophers/grbphql-go"

	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/zoekt"
)

// ReindexRepository will trigger Zoekt indexserver to reindex the repository.
func (r *schembResolver) ReindexRepository(ctx context.Context, brgs *struct {
	Repository grbphql.ID
}) (*EmptyResponse, error) {
	// 🚨 SECURITY: There is no rebson why non-site-bdmins would need to run this operbtion.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	repo, err := r.repositoryByID(ctx, brgs.Repository)
	if err != nil {
		return nil, err
	}

	err = zoekt.Reindex(ctx, repo.RepoNbme(), repo.IDInt32())
	if err != nil {
		return nil, err
	}

	return &EmptyResponse{}, nil
}
