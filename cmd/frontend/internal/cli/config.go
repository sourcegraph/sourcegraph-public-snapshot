package cli

import (
	"context"
	"io/ioutil"
	"log"
	"os"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/pkg/db/confdb"

	log15 "gopkg.in/inconshreveable/log15.v2"
)

func printConfigValidation() {
	messages, err := conf.Validate(globals.ConfigurationServerFrontendOnly.Raw())
	if err != nil {
		log.Printf("Warning: Unable to validate Sourcegraph site configuration: %s", err)
		return
	}

	if len(messages) > 0 {
		log15.Warn("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		log15.Warn("⚠️ Warnings related to the Sourcegraph site configuration:")
		for _, verr := range messages {
			log15.Warn(verr)
		}
		log15.Warn("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	}
}

// handleConfigOverrides handles allowing dev environments to forcibly override
// the configuration in the database upon startup. This is used to e.g. ensure
// dev environments have a consistent configuration and to load secrets from a
// separate private repository.
func handleConfigOverrides() {
	if conf.IsDev(conf.DeployType()) {
		raw := conf.Raw()

		devOverrideCriticalConfig := os.Getenv("DEV_OVERRIDE_CRITICAL_CONFIG")
		if devOverrideCriticalConfig != "" {
			critical, err := ioutil.ReadFile(devOverrideCriticalConfig)
			if err != nil {
				log.Fatal(err)
			}
			raw.Critical = string(critical)
		}

		devOverrideSiteConfig := os.Getenv("DEV_OVERRIDE_SITE_CONFIG")
		if devOverrideSiteConfig != "" {
			site, err := ioutil.ReadFile(devOverrideSiteConfig)
			if err != nil {
				log.Fatal(err)
			}
			raw.Site = string(site)
		}

		if devOverrideCriticalConfig != "" || devOverrideSiteConfig != "" {
			err := (&configurationSource{}).Write(context.Background(), raw)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

type configurationSource struct{}

func (c configurationSource) Read(ctx context.Context) (conftypes.RawUnified, error) {
	critical, err := confdb.CriticalGetLatest(ctx)
	if err != nil {
		return conftypes.RawUnified{}, errors.Wrap(err, "confdb.CriticalGetLatest")
	}
	site, err := confdb.SiteGetLatest(ctx)
	if err != nil {
		return conftypes.RawUnified{}, errors.Wrap(err, "confdb.SiteGetLatest")
	}
	return conftypes.RawUnified{
		Critical: critical.Contents,
		Site:     site.Contents,

		// TODO(slimsag): future: pass GitServers list via this.
		ServiceConnections: conftypes.ServiceConnections{},
	}, nil
}

func (c configurationSource) Write(ctx context.Context, input conftypes.RawUnified) error {
	// TODO(slimsag): future: pass lastID through for race prevention
	critical, err := confdb.CriticalGetLatest(ctx)
	if err != nil {
		return errors.Wrap(err, "confdb.CriticalGetLatest")
	}
	site, err := confdb.SiteGetLatest(ctx)
	if err != nil {
		return errors.Wrap(err, "confdb.SiteGetLatest")
	}

	_, err = confdb.CriticalCreateIfUpToDate(ctx, &critical.ID, input.Critical)
	if err != nil {
		return errors.Wrap(err, "confdb.CriticalCreateIfUpToDate")
	}
	_, err = confdb.SiteCreateIfUpToDate(ctx, &site.ID, input.Site)
	if err != nil {
		return errors.Wrap(err, "confdb.SiteCreateIfUpToDate")
	}
	return nil
}
