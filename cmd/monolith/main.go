package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/monolith/internal/goreman"
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
	"SRC_HTTP_ADDR":           ":80",
	"SRC_FRONTEND_INTERNAL":   "127.0.0.1:3090",

	// We disable google analytics, etc
	"SRC_APP_DISABLE_SUPPORT_SERVICES": "true",

	// We adjust behaviour for on-prem vs prod
	"DEPLOYMENT_ON_PREM": "true",

	// TODO environment variables we need to support
	// CACHE_DIR
	// GITHUB_CONFIG
	// GITHUB_PERSONAL_ACCESS_TOKEN   deprecated??
	// GITOLITE_HOSTS
	// ORIGIN_MAP
	// SEARCHER_CACHE_SIZE_MB

	// TODO this is true in sourcegraph-server-gen, but seems uneeded in
	// practice? It causes requests to be logged in github-proxy
	"LOG_REQUESTS": "true",

	// Enable our repo-list-updater to run every minute. Currently this is
	// only used to sync from gitolite.
	"REPO_LIST_UPDATE_INTERVAL": "1",

	// TODO wizard which may avoid setting this. This is setup to auto-clone
	// github using our github dev secrets.
	//"GITHUB_BASE_URL":       "http://127.0.0.1:3180",
	//"GITHUB_CLIENT_ID":      "6f2a43bd8877ff5fd1d5",
	//"GITHUB_CLIENT_SECRET":  "c5ff37d80e3736924cbbdf2922a50cac31963e43",
	//"PUBLIC_REPO_REDIRECTS": "false",
	"AUTO_REPO_ADD": "true", // false in server-gen, but until we have a nice way to setup repo cloning this is best

	// TODO other bits
	// * should we guess SRC_APP_URL?
	// * Do we need BI_LOGGER?
	// * SRC_LOG_LEVEL and DEBUG
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

		// As a convenience some environment variables can stored as a file
		envFiles := map[string]string{
			"license.sgl": "LICENSE_KEY",
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

		// Set config
		if b, err := ioutil.ReadFile(filepath.Join(configDir, "config.json")); err == nil {
			setDefaultEnvFromConfig(string(b))
		} else if !os.IsNotExist(err) {
			log.Fatal("failed to read contents of `config.json`")
		}
	}

	// Next persistence
	{
		dataDir := setDefaultEnv("DATA_DIR", "/var/opt/sourcegraph")
		setDefaultEnv("SRC_REPOS_DIR", filepath.Join(dataDir, "repos"))
	}

	// Special case some convenience environment variables
	if redis, ok := os.LookupEnv("REDIS"); ok {
		setDefaultEnv("REDIS_MASTER_ENDPOINT", redis)
		setDefaultEnv("SRC_SESSION_STORE_REDIS", redis)
	}
	// TODO is this an alright idea for skipping this bit of config?
	setDefaultEnv("SRC_APP_SECRET_KEY", os.Getenv("LICENSE_KEY"))

	for k, v := range defaultEnv {
		setDefaultEnv(k, v)
	}

	// More convenient to fail now than when the page is loaded if the license
	// is missing
	if _, ok := os.LookupEnv("LICENSE_KEY"); !ok {
		log.Fatal("Please set the environment variable LICENSE_KEY. Please contact sales@sourcegraph.com to obtain a license.")
	}

	// TODO we are requiring origin map, but if you have github or ghe setup
	// you don't need it.
	if originMap := os.Getenv("ORIGIN_MAP"); originMap == "" {
		log.Println("Please set the environment variable ORIGIN_MAP.")
		log.Println("ORIGIN_MAP tells sourcegraph how to map repo names into cloneable URLs.")
		log.Printf("The format is prefix!cloneURL with a special '%%' denoting the suffix. Example:")
		// TODO (give example URLs they are accessible from)
		log.Println()
		log.Printf(" gitlab.mycompany.com/!git@gitlab.mycompany.com:%%.git")
		log.Fatal()
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
