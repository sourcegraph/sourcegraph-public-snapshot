// Package globals contains global variables that should be set by the frontend's main function on initialization.
package globals

import (
	"io/ioutil"
	"net/url"
	"os"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
)

// ExternalURL is the fully-resolved, externally accessible frontend URL.
var ExternalURL = &url.URL{Scheme: "http", Host: "example.com"}

// ConfigurationServerFrontendOnly provides the contents of the site configuration
// to other services and manages modifications to it.
//
// Any another service that attempts to use this variable will panic.
var ConfigurationServerFrontendOnly = conf.InitConfigurationServerFrontendOnly(configurationSource{
	configFilePath: os.Getenv("SOURCEGRAPH_CONFIG_FILE"),
})

type configurationSource struct {
	configFilePath string
}

func (c configurationSource) Read() (string, error) {
	data, err := ioutil.ReadFile(c.FilePath())
	if err != nil {
		return "", errors.Wrapf(err, "unable to read config file from %q", c.FilePath())
	}

	return string(data), err
}

func (c configurationSource) Write(input string) error {
	return ioutil.WriteFile(c.FilePath(), []byte(input), 0600)
}

func (c configurationSource) FilePath() string {
	filePath := c.configFilePath
	if filePath == "" {
		filePath = "/etc/sourcegraph/config.json"
	}

	return filePath
}
