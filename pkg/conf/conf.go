package conf

import (
	"os"

	"github.com/sourcegraph/jsonx"
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
		return modeTest
	}

}

func init() {
	defaultClient = &client{
		store:   Store(),
		fetcher: httpFetcher{},
	}

	mode := getMode()

	// Don't kickoff the background updaters for the client/server
	// when running test cases.
	if mode == modeTest {
		return
	}

	if mode == modeServer {
		DefaultServerFrontendOnly = &server{
			configFilePath: os.Getenv("SOURCEGRAPH_CONFIG_FILE"),
			store:          Store(),
			fileWrite:      make(chan chan struct{}, 1),
		}

		go DefaultServerFrontendOnly.watchDisk()
		defaultClient.fetcher = passthroughFetcherFrontendOnly{}
	}

	go defaultClient.continouslyUpdate()
}

// FormatOptions is the default format options that should be used for jsonx
// edit computation.
var FormatOptions = jsonx.FormatOptions{InsertSpaces: true, TabSize: 2, EOL: "\n"}
