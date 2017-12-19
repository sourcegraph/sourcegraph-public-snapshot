package graphqlbackend

import (
	"context"
	"fmt"
	"time"

	graphql "github.com/neelance/graphql-go"
	"github.com/neelance/graphql-go/relay"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
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

func (r *siteResolver) LatestSettings() *settingsResolver {
	return &settingsResolver{
		subject: &configurationSubject{site: r},
		settings: &sourcegraph.Settings{
			ID:        1,
			Contents:  "{}", // TODO(sqs): put actual config here
			CreatedAt: serverStart,
			Subject:   sourcegraph.ConfigurationSubject{Site: &r.id},
		},
	}
}
