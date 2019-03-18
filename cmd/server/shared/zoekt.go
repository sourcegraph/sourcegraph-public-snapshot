package shared

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

const zoektHost = "127.0.0.1:3070"

func maybeZoektProcfile(dataDir string) ([]string, error) {
	if !zoektEnabled() {
		// Explicitly disable zoekt
		return nil, os.Setenv("ZOEKT_HOST", "")
	}

	SetDefaultEnv("ZOEKT_HOST", zoektHost)
	indexDir := filepath.Join(dataDir, "zoekt/index")
	return []string{
		fmt.Sprintf("zoekt-indexserver: zoekt-sourcegraph-indexserver -sourcegraph_url http://%s -index %s -interval 1m -listen 127.0.0.1:6072", FrontendInternalHost, indexDir),
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
