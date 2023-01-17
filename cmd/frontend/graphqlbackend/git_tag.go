package graphqlbackend

import (
	"context"
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
)

type GitTagResolver struct {
	logger          log.Logger
	db              database.DB
	gitserverClient gitserver.Client
	repoResolver    *RepositoryResolver

	// oid MUST be specified and a 40-character Git SHA.
	oid GitObjectID

	gitRepo api.RepoName

	// tag should not be accessed directly since it might not be initialized.
	// Use the resolver methods instead.
	tag     *gitdomain.Tag
	tagOnce sync.Once
	tagErr  error
}

func NewGitTagResolver(db database.DB, gsClient gitserver.Client, repo *RepositoryResolver, id api.TagID) *GitTagResolver {
	return &GitTagResolver{
		logger: log.Scoped("gitTagResolver", "resolve a specific tag").
			With(log.String("repo", string(repo.RepoName())),
				log.String("tagID", string(id))),
		db:              db,
		gitserverClient: gsClient,
		repoResolver:    repo,
		gitRepo:         repo.RepoName(),
		oid:             GitObjectID(id),
	}
}

func (r *GitTagResolver) resolveTag(ctx context.Context) (*gitdomain.Tag, error) {
	r.tagOnce.Do(func() {
		var tags []*gitdomain.Tag
		tags, r.tagErr = r.gitserverClient.ListTags(ctx, r.gitRepo, string(r.oid))
		if len(tags) > 1 {
			panic("impossible state: a tags oid can't point to more than one tag")
		}

		if len(tags) == 1 {
			r.tag = tags[0]
		}
	})
	return r.tag, r.tagErr
}

// gitTagGQLID is a type used for marshaling and unmarshalling a Git tag's
// GraphQL ID.
type gitTagGQLID struct {
	Repository graphql.ID  `json:"r"`
	TagID      GitObjectID `json:"c"`
}

func marshalGitTagID(repo graphql.ID, tagID GitObjectID) graphql.ID {
	return relay.MarshalID("GitTag", gitTagGQLID{Repository: repo, TagID: tagID})
}

func unmarshalGitTagID(id graphql.ID) (repoID graphql.ID, tagID GitObjectID, err error) {
	var spec gitTagGQLID
	err = relay.UnmarshalSpec(id, &spec)
	return spec.Repository, spec.TagID, err
}

func (r *GitTagResolver) ID() graphql.ID {
	return marshalGitCommitID(r.repoResolver.ID(), r.oid)
}

func (r *GitTagResolver) Repository() *RepositoryResolver { return r.repoResolver }

func (r *GitTagResolver) OID() GitObjectID { return r.oid }

func (r *GitTagResolver) AbbreviatedOID() string {
	return string(r.oid)[:7]
}

// TODO: nsc - all this lol
func (r *GitTagResolver) Commit() *GitCommitResolver {
	return nil
}

func (r *GitTagResolver) Tagger(ctx context.Context) (*signatureResolver, error) {
	tag, err := r.resolveTag(ctx)
	if err != nil {
		return nil, err
	}
	return toSignatureResolver(r.db, tag.Tagger, true), nil
}

func (r *GitTagResolver) Message(ctx context.Context) (string, error) {
	return "", nil
}

func (r *GitTagResolver) Subject(ctx context.Context) (string, error) {
	return "", nil
}

func (r *GitTagResolver) Body(ctx context.Context) (*string, error) {
	return nil, nil
}
