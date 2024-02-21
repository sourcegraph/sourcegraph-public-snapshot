package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
)

type hunkResolver struct {
	db   database.DB
	repo *RepositoryResolver
	hunk *gitdomain.Hunk
}

func (r *hunkResolver) Author() signatureResolver {
	return signatureResolver{
		person: &PersonResolver{
			db:              r.db,
			name:            r.hunk.Author.Name,
			email:           r.hunk.Author.Email,
			includeUserInfo: true,
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
	return NewGitCommitResolver(r.db, gitserver.NewClient("graphql.diff.hunk"), r.repo, r.hunk.CommitID, nil), nil
}

func (r *hunkResolver) Filename() string {
	return r.hunk.Filename
}
