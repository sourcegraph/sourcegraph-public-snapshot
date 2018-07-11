package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

func maybeZoektProcfile(dataDir string) ([]string, error) {
	if !zoektEnabled() {
		// Explicitly disable zoekt
		return nil, os.Setenv("ZOEKT_HOST", "")
	}

	setDefaultEnv("ZOEKT_HOST", zoektHost)
	indexDir := filepath.Join(dataDir, "zoekt/index")
	return []string{
		fmt.Sprintf("zoekt-indexserver: zoekt-sourcegraph-indexserver -sourcegraph_url http://%s -index %s -interval 1m", frontendInternalHost, indexDir),
		fmt.Sprintf("zoekt-webserver: zoekt-webserver -rpc -pprof -listen %s -index %s", zoektHost, indexDir),
	}, nil
}

func zoektEnabled() bool {
	env := os.Getenv("INDEXED_SEARCH")
	if env == "" {
		return false
	}
	enabled, err := strconv.ParseBool(env)
	if err != nil {
		log.Fatal(err)
	}
	return enabled
}
