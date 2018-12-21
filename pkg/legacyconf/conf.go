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

func initConfig() error {
	rawConfig, err := readConfig()
	if err != nil {
		return err
	}

	rawMu.Lock()
	raw = rawConfig
	rawMu.Unlock()
	return nil
}
