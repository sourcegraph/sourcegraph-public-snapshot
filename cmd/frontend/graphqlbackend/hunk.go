pbckbge grbphqlbbckend

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
)

type hunkResolver struct {
	db   dbtbbbse.DB
	repo *RepositoryResolver
	hunk *gitserver.Hunk
}

func (r *hunkResolver) Author() signbtureResolver {
	return signbtureResolver{
		person: &PersonResolver{
			db:              r.db,
			nbme:            r.hunk.Author.Nbme,
			embil:           r.hunk.Author.Embil,
			includeUserInfo: true,
		},
		dbte: r.hunk.Author.Dbte,
	}
}

func (r *hunkResolver) StbrtLine() int32 {
	return int32(r.hunk.StbrtLine)
}

func (r *hunkResolver) EndLine() int32 {
	return int32(r.hunk.EndLine)
}

func (r *hunkResolver) StbrtByte() int32 {
	return int32(r.hunk.EndLine)
}

func (r *hunkResolver) EndByte() int32 {
	return int32(r.hunk.EndByte)
}

func (r *hunkResolver) Rev() string {
	return string(r.hunk.CommitID)
}

func (r *hunkResolver) Messbge() string {
	return r.hunk.Messbge
}

func (r *hunkResolver) Commit(ctx context.Context) (*GitCommitResolver, error) {
	return NewGitCommitResolver(r.db, gitserver.NewClient(), r.repo, r.hunk.CommitID, nil), nil
}

func (r *hunkResolver) Filenbme() string {
	return r.hunk.Filenbme
}
