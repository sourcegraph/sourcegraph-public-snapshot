package shared

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"

	"github.com/sourcegraph/sourcegraph/cmd/server/internal/goreman"
)

// FrontendInternalHost is the value of SRC_FRONTEND_INTERNAL.
const FrontendInternalHost = "127.0.0.1:3090"

// zoektHost is the address zoekt webserver is listening on.
const zoektHost = "127.0.0.1:3070"

// defaultEnv is environment variables that will be set if not already set.
var defaultEnv = map[string]string{
	// Sourcegraph services running in this container
	"SRC_GIT_SERVERS":       "127.0.0.1:3178",
	"SEARCHER_URL":          "http://127.0.0.1:3181",
	"REPO_UPDATER_URL":      "http://127.0.0.1:3182",
	"SRC_INDEXER":           "127.0.0.1:3179",
	"QUERY_RUNNER_URL":      "http://127.0.0.1:3183",
	"SRC_SYNTECT_SERVER":    "http://127.0.0.1:9238",
	"SYMBOLS_URL":           "http://127.0.0.1:3184",
	"SRC_HTTP_ADDR":         ":7080",
	"SRC_HTTPS_ADDR":        ":7443",
	"SRC_FRONTEND_INTERNAL": FrontendInternalHost,
	"GITHUB_BASE_URL":       "http://127.0.0.1:3180", // points to github-proxy
	"LSP_PROXY":             "127.0.0.1:4388",
	"ZOEKT_HOST":            zoektHost,

	// Limit our cache size to 100GB, same as prod. We should probably update
	// searcher/symbols to ensure this value isn't larger than the volume for
	// CACHE_DIR.
	"SEARCHER_CACHE_SIZE_MB": "50000",
	"SYMBOLS_CACHE_SIZE_MB":  "50000",

	// Used to differentiate between deployments on dev, Docker, and Kubernetes.
	"DEPLOY_TYPE": "docker-container",

	// enables the debug proxy (/-/debug)
	"SRC_PROF_HTTP": "",

	"LOGO":          "t",
	"SRC_LOG_LEVEL": "warn",

	// TODO other bits
	// * Guess SRC_APP_URL based on hostname
	// * DEBUG LOG_REQUESTS https://github.com/sourcegraph/sourcegraph/issues/8458
}

// Set verbosity based on simple interpretation of env var to avoid external dependencies (such as
// on github.com/sourcegraph/sourcegraph/pkg/env).
var verbose = os.Getenv("SRC_LOG_LEVEL") == "dbug" || os.Getenv("SRC_LOG_LEVEL") == "info"

// Main is the main server command function which is shared between Sourcegraph
// server's open-source and enterprise variant.
func Main() {
	log.SetFlags(0)

	// Load $CONFIG_DIR/env before we set any defaults
	{
		configDir := SetDefaultEnv("CONFIG_DIR", "/etc/sourcegraph")
		err := godotenv.Load(filepath.Join(configDir, "env"))
		if err != nil && !os.IsNotExist(err) {
			log.Fatalf("failed to load %s: %s", filepath.Join(configDir, "env"), err)
		}

		// Load the config file, or generate a new one if it doesn't exist.
		configPath := os.Getenv("SOURCEGRAPH_CONFIG_FILE")
		if configPath == "" {
			configPath = filepath.Join(configDir, "sourcegraph-config.json")
		}
		_, configIsWritable, err := readOrGenerateConfig(configPath)
		if err != nil {
			log.Fatalf("failed to load config: %s", err)
		}
		if configIsWritable {
			if err := os.Setenv("SOURCEGRAPH_CONFIG_FILE", configPath); err != nil {
				log.Fatal(err)
			}
		}
	}

	// Next persistence
	{
		SetDefaultEnv("SRC_REPOS_DIR", filepath.Join(DataDir, "repos"))
		SetDefaultEnv("CACHE_DIR", filepath.Join(DataDir, "cache"))
	}

	// Special case some convenience environment variables
	if redis, ok := os.LookupEnv("REDIS"); ok {
		SetDefaultEnv("REDIS_ENDPOINT", redis)
	}

	data, err := json.MarshalIndent(SrcProfServices, "", "  ")
	if err != nil {
		log.Println("Failed to marshal default SRC_PROF_SERVICES")
	} else {
		SetDefaultEnv("SRC_PROF_SERVICES", string(data))
	}

	for k, v := range defaultEnv {
		SetDefaultEnv(k, v)
	}

	// Now we put things in the right place on the FS
	if err := copySSH(); err != nil {
		// TODO There are likely several cases where we don't need SSH
		// working, we shouldn't prevent setup in those cases. The main one
		// that comes to mind is an ORIGIN_MAP which creates https clone URLs.
		log.Println("Failed to setup SSH authorization:", err)
		log.Fatal("SSH authorization required for cloning from your codehost. Please see README.")
	}
	if err := copyNetrc(); err != nil {
		log.Fatal("Failed to copy netrc:", err)
	}

	// TODO validate known_hosts contains all code hosts in config.

	zoektIndexDir := filepath.Join(DataDir, "zoekt/index")

	procfile := []string{
		`gitserver: gitserver`,
		`query-runner: query-runner`,
		`symbols: symbols`,
		`lsp-proxy: lsp-proxy`,
		`searcher: searcher`,
		`github-proxy: github-proxy`,
		`frontend: frontend`,
		`repo-updater: repo-updater`,
		`indexer: indexer`,
		`syntect_server: sh -c 'env QUIET=true ROCKET_LIMITS='"'"'{json=10485760}'"'"' ROCKET_PORT=9238 ROCKET_ADDRESS='"'"'"127.0.0.1"'"'"' ROCKET_ENV=production syntect_server | grep -v "Rocket has launched" | grep -v "Warning: environment is"'`,
		fmt.Sprintf("zoekt-indexserver: zoekt-sourcegraph-indexserver -sourcegraph_url http://%s -index %s -interval 1m -listen 127.0.0.1:6072", FrontendInternalHost, zoektIndexDir),
		fmt.Sprintf("zoekt-webserver: zoekt-webserver -rpc -pprof -listen %s -index %s", zoektHost, zoektIndexDir),
	}
	procfile = append(procfile, ProcfileAdditions...)
	if line, err := maybeRedisProcFile(); err != nil {
		log.Fatal(err)
	} else if line != "" {
		procfile = append(procfile, line)
	}
	if line, err := maybePostgresProcFile(); err != nil {
		log.Fatal(err)
	} else if line != "" {
		procfile = append(procfile, line)
	}

	const goremanAddr = "127.0.0.1:5005"
	if err := os.Setenv("GOREMAN_RPC_ADDR", goremanAddr); err != nil {
		log.Fatal(err)
	}

	err = goreman.Start(goremanAddr, []byte(strings.Join(procfile, "\n")))
	if err != nil {
		log.Fatal(err)
	}
}
