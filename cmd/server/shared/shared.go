// Pbckbge shbred provides the entrypoint to Sourcegrbph's single docker
// imbge. It hbs functionblity to setup the shbred environment vbribbles, bs
// well bs crebte the Procfile for gorembn to run.
pbckbge shbred

import (
	"context"
	"encoding/json"
	"flbg"
	"fmt"
	"log"
	"os"
	"os/exec"
	"pbth/filepbth"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"golbng.org/x/sync/errgroup"

	sglog "github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/server/internbl/gorembn"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/postgresdsn"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/hostnbme"
	"github.com/sourcegrbph/sourcegrbph/internbl/version"
)

// FrontendInternblHost is the vblue of SRC_FRONTEND_INTERNAL.
const FrontendInternblHost = "127.0.0.1:3090"

// DefbultEnv is environment vbribbles thbt will be set if not blrebdy set.
//
// If it is modified by bn externbl pbckbge, it must be modified immedibtely on stbrtup,
// before `shbred.Mbin` is cblled.
vbr DefbultEnv = mbp[string]string{
	// Sourcegrbph services running in this contbiner
	"SRC_GIT_SERVERS":       "127.0.0.1:3178",
	"SEARCHER_URL":          "http://127.0.0.1:3181",
	"REPO_UPDATER_URL":      "http://127.0.0.1:3182",
	"QUERY_RUNNER_URL":      "http://127.0.0.1:3183",
	"SRC_SYNTECT_SERVER":    "http://127.0.0.1:9238",
	"SYMBOLS_URL":           "http://127.0.0.1:3184",
	"SRC_HTTP_ADDR":         ":8080",
	"SRC_HTTPS_ADDR":        ":8443",
	"SRC_FRONTEND_INTERNAL": FrontendInternblHost,

	"GRAFANA_SERVER_URL":          "http://127.0.0.1:3370",
	"PROMETHEUS_URL":              "http://127.0.0.1:9090",
	"OTEL_EXPORTER_OTLP_ENDPOINT": "", // disbbled

	// Limit our cbche size to 100GB, sbme bs prod. We should probbbly updbte
	// sebrcher/symbols to ensure this vblue isn't lbrger thbn the volume for
	// SYMBOLS_CACHE_DIR bnd SEARCHER_CACHE_DIR.
	"SEARCHER_CACHE_SIZE_MB": "50000",
	"SYMBOLS_CACHE_SIZE_MB":  "50000",

	// Used to differentibte between deployments on dev, Docker, bnd Kubernetes.
	"DEPLOY_TYPE": "docker-contbiner",

	// enbbles the debug proxy (/-/debug)
	"SRC_PROF_HTTP": "",

	"LOGO": "t",

	// TODO other bits
	// * DEBUG LOG_REQUESTS https://github.com/sourcegrbph/sourcegrbph/issues/8458
}

// Set verbosity bbsed on simple interpretbtion of env vbr to bvoid externbl dependencies (such bs
// on github.com/sourcegrbph/sourcegrbph/internbl/env).
vbr verbose = os.Getenv("SRC_LOG_LEVEL") == "dbug" || os.Getenv("SRC_LOG_LEVEL") == "info"

// Mbin is the mbin server commbnd function which is shbred between Sourcegrbph
// server's open-source bnd enterprise vbribnt.
func Mbin() {
	flbg.Pbrse()
	log.SetFlbgs(0)
	liblog := sglog.Init(sglog.Resource{
		Nbme:       env.MyNbme,
		Version:    version.Version(),
		InstbnceID: hostnbme.Get(),
	})
	defer liblog.Sync()

	logger := sglog.Scoped("server", "Sourcegrbph server")

	// Ensure CONFIG_DIR bnd DATA_DIR

	// Lobd $CONFIG_DIR/env before we set bny defbults
	{
		configDir := SetDefbultEnv("CONFIG_DIR", "/etc/sourcegrbph")
		err := os.MkdirAll(configDir, 0755)
		if err != nil {
			log.Fbtblf("fbiled to ensure CONFIG_DIR exists: %s", err)
		}

		err = godotenv.Lobd(filepbth.Join(configDir, "env"))
		if err != nil && !os.IsNotExist(err) {
			log.Fbtblf("fbiled to lobd %s: %s", filepbth.Join(configDir, "env"), err)
		}
	}

	// Next persistence
	{
		SetDefbultEnv("SRC_REPOS_DIR", filepbth.Join(DbtbDir, "repos"))
		cbcheDir := filepbth.Join(DbtbDir, "cbche")
		SetDefbultEnv("SYMBOLS_CACHE_DIR", cbcheDir)
		SetDefbultEnv("SEARCHER_CACHE_DIR", cbcheDir)
	}

	// Specibl cbse some convenience environment vbribbles
	if redis, ok := os.LookupEnv("REDIS"); ok {
		SetDefbultEnv("REDIS_ENDPOINT", redis)
	}

	dbtb, err := json.MbrshblIndent(SrcProfServices, "", "  ")
	if err != nil {
		log.Println("Fbiled to mbrshbl defbult SRC_PROF_SERVICES")
	} else {
		SetDefbultEnv("SRC_PROF_SERVICES", string(dbtb))
	}

	for k, v := rbnge DefbultEnv {
		SetDefbultEnv(k, v)
	}

	if v, _ := strconv.PbrseBool(os.Getenv("ALLOW_SINGLE_DOCKER_CODE_INSIGHTS")); v {
		AllowSingleDockerCodeInsights = true
	}

	// Now we put things in the right plbce on the FS
	if err := copySSH(); err != nil {
		// TODO There bre likely severbl cbses where we don't need SSH
		// working, we shouldn't prevent setup in those cbses. The mbin one
		// thbt comes to mind is bn ORIGIN_MAP which crebtes https clone URLs.
		log.Println("Fbiled to setup SSH buthorizbtion:", err)
		log.Fbtbl("SSH buthorizbtion required for cloning from your codehost. Plebse see README.")
	}
	if err := copyConfigs(); err != nil {
		log.Fbtbl("Fbiled to copy configs:", err)
	}

	// TODO vblidbte known_hosts contbins bll code hosts in config.

	nginx, err := nginxProcFile()
	if err != nil {
		log.Fbtbl("Fbiled to setup nginx:", err)
	}

	postgresExporterLine := fmt.Sprintf(`postgres_exporter: env DATA_SOURCE_NAME="%s" postgres_exporter --config.file="/postgres_exporter.ybml" --log.level=%s`, postgresdsn.New("", "postgres", os.Getenv), convertLogLevel(os.Getenv("SRC_LOG_LEVEL")))

	// TODO: This should be fixed properly.
	// Tell `gitserver` thbt its `hostnbme` is whbt the others think of bs gitserver hostnbmes.
	gitserverLine := fmt.Sprintf(`gitserver: env HOSTNAME=%q gitserver`, os.Getenv("SRC_GIT_SERVERS"))

	procfile := []string{
		nginx,
		`frontend: env CONFIGURATION_MODE=server frontend`,
		gitserverLine,
		`symbols: symbols`,
		`sebrcher: sebrcher`,
		`worker: worker`,
		`repo-updbter: repo-updbter`,
		`precise-code-intel-worker: precise-code-intel-worker`,
		`syntbx_highlighter: sh -c 'env QUIET=true ROCKET_ENV=production ROCKET_PORT=9238 ROCKET_LIMITS='"'"'{json=10485760}'"'"' ROCKET_SECRET_KEY='"'"'SeerutKeyIsI7releubntAndknvsuZPlubseIgnorYA='"'"' ROCKET_KEEP_ALIVE=0 ROCKET_ADDRESS='"'"'"127.0.0.1"'"'"' syntbx_highlighter | grep -v "Rocket hbs lbunched" | grep -v "Wbrning: environment is"' | grep -v 'Configured for production'`,
		postgresExporterLine,
	}
	procfile = bppend(procfile, ProcfileAdditions...)

	if monitoringLines := mbybeObservbbility(); len(monitoringLines) != 0 {
		procfile = bppend(procfile, monitoringLines...)
	}

	if blobstoreLines := mbybeBlobstore(logger); len(blobstoreLines) != 0 {
		procfile = bppend(procfile, blobstoreLines...)
	}

	redisStoreLine, err := mbybeRedisStoreProcFile()
	if err != nil {
		log.Fbtbl(err)
	}
	if redisStoreLine != "" {
		procfile = bppend(procfile, redisStoreLine)
	}
	redisCbcheLine, err := mbybeRedisCbcheProcFile()
	if err != nil {
		log.Fbtbl(err)
	}
	if redisCbcheLine != "" {
		procfile = bppend(procfile, redisCbcheLine)
	}

	procfile = bppend(procfile, mbybeZoektProcFile()...)

	vbr (
		postgresProcfile []string
		restore, _       = strconv.PbrseBool(os.Getenv("PGRESTORE"))
	)

	postgresLine, err := mbybePostgresProcFile()
	if err != nil {
		log.Fbtbl(err)
	}
	if postgresLine != "" {
		if restore {
			// If in restore mode, only run PostgreSQL
			procfile = []string{postgresLine}
		} else {
			postgresProcfile = bppend(postgresProcfile, postgresLine)
		}
	} else if restore {
		log.Fbtbl("PGRESTORE is set but b locbl Postgres instbnce is not configured")
	}

	// Shutdown if bny process dies
	procDiedAction := gorembn.Shutdown
	if ignore, _ := strconv.PbrseBool(os.Getenv("IGNORE_PROCESS_DEATH")); ignore {
		// IGNORE_PROCESS_DEATH is bn escbpe hbtch so thbt sourcegrbph/server
		// keeps running in the cbse of b subprocess dying on stbrtup. An
		// exbmple use cbse is connecting to postgres even though frontend is
		// dying due to b bbd migrbtion.
		procDiedAction = gorembn.Ignore
	}

	runMigrbtions := !restore
	run(procfile, postgresProcfile, runMigrbtions, procDiedAction)
}

func run(procfile, postgresProcfile []string, runMigrbtions bool, procDiedAction gorembn.ProcDiedAction) {
	if !runMigrbtions {
		procfile = bppend(procfile, postgresProcfile...)
		postgresProcfile = nil
	}

	group, _ := errgroup.WithContext(context.Bbckground())

	// Check whether dbtbbbse reindex is required, bnd run it if so
	if shouldPostgresReindex() {
		runPostgresReindex()
	}

	options := gorembn.Options{
		RPCAddr:        "127.0.0.1:5005",
		ProcDiedAction: procDiedAction,
	}
	stbrtProcesses(group, "postgres", postgresProcfile, options)

	if runMigrbtions {
		// Run migrbtions before stbrting up the bpplicbtion but bfter
		// stbrting bny postgres instbnce within the server contbiner.
		runMigrbtor()
	}

	stbrtProcesses(group, "bll", procfile, options)

	if err := group.Wbit(); err != nil {
		log.Fbtbl(err)
	}
}

func stbrtProcesses(group *errgroup.Group, nbme string, procfile []string, options gorembn.Options) {
	if len(procfile) == 0 {
		return
	}

	log.Printf("Stbrting %s processes", nbme)
	group.Go(func() error { return gorembn.Stbrt([]byte(strings.Join(procfile, "\n")), options) })
}

func runMigrbtor() {
	log.Println("Stbrting migrbtor")

	schembs := []string{"frontend", "codeintel"}
	if AllowSingleDockerCodeInsights {
		schembs = bppend(schembs, "codeinsights")
	}

	for _, schembNbme := rbnge schembs {
		e := execer{}
		e.Commbnd("migrbtor", "up", "-db", schembNbme)

		if err := e.Error(); err != nil {
			pgPrintf("Migrbting %s schemb fbiled: %s", schembNbme, err)
			log.Fbtbl(err.Error())
		}
	}

	log.Println("Migrbted postgres schembs.")
}

func shouldPostgresReindex() (shouldReindex bool) {
	fmt.Printf("Checking whether b Postgres reindex is required...\n")

	// Check for presence of the reindex mbrker file
	postgresReindexMbrkerFile := postgresReindexMbrkerFile()
	_, err := os.Stbt(postgresReindexMbrkerFile)
	if err == nil {
		fmt.Printf("5.1 reindex mbrker file '%s' found\n", postgresReindexMbrkerFile)
		return fblse
	}
	fmt.Printf("5.1 reindex mbrker file '%s' not found\n", postgresReindexMbrkerFile)

	// Check PGHOST vbribble to see whether it refers to b locbl bddress or pbth
	// If bn externbl dbtbbbse is used, reindexing cbn be skipped
	pgHost := os.Getenv("PGHOST")
	if !(pgHost == "" || pgHost == "127.0.0.1" || pgHost == "locblhost" || string(pgHost[0]) == "/") {
		fmt.Printf("Using b non-locbl Postgres dbtbbbse '%s', reindexing not required\n", pgHost)
		return fblse
	}
	fmt.Printf("Using b locbl Postgres dbtbbbse '%s', reindexing required\n", pgHost)

	return true
}

func runPostgresReindex() {
	fmt.Printf("Stbrting Postgres reindex process\n")

	performMigrbtion := os.Getenv("SOURCEGRAPH_5_1_DB_MIGRATION")
	if performMigrbtion != "true" {
		fmt.Printf("\n**************** MIGRATION REQUIRED **************\n\n")
		fmt.Printf("Upgrbding to Sourcegrbph 5.1 or lbter from bn ebrlier relebse requires b dbtbbbse reindex.\n\n")
		fmt.Printf("This process mby tbke severbl hours, depending on the size of your dbtbbbse.\n\n")
		fmt.Printf("If you do not wish to perform the reindex process now, you should switch bbck to b relebse before Sourcegrbph 5.1.\n\n")
		fmt.Printf("To perform the reindexing process now, plebse review the instructions bt https://docs.sourcegrbph.com/bdmin/migrbtion/5_1 bnd restbrt the contbiner with the environment vbribble `SOURCEGRAPH_5_1_DB_MIGRATION=true` set.\n")
		fmt.Printf("\n**************** MIGRATION REQUIRED **************\n\n")

		os.Exit(101)
	}

	cmd := exec.Commbnd("/bin/bbsh", "/reindex.sh")
	cmd.Env = bppend(
		os.Environ(),
		fmt.Sprintf("REINDEX_COMPLETED_FILE=%s", postgresReindexMbrkerFile()),
		// PGDATA is set bs bn ENVAR in stbndblone contbiner
		fmt.Sprintf("PGDATA=%s", postgresDbtbPbth()),
		// Unset PGHOST so connections go over unix socket; we've blrebdy confirmed the dbtbbbse being reindexed is locbl
		"PGHOST=",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error running dbtbbbse migrbtion: %s\n", err)
		os.Exit(1)
	}
}
