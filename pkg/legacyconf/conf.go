package legacyconf

import (
	"io/ioutil"
	"log"
	"os"
	"sync"
)

var (
	configFilePath string

	rawMu sync.RWMutex
	raw   string
)

// Raw returns the raw site configuration JSON.
func Raw() string {
	rawMu.RLock()
	defer rawMu.RUnlock()
	return raw
}

func init() {
	configFilePath = os.Getenv("SOURCEGRAPH_CONFIG_FILE")
	if configFilePath == "" {
		return
	}

	// Read configuration initially.
	if err := initConfig(); err != nil {
		log.Fatalf("error reading SOURCEGRAPH_CONFIG_FILE: %s (%s)", err, configFilePath)
	}
}

func readConfig() (string, error) {
	if configFilePath == "" {
		return "", nil
	}
	data, err := ioutil.ReadFile(configFilePath)
	return string(data), err
}

/*
// ParseConfigData reads the provided config string, but NOT the environment
func ParseConfigData(data string) (*schema.SiteConfiguration, error) {
	var tmpConfig schema.SiteConfiguration

	if data != "" {
		data, err := jsonc.Parse(data)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(data, &tmpConfig); err != nil {
			fmt.Println("HERE2", err)
			return nil, err
		}
	}

	// For convenience, make sure this is not nil.
	if tmpConfig.ExperimentalFeatures == nil {
		tmpConfig.ExperimentalFeatures = &schema.ExperimentalFeatures{}
	}
	return &tmpConfig, nil
}

// parseConfig reads the provided string, then merges in additional
// data from the (deprecated) environment.
func parseConfig(data string) (*schema.SiteConfiguration, error) {
	tmpConfig, err := ParseConfigData(data)
	if err != nil {
		return nil, err
	}

	// Env var config takes highest precedence but is deprecated.
	if v, envVarNames, err := configFromEnv(); err != nil {
		return nil, err
	} else if len(envVarNames) > 0 {
		// We don't care about any errors unmarshalling.
		_ = json.Unmarshal(v, tmpConfig)
	}
	return tmpConfig, nil
}
*/

func initConfig() error {
	rawConfig, err := readConfig()
	if err != nil {
		return err
	}

	/*
		cfg, err := parseConfig(rawConfig)
		if err != nil {
			return err
		}
		_ = cfg
	*/

	/*if err := jsonc.Unmarshal(oldConfig, &config); err != nil {
		return nil, err
	}*/

	rawMu.Lock()
	raw = rawConfig
	rawMu.Unlock()
	return nil
}
