package shared

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/cmd/worker/internal/encryption"
	"github.com/sourcegraph/sourcegraph/cmd/worker/internal/gitserver"
	workermigrations "github.com/sourcegraph/sourcegraph/cmd/worker/internal/migrations"
	"github.com/sourcegraph/sourcegraph/cmd/worker/internal/repostatistics"
	"github.com/sourcegraph/sourcegraph/cmd/worker/internal/webhooks"
	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration/migrations"
	"github.com/sourcegraph/sourcegraph/internal/profiler"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/tracer"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const addr = ":3189"

// Start runs the worker.
func Start(logger log.Logger, additionalJobs map[string]job.Job, registerEnterpriseMigrators oobmigration.RegisterMigratorsFunc) error {
	registerMigrators := oobmigration.ComposeRegisterMigratorsFuncs(migrations.RegisterOSSMigrators, registerEnterpriseMigrators)

	builtins := map[string]job.Job{
		"webhook-log-janitor":                   webhooks.NewJanitor(),
		"out-of-band-migrations":                workermigrations.NewMigrator(registerMigrators),
		"codeintel-documents-indexer":           codeintel.NewDocumentsIndexerJob(),
		"codeintel-policies-repository-matcher": codeintel.NewPoliciesRepositoryMatcherJob(),
		"codeintel-crates-syncer":               codeintel.NewCratesSyncerJob(),
		"gitserver-metrics":                     gitserver.NewMetricsJob(),
		"record-encrypter":                      encryption.NewRecordEncrypterJob(),
		"repo-statistics-compactor":             repostatistics.NewCompactor(),
	}

	jobs := map[string]job.Job{}
	for name, job := range builtins {
		jobs[name] = job
	}
	for name, job := range additionalJobs {
		jobs[name] = job
	}

	// Setup environment variables
	loadConfigs(jobs)

	env.Lock()
	env.HandleHelpFlag()
	conf.Init()
	logging.Init()
	tracer.Init(log.Scoped("tracer", "internal tracer package"), conf.DefaultClient())
	trace.Init()
	profiler.Init()

	if err := keyring.Init(context.Background()); err != nil {
		return errors.Wrap(err, "Failed to intialise keyring")
	}

	// Start debug server
	ready := make(chan struct{})
	go debugserver.NewServerRoutine(ready).Start()

	// Validate environment variables
	if err := validateConfigs(jobs); err != nil {
		return err
	}

	// Emit metrics to help site admins detect instances that accidentally
	// omit a job from from the instance's deployment configuration.
	emitJobCountMetrics(jobs)

	// Create the background routines that the worker will monitor for its
	// lifetime. There may be a non-trivial startup time on this step as we
	// connect to external databases, wait for migrations, etc.
	allRoutines, err := createBackgroundRoutines(logger, jobs)
	if err != nil {
		return err
	}

	// Initialize health server
	server := httpserver.NewFromAddr(addr, &http.Server{
		ReadTimeout:  75 * time.Second,
		WriteTimeout: 10 * time.Minute,
		Handler:      httpserver.NewHandler(nil),
	})
	allRoutines = append(allRoutines, server)

	// We're all set up now
	// Respond positively to ready checks
	close(ready)

	// This method blocks while the app is live - the following return is only to appease
	// the type checker.
	goroutine.MonitorBackgroundRoutines(context.Background(), allRoutines...)
	return nil
}

// loadConfigs calls Load on the configs of each of the jobs registered in this binary.
// All configs will be loaded regardless if they would later be validated - this is the
// best place we have to manipulate the environment before the call to env.Lock.
func loadConfigs(jobs map[string]job.Job) {
	// Load the worker config
	config.names = jobNames(jobs)
	config.Load()

	// Load all other registered configs
	for _, j := range jobs {
		for _, c := range j.Config() {
			c.Load()
		}
	}
}

// validateConfigs calls Validate on the configs of each of the jobs that will be run
// by this instance of the worker. If any config has a validation error, an error is
// returned.
func validateConfigs(jobs map[string]job.Job) error {
	validationErrors := map[string][]error{}
	if err := config.Validate(); err != nil {
		return errors.Wrap(err, "Failed to load configuration")
	}

	if len(validationErrors) == 0 {
		// If the worker config is valid, validate the children configs. We guard this
		// in the case of worker config errors because we don't want to spew validation
		// errors for things that should be disabled.
		for name, job := range jobs {
			if !shouldRunJob(name) {
				continue
			}

			for _, c := range job.Config() {
				if err := c.Validate(); err != nil {
					validationErrors[name] = append(validationErrors[name], err)
				}
			}
		}
	}

	if len(validationErrors) != 0 {
		var descriptions []string
		for name, errs := range validationErrors {
			for _, err := range errs {
				descriptions = append(descriptions, fmt.Sprintf("  - %s: %s ", name, err))
			}
		}
		sort.Strings(descriptions)

		return errors.Newf("Failed to load configuration:\n%s", strings.Join(descriptions, "\n"))
	}

	return nil
}

// emitJobCountMetrics registers and emits an initial value for gauges referencing each of
// the jobs that will be run by this instance of the worker. Since these metrics are summed
// over all instances (and we don't change the jobs that are registered to a running worker),
// we only need to emit an initial count once.
func emitJobCountMetrics(jobs map[string]job.Job) {
	gauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "src_worker_jobs",
		Help: "Total number of jobs running in the worker.",
	}, []string{"job_name"})

	prometheus.DefaultRegisterer.MustRegister(gauge)

	for name := range jobs {
		if !shouldRunJob(name) {
			continue
		}

		gauge.WithLabelValues(name).Set(1)
	}
}

// createBackgroundRoutines runs the Routines function of each of the given jobs concurrently.
// If an error occurs from any of them, a fatal log message will be emitted. Otherwise, the set
// of background routines from each job will be returned.
func createBackgroundRoutines(logger log.Logger, jobs map[string]job.Job) ([]goroutine.BackgroundRoutine, error) {
	var (
		allRoutines  []goroutine.BackgroundRoutine
		descriptions []string
	)

	for result := range runRoutinesConcurrently(logger, jobs) {
		if result.err == nil {
			allRoutines = append(allRoutines, result.routines...)
		} else {
			descriptions = append(descriptions, fmt.Sprintf("  - %s: %s", result.name, result.err))
		}
	}
	sort.Strings(descriptions)

	if len(descriptions) != 0 {
		return nil, errors.Newf("Failed to initialize worker:\n%s", strings.Join(descriptions, "\n"))
	}

	return allRoutines, nil
}

type routinesResult struct {
	name     string
	routines []goroutine.BackgroundRoutine
	err      error
}

// runRoutinesConcurrently returns a channel that will be populated with the return value of
// the Routines function from each given job. Each function is called concurrently. If an
// error occurs in one function, the context passed to all its siblings will be canceled.
func runRoutinesConcurrently(logger log.Logger, jobs map[string]job.Job) chan routinesResult {
	results := make(chan routinesResult, len(jobs))
	defer close(results)

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, name := range jobNames(jobs) {
		jobLogger := logger.Scoped(name, jobs[name].Description())

		if !shouldRunJob(name) {
			jobLogger.Info("Skipping job")
			continue
		}

		wg.Add(1)
		jobLogger.Info("Running job")

		go func(name string) {
			defer wg.Done()

			routines, err := jobs[name].Routines(ctx, jobLogger)
			results <- routinesResult{name, routines, err}

			if err == nil {
				jobLogger.Info("Finished initializing job")
			} else {
				cancel()
			}
		}(name)
	}

	wg.Wait()
	return results
}

// jobNames returns an ordered slice of keys from the given map.
func jobNames(jobs map[string]job.Job) []string {
	names := make([]string, 0, len(jobs))
	for name := range jobs {
		names = append(names, name)
	}
	sort.Strings(names)

	return names
}
