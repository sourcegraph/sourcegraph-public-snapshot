package resolvers

import (
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

var _ graphqlbackend.CampaignSpecResolver = &campaignSpecResolver{}

type campaignSpecResolver struct {
}

func (r *campaignSpecResolver) ID() (graphql.ID, error) {
	return "", errors.New("TODO: not implemented")
}

func (r *campaignSpecResolver) OriginalInput() (string, error) {
	return "", errors.New("TODO: not implemented")
}

func (r *campaignSpecResolver) ParsedInput() (graphqlbackend.JSONValue, error) {
	return graphqlbackend.JSONValue{}, errors.New("TODO: not implemented")
}

func (r *campaignSpecResolver) ChangesetSpecs() ([]graphqlbackend.ChangesetSpecResolver, error) {
	return []graphqlbackend.ChangesetSpecResolver{}, errors.New("TODO: not implemented")
}

func (r *campaignSpecResolver) Creator() (*graphqlbackend.UserResolver, error) {
	return nil, errors.New("TODO: not implemented")
}

func (r *campaignSpecResolver) Namespace() (*graphqlbackend.NamespaceResolver, error) {
	return nil, errors.New("TODO: not implemented")
}

func (r *campaignSpecResolver) PreviewURL() (string, error) {
	return "", errors.New("TODO: not implemented")
}

func (r *campaignSpecResolver) CreatedAt() *graphqlbackend.DateTime {
	// TODO: not implemented
	return &graphqlbackend.DateTime{Time: time.Now()}
}

func (r *campaignSpecResolver) ExpiresAt() *graphqlbackend.DateTime {
	// TODO: not implemented
	return &graphqlbackend.DateTime{Time: time.Now()}
}
