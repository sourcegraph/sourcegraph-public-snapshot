package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/user"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/pkg/db/confdb"
	"github.com/sourcegraph/sourcegraph/pkg/jsonc"
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

		devOverrideExtSvcConfig := os.Getenv("DEV_OVERRIDE_EXTSVC_CONFIG")
		if devOverrideExtSvcConfig != "" {
			existing, err := db.ExternalServices.List(context.Background(), db.ExternalServicesListOptions{})
			if err != nil {
				log.Fatal(err)
			}
			if len(existing) > 0 {
				return
			}

			extsvc, err := ioutil.ReadFile(devOverrideExtSvcConfig)
			if err != nil {
				log.Fatal(err)
			}
			var configs map[string][]*json.RawMessage
			if err := jsonc.Unmarshal(string(extsvc), &configs); err != nil {
				log.Fatal(err)
			}
			for key, cfgs := range configs {
				for i, cfg := range cfgs {
					marshaledCfg, err := json.MarshalIndent(cfg, "", "  ")
					if err != nil {
						log.Fatal(err)
					}
					if err := db.ExternalServices.Create(context.Background(), &types.ExternalService{
						Kind:        key,
						DisplayName: fmt.Sprintf("Dev %s #%d", key, i+1),
						Config:      string(marshaledCfg),
					}); err != nil {
						log.Fatal(err)
					}
				}
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
		ServiceConnections: conftypes.ServiceConnections{
			PostgresDSN: postgresDSN(),
		},
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

func postgresDSN() string {
	username := ""
	if user, err := user.Current(); err == nil {
		username = user.Username
	}
	return doPostgresDSN(username, os.Getenv)
}

func doPostgresDSN(currentUser string, getenv func(string) string) string {
	// PGDATASOURCE is a sourcegraph specific variable for just setting the DSN
	if dsn := getenv("PGDATASOURCE"); dsn != "" {
		return dsn
	}

	// TODO match logic in lib/pq
	// https://sourcegraph.com/github.com/lib/pq@d6156e141ac6c06345c7c73f450987a9ed4b751f/-/blob/connector.go#L42
	dsn := &url.URL{
		Scheme:   "postgres",
		Host:     "127.0.0.1:5432",
		RawQuery: "sslmode=disable",
	}

	// Username preference: PGUSER, $USER, postgres
	username := "postgres"
	if currentUser != "" {
		username = currentUser
	}
	if user := getenv("PGUSER"); user != "" {
		username = user
	}

	if password := getenv("PGPASSWORD"); password != "" {
		dsn.User = url.UserPassword(username, password)
	} else {
		dsn.User = url.User(username)
	}

	if host := getenv("PGHOST"); host != "" {
		dsn.Host = host
	}

	if port := getenv("PGPORT"); port != "" {
		dsn.Host += ":" + port
	}

	if db := getenv("PGDATABASE"); db != "" {
		dsn.Path = db
	}

	if sslmode := getenv("PGSSLMODE"); sslmode != "" {
		qry := dsn.Query()
		qry.Set("sslmode", sslmode)
		dsn.RawQuery = qry.Encode()
	}

	return dsn.String()
}
