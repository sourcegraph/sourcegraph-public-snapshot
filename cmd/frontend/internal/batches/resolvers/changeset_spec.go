package resolvers

import (
	"context"
	"strings"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/go-diff/diff"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const changesetSpecIDKind = "ChangesetSpec"

func marshalChangesetSpecRandID(id string) graphql.ID {
	return relay.MarshalID(changesetSpecIDKind, id)
}

func unmarshalChangesetSpecID(id graphql.ID) (changesetSpecRandID string, err error) {
	err = relay.UnmarshalSpec(id, &changesetSpecRandID)
	return
}

var _ graphqlbackend.ChangesetSpecResolver = &changesetSpecResolver{}

type changesetSpecResolver struct {
	store *store.Store

	changesetSpec *btypes.ChangesetSpec

	repo *types.Repo
}

func NewChangesetSpecResolver(ctx context.Context, store *store.Store, changesetSpec *btypes.ChangesetSpec) (*changesetSpecResolver, error) {
	resolver := &changesetSpecResolver{
		store:         store,
		changesetSpec: changesetSpec,
	}

	// 🚨 SECURITY: database.Repos.GetByIDs uses the authzFilter under the hood and
	// filters out repositories that the user doesn't have access to.
	// In case we don't find a repository, it might be because it's deleted
	// or because the user doesn't have access.
	rs, err := store.Repos().GetByIDs(ctx, changesetSpec.BaseRepoID)
	if err != nil {
		return nil, err
	}

	// Not found is ok, the resolver will disguise as a HiddenChangesetResolver.
	if len(rs) == 1 {
		resolver.repo = rs[0]
	}

	return resolver, nil
}

func NewChangesetSpecResolverWithRepo(store *store.Store, repo *types.Repo, changesetSpec *btypes.ChangesetSpec) *changesetSpecResolver {
	return &changesetSpecResolver{
		store:         store,
		repo:          repo,
		changesetSpec: changesetSpec,
	}
}

func (r *changesetSpecResolver) ID() graphql.ID {
	// 🚨 SECURITY: This needs to be the RandID! We can't expose the
	// sequential, guessable ID.
	return marshalChangesetSpecRandID(r.changesetSpec.RandID)
}

func (r *changesetSpecResolver) Type() string {
	return strings.ToUpper(string(r.changesetSpec.Type))
}

func (r *changesetSpecResolver) Description(ctx context.Context) (graphqlbackend.ChangesetDescription, error) {
	db := r.store.DatabaseDB()
	descriptionResolver := &changesetDescriptionResolver{
		store: r.store,
		spec:  r.changesetSpec,
		// Note: r.repo can never be nil, because Description is a VisibleChangesetSpecResolver-only field.
		repoResolver: graphqlbackend.NewRepositoryResolver(db, gitserver.NewClient("graphql.batches.changesetspecrepo"), r.repo),
		diffStat:     r.changesetSpec.DiffStat(),
	}

	return descriptionResolver, nil
}

func (r *changesetSpecResolver) ExpiresAt() *gqlutil.DateTime {
	return &gqlutil.DateTime{Time: r.changesetSpec.ExpiresAt()}
}

func (r *changesetSpecResolver) ForkTarget() graphqlbackend.ForkTargetInterface {
	return &forkTargetResolver{changesetSpec: r.changesetSpec}
}

func (r *changesetSpecResolver) repoAccessible() bool {
	// If the repository is not nil, it's accessible
	return r.repo != nil
}

func (r *changesetSpecResolver) Workspace(ctx context.Context) (graphqlbackend.BatchSpecWorkspaceResolver, error) {
	// TODO(ssbc): not implemented
	return nil, errors.New("not implemented")
}

func (r *changesetSpecResolver) ToHiddenChangesetSpec() (graphqlbackend.HiddenChangesetSpecResolver, bool) {
	if r.repoAccessible() {
		return nil, false
	}

	return r, true
}

func (r *changesetSpecResolver) ToVisibleChangesetSpec() (graphqlbackend.VisibleChangesetSpecResolver, bool) {
	if !r.repoAccessible() {
		return nil, false
	}

	return r, true
}

var _ graphqlbackend.ChangesetDescription = &changesetDescriptionResolver{}

// changesetDescriptionResolver implements both ChangesetDescription
// interfaces: ExistingChangesetReferenceResolver and
// GitBranchChangesetDescriptionResolver.
type changesetDescriptionResolver struct {
	store        *store.Store
	repoResolver *graphqlbackend.RepositoryResolver
	spec         *btypes.ChangesetSpec
	diffStat     diff.Stat
}

func (r *changesetDescriptionResolver) ToExistingChangesetReference() (graphqlbackend.ExistingChangesetReferenceResolver, bool) {
	if r.spec.Type == btypes.ChangesetSpecTypeExisting {
		return r, true
	}
	return nil, false
}
func (r *changesetDescriptionResolver) ToGitBranchChangesetDescription() (graphqlbackend.GitBranchChangesetDescriptionResolver, bool) {
	if r.spec.Type == btypes.ChangesetSpecTypeBranch {
		return r, true
	}
	return nil, false
}

func (r *changesetDescriptionResolver) BaseRepository() *graphqlbackend.RepositoryResolver {
	return r.repoResolver
}
func (r *changesetDescriptionResolver) ExternalID() string { return r.spec.ExternalID }
func (r *changesetDescriptionResolver) BaseRef() string {
	return gitdomain.AbbreviateRef(r.spec.BaseRef)
}
func (r *changesetDescriptionResolver) BaseRev() string { return r.spec.BaseRev }
func (r *changesetDescriptionResolver) HeadRef() string {
	return gitdomain.AbbreviateRef(r.spec.HeadRef)
}
func (r *changesetDescriptionResolver) Title() string { return r.spec.Title }
func (r *changesetDescriptionResolver) Body() string  { return r.spec.Body }
func (r *changesetDescriptionResolver) Published() *batcheslib.PublishedValue {
	if published := r.spec.Published; !published.Nil() {
		return &published
	}
	return nil
}

func (r *changesetDescriptionResolver) DiffStat() *graphqlbackend.DiffStat {
	return graphqlbackend.NewDiffStat(r.diffStat)
}

func (r *changesetDescriptionResolver) Diff(ctx context.Context) (graphqlbackend.PreviewRepositoryComparisonResolver, error) {
	return graphqlbackend.NewPreviewRepositoryComparisonResolver(ctx, r.store.DatabaseDB(), gitserver.NewClient("graphql.batches.changesetdescriptiondiff"), r.repoResolver, r.spec.BaseRev, r.spec.Diff)
}

func (r *changesetDescriptionResolver) Commits() []graphqlbackend.GitCommitDescriptionResolver {
	return []graphqlbackend.GitCommitDescriptionResolver{&gitCommitDescriptionResolver{
		store:       r.store,
		message:     r.spec.CommitMessage,
		diff:        r.spec.Diff,
		authorName:  r.spec.CommitAuthorName,
		authorEmail: r.spec.CommitAuthorEmail,
	}}
}

var _ graphqlbackend.GitCommitDescriptionResolver = &gitCommitDescriptionResolver{}

type gitCommitDescriptionResolver struct {
	store       *store.Store
	message     string
	diff        []byte
	authorName  string
	authorEmail string
}

func (r *gitCommitDescriptionResolver) Author() *graphqlbackend.PersonResolver {
	return graphqlbackend.NewPersonResolver(
		r.store.DatabaseDB(),
		r.authorName,
		r.authorEmail,
		// Try to find the corresponding Sourcegraph user.
		true,
	)
}
func (r *gitCommitDescriptionResolver) Message() string { return r.message }
func (r *gitCommitDescriptionResolver) Subject() string {
	return gitdomain.Message(r.message).Subject()
}
func (r *gitCommitDescriptionResolver) Body() *string {
	body := gitdomain.Message(r.message).Body()
	if body == "" {
		return nil
	}
	return &body
}
func (r *gitCommitDescriptionResolver) Diff() string { return string(r.diff) }

type forkTargetResolver struct {
	changesetSpec *btypes.ChangesetSpec
}

var _ graphqlbackend.ForkTargetInterface = &forkTargetResolver{}

func (r *forkTargetResolver) PushUser() bool {
	return r.changesetSpec.IsFork()
}

func (r *forkTargetResolver) Namespace() *string {
	// We don't use `changesetSpec.GetForkNamespace()` here because it returns `nil` if
	// the namespace matches the user default namespace. This is a perfectly reasonable
	// thing to do for the way we use the method internally, but for resolving this field
	// on the GraphQL scehma, we want to return the namespace regardless of what it is.
	return r.changesetSpec.ForkNamespace
}
