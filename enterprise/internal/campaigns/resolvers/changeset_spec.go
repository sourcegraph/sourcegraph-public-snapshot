package resolvers

import (
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	ee "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

func marshalChangesetSpecRandID(id string) graphql.ID {
	return relay.MarshalID("ChangesetSpec", id)
}

func unmarshalChangesetSpecID(id graphql.ID) (changesetSpecRandID string, err error) {
	err = relay.UnmarshalSpec(id, &changesetSpecRandID)
	return
}

var _ graphqlbackend.ChangesetSpecResolver = &changesetSpecResolver{}

type changesetSpecResolver struct {
	store       *ee.Store
	httpFactory *httpcli.Factory

	changesetSpec *campaigns.ChangesetSpec
}

func (r *changesetSpecResolver) ID() graphql.ID {
	// 🚨 SECURITY: This needs to be the RandID! We can't expose the
	// sequential, guessable ID.
	return marshalChangesetSpecRandID(r.changesetSpec.RandID)
}

func (r *changesetSpecResolver) Description() graphqlbackend.ChangesetDescription {
	return &changesetDescriptionResolver{desc: &r.changesetSpec.Spec}
}

func (r *changesetSpecResolver) ExpiresAt() *graphqlbackend.DateTime {
	// TODO: This is a bogus value and needs to be implemented properly
	expiresAt := r.changesetSpec.CreatedAt.Add(2 * time.Hour)
	return &graphqlbackend.DateTime{Time: expiresAt}
}

var _ graphqlbackend.ChangesetDescription = &changesetDescriptionResolver{}

// changesetDescriptionResolver implements both ChangesetDescription
// interfaces: ExistingChangesetReferenceResolver and
// GitBranchChangesetDescriptionResolver.
type changesetDescriptionResolver struct {
	desc *campaigns.ChangesetSpecDescription
}

func (r *changesetDescriptionResolver) isExistingChangesetRef() bool {
	return r.desc.ExternalID != ""
}

func (r *changesetDescriptionResolver) ToExistingChangesetReference() (graphqlbackend.ExistingChangesetReferenceResolver, bool) {
	if r.isExistingChangesetRef() {
		return r, true
	}
	return nil, false

}
func (r *changesetDescriptionResolver) ToGitBranchChangesetDescription() (graphqlbackend.GitBranchChangesetDescriptionResolver, bool) {
	if r.isExistingChangesetRef() {
		return nil, false
	}
	return r, true
}

func (r *changesetDescriptionResolver) BaseRepository() graphql.ID { return r.desc.BaseRepository }
func (r *changesetDescriptionResolver) ExternalID() string         { return r.desc.ExternalID }
func (r *changesetDescriptionResolver) BaseRef() string            { return r.desc.BaseRef }
func (r *changesetDescriptionResolver) BaseRev() string            { return r.desc.BaseRev }
func (r *changesetDescriptionResolver) HeadRepository() graphql.ID { return r.desc.HeadRepository }
func (r *changesetDescriptionResolver) HeadRef() string            { return r.desc.HeadRef }
func (r *changesetDescriptionResolver) Title() string              { return r.desc.Title }
func (r *changesetDescriptionResolver) Body() string               { return r.desc.Body }
func (r *changesetDescriptionResolver) Published() bool            { return r.desc.Published }

func (r *changesetDescriptionResolver) Commits() []graphqlbackend.GitCommitDescriptionResolver {
	var resolvers []graphqlbackend.GitCommitDescriptionResolver
	for _, c := range r.desc.Commits {
		resolvers = append(resolvers, &gitCommitDescriptionResolver{
			message: c.Message,
			diff:    c.Diff,
		})
	}
	return resolvers
}

var _ graphqlbackend.GitCommitDescriptionResolver = &gitCommitDescriptionResolver{}

type gitCommitDescriptionResolver struct {
	message string
	diff    string
}

func (r *gitCommitDescriptionResolver) Message() string { return r.message }
func (r *gitCommitDescriptionResolver) Diff() string    { return r.diff }
