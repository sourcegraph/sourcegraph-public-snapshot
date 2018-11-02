package conf

import (
	"log"
	"os"
	"path/filepath"

	"github.com/sourcegraph/jsonx"
	"github.com/sourcegraph/sourcegraph/pkg/conf/conftypes"
)

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
		if err == nil && filepath.Ext(p) == ".test" {
			return modeTest
		}

		// Otherwise we default to client mode, so that most services need not
		// specify CONFIGURATION_MODE=client explicitly.
		return modeClient
	}
}

func init() {
	clientBasicStore := NewStore()
	clientCoreStore := NewStore()

	defaultClient = &client{
		basicStore:   clientBasicStore,
		coreStore:    clientCoreStore,
		basicFetcher: httpBasicFetcher{},
		coreFetcher:  httpCoreFetcher{},
	}

	mode := getMode()

	// Don't kickoff the background updaters for the client/server
	// when running test cases.
	if mode == modeTest {
		// Seed the client store with a dummy configuration for test cases.
		dummyConfig := `
		{
			// This is an empty configuration to run test cases.
		}`

		_, err := clientBasicStore.MaybeUpdate(dummyConfig, conftypes.ParseBasic)

		if err != nil {
			log.Fatalf("received error when setting up the basic store for the default client during test, err :%s", err)
		}

		_, err = clientCoreStore.MaybeUpdate(dummyConfig, conftypes.ParseCore)

		if err != nil {
			log.Fatalf("received error when setting up the basic store for the default client during test, err :%s", err)
		}

		return
	}

	// The default client is started in NewConfigurationServerFrontendOnly in
	// the case of server mode.
	if mode == modeClient {
		go defaultClient.continuouslyUpdate()
	}
}

// TODO(slimsag): remove this by passing an argument through?
var (
	configurationServerFrontendOnly *Server

	configurationServerFrontendOnlyInitialized = make(chan struct{})
)

// InitConfigurationServerFrontendOnly creates and returns a configuration
// server. This should only be invoked by the frontend, or else a panic will
// occur. This function should only ever be called once.
func InitConfigurationServerFrontendOnly() *Server {
	mode := getMode()
	if mode != modeServer {
		panic("cannot call this function except in server mode")
	}

	var server *Server
	server = NewServer(os.Getenv("SOURCEGRAPH_CONFIG_FILE"), os.Getenv("SOURCEGRAPH_CONFIG_CORE_FILE"))
	server.Start()

	// Install the passthrough fetcher for defaultClient in order to avoid deadlock issues.
	defaultClient.basicFetcher = passthroughBasicFetcherFrontendOnly{}
	defaultClient.coreFetcher = passthroughCoreFetcherFrontendOnly{}
	go defaultClient.continuouslyUpdate()

	close(configurationServerFrontendOnlyInitialized)
	configurationServerFrontendOnly = server
	return server
}

func detectDeadlock() {
	mode := getMode()
	if mode != modeServer {
		return
	}
	select {
	case <-configurationServerFrontendOnlyInitialized:
		// Configuration server is initialized.
	default:
		panic("deadlock detected: you have called conf.Get or conf.Watch before the frontend has been initialized (you may need to use conf.AsyncWatch instead)")
	}
}

// FormatOptions is the default format options that should be used for jsonx
// edit computation.
var FormatOptions = jsonx.FormatOptions{InsertSpaces: true, TabSize: 2, EOL: "\n"}
