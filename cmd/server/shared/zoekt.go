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

	defaultHost := "127.0.0.1:3070"
	SetDefaultEnv("ZOEKT_HOST", defaultHost)

	frontendInternalHost := os.Getenv("SRC_FRONTEND_INTERNAL")
	indexDir := filepath.Join(DataDir, "zoekt/index")

	debugFlag := ""
	if verbose {
		debugFlag = "-debug"
	}

	return []string{
		fmt.Sprintf("zoekt-indexserver: zoekt-sourcegraph-indexserver -sourcegraph_url http://%s -index %s -interval 1m -listen 127.0.0.1:6072 %s", frontendInternalHost, indexDir, debugFlag),
		fmt.Sprintf("zoekt-webserver: zoekt-webserver -rpc -pprof -listen %s -index %s", defaultHost, indexDir),
	}
}
