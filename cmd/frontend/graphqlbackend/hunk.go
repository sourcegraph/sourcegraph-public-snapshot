package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

type hunkResolver struct {
	db   dbutil.DB
	repo *RepositoryResolver
	hunk *git.Hunk
}

func (r *hunkResolver) Author() signatureResolver {
	return signatureResolver{
		person: &PersonResolver{
			db:    r.db,
			name:  r.hunk.Author.Name,
			email: r.hunk.Author.Email,
		},
		date: r.hunk.Author.Date,
	}
}

func (r *hunkResolver) StartLine() int32 {
	return int32(r.hunk.StartLine)
}

func (r *hunkResolver) EndLine() int32 {
	return int32(r.hunk.EndLine)
}

func (r *hunkResolver) StartByte() int32 {
	return int32(r.hunk.EndLine)
}

func (r *hunkResolver) EndByte() int32 {
	return int32(r.hunk.EndByte)
}

func (r *hunkResolver) Rev() string {
	return string(r.hunk.CommitID)
}

func (r *hunkResolver) Message() string {
	return r.hunk.Message
}

func (r *hunkResolver) Commit(ctx context.Context) (*GitCommitResolver, error) {
	return toGitCommitResolver(r.repo, r.db, r.hunk.CommitID, nil), nil
}
