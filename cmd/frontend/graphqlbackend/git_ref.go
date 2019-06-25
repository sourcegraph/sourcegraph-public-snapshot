package graphqlbackend

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
)

const (
	gitRefTypeBranch = "GIT_BRANCH"
	gitRefTypeTag    = "GIT_TAG"
	gitRefTypeOther  = "GIT_REF_OTHER"

	gitRefOrderAuthoredOrCommittedAt = "AUTHORED_OR_COMMITTED_AT"
)

func gitRefPrefix(ref string) string {
	if strings.HasPrefix(ref, "refs/heads/") {
		return "refs/heads/"
	}
	if strings.HasPrefix(ref, "refs/tags/") {
		return "refs/tags/"
	}
	if strings.HasPrefix(ref, "refs/pull/") {
		return "refs/pull/"
	}
	if strings.HasPrefix(ref, "refs/") {
		return "refs/"
	}
	return ""
}

func gitRefType(ref string) string {
	if strings.HasPrefix(ref, "refs/heads/") {
		return gitRefTypeBranch
	}
	if strings.HasPrefix(ref, "refs/tags/") {
		return gitRefTypeTag
	}
	return gitRefTypeOther
}

func gitRefDisplayName(ref string) string {
	prefix := gitRefPrefix(ref)

	if prefix == "refs/pull/" && (strings.HasSuffix(ref, "/head") || strings.HasSuffix(ref, "/merge")) {
		// Special-case GitHub pull requests for a nicer display name.
		numberStr := ref[len(prefix) : len(prefix)+strings.Index(ref[len(prefix):], "/")]
		number, err := strconv.Atoi(numberStr)
		if err == nil {
			return fmt.Sprintf("#%d", number)
		}
	}

	return strings.TrimPrefix(ref, prefix)
}

func gitRefByID(ctx context.Context, id graphql.ID) (*GitRefResolver, error) {
	repoID, rev, err := unmarshalGitRefID(id)
	if err != nil {
		return nil, err
	}
	repo, err := repositoryByID(ctx, repoID)
	if err != nil {
		return nil, err
	}
	return &GitRefResolver{
		repo: repo,
		name: rev,
	}, nil
}

func NewGitRefResolver(repo *repositoryResolver, name string, target gitObjectID) *GitRefResolver {
	return &GitRefResolver{repo: repo, name: name, target: target}
}

type GitRefResolver struct {
	repo *repositoryResolver
	name string

	target gitObjectID // the target's OID, if known (otherwise computed on demand)
}

// gitRefGQLID is a type used for marshaling and unmarshaling a Git ref's
// GraphQL ID.
type gitRefGQLID struct {
	Repository graphql.ID `json:"r"`
	Rev        string     `json:"v"`
}

func marshalGitRefID(repo graphql.ID, rev string) graphql.ID {
	return relay.MarshalID("GitRef", gitRefGQLID{Repository: repo, Rev: rev})
}

func unmarshalGitRefID(id graphql.ID) (repoID graphql.ID, rev string, err error) {
	var spec gitRefGQLID
	err = relay.UnmarshalSpec(id, &spec)
	return spec.Repository, spec.Rev, err
}

func (r *GitRefResolver) ID() graphql.ID      { return marshalGitRefID(r.repo.ID(), r.name) }
func (r *GitRefResolver) Name() string        { return r.name }
func (r *GitRefResolver) AbbrevName() string  { return strings.TrimPrefix(r.name, gitRefPrefix(r.name)) }
func (r *GitRefResolver) DisplayName() string { return gitRefDisplayName(r.name) }
func (r *GitRefResolver) Prefix() string      { return gitRefPrefix(r.name) }
func (r *GitRefResolver) Type() string        { return gitRefType(r.name) }
func (r *GitRefResolver) Target() interface {
	OID(context.Context) (gitObjectID, error)
	//lint:ignore U1000 is used by graphql via reflection
	AbbreviatedOID(context.Context) (string, error)
	//lint:ignore U1000 is used by graphql via reflection
	Commit(context.Context) (*gitCommitResolver, error)
	//lint:ignore U1000 is used by graphql via reflection
	Type(context.Context) (gitObjectType, error)
} {
	if r.target != "" {
		return &gitObject{repo: r.repo, oid: r.target, typ: gitObjectTypeCommit}
	}
	return &gitObjectResolver{repo: r.repo, revspec: r.name}
}
func (r *GitRefResolver) Repository() *repositoryResolver { return r.repo }

func (r *GitRefResolver) URL() string { return r.repo.URL() + "@" + escapeRevspecForURL(r.AbbrevName()) }
