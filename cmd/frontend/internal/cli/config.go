package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/database/postgresdsn"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func printConfigValidation(logger log.Logger) {
	logger = logger.Scoped("configValidation", "")
	messages, err := conf.Validate(conf.Raw())
	if err != nil {
		logger.Warn("unable to validate Sourcegraph site configuration", log.Error(err))
		return
	}

	if len(messages) > 0 {
		logger.Warn("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		logger.Warn("⚠️ Warnings related to the Sourcegraph site configuration:")
		for _, verr := range messages {
			logger.Warn(verr.String())
		}
		logger.Warn("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	}
}

var metricConfigOverrideUpdates = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "src_frontend_config_file_watcher_updates",
	Help: "Incremented each time the config file is updated.",
}, []string{"status"})

// readSiteConfigFile reads and merges the paths. paths is the value of the
// envvar SITE_CONFIG_FILE seperated by os.ListPathSeparator (":"). The
// merging just concats the objects together. So does not check for things
// like duplicate keys between files.
func readSiteConfigFile(paths []string) ([]byte, error) {
	// special case 1
	if len(paths) == 1 {
		return os.ReadFile(paths[0])
	}

	var merged bytes.Buffer
	merged.WriteString("// merged SITE_CONFIG_FILE\n{\n")

	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			return nil, err
		}

		var m map[string]*json.RawMessage
		err = jsonc.Unmarshal(string(b), &m)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse JSON in %s", p)
		}

		var keys []string
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		_, _ = fmt.Fprintf(&merged, "\n// BEGIN %s\n", p)
		for _, k := range keys {
			keyB, _ := json.Marshal(k)
			valB, _ := json.Marshal(m[k])
			_, _ = fmt.Fprintf(&merged, "  %s: %s,\n", keyB, valB)
		}
		_, _ = fmt.Fprintf(&merged, "// END %s\n", p)
	}

	merged.WriteString("}\n")
	formatted, err := jsonc.Format(merged.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to format JSONC")
	}
	return []byte(formatted), nil
}

func overrideSiteConfig(ctx context.Context, logger log.Logger, db database.DB) error {
	logger = logger.Scoped("overrideSiteConfig", "")
	path := os.Getenv("SITE_CONFIG_FILE")
	if path == "" {
		return nil
	}
	cs := newConfigurationSource(logger, db)
	paths := filepath.SplitList(path)
	updateFunc := func(ctx context.Context) error {
		raw, err := cs.Read(ctx)
		if err != nil {
			return err
		}
		site, err := readSiteConfigFile(paths)
		if err != nil {
			return errors.Wrap(err, "reading SITE_CONFIG_FILE")
		}
		raw.Site = string(site)

		err = cs.Write(ctx, raw)
		if err != nil {
			return errors.Wrap(err, "writing site config overrides to database")
		}
		return nil
	}
	err := updateFunc(ctx)
	if err != nil {
		return err
	}

	go watchUpdate(ctx, logger, path, updateFunc)
	return nil
}

func overrideGlobalSettings(ctx context.Context, logger log.Logger, db database.DB) error {
	logger = logger.Scoped("overrideGlobalSettings", "")
	path := os.Getenv("GLOBAL_SETTINGS_FILE")
	if path == "" {
		return nil
	}
	settings := db.Settings()
	update := func(ctx context.Context) error {
		globalSettingsBytes, err := os.ReadFile(path)
		if err != nil {
			return errors.Wrap(err, "reading GLOBAL_SETTINGS_FILE")
		}
		currentSettings, err := settings.GetLatest(ctx, api.SettingsSubject{Site: true})
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
			_, err = settings.CreateIfUpToDate(ctx, api.SettingsSubject{Site: true}, lastID, nil, globalSettings)
			if err != nil {
				return errors.Wrap(err, "writing global setting override to database")
			}
		}
		return nil
	}
	if err := update(ctx); err != nil {
		return err
	}
	go watchUpdate(ctx, logger, path, update)

	return nil
}

func overrideExtSvcConfig(ctx context.Context, logger log.Logger, db database.DB) error {
	logger = logger.Scoped("overrideExtSvcConfig", "")
	path := os.Getenv("EXTSVC_CONFIG_FILE")
	if path == "" {
		return nil
	}
	extsvcs := db.ExternalServices()
	cs := newConfigurationSource(logger, db)

	update := func(ctx context.Context) error {
		raw, err := cs.Read(ctx)
		if err != nil {
			return err
		}
		parsed, err := conf.ParseConfig(raw)
		if err != nil {
			return errors.Wrap(err, "parsing extsvc config")
		}
		confGet := func() *conf.Unified { return parsed }

		extsvcConfig, err := os.ReadFile(path)
		if err != nil {
			return errors.Wrap(err, "reading EXTSVC_CONFIG_FILE")
		}
		var rawConfigs map[string][]*json.RawMessage
		if err := jsonc.Unmarshal(string(extsvcConfig), &rawConfigs); err != nil {
			return errors.Wrap(err, "parsing EXTSVC_CONFIG_FILE")
		}
		if len(rawConfigs) == 0 {
			logger.Warn("EXTSVC_CONFIG_FILE contains zero external service configurations")
		}

		existing, err := extsvcs.List(ctx, database.ExternalServicesListOptions{
			// NOTE: External services loaded from config file do not have namespace specified.
			// Therefore, we only need to load those from database.
			NoNamespace: true,
		})
		if err != nil {
			return errors.Wrap(err, "ExternalServices.List")
		}

		// Perform delta update for external services. We don't want to just delete all
		// external services and re-add all of them, because that would cause
		// repo-updater to need to update repositories and reassociate them with external
		// services each time the frontend restarts.
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

				// When overriding external service config from a file we allow setting the value
				// of the cloud_default column.
				var cloudDefault bool
				switch key {
				case extsvc.KindGitHub:
					var c schema.GitHubConnection
					if err = json.Unmarshal(marshaledCfg, &c); err != nil {
						return err
					}
					cloudDefault = c.CloudDefault

				case extsvc.KindGitLab:
					var c schema.GitLabConnection
					if err = json.Unmarshal(marshaledCfg, &c); err != nil {
						return err
					}
					cloudDefault = c.CloudDefault

				}

				toAdd[&types.ExternalService{
					Kind:         key,
					DisplayName:  fmt.Sprintf("%s #%d", key, i+1),
					Config:       extsvc.NewUnencryptedConfig(string(marshaledCfg)),
					CloudDefault: cloudDefault,
				}] = true
			}
		}
		// Now eliminate operations from toAdd/toRemove where the config
		// file and DB describe an equivalent external service.
		isEquiv := func(a, b *types.ExternalService) (bool, error) {
			aConfig, err := a.Config.Decrypt(ctx)
			if err != nil {
				return false, err
			}

			bConfig, err := b.Config.Decrypt(ctx)
			if err != nil {
				return false, err
			}

			return a.Kind == b.Kind && a.DisplayName == b.DisplayName && aConfig == bConfig, nil
		}
		shouldUpdate := func(a, b *types.ExternalService) (bool, error) {
			aConfig, err := a.Config.Decrypt(ctx)
			if err != nil {
				return false, err
			}

			bConfig, err := b.Config.Decrypt(ctx)
			if err != nil {
				return false, err
			}

			return a.Kind == b.Kind && a.DisplayName == b.DisplayName && aConfig != bConfig, nil
		}
		for a := range toAdd {
			for b := range toRemove {
				if ok, err := isEquiv(a, b); err != nil {
					return err
				} else if ok {
					// Nothing changed
					delete(toAdd, a)
					delete(toRemove, b)
					continue
				}

				if ok, err := shouldUpdate(a, b); err != nil {
					return err
				} else if ok {
					delete(toAdd, a)
					delete(toRemove, b)
					toUpdate[b.ID] = a
				}
			}
		}

		// Apply the delta update.
		for extSvc := range toRemove {
			logger.Debug("Deleting external service", log.Int64("id", extSvc.ID), log.String("displayName", extSvc.DisplayName))
			err := extsvcs.Delete(ctx, extSvc.ID)
			if err != nil {
				return errors.Wrap(err, "ExternalServices.Delete")
			}
		}
		for extSvc := range toAdd {
			logger.Debug("Adding external service", log.String("displayName", extSvc.DisplayName))
			if err := extsvcs.Create(ctx, confGet, extSvc); err != nil {
				return errors.Wrap(err, "ExternalServices.Create")
			}
		}

		ps := confGet().AuthProviders
		for id, extSvc := range toUpdate {
			logger.Debug("Updating external service", log.Int64("id", id), log.String("displayName", extSvc.DisplayName))

			rawConfig, err := extSvc.Config.Decrypt(ctx)
			if err != nil {
				return err
			}

			update := &database.ExternalServiceUpdate{DisplayName: &extSvc.DisplayName, Config: &rawConfig, CloudDefault: &extSvc.CloudDefault}
			if err := extsvcs.Update(ctx, ps, id, update); err != nil {
				return errors.Wrap(err, "ExternalServices.Update")
			}
		}
		return nil
	}
	if err := update(ctx); err != nil {
		return err
	}

	go watchUpdate(ctx, logger, path, update)

	return nil
}

func watchUpdate(ctx context.Context, logger log.Logger, path string, update func(context.Context) error) {
	logger = logger.Scoped("watch", "")
	events, err := watchPaths(ctx, path)
	if err != nil {
		logger.Error("failed to watch config override files", log.Error(err))
		return
	}
	for err := range events {
		if err != nil {
			logger.Warn("error while watching config override files", log.Error(err))
			metricConfigOverrideUpdates.WithLabelValues("watch_failed").Inc()
			continue
		}

		if err := update(ctx); err != nil {
			logger.Error("failed to update configuration from modified config override file", log.Error(err), log.String("file", path))
			metricConfigOverrideUpdates.WithLabelValues("update_failed").Inc()
		} else {
			logger.Info("updated configuration from modified config override files", log.String("file", path))
			metricConfigOverrideUpdates.WithLabelValues("success").Inc()
		}
	}
}

// watchPaths returns a channel which watches the non-empty paths. Whenever
// any path changes a nil error is sent down chan. If an error occurs it is
// sent. chan is closed when ctx is Done.
//
// Note: This can send many events even if the file content hasn't
// changed. For example chmod events are sent. Another is a rename is two
// events for watcher (remove and create). Additionally if a file is removed
// the watch is removed. Even if a file with the same name is created in its
// place later.
func watchPaths(ctx context.Context, paths ...string) (<-chan error, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	for _, p := range paths {
		// as a convenience ignore empty paths
		if p == "" {
			continue
		}
		if err := watcher.Add(p); err != nil {
			return nil, errors.Wrapf(err, "failed to add %s to watcher", p)
		}
	}

	out := make(chan error)
	go func() {
		for {
			select {
			case <-ctx.Done():
				err := watcher.Close()
				if err != nil {
					out <- err
				}
				close(out)
				return

			case <-watcher.Events:
				out <- nil

			case err := <-watcher.Errors:
				out <- err

			}
		}
	}()

	return out, nil
}

func newConfigurationSource(logger log.Logger, db database.DB) *configurationSource {
	return &configurationSource{
		logger: logger.Scoped("configurationSource", ""),
		db:     db,
	}
}

type configurationSource struct {
	logger log.Logger
	db     database.DB
}

func (c *configurationSource) Read(ctx context.Context) (conftypes.RawUnified, error) {
	site, err := c.db.Conf().SiteGetLatest(ctx)
	if err != nil {
		return conftypes.RawUnified{}, errors.Wrap(err, "ConfStore.SiteGetLatest")
	}

	return conftypes.RawUnified{
		Site:               site.Contents,
		ServiceConnections: serviceConnections(c.logger),
	}, nil
}

func (c *configurationSource) Write(ctx context.Context, input conftypes.RawUnified) error {
	// TODO(slimsag): future: pass lastID through for race prevention
	site, err := c.db.Conf().SiteGetLatest(ctx)
	if err != nil {
		return errors.Wrap(err, "ConfStore.SiteGetLatest")
	}
	_, err = c.db.Conf().SiteCreateIfUpToDate(ctx, &site.ID, input.Site)
	if err != nil {
		return errors.Wrap(err, "ConfStore.SiteCreateIfUpToDate")
	}
	return nil
}

var (
	serviceConnectionsVal  conftypes.ServiceConnections
	serviceConnectionsOnce sync.Once

	gitservers = endpoint.New(func() string {
		v := os.Getenv("SRC_GIT_SERVERS")
		if v == "" {
			// Detect 'go test' and setup default addresses in that case.
			p, err := os.Executable()
			if err == nil && strings.HasSuffix(p, ".test") {
				return "gitserver:3178"
			}
			return "k8s+rpc://gitserver:3178?kind=sts"
		}
		return v
	}())
)

func serviceConnections(logger log.Logger) conftypes.ServiceConnections {
	serviceConnectionsOnce.Do(func() {
		dsns, err := postgresdsn.DSNsBySchema(schemas.SchemaNames)
		if err != nil {
			panic(err.Error())
		}

		serviceConnectionsVal = conftypes.ServiceConnections{
			PostgresDSN:          dsns["frontend"],
			CodeIntelPostgresDSN: dsns["codeintel"],
			CodeInsightsDSN:      dsns["codeinsights"],
		}
	})

	addrs, err := gitservers.Endpoints()
	if err != nil {
		logger.Error("failed to get gitserver endpoints for service connections", log.Error(err))
	}

	return conftypes.ServiceConnections{
		GitServers:           addrs,
		PostgresDSN:          serviceConnectionsVal.PostgresDSN,
		CodeIntelPostgresDSN: serviceConnectionsVal.CodeIntelPostgresDSN,
		CodeInsightsDSN:      serviceConnectionsVal.CodeInsightsDSN,
	}
}
