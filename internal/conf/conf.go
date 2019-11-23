// Package conf provides functions for accessing the Site Configuration.
package conf

import (
	"context"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/sourcegraph/jsonx"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/schema"
)

// Unified represents the overall global Sourcegraph configuration from various
// sources:
//
// - The critical configuration, from the database (from the management console).
// - The site configuration, from the database (from the site-admin panel).
// - Service connections, from the frontend (e.g. which gitservers to talk to).
//
type Unified struct {
	schema.SiteConfiguration
	Critical           schema.CriticalConfiguration
	ServiceConnections conftypes.ServiceConnections
}

type configurationMode int

const (
	// The user of pkg/conf reads and writes to the configuration file.
	// This should only ever be used by frontend.
	modeServer configurationMode = iota

	// The user of pkg/conf only reads the configuration file.
	modeClient

	// The user of pkg/conf is a test case.
	modeTest
)

func getMode() configurationMode {
	mode := os.Getenv("CONFIGURATION_MODE")

	switch mode {
	case "server":
		return modeServer
	case "client":
		return modeClient
	default:
		// Detect 'go test' and default to test mode in that case.
		p, err := os.Executable()
		if err == nil && strings.Contains(strings.ToLower(p), "test") {
			return modeTest
		}

		// Otherwise we default to client mode, so that most services need not
		// specify CONFIGURATION_MODE=client explicitly.
		return modeClient
	}
}

var (
	configurationServerFrontendOnlyInitialized = make(chan struct{})
)

func init() {
	clientStore := NewStore()
	defaultClient = &client{store: clientStore}

	mode := getMode()

	// Don't kickoff the background updaters for the client/server
	// when running test cases.
	if mode == modeTest {
		close(configurationServerFrontendOnlyInitialized)

		// Seed the client store with a dummy configuration for test cases.
		_, err := clientStore.MaybeUpdate(conftypes.RawUnified{
			Critical:           "{}",
			Site:               "{}",
			ServiceConnections: conftypes.ServiceConnections{},
		})
		if err != nil {
			log.Fatalf("received error when setting up the store for the default client during test, err :%s", err)
		}
		return
	}

	// The default client is started in InitConfigurationServerFrontendOnly in
	// the case of server mode.
	if mode == modeClient {
		go defaultClient.continuouslyUpdate(nil)
		close(configurationServerFrontendOnlyInitialized)
	}
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

	if mode == modeTest {
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
	defaultClient.passthrough = source

	go defaultClient.continuouslyUpdate(nil)
	close(configurationServerFrontendOnlyInitialized)
	return server
}

// FormatOptions is the default format options that should be used for jsonx
// edit computation.
var FormatOptions = jsonx.FormatOptions{InsertSpaces: true, TabSize: 2, EOL: "\n"}
