package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"sourcegraph.com/sourcegraph/sourcegraph/schema"
)

// readOrGenerateConfig supports the following cases:
//
// - There is an existing config file.
// - There is a SOURCEGRAPH_CONFIG env var.
// - There is neither a config file nor a SOURCEGRAPH_CONFIG env var.
//
// It returns the raw config JSON (which can have comments and trailing commas)
// and whether the config is writable. Iff the SOURCEGRAPH_CONFIG env var is used,
// the config is not writable.
//
// Any other case results in an error (such as having both a config file
// and a SOURCEGRAPH_CONFIG env var). This is to avoid needing to merge configs
// from multiple sources, which is complex because it's not clear which should
// take precedence and how to edit it.
func readOrGenerateConfig(path string) (configJSON string, writable bool, err error) {
	fileData, err := ioutil.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return "", false, err
	}
	hasFile := err == nil

	envData, hasEnvVar := os.LookupEnv("SOURCEGRAPH_CONFIG")

	if hasFile && hasEnvVar {
		return "", false, fmt.Errorf("multiple configuration sources are not allowed; use only one of SOURCEGRAPH_CONFIG and %s", path)
	}

	if hasEnvVar {
		configJSON = envData
	} else if hasFile {
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
		AutoRepoAdd:     true,
		SecretKey:       string(mustCryptoRand()),
		AuthAllowSignup: true,
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
