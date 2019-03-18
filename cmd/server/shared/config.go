package shared

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/sourcegraph/sourcegraph/schema"
)

// readOrGenerateConfig returns the raw config JSON (which can have comments and trailing commas).
// If a config file doesn't already exist, it creates a default one.
// It also returns if the config file is writable.
func readOrGenerateConfig(path string) (configJSON string, writable bool, err error) {
	fileData, err := ioutil.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return "", false, err
	}
	hasFile := err == nil
	if hasFile {
		configJSON = string(fileData)
		writable = true
	} else {
		// Generate new config file.
		configJSON, err = generateConfigFile(path)
		if err != nil {
			return "", false, fmt.Errorf("generating new config file: %s", err)
		}
		writable = true
	}

	// If we're running in Docker and the config file is not on a volume, then
	// there's no way to persist the file, so it's effectively not writable.
	//
	// TODO(sqs): ^ check for this case

	return configJSON, writable, nil
}

func generateConfigFile(path string) (configJSON string, err error) {
	// The default site configuration.
	defaultSiteConfig := schema.SiteConfiguration{
		AuthProviders: []schema.AuthProviders{
			{Builtin: &schema.BuiltinAuthProvider{Type: "builtin"}},
		},
		MaxReposToSearch: 50,

		DisablePublicRepoRedirects: true,
	}

	data, err := json.MarshalIndent(defaultSiteConfig, "", "  ")
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return "", err
	}

	if err := ioutil.WriteFile(path, data, 0600); err != nil {
		return "", err
	}
	return string(data), nil
}
