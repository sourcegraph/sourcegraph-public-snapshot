package shared

import (
	"fmt"
	"os"
	"path/filepath"
)

func maybeZoektProcFile() []string {
	// Zoekt is alreay configured
	if os.Getenv("ZOEKT_HOST") != "" {
		return nil
	}
	if os.Getenv("INDEXED_SEARCH_SERVERS") != "" {
		return nil
	}

	defaultHost := "127.0.0.1:3070"
	SetDefaultEnv("INDEXED_SEARCH_SERVERS", defaultHost)

	frontendInternalHost := os.Getenv("SRC_FRONTEND_INTERNAL")
	indexDir := filepath.Join(DataDir, "zoekt/index")

	return []string{
		fmt.Sprintf("zoekt-indexserver: env GOGC=25 HOSTNAME=%s zoekt-sourcegraph-indexserver -sourcegraph_url http://%s -index %s -interval 1m -listen 127.0.0.1:6072 -cpu_fraction 0.25", defaultHost, frontendInternalHost, indexDir),
		fmt.Sprintf("zoekt-webserver: env GOGC=25 zoekt-webserver -rpc -pprof -indexserver_proxy -listen %s -index %s", defaultHost, indexDir),
	}
}
