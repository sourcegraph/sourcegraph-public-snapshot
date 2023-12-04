// Package shared provides the entrypoint to Sourcegraph's single docker
// image. It has functionality to setup the shared environment variables, as
// well as create the Procfile for goreman to run.
package shared

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"golang.org/x/sync/errgroup"

	sglog "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/server/internal/goreman"
	"github.com/sourcegraph/sourcegraph/internal/database/postgresdsn"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/hostname"
	"github.com/sourcegraph/sourcegraph/internal/version"
)

// FrontendInternalHost is the value of SRC_FRONTEND_INTERNAL.
const FrontendInternalHost = "127.0.0.1:3090"

// DefaultEnv is environment variables that will be set if not already set.
//
// If it is modified by an external package, it must be modified immediately on startup,
// before `shared.Main` is called.
var DefaultEnv = map[string]string{
	// Sourcegraph services running in this container
	"SRC_GIT_SERVERS":       "127.0.0.1:3178",
	"SEARCHER_URL":          "http://127.0.0.1:3181",
	"REPO_UPDATER_URL":      "http://127.0.0.1:3182",
	"QUERY_RUNNER_URL":      "http://127.0.0.1:3183",
	"SRC_SYNTECT_SERVER":    "http://127.0.0.1:9238",
	"SYMBOLS_URL":           "http://127.0.0.1:3184",
	"SRC_HTTP_ADDR":         ":8080",
	"SRC_HTTPS_ADDR":        ":8443",
	"SRC_FRONTEND_INTERNAL": FrontendInternalHost,

	"GRAFANA_SERVER_URL":          "http://127.0.0.1:3370",
	"PROMETHEUS_URL":              "http://127.0.0.1:9090",
	"OTEL_EXPORTER_OTLP_ENDPOINT": "", // disabled

	// Limit our cache size to 100GB, same as prod. We should probably update
	// searcher/symbols to ensure this value isn't larger than the volume for
	// SYMBOLS_CACHE_DIR and SEARCHER_CACHE_DIR.
	"SEARCHER_CACHE_SIZE_MB": "50000",
	"SYMBOLS_CACHE_SIZE_MB":  "50000",

	// Used to differentiate between deployments on dev, Docker, and Kubernetes.
	"DEPLOY_TYPE": "docker-container",

	// enables the debug proxy (/-/debug)
	"SRC_PROF_HTTP": "",

	"LOGO": "t",

	// TODO other bits
	// * DEBUG LOG_REQUESTS https://github.com/sourcegraph/sourcegraph/issues/8458
}

// Set verbosity based on simple interpretation of env var to avoid external dependencies (such as
// on github.com/sourcegraph/sourcegraph/internal/env).
var verbose = os.Getenv("SRC_LOG_LEVEL") == "dbug" || os.Getenv("SRC_LOG_LEVEL") == "info"

// Main is the main server command function which is shared between Sourcegraph
// server's open-source and enterprise variant.
func Main() {
	flag.Parse()
	log.SetFlags(0)
	liblog := sglog.Init(sglog.Resource{
		Name:       env.MyName,
		Version:    version.Version(),
		InstanceID: hostname.Get(),
	})
	defer liblog.Sync()

	logger := sglog.Scoped("server")

	// Ensure CONFIG_DIR and DATA_DIR

	// Load $CONFIG_DIR/env before we set any defaults
	{
		configDir := SetDefaultEnv("CONFIG_DIR", "/etc/sourcegraph")
		err := os.MkdirAll(configDir, 0755)
		if err != nil {
			log.Fatalf("failed to ensure CONFIG_DIR exists: %s", err)
		}

		err = godotenv.Load(filepath.Join(configDir, "env"))
		if err != nil && !os.IsNotExist(err) {
			log.Fatalf("failed to load %s: %s", filepath.Join(configDir, "env"), err)
		}
	}

	// Next persistence
	{
		SetDefaultEnv("SRC_REPOS_DIR", filepath.Join(DataDir, "repos"))
		cacheDir := filepath.Join(DataDir, "cache")
		SetDefaultEnv("SYMBOLS_CACHE_DIR", cacheDir)
		SetDefaultEnv("SEARCHER_CACHE_DIR", cacheDir)
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

	for k, v := range DefaultEnv {
		SetDefaultEnv(k, v)
	}

	if v, _ := strconv.ParseBool(os.Getenv("ALLOW_SINGLE_DOCKER_CODE_INSIGHTS")); v {
		AllowSingleDockerCodeInsights = true
	}

	// Now we put things in the right place on the FS
	if err := copySSH(); err != nil {
		// TODO There are likely several cases where we don't need SSH
		// working, we shouldn't prevent setup in those cases. The main one
		// that comes to mind is an ORIGIN_MAP which creates https clone URLs.
		log.Println("Failed to setup SSH authorization:", err)
		log.Fatal("SSH authorization required for cloning from your codehost. Please see README.")
	}
	if err := copyConfigs(); err != nil {
		log.Fatal("Failed to copy configs:", err)
	}

	// TODO validate known_hosts contains all code hosts in config.

	nginx, err := nginxProcFile()
	if err != nil {
		log.Fatal("Failed to setup nginx:", err)
	}

	postgresExporterLine := fmt.Sprintf(`postgres_exporter: env DATA_SOURCE_NAME="%s" postgres_exporter --config.file="/postgres_exporter.yaml" --log.level=%s`, postgresdsn.New("", "postgres", os.Getenv), convertLogLevel(os.Getenv("SRC_LOG_LEVEL")))

	// TODO: This should be fixed properly.
	// Tell `gitserver` that its `hostname` is what the others think of as gitserver hostnames.
	gitserverLine := fmt.Sprintf(`gitserver: env HOSTNAME=%q gitserver`, os.Getenv("SRC_GIT_SERVERS"))

	procfile := []string{
		nginx,
		`frontend: env CONFIGURATION_MODE=server frontend`,
		gitserverLine,
		`symbols: symbols`,
		`searcher: searcher`,
		`worker: worker`,
		`repo-updater: repo-updater`,
		`precise-code-intel-worker: precise-code-intel-worker`,
		`syntax_highlighter: sh -c 'env QUIET=true ROCKET_ENV=production ROCKET_PORT=9238 ROCKET_LIMITS='"'"'{json=10485760}'"'"' ROCKET_SECRET_KEY='"'"'SeerutKeyIsI7releuantAndknvsuZPluaseIgnorYA='"'"' ROCKET_KEEP_ALIVE=0 ROCKET_ADDRESS='"'"'"127.0.0.1"'"'"' syntax_highlighter | grep -v "Rocket has launched" | grep -v "Warning: environment is"' | grep -v 'Configured for production'`,
		postgresExporterLine,
	}
	procfile = append(procfile, ProcfileAdditions...)

	if monitoringLines := maybeObservability(); len(monitoringLines) != 0 {
		procfile = append(procfile, monitoringLines...)
	}

	if blobstoreLines := maybeBlobstore(logger); len(blobstoreLines) != 0 {
		procfile = append(procfile, blobstoreLines...)
	}

	redisStoreLine, err := maybeRedisStoreProcFile()
	if err != nil {
		log.Fatal(err)
	}
	if redisStoreLine != "" {
		procfile = append(procfile, redisStoreLine)
	}
	redisCacheLine, err := maybeRedisCacheProcFile()
	if err != nil {
		log.Fatal(err)
	}
	if redisCacheLine != "" {
		procfile = append(procfile, redisCacheLine)
	}

	procfile = append(procfile, maybeZoektProcFile()...)

	var (
		postgresProcfile []string
		restore, _       = strconv.ParseBool(os.Getenv("PGRESTORE"))
	)

	postgresLine, err := maybePostgresProcFile()
	if err != nil {
		log.Fatal(err)
	}
	if postgresLine != "" {
		if restore {
			// If in restore mode, only run PostgreSQL
			procfile = []string{postgresLine}
		} else {
			postgresProcfile = append(postgresProcfile, postgresLine)
		}
	} else if restore {
		log.Fatal("PGRESTORE is set but a local Postgres instance is not configured")
	}

	// Shutdown if any process dies
	procDiedAction := goreman.Shutdown
	if ignore, _ := strconv.ParseBool(os.Getenv("IGNORE_PROCESS_DEATH")); ignore {
		// IGNORE_PROCESS_DEATH is an escape hatch so that sourcegraph/server
		// keeps running in the case of a subprocess dying on startup. An
		// example use case is connecting to postgres even though frontend is
		// dying due to a bad migration.
		procDiedAction = goreman.Ignore
	}

	runMigrations := !restore
	run(procfile, postgresProcfile, runMigrations, procDiedAction)
}

func run(procfile, postgresProcfile []string, runMigrations bool, procDiedAction goreman.ProcDiedAction) {
	if !runMigrations {
		procfile = append(procfile, postgresProcfile...)
		postgresProcfile = nil
	}

	group, _ := errgroup.WithContext(context.Background())

	// Check whether database reindex is required, and run it if so
	if shouldPostgresReindex() {
		runPostgresReindex()
	}

	options := goreman.Options{
		RPCAddr:        "127.0.0.1:5005",
		ProcDiedAction: procDiedAction,
	}
	startProcesses(group, "postgres", postgresProcfile, options)

	if runMigrations {
		// Run migrations before starting up the application but after
		// starting any postgres instance within the server container.
		runMigrator()
	}

	startProcesses(group, "all", procfile, options)

	if err := group.Wait(); err != nil {
		log.Fatal(err)
	}
}

func startProcesses(group *errgroup.Group, name string, procfile []string, options goreman.Options) {
	if len(procfile) == 0 {
		return
	}

	log.Printf("Starting %s processes", name)
	group.Go(func() error { return goreman.Start([]byte(strings.Join(procfile, "\n")), options) })
}

func runMigrator() {
	log.Println("Starting migrator")

	schemas := []string{"frontend", "codeintel"}
	if AllowSingleDockerCodeInsights {
		schemas = append(schemas, "codeinsights")
	}

	for _, schemaName := range schemas {
		e := execer{}
		e.Command("migrator", "up", "-db", schemaName)

		if err := e.Error(); err != nil {
			pgPrintf("Migrating %s schema failed: %s", schemaName, err)
			log.Fatal(err.Error())
		}
	}

	log.Println("Migrated postgres schemas.")
}

func shouldPostgresReindex() (shouldReindex bool) {
	fmt.Printf("Checking whether a Postgres reindex is required...\n")

	// Check for presence of the reindex marker file
	postgresReindexMarkerFile := postgresReindexMarkerFile()
	_, err := os.Stat(postgresReindexMarkerFile)
	if err == nil {
		fmt.Printf("5.1 reindex marker file '%s' found\n", postgresReindexMarkerFile)
		return false
	}
	fmt.Printf("5.1 reindex marker file '%s' not found\n", postgresReindexMarkerFile)

	// Check PGHOST variable to see whether it refers to a local address or path
	// If an external database is used, reindexing can be skipped
	pgHost := os.Getenv("PGHOST")
	if !(pgHost == "" || pgHost == "127.0.0.1" || pgHost == "localhost" || string(pgHost[0]) == "/") {
		fmt.Printf("Using a non-local Postgres database '%s', reindexing not required\n", pgHost)
		return false
	}
	fmt.Printf("Using a local Postgres database '%s', reindexing required\n", pgHost)

	return true
}

func runPostgresReindex() {
	fmt.Printf("Starting Postgres reindex process\n")

	performMigration := os.Getenv("SOURCEGRAPH_5_1_DB_MIGRATION")
	if performMigration != "true" {
		fmt.Printf("\n**************** MIGRATION REQUIRED **************\n\n")
		fmt.Printf("Upgrading to Sourcegraph 5.1 or later from an earlier release requires a database reindex.\n\n")
		fmt.Printf("This process may take several hours, depending on the size of your database.\n\n")
		fmt.Printf("If you do not wish to perform the reindex process now, you should switch back to a release before Sourcegraph 5.1.\n\n")
		fmt.Printf("To perform the reindexing process now, please review the instructions at https://docs.sourcegraph.com/admin/migration/5_1 and restart the container with the environment variable `SOURCEGRAPH_5_1_DB_MIGRATION=true` set.\n")
		fmt.Printf("\n**************** MIGRATION REQUIRED **************\n\n")

		os.Exit(101)
	}

	cmd := exec.Command("/bin/bash", "/reindex.sh")
	cmd.Env = append(
		os.Environ(),
		fmt.Sprintf("REINDEX_COMPLETED_FILE=%s", postgresReindexMarkerFile()),
		// PGDATA is set as an ENVAR in standalone container
		fmt.Sprintf("PGDATA=%s", postgresDataPath()),
		// Unset PGHOST so connections go over unix socket; we've already confirmed the database being reindexed is local
		"PGHOST=",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error running database migration: %s\n", err)
		os.Exit(1)
	}
}
