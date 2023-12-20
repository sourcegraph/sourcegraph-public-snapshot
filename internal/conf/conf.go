// Package conf provides functions for accessing the Site Configuration.
package conf

import (
	"context"
	"log" //nolint:logging // TODO move all logging to sourcegraph/log
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/jsonx"
	sglog "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/schema"
)

// Unified represents the overall global Sourcegraph configuration from various
// sources:
//
// - The site configuration, from the database (from the site-admin panel).
// - Service connections, from the frontend (e.g. which gitservers to talk to).
type Unified struct {
	schema.SiteConfiguration
	ServiceConnectionConfig conftypes.ServiceConnections
}

var _ conftypes.UnifiedQuerier = Unified{}

func (u Unified) SiteConfig() schema.SiteConfiguration {
	return u.SiteConfiguration
}

func (u Unified) ServiceConnections() conftypes.ServiceConnections {
	return u.ServiceConnectionConfig
}

type configurationMode int

const (
	// The user of pkg/conf reads and writes to the configuration file.
	// This should only ever be used by frontend.
	modeServer configurationMode = iota

	// The user of pkg/conf only reads the configuration file.
	modeClient

	// The user of pkg/conf is a test case or explicitly opted to have no
	// configuration.
	modeEmpty
)

var (
	cachedModeOnce sync.Once
	cachedMode     configurationMode
)

func getMode() configurationMode {
	cachedModeOnce.Do(func() {
		cachedMode = getModeUncached()
	})
	return cachedMode
}

func getModeUncached() configurationMode {
	if deploy.IsSingleBinary() {
		// When running everything in the same process, use server mode.
		return modeServer
	}

	mode := os.Getenv("CONFIGURATION_MODE")

	switch mode {
	case "server":
		return modeServer
	case "client":
		return modeClient
	case "empty":
		return modeEmpty
	default:
		p, err := os.Executable()
		if err == nil && filepath.Base(p) == "sg" {
			// If we're  running `sg`, force the configuration mode to empty so `sg`
			// can make use of the `internal/database` package without configuration
			// side effects taking place.
			//
			// See https://github.com/sourcegraph/sourcegraph/issues/29222.
			return modeEmpty
		}

		if err == nil && strings.Contains(strings.ToLower(filepath.Base(p)), "test") {
			// If we detect 'go test', defaults to empty mode in that case.
			return modeEmpty
		}

		// Otherwise we default to client mode, so that most services need not
		// specify CONFIGURATION_MODE=client explicitly.
		return modeClient
	}
}

var configurationServerFrontendOnlyInitialized = make(chan struct{})

func initDefaultClient() *client {
	defaultClient := &client{
		store:          newStore(),
		lastUpdateTime: time.Now(),
	}

	mode := getMode()
	// Don't kickoff the background updaters for the client/server
	// when in empty mode.
	if mode == modeEmpty {
		close(configurationServerFrontendOnlyInitialized)

		// Seed the client store with an empty configuration.
		_, err := defaultClient.store.MaybeUpdate(conftypes.RawUnified{
			Site:               "{}",
			ServiceConnections: conftypes.ServiceConnections{},
		})
		if err != nil {
			log.Fatalf("received error when setting up the store for the default client during test, err :%s", err)
		}
	}

	m := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "src_conf_client_time_since_last_successful_update_seconds",
		Help: "Time since the last successful update of the configuration by the conf client",
	}, func() float64 {
		defaultClient.lastUpdateTimeMu.RLock()
		defer defaultClient.lastUpdateTimeMu.RUnlock()

		return time.Since(defaultClient.lastUpdateTime).Seconds()
	})

	prometheus.DefaultRegisterer.MustRegister(m)

	return defaultClient
}

// cachedConfigurationSource caches reads for a specified duration to reduce
// the number of reads against the underlying configuration source (e.g. a
// Postgres DB).
type cachedConfigurationSource struct {
	source ConfigurationSource

	ttl       time.Duration
	entryMu   sync.Mutex
	entry     *conftypes.RawUnified
	entryTime time.Time
}

func (c *cachedConfigurationSource) Read(ctx context.Context) (conftypes.RawUnified, error) {
	c.entryMu.Lock()
	defer c.entryMu.Unlock()
	if c.entry == nil || time.Since(c.entryTime) > c.ttl {
		updatedEntry, err := c.source.Read(ctx)
		if err != nil {
			return updatedEntry, err
		}
		c.entry = &updatedEntry
		c.entryTime = time.Now()
	}
	return *c.entry, nil
}

func (c *cachedConfigurationSource) Write(ctx context.Context, input conftypes.RawUnified, lastID int32, authorUserID int32) error {
	c.entryMu.Lock()
	defer c.entryMu.Unlock()
	if err := c.source.Write(ctx, input, lastID, authorUserID); err != nil {
		return err
	}
	c.entry = &input
	c.entryTime = time.Now()
	return nil
}

// InitConfigurationServerFrontendOnly creates and returns a configuration
// server. This should only be invoked by the frontend, or else a panic will
// occur. This function should only ever be called once.
func InitConfigurationServerFrontendOnly(source ConfigurationSource) *Server {
	mode := getMode()

	if mode == modeEmpty {
		return nil
	}

	if mode == modeClient {
		panic("cannot call this function while in client mode")
	}

	server := NewServer(&cachedConfigurationSource{
		source: source,
		// conf.Watch poll rate is 5s, so we use half that.
		ttl: 2500 * time.Millisecond,
	})
	server.Start()

	// Install the passthrough configuration source for defaultClient. This is
	// so that the frontend does not request configuration from itself via HTTP
	// and instead only relies on the DB.
	DefaultClient().passthrough = source

	// Notify the default client of updates to the source to ensure updates
	// propagate quickly.
	DefaultClient().sourceUpdates = server.sourceWrites

	go DefaultClient().continuouslyUpdate(nil)
	close(configurationServerFrontendOnlyInitialized)

	startSiteConfigEscapeHatchWorker(source)
	return server
}

// FormatOptions is the default format options that should be used for jsonx
// edit computation.
var FormatOptions = jsonx.FormatOptions{InsertSpaces: true, TabSize: 2, EOL: "\n"}

var siteConfigEscapeHatchPath = env.Get("SITE_CONFIG_ESCAPE_HATCH_PATH", "$HOME/site-config.json", "Path where the site-config.json escape-hatch file will be written.")

// startSiteConfigEscapeHatchWorker handles ensuring that edits to the ephemeral on-disk
// site-config.json file are propagated to the persistent DB and vice-versa. This acts as
// an escape hatch such that if a site admin configures their instance in a way that they
// cannot access the UI (for example by configuring auth in a way that locks them out)
// they can simply edit this file in any of the frontend containers to undo the change.
func startSiteConfigEscapeHatchWorker(c ConfigurationSource) {
	if os.Getenv("NO_SITE_CONFIG_ESCAPE_HATCH") == "1" {
		return
	}

	siteConfigEscapeHatchPath = os.ExpandEnv(siteConfigEscapeHatchPath)
	if deploy.IsSingleBinary() {
		// For single-binary mode, always store the site config on disk, and this is achieved through
		// making the "escape hatch file" point to our desired location on disk.
		// The concept of an escape hatch file is not something users care
		// about (it only makes sense in Docker/Kubernetes, e.g. to edit the config
		// file if the sourcegraph-frontend container is crashing) - it runs
		// natively and this mechanism is just a convenient way for us to keep
		// the file on disk as our source of truth.
		siteConfigEscapeHatchPath = os.Getenv("SITE_CONFIG_FILE")
	}

	var (
		ctx                                        = context.Background()
		lastKnownFileContents, lastKnownDBContents string
		lastKnownConfigID                          int32
		logger                                     = sglog.Scoped("SiteConfigEscapeHatch").With(sglog.String("path", siteConfigEscapeHatchPath))
	)
	go func() {
		// First, ensure we populate the file with what is currently in the DB.
		for {
			config, err := c.Read(ctx)
			if err != nil {
				logger.Warn("failed to read config from database, trying again in 1s", sglog.Error(err))
				time.Sleep(1 * time.Second)
				continue
			}
			if err := os.WriteFile(siteConfigEscapeHatchPath, []byte(config.Site), 0644); err != nil {
				logger.Warn("failed to write site config file, trying again in 1s", sglog.Error(err))
				time.Sleep(1 * time.Second)
				continue
			}
			lastKnownDBContents = config.Site
			lastKnownFileContents = config.Site
			lastKnownConfigID = config.ID
			break
		}

		// Watch for changes to the file AND the database.
		for {
			// If the file changes from what we last wrote, an admin made an edit to the file and
			// we should propagate it to the database for them.
			newFileContents, err := os.ReadFile(siteConfigEscapeHatchPath)
			if err != nil {
				logger.Warn("failed to read site config from disk, trying again in 1s")
				time.Sleep(1 * time.Second)
				continue
			}
			if string(newFileContents) != lastKnownFileContents {
				logger.Info("detected site config file edit, saving edit to database")
				config, err := c.Read(ctx)
				if err != nil {
					logger.Warn("failed to save edit to database, trying again in 1s (read error)", sglog.Error(err))
					time.Sleep(1 * time.Second)
					continue
				}
				config.Site = string(newFileContents)

				// NOTE: authorUserID is 0 because this code is on the start-up path and we will
				// never have a non-nil actor available here to determine the user ID. This is
				// consistent with the behaviour of site config creation via SITE_CONFIG_FILE.
				//
				// A value of 0 will be treated as null when writing to the the database for this column.
				err = c.Write(ctx, config, lastKnownConfigID, 0)
				if err != nil {
					logger.Warn("failed to save edit to database, trying again in 1s (write error)", sglog.Error(err))
					time.Sleep(1 * time.Second)
					continue
				}
				lastKnownFileContents = config.Site
				continue
			}

			// If the database changes from what we last remember, an admin made an edit to the
			// database (e.g. through the web UI or by editing the file of another frontend
			// process), and we should propagate it to the file on disk.
			newDBConfig, err := c.Read(ctx)
			if err != nil {
				logger.Warn("failed to read config from database(2), trying again in 1s (read error)", sglog.Error(err))
				time.Sleep(1 * time.Second)
				continue
			}
			if newDBConfig.Site != lastKnownDBContents {
				if err := os.WriteFile(siteConfigEscapeHatchPath, []byte(newDBConfig.Site), 0644); err != nil {
					logger.Warn("failed to write site config file, trying again in 1s", sglog.Error(err))
					time.Sleep(1 * time.Second)
					continue
				}
				lastKnownDBContents = newDBConfig.Site
				lastKnownFileContents = newDBConfig.Site
				lastKnownConfigID = newDBConfig.ID
			}
			time.Sleep(1 * time.Second)
		}
	}()
}
