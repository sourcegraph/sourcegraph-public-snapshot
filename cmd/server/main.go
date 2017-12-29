package main

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/server/internal/goreman"
)

//docker:install curl
//docker:run curl -o /usr/local/bin/syntect_server https://storage.googleapis.com/sourcegraph-artifacts/syntect_server/c04f24fcc8c0a04e476ecd1970445dba && chmod +x /usr/local/bin/syntect_server

// defaultEnv is environment variables that will be set if not already set.
var defaultEnv = map[string]string{
	// Sourcegraph services running in this container
	"SRC_GIT_SERVERS":         "127.0.0.1:3178",
	"SEARCHER_URL":            "http://127.0.0.1:3181",
	"SRC_SESSION_STORE_REDIS": "127.0.0.1:6379",
	"SRC_INDEXER":             "127.0.0.1:3179",
	"SRC_SYNTECT_SERVER":      "http://localhost:3700",
	"SRC_HTTP_ADDR":           ":7080",
	"SRC_FRONTEND_INTERNAL":   "127.0.0.1:3090",

	// We disable google analytics, etc
	"SRC_APP_DISABLE_SUPPORT_SERVICES": "true",

	// We adjust behaviour for on-prem vs prod
	"DEPLOYMENT_ON_PREM": "true",

	// Limit our cache size to 100GB, same as prod. We should probably update
	// searcher to ensure this value isn't larger than the volume for
	// CACHE_DIR.
	"SEARCHER_CACHE_SIZE_MB": "100000",

	// Enable our repo-updater to run every minute. Currently this is
	// only used to sync from gitolite.
	"REPO_LIST_UPDATE_INTERVAL": "1",

	// Env vars for higher rate limits to api.github.com
	"GITHUB_BASE_URL":      "http://localhost:3180",
	"GITHUB_CLIENT_ID":     "a359c6590ebe783800b1",
	"GITHUB_CLIENT_SECRET": "2f8c304d01ad3b23c1f6c0a3b42bd09a8694262f",

	// TODO other bits
	// * Guess SRC_APP_URL based on hostname
	// * SRC_LOG_LEVEL, DEBUG LOG_REQUESTS https://github.com/sourcegraph/sourcegraph/issues/8458
}

func main() {
	log.SetFlags(0)

	// Load $CONFIG_DIR/env before we set any defaults
	{
		configDir := setDefaultEnv("CONFIG_DIR", "/etc/sourcegraph")
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
		dataDir := setDefaultEnv("DATA_DIR", "/var/opt/sourcegraph")
		setDefaultEnv("SRC_REPOS_DIR", filepath.Join(dataDir, "repos"))
		setDefaultEnv("CACHE_DIR", filepath.Join(dataDir, "cache"))
	}

	// Special case some convenience environment variables
	if redis, ok := os.LookupEnv("REDIS"); ok {
		setDefaultEnv("REDIS_MASTER_ENDPOINT", redis)
		setDefaultEnv("SRC_SESSION_STORE_REDIS", redis)
	}

	for k, v := range defaultEnv {
		setDefaultEnv(k, v)
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

	// Should be kept in sync with package processrestart's goreman_server.go
	// process list.
	procfile := []string{
		`gitserver: gitserver`,
		`indexer: sh -c "sleep 5 && exec indexer"`, // Sleep to avoid migration race with frontend"
		`searcher: searcher`,
		`github-proxy: github-proxy`,
		`frontend: frontend`,
		`repo-updater: repo-updater`,
		`syntect_server: env QUIET=true ROCKET_LIMITS='{json=10485760}' ROCKET_PORT=3700 ROCKET_ADDRESS='"127.0.0.1"' ROCKET_ENV=production syntect_server`,
	}
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

	const goremanAddr = "localhost:5005"
	if err := os.Setenv("GOREMAN_RPC_ADDR", goremanAddr); err != nil {
		log.Fatal(err)
	}

	err := goreman.Start(goremanAddr, []byte(strings.Join(procfile, "\n")))
	if err != nil {
		log.Fatal(err)
	}
}

func mustCryptoRand() []byte {
	var b [80]byte
	if _, err := rand.Read(b[:]); err != nil {
		log.Fatalf("could not generate generate random value for SRC_APP_SECRET_KEY: %s", err)
	}
	buf := make([]byte, base64.StdEncoding.EncodedLen(len(b)))
	base64.StdEncoding.Encode(buf, b[:])
	return buf
}
