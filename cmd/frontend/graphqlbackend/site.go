package graphqlbackend

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/siteid"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/useractivity"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/env"
	"github.com/sourcegraph/sourcegraph/pkg/processrestart"
	"github.com/sourcegraph/sourcegraph/pkg/version"
)

const singletonSiteGQLID = "site"

func siteByGQLID(ctx context.Context, id graphql.ID) (node, error) {
	siteGQLID, err := unmarshalSiteGQLID(id)
	if err != nil {
		return nil, err
	}
	if siteGQLID != singletonSiteGQLID {
		return nil, fmt.Errorf("site not found: %q", siteGQLID)
	}
	return &siteResolver{gqlID: siteGQLID}, nil
}

func marshalSiteGQLID(siteID string) graphql.ID { return relay.MarshalID("Site", siteID) }

// SiteGQLID is the GraphQL ID of the Sourcegraph site. It is a constant across all Sourcegraph
// instances.
func SiteGQLID() graphql.ID { return singletonSiteResolver.ID() }

func unmarshalSiteGQLID(id graphql.ID) (siteID string, err error) {
	err = relay.UnmarshalSpec(id, &siteID)
	return
}

func (*schemaResolver) Site() *siteResolver {
	return &siteResolver{gqlID: singletonSiteGQLID}
}

type siteResolver struct {
	gqlID string // == singletonSiteGQLID, not the site ID
}

var singletonSiteResolver = &siteResolver{gqlID: singletonSiteGQLID}

func (r *siteResolver) ID() graphql.ID { return marshalSiteGQLID(r.gqlID) }

func (r *siteResolver) SiteID() string { return siteid.Get() }

func (r *siteResolver) Configuration(ctx context.Context) (*siteConfigurationResolver, error) {
	// 🚨 SECURITY: The site configuration contains secret tokens and credentials,
	// so only admins may view it.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}
	return &siteConfigurationResolver{}, nil
}

func (r *siteResolver) ViewerCanAdminister(ctx context.Context) (bool, error) {
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err == backend.ErrMustBeSiteAdmin || err == backend.ErrNotAuthenticated {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (r *siteResolver) configurationSubject() api.ConfigurationSubject {
	return api.ConfigurationSubject{Site: true}
}

func (r *siteResolver) DeprecatedSiteConfigurationSettings() (*string, error) {
	// The site configuration (which is only visible to admins) contains a field "settings"
	// that is visible to all users. So, this does not need a permissions check.
	settings := conf.Get().Settings
	if settings == nil {
		return nil, nil
	}
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return nil, err
	}
	return strptr(string(data)), nil
}

func (r *siteResolver) LatestSettings(ctx context.Context) (*settingsResolver, error) {
	settings, err := db.Settings.GetLatest(ctx, r.configurationSubject())
	if err != nil {
		return nil, err
	}
	if settings == nil {
		return nil, nil
	}
	return &settingsResolver{&configurationSubject{site: r}, settings, nil}, nil
}

func (r *siteResolver) ConfigurationCascade() *configurationCascadeResolver {
	return &configurationCascadeResolver{subject: &configurationSubject{site: r}}
}

func (r *siteResolver) SettingsURL() string { return "/site-admin/global-settings" }

func (r *siteResolver) CanReloadSite(ctx context.Context) bool {
	err := backend.CheckCurrentUserIsSiteAdmin(ctx)
	return canReloadSite && err == nil
}

func (r *siteResolver) BuildVersion() string { return env.Version }

func (r *siteResolver) ProductVersion() string { return version.Version() }

func (r *siteResolver) HasCodeIntelligence() bool {
	return envvar.HasCodeIntelligence()
}

func (r *siteResolver) ProductSubscription() *productSubscriptionStatus {
	return &productSubscriptionStatus{}
}

func (r *siteResolver) Activity(ctx context.Context, args *struct {
	Days   *int32
	Weeks  *int32
	Months *int32
}) (*siteActivityResolver, error) {
	// 🚨 SECURITY
	// TODO(Dan, Beyang): this endpoint should eventually only be accessible by site admins.
	// It is temporarily exposed to all users on an instance.
	if envvar.SourcegraphDotComMode() {
		return nil, errors.New("site analytics is not available on sourcegraph.com")
	}
	opt := &useractivity.SiteActivityOptions{}
	if args.Days != nil {
		d := int(*args.Days)
		opt.DayPeriods = &d
	}
	if args.Weeks != nil {
		w := int(*args.Weeks)
		opt.WeekPeriods = &w
	}
	if args.Months != nil {
		m := int(*args.Months)
		opt.MonthPeriods = &m
	}
	activity, err := useractivity.GetSiteActivity(opt)
	if err != nil {
		return nil, err
	}
	return &siteActivityResolver{activity}, nil
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

func (r *siteConfigurationResolver) ValidationMessages(ctx context.Context) ([]string, error) {
	contents, err := r.pendingOrEffectiveContents(ctx)
	if err != nil {
		return nil, err
	}
	return conf.Validate(contents)
}

func (r *siteConfigurationResolver) CanUpdate() bool {
	// We assume the is-admin check has already been performed before constructing
	// our receiver.
	return conf.IsWritable() && processrestart.CanRestart()
}

func (r *siteConfigurationResolver) Source() string {
	s := conf.FilePath()
	if !conf.IsWritable() {
		s += " (read-only)"
	}
	return s
}

func (r *schemaResolver) UpdateSiteConfiguration(ctx context.Context, args *struct {
	Input string
}) (bool, error) {
	// 🚨 SECURITY: The site configuration contains secret tokens and credentials,
	// so only admins may view it.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return false, err
	}
	if err := conf.Write(args.Input); err != nil {
		return false, err
	}
	return conf.NeedServerRestart(), nil
}
