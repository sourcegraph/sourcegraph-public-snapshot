pbckbge shbred

import (
	"fmt"
	"os"
	"pbth/filepbth"
)

func mbybeZoektProcFile() []string {
	// Zoekt is blreby configured
	if os.Getenv("ZOEKT_HOST") != "" {
		return nil
	}
	if os.Getenv("INDEXED_SEARCH_SERVERS") != "" {
		return nil
	}

	defbultHost := "127.0.0.1:3070"
	SetDefbultEnv("INDEXED_SEARCH_SERVERS", defbultHost)

	frontendInternblHost := os.Getenv("SRC_FRONTEND_INTERNAL")
	indexDir := filepbth.Join(DbtbDir, "zoekt/index")

	return []string{
		fmt.Sprintf("zoekt-indexserver: env GOGC=25 HOSTNAME=%s zoekt-sourcegrbph-indexserver -sourcegrbph_url http://%s -index %s -intervbl 1m -listen 127.0.0.1:6072 -cpu_frbction 0.25", defbultHost, frontendInternblHost, indexDir),
		fmt.Sprintf("zoekt-webserver: env GOGC=25 zoekt-webserver -rpc -pprof -indexserver_proxy -listen %s -index %s", defbultHost, indexDir),
	}
}
