package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/db/confdb"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
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
			log15.Warn(verr.String())
		}
		log15.Warn("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	}
}

// handleConfigOverrides handles allowing dev environments to forcibly override
// the configuration in the database upon startup. This is used to e.g. ensure
// dev environments have a consistent configuration and to load secrets from a
// separate private repository.
//
// As this method writes to the configuration DB, it should be invoked before
// the configuration server is started but after PostgreSQL is connected.
func handleConfigOverrides() error {
	ctx := context.Background()

	overrideCriticalConfig := os.Getenv("CRITICAL_CONFIG_FILE")
	overrideSiteConfig := os.Getenv("SITE_CONFIG_FILE")
	overrideExtSvcConfig := os.Getenv("EXTSVC_CONFIG_FILE")
	overrideGlobalSettings := os.Getenv("GLOBAL_SETTINGS_FILE")
	overrideAny := overrideCriticalConfig != "" || overrideSiteConfig != "" || overrideExtSvcConfig != "" || overrideGlobalSettings != ""
	if overrideAny || conf.IsDev(conf.DeployType()) {
		raw, err := (&configurationSource{}).Read(ctx)
		if err != nil {
			return errors.Wrap(err, "reading existing config for applying overrides")
		}

		if overrideCriticalConfig != "" {
			critical, err := ioutil.ReadFile(overrideCriticalConfig)
			if err != nil {
				return errors.Wrap(err, "reading CRITICAL_CONFIG_FILE")
			}
			raw.Critical = string(critical)
		}

		if overrideSiteConfig != "" {
			site, err := ioutil.ReadFile(overrideSiteConfig)
			if err != nil {
				return errors.Wrap(err, "reading SITE_CONFIG_FILE")
			}
			raw.Site = string(site)
		}

		if overrideCriticalConfig != "" || overrideSiteConfig != "" {
			err := (&configurationSource{}).Write(ctx, raw)
			if err != nil {
				return errors.Wrap(err, "writing critical/site config overrides to database")
			}
		}

		if overrideGlobalSettings != "" {
			globalSettingsBytes, err := ioutil.ReadFile(overrideGlobalSettings)
			if err != nil {
				return errors.Wrap(err, "reading GLOBAL_SETTINGS_FILE")
			}
			currentSettings, err := db.Settings.GetLatest(ctx, api.SettingsSubject{Site: true})
			if err != nil {
				return errors.Wrap(err, "could not fetch current settings")
			}
			// Only overwrite the settings if the current settings differ, don't exist, or were
			// created by a human user to prevent creating unnecessary rows in the DB.
			globalSettings := string(globalSettingsBytes)
			if currentSettings == nil || currentSettings.AuthorUserID != nil || currentSettings.Contents != globalSettings {
				var lastID *int32 = nil
				if currentSettings != nil {
					lastID = &currentSettings.ID
				}
				_, err = db.Settings.CreateIfUpToDate(ctx, api.SettingsSubject{Site: true}, lastID, nil, globalSettings)
				if err != nil {
					return errors.Wrap(err, "writing global setting override to database")
				}
			}
		}

		if overrideExtSvcConfig != "" {
			parsed, err := conf.ParseConfig(raw)
			if err != nil {
				return errors.Wrap(err, "parsing critical/site config")
			}
			confGet := func() *conf.Unified { return parsed }

			extsvc, err := ioutil.ReadFile(overrideExtSvcConfig)
			if err != nil {
				return errors.Wrap(err, "reading EXTSVC_CONFIG_FILE")
			}
			var rawConfigs map[string][]*json.RawMessage
			if err := jsonc.Unmarshal(string(extsvc), &rawConfigs); err != nil {
				return errors.Wrap(err, "parsing EXTSVC_CONFIG_FILE")
			}
			if len(rawConfigs) == 0 {
				log15.Warn("EXTSVC_CONFIG_FILE contains zero external service configurations")
			}

			existing, err := db.ExternalServices.List(ctx, db.ExternalServicesListOptions{})
			if err != nil {
				return errors.Wrap(err, "ExternalServices.List")
			}

			// Perform delta update for external services. We don't want to
			// just delete all external services and re-add all of them,
			// because that would cause repo-updater to need to update
			// repositories and reassociate them with external services each
			// time the frontend restarts.
			//
			// Start out by assuming we will remove all and re-add all.
			var (
				toAdd    = make(map[*types.ExternalService]bool)
				toRemove = make(map[*types.ExternalService]bool)
				toUpdate = make(map[int64]*types.ExternalService)
			)
			for _, existing := range existing {
				toRemove[existing] = true
			}
			for key, cfgs := range rawConfigs {
				for i, cfg := range cfgs {
					marshaledCfg, err := json.MarshalIndent(cfg, "", "  ")
					if err != nil {
						return errors.Wrap(err, fmt.Sprintf("marshaling extsvc config ([%v][%v])", key, i))
					}
					toAdd[&types.ExternalService{
						Kind:        key,
						DisplayName: fmt.Sprintf("%s #%d", key, i+1),
						Config:      string(marshaledCfg),
					}] = true
				}
			}
			// Now eliminate operations from toAdd/toRemove where the config
			// file and DB describe an equivalent external service.
			isEquiv := func(a, b *types.ExternalService) bool {
				return a.Kind == b.Kind && a.DisplayName == b.DisplayName && a.Config == b.Config
			}
			shouldUpdate := func(a, b *types.ExternalService) bool {
				return a.Kind == b.Kind && a.DisplayName == b.DisplayName && a.Config != b.Config
			}
			for a := range toAdd {
				for b := range toRemove {
					if isEquiv(a, b) { // Nothing changed
						delete(toAdd, a)
						delete(toRemove, b)
					} else if shouldUpdate(a, b) {
						delete(toAdd, a)
						delete(toRemove, b)
						toUpdate[b.ID] = a
					}
				}
			}

			// Apply the delta update.
			for extSvc := range toRemove {
				log15.Debug("Deleting external service", "id", extSvc.ID, "displayName", extSvc.DisplayName)
				err := db.ExternalServices.Delete(ctx, extSvc.ID)
				if err != nil {
					return errors.Wrap(err, "ExternalServices.Delete")
				}
			}
			for extSvc := range toAdd {
				log15.Debug("Adding external service", "displayName", extSvc.DisplayName)
				if err := db.ExternalServices.Create(ctx, confGet, extSvc); err != nil {
					return errors.Wrap(err, "ExternalServices.Create")
				}
			}

			ps := confGet().AuthProviders
			for id, extSvc := range toUpdate {
				log15.Debug("Updating external service", "id", id, "displayName", extSvc.DisplayName)

				update := &db.ExternalServiceUpdate{DisplayName: &extSvc.DisplayName, Config: &extSvc.Config}
				if err := db.ExternalServices.Update(ctx, ps, id, update); err != nil {
					return errors.Wrap(err, "ExternalServices.Update")
				}
			}
		}
	}
	return nil
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

		ServiceConnections: serviceConnections(),
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

var (
	serviceConnectionsVal  conftypes.ServiceConnections
	serviceConnectionsOnce sync.Once
)

func serviceConnections() conftypes.ServiceConnections {
	serviceConnectionsOnce.Do(func() {
		username := ""
		if user, err := user.Current(); err == nil {
			username = user.Username
		}

		serviceConnectionsVal = conftypes.ServiceConnections{
			GitServers:  gitServers(),
			PostgresDSN: dbutil.PostgresDSN(username, os.Getenv),
		}
	})
	return serviceConnectionsVal
}

func gitServers() []string {
	v := os.Getenv("SRC_GIT_SERVERS")
	if v == "" {
		// Detect 'go test' and setup default addresses in that case.
		p, err := os.Executable()
		if err == nil && strings.HasSuffix(p, ".test") {
			v = "gitserver:3178"
		}
	}
	return strings.Fields(v)
}
