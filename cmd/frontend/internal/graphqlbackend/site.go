package graphqlbackend

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	graphql "github.com/neelance/graphql-go"
	"github.com/neelance/graphql-go/relay"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
)

const singletonSiteID = "site"

var serverStart = time.Now()

func siteByID(ctx context.Context, id graphql.ID) (node, error) {
	siteID, err := unmarshalSiteID(id)
	if err != nil {
		return nil, err
	}
	if siteID != singletonSiteID {
		return nil, fmt.Errorf("site not found: %q", siteID)
	}
	return &siteResolver{id: siteID}, nil
}

func marshalSiteID(siteID string) graphql.ID { return relay.MarshalID("Site", siteID) }

func unmarshalSiteID(id graphql.ID) (siteID string, err error) {
	err = relay.UnmarshalSpec(id, &siteID)
	return
}

func (*schemaResolver) Site() *siteResolver {
	return &siteResolver{id: singletonSiteID}
}

type siteResolver struct {
	id string
}

var singletonSiteResolver = &siteResolver{id: singletonSiteID}

func (r *siteResolver) ID() graphql.ID { return marshalSiteID(r.id) }

func (r *siteResolver) Configuration(ctx context.Context) (string, error) {
	// 🚨 SECURITY: The site configuration contains secret tokens and credentials,
	// so only admins may view it.
	if !actor.FromContext(ctx).IsAdmin() {
		return "", errors.New("must be admin to view site configuration")
	}

	siteConfigJSON, err := json.MarshalIndent(conf.Get(), "", "  ")
	if err != nil {
		return "", err
	}

	return string(siteConfigJSON), nil
}

func (r *siteResolver) LatestSettings() (*settingsResolver, error) {
	// The site configuration (which is only visible to admins) contains a field "settings"
	// that is visible to all users. So, this does not need a permissions check.
	siteConfigJSON, err := json.MarshalIndent(conf.Get().Settings, "", "  ")
	if err != nil {
		return nil, err
	}

	return &settingsResolver{
		subject: &configurationSubject{site: r},
		settings: &sourcegraph.Settings{
			ID:        1,
			Contents:  string(siteConfigJSON),
			CreatedAt: serverStart,
			Subject:   sourcegraph.ConfigurationSubject{Site: &r.id},
		},
	}, nil
}
