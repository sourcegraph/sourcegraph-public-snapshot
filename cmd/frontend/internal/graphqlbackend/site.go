package graphqlbackend

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	graphql "github.com/neelance/graphql-go"
	"github.com/neelance/graphql-go/relay"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/processrestart"
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

func (r *siteResolver) Configuration(ctx context.Context) (*siteConfigurationResolver, error) {
	// 🚨 SECURITY: The site configuration contains secret tokens and credentials,
	// so only admins may view it.
	if err := checkCanViewOrUpdateSiteConfiguration(ctx); err != nil {
		return nil, err
	}
	return &siteConfigurationResolver{}, nil
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

func (r *siteResolver) CanReloadSite(ctx context.Context) bool {
	return canReloadSite && actor.FromContext(ctx).IsAdmin()
}

type siteConfigurationResolver struct{}

func checkCanViewOrUpdateSiteConfiguration(ctx context.Context) error {
	// checkIsAdmin returns an error if the actor is not an admin. The site configuration
	// contains secret tokens and credentials, so only admins may view/update it.
	//
	// 🚨 SECURITY: checkIsAdmin MUST be called anytime a siteConfigurationResolver struct
	// value is created. To be extra safe, other *siteConfigurationResolver methods that
	// edit/return the config should also call checkIsAdmin (in case new code is committed
	// that, e.g., accidentally constructs a *siteConfigurationResolver without performing
	// the is-admin check).
	if !actor.FromContext(ctx).IsAdmin() {
		return errors.New("must be admin to view/update site configuration")
	}
	return nil
}

func (r *siteConfigurationResolver) EffectiveContents(ctx context.Context) (string, error) {
	// 🚨 SECURITY: The site configuration contains secret tokens and credentials,
	// so only admins may view it.
	if err := checkCanViewOrUpdateSiteConfiguration(ctx); err != nil {
		return "", err
	}
	return conf.Raw(), nil
}

func (r *siteConfigurationResolver) PendingContents(ctx context.Context) (*string, error) {
	// 🚨 SECURITY: The site configuration contains secret tokens and credentials,
	// so only admins may view it.
	if err := checkCanViewOrUpdateSiteConfiguration(ctx); err != nil {
		return nil, err
	}

	if !conf.IsDirty() {
		return nil, nil
	}

	rawContents, err := ioutil.ReadFile(conf.FilePath())
	if err != nil {
		if os.IsNotExist(err) {
			s := "// The site configuration file does not exist."
			return &s, nil
		}
		return nil, err
	}

	s := string(rawContents)
	return &s, nil
}

func (r *siteConfigurationResolver) CanUpdate() bool {
	// We assume the is-admin check has already been performed before constructing
	// our receiver, so we just need to check if the file itself is writable, not
	// the viewer's authorization.
	//
	// Also, we disallow updating if the site can't be auto-restarted.
	return conf.IsWritable() && processrestart.CanRestart()
}

func (r *siteConfigurationResolver) Source() string {
	if conf.FilePath() != "" {
		s := conf.FilePath()
		if !conf.IsWritable() {
			s += " (read-only)"
		}
		return s
	}
	return "SOURCEGRAPH_CONFIG (environment variable)"
}

func (r *schemaResolver) UpdateSiteConfiguration(ctx context.Context, args *struct {
	Input string
}) (*EmptyResponse, error) {
	// 🚨 SECURITY: The site configuration contains secret tokens and credentials,
	// so only admins may view it.
	if err := checkCanViewOrUpdateSiteConfiguration(ctx); err != nil {
		return nil, err
	}

	if _, err := conf.Write(args.Input); err != nil {
		return nil, err
	}
	return &EmptyResponse{}, nil
}
