package graphqlbackend

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	graphql "github.com/neelance/graphql-go"
	"github.com/neelance/graphql-go/relay"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
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
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
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
	err := backend.CheckCurrentUserIsSiteAdmin(ctx)
	return canReloadSite && err == nil
}

type siteConfigurationResolver struct{}

func (r *siteConfigurationResolver) EffectiveContents(ctx context.Context) (string, error) {
	// 🚨 SECURITY: The site configuration contains secret tokens and credentials,
	// so only admins may view it.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return "", err
	}
	return conf.Raw(), nil
}

func (r *siteConfigurationResolver) PendingContents(ctx context.Context) (*string, error) {
	// 🚨 SECURITY: The site configuration contains secret tokens and credentials,
	// so only admins may view it.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
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

// pendingOrEffectiveContents returns pendingContents if it exists, or else effectiveContents.
func (r *siteConfigurationResolver) pendingOrEffectiveContents(ctx context.Context) (string, error) {
	// 🚨 SECURITY: Site admin status is checked in both r.PendingContents and r.EffectiveContents,
	// so we don't need to check it in this method.
	pendingContents, err := r.PendingContents(ctx)
	if err != nil {
		return "", err
	}
	if pendingContents != nil {
		return *pendingContents, nil
	}
	return r.EffectiveContents(ctx)
}

func (r *siteConfigurationResolver) ExtraValidationErrors(ctx context.Context) ([]string, error) {
	contents, err := r.pendingOrEffectiveContents(ctx)
	if err != nil {
		return nil, err
	}
	return conf.ValidateCustom(normalizeJSON(contents))
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
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	if _, err := conf.Write(args.Input); err != nil {
		return nil, err
	}
	return &EmptyResponse{}, nil
}
