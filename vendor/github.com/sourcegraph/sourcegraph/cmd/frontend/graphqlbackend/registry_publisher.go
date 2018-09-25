package graphqlbackend

import (
	"context"
	"errors"
	"fmt"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
)

var errRegistryUnknownPublisher = errors.New("unknown registry extension publisher")

func (r *extensionRegistryResolver) ViewerPublishers(ctx context.Context) ([]*registryPublisher, error) {
	var publishers []*registryPublisher
	user, err := currentUser(ctx)
	if err != nil || user == nil {
		return nil, err
	}
	publishers = append(publishers, &registryPublisher{user: user})

	orgs, err := db.Orgs.GetByUserID(ctx, user.user.ID)
	if err != nil {
		return nil, err
	}
	for _, org := range orgs {
		publishers = append(publishers, &registryPublisher{org: &orgResolver{org: org}})
	}
	return publishers, nil
}

type registryPublisher struct {
	user *UserResolver
	org  *orgResolver
}

func (r *registryPublisher) ToUser() (*UserResolver, bool) { return r.user, r.user != nil }
func (r *registryPublisher) ToOrg() (*orgResolver, bool)   { return r.org, r.org != nil }

func (r *registryPublisher) toDBRegistryPublisher() db.RegistryPublisher {
	switch {
	case r.user != nil:
		return db.RegistryPublisher{UserID: r.user.user.ID, NonCanonicalName: r.user.user.Username}
	case r.org != nil:
		return db.RegistryPublisher{OrgID: r.org.org.ID, NonCanonicalName: r.org.org.Name}
	default:
		return db.RegistryPublisher{}
	}
}

func getRegistryPublisher(ctx context.Context, publisher db.RegistryPublisher) (*registryPublisher, error) {
	switch {
	case publisher.UserID != 0:
		user, err := UserByIDInt32(ctx, publisher.UserID)
		if err != nil {
			return nil, err
		}
		return &registryPublisher{user: user}, nil
	case publisher.OrgID != 0:
		org, err := orgByIDInt32(ctx, publisher.OrgID)
		if err != nil {
			return nil, err
		}
		return &registryPublisher{org: org}, nil
	default:
		return nil, errRegistryUnknownPublisher
	}
}

type registryPublisherID struct {
	userID, orgID int32
}

func toRegistryPublisherID(extension *db.RegistryExtension) *registryPublisherID {
	return &registryPublisherID{
		userID: extension.Publisher.UserID,
		orgID:  extension.Publisher.OrgID,
	}
}

// unmarshalRegistryPublisherID unmarshals the GraphQL ID into the possible publisher ID
// types.
//
// 🚨 SECURITY
func unmarshalRegistryPublisherID(id graphql.ID) (*registryPublisherID, error) {
	var (
		p   registryPublisherID
		err error
	)
	switch kind := relay.UnmarshalKind(id); kind {
	case "User":
		p.userID, err = unmarshalUserID(id)
	case "Org":
		p.orgID, err = unmarshalOrgID(id)
	default:
		return nil, fmt.Errorf("unknown registry extension publisher type: %q", kind)
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// viewerCanAdminister returns whether the current user is allowed to perform mutations on a
// registry extension with the given publisher.
//
// 🚨 SECURITY
func (p *registryPublisherID) viewerCanAdminister(ctx context.Context) error {
	switch {
	case p.userID != 0:
		// 🚨 SECURITY: Check that the current user is either the publisher or a site admin.
		return backend.CheckSiteAdminOrSameUser(ctx, p.userID)
	case p.orgID != 0:
		// 🚨 SECURITY: Check that the current user is a member of the publisher org.
		return backend.CheckOrgAccess(ctx, p.orgID)
	default:
		return errRegistryUnknownPublisher
	}
}
