package main

import (
	"crypto/rand"
	"encoding/base64"
	"io/ioutil"
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

	// TODO environment variables we need to support related to codehosts
	// GITHUB_CONFIG
	// GITOLITE_HOSTS
	// ORIGIN_MAP
	// PUBLIC_REPO_REDIRECTS
	"AUTO_REPO_ADD": "true", // false in server-gen, but until we have a nice way to setup repo cloning this is best

	// Limit our cache size to 100GB, same as prod. We should probably update
	// searcher to ensure this value isn't larger than the volume for
	// CACHE_DIR.
	"SEARCHER_CACHE_SIZE_MB": "100000",

	// Enable our repo-list-updater to run every minute. Currently this is
	// only used to sync from gitolite.
	"REPO_LIST_UPDATE_INTERVAL": "1",

	// We don't want to require users to have a license. So we use the magic
	// license which bypasses license checks.
	"LICENSE_KEY": "24348deeb9916a070914b5617a9a4e2c7bec0d313ca6ae11545ef034c7138d4d8710cddac80980b00426fb44830263268f028c9735",

	// Env vars for higher rate limits to api.github.com
	"GITHUB_BASE_URL":      "http://localhost:3180",
	"GITHUB_CLIENT_ID":     "a359c6590ebe783800b1",
	"GITHUB_CLIENT_SECRET": "2f8c304d01ad3b23c1f6c0a3b42bd09a8694262f",

	"MANDRILL_KEY": "HUBvkcrqJAJ5ufBK52nEJQ",
	"USE_AUTH0":    "false",

	// TODO other bits
	// * Guess SRC_APP_URL based on hostname
	// * SRC_LOG_LEVEL, DEBUG LOG_REQUESTS https://github.com/sourcegraph/sourcegraph/issues/8458
	// * TRACKING_APP_ID can be guessed from LICENSE_KEY https://github.com/sourcegraph/sourcegraph/issues/8377
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

		// As a convenience some environment variables can be stored as a file
		envFiles := map[string]string{
			"license.sgl": "LICENSE_KEY",
			"config.json": "SOURCEGRAPH_CONFIG",
		}
		for name, key := range envFiles {
			b, err := ioutil.ReadFile(filepath.Join(configDir, name))
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				log.Fatalf("could not read file %q into environment variable %s: %s", name, key, err)
			}
			setDefaultEnv(key, strings.TrimSpace(string(b)))
		}

		// Convert SOURCEGRAPH_CONFIG into env vars
		if config, ok := os.LookupEnv("SOURCEGRAPH_CONFIG"); ok {
			setDefaultEnvFromConfig(config)
		}

		if _, ok := os.LookupEnv("SRC_APP_SECRET_KEY"); !ok {
			appSecretKeyFile := filepath.Join(configDir, "srcAppSecretKey")
			appSecretKey, err := ioutil.ReadFile(appSecretKeyFile)
			if os.IsNotExist(err) {
				appSecretKey := mustCryptoRand()
				if err := os.MkdirAll(configDir, os.FileMode(0755)); err != nil {
					log.Fatalf("could not create config directory %s: %s", configDir, err)
				}
				if err := ioutil.WriteFile(appSecretKeyFile, appSecretKey, 0644); err != nil {
					log.Fatalf("could not write secret key file: %s", err)
				}
			} else if err != nil {
				log.Fatalf("could not read app secret key file: %s", err)
			}
			setDefaultEnv("SRC_APP_SECRET_KEY", string(appSecretKey))
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
	// TODO Most users are using the same LICENSE_KEY, so we need to use
	// something that can't be guessed across installations (crypto random
	// number we save?)
	setDefaultEnv("SRC_APP_SECRET_KEY", os.Getenv("LICENSE_KEY"))

	for k, v := range defaultEnv {
		setDefaultEnv(k, v)
	}

	// More convenient to fail now than when the page is loaded if the license
	// is missing
	if _, ok := os.LookupEnv("LICENSE_KEY"); !ok {
		log.Fatal("Please set the environment variable LICENSE_KEY. Please contact sales@sourcegraph.com to obtain a license.")
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

	procfile := []string{
		`gitserver: gitserver`,
		`indexer: sh -c "sleep 5 && exec indexer"`, // Sleep to avoid migration race with frontend"
		`searcher: searcher`,
		`github-proxy: github-proxy`,
		`frontend: frontend`,
		`repo-list-updater: repo-list-updater`,
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

	err := goreman.Start([]byte(strings.Join(procfile, "\n")))
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
