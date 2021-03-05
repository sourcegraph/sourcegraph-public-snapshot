// Package conf provides functions for accessing the Site Configuration.
package conf

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/jsonx"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/schema"
)

// Unified represents the overall global Sourcegraph configuration from various
// sources:
//
// - The site configuration, from the database (from the site-admin panel).
// - Service connections, from the frontend (e.g. which gitservers to talk to).
//
type Unified struct {
	schema.SiteConfiguration
	ServiceConnections conftypes.ServiceConnections
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

func getMode() configurationMode {
	mode := os.Getenv("CONFIGURATION_MODE")

	switch mode {
	case "server":
		return modeServer
	case "client":
		return modeClient
	case "empty":
		return modeEmpty
	default:
		// Detect 'go test' and default to empty mode in that case.
		p, err := os.Executable()
		if err == nil && strings.Contains(strings.ToLower(p), "test") {
			return modeEmpty
		}

		// Otherwise we default to client mode, so that most services need not
		// specify CONFIGURATION_MODE=client explicitly.
		return modeClient
	}
}

var (
	configurationServerFrontendOnlyInitialized = make(chan struct{})
)

func initDefaultClient() *client {
	clientStore := newStore()
	defaultClient := &client{store: clientStore}

	mode := getMode()

	// Don't kickoff the background updaters for the client/server
	// when in empty mode.
	if mode == modeEmpty {
		close(configurationServerFrontendOnlyInitialized)

		// Seed the client store with an empty configuration.
		_, err := clientStore.MaybeUpdate(conftypes.RawUnified{
			Site:               "{}",
			ServiceConnections: conftypes.ServiceConnections{},
		})
		if err != nil {
			log.Fatalf("received error when setting up the store for the default client during test, err :%s", err)
		}
		return defaultClient
	}

	// The default client is started in InitConfigurationServerFrontendOnly in
	// the case of server mode.
	if mode == modeClient {
		go defaultClient.continuouslyUpdate(nil)
		close(configurationServerFrontendOnlyInitialized)
	}

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

func (c *cachedConfigurationSource) Write(ctx context.Context, input conftypes.RawUnified) error {
	c.entryMu.Lock()
	defer c.entryMu.Unlock()
	if err := c.source.Write(ctx, input); err != nil {
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
	defaultClient().passthrough = source

	go defaultClient().continuouslyUpdate(nil)
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
	siteConfigEscapeHatchPath = os.ExpandEnv(siteConfigEscapeHatchPath)

	var (
		ctx                                        = context.Background()
		lastKnownFileContents, lastKnownDBContents string
	)
	go func() {
		// First, ensure we populate the file with what is currently in the DB.
		for {
			config, err := c.Read(ctx)
			if err != nil {
				log15.Error("config: failed to read config from database, trying again in 1s", "error", err)
				time.Sleep(1 * time.Second)
				continue
			}
			if err := ioutil.WriteFile(siteConfigEscapeHatchPath, []byte(config.Site), 0644); err != nil {
				log15.Error("config: failed to write site config file, trying again in 1s", "error", err)
				time.Sleep(1 * time.Second)
				continue
			}
			lastKnownDBContents = config.Site
			lastKnownFileContents = config.Site
			break
		}

		// Watch for changes to the file AND the database.
		for {
			// If the file changes from what we last wrote, an admin made an edit to the file and
			// we should propagate it to the database for them.
			newFileContents, err := ioutil.ReadFile(siteConfigEscapeHatchPath)
			if err != nil {
				log15.Error("config: failed to read site config from disk, trying again in 1s", "path", siteConfigEscapeHatchPath)
				time.Sleep(1 * time.Second)
				continue
			}
			if string(newFileContents) != lastKnownFileContents {
				log15.Info("config: detected site config file edit, saving edit to database", "path", siteConfigEscapeHatchPath)
				config, err := c.Read(ctx)
				if err != nil {
					log15.Error("config: failed to save edit to database, trying again in 1s (read error)", "error", err)
					time.Sleep(1 * time.Second)
					continue
				}
				config.Site = string(newFileContents)
				err = c.Write(ctx, config)
				if err != nil {
					log15.Error("config: failed to save edit to database, trying again in 1s (write error)", "error", err)
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
				log15.Error("config: failed to read config from database(2), trying again in 1s (read error)", "error", err)
				time.Sleep(1 * time.Second)
				continue
			}
			if newDBConfig.Site != lastKnownDBContents {
				if err := ioutil.WriteFile(siteConfigEscapeHatchPath, []byte(newDBConfig.Site), 0644); err != nil {
					log15.Error("config: failed to write site config file, trying again in 1s", "error", err)
					time.Sleep(1 * time.Second)
					continue
				}
				lastKnownFileContents = newDBConfig.Site
			}
			time.Sleep(1 * time.Second)
		}
	}()
}
