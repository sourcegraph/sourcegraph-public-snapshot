package shared

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/profiler"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/tracer"
)

const addr = ":3189"

// Start runs the worker. This method does not return.
func Start(additionalJobs map[string]Job) {
	jobs := map[string]Job{}
	for name, job := range builtins {
		jobs[name] = job
	}
	for name, job := range additionalJobs {
		jobs[name] = job
	}

	// Setup environment variables
	loadConfigs(jobs)

	// Set up Google Cloud Profiler when running in Cloud
	if err := profiler.Init(); err != nil {
		log.Fatalf("Failed to start profiler: %v", err)
	}

	env.Lock()
	env.HandleHelpFlag()
	logging.Init()
	tracer.Init()
	trace.Init(true)

	// Start debug server
	ready := make(chan struct{})
	go debugserver.NewServerRoutine(ready).Start()

	// Validate environment variables
	mustValidateConfigs(jobs)

	// Create the background routines that the worker will monitor for its
	// lifetime. There may be a non-trivial startup time on this step as we
	// connect to external databases, wait for migrations, etc.
	allRoutines := mustCreateBackgroundRoutines(jobs)

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

	goroutine.MonitorBackgroundRoutines(context.Background(), allRoutines...)
}

// loadConfigs calls Load on the configs of each of the jobs registered in this binary.
// All configs will be loaded regardless if they would later be validated - this is the
// best place we have to manipulate the environment before the call to env.Lock.
func loadConfigs(jobs map[string]Job) {
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

// mustValidateConfigs calls Validate on the configs of each of the jobs that will be run
// by this instance of the worker. If any config has a validation error, a fatal log message
// will be emitted.
func mustValidateConfigs(jobs map[string]Job) {
	validationErrors := map[string][]error{}
	if err := config.Validate(); err != nil {
		log.Fatalf("Failed to load configuration: %s", err)
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

		log.Fatalf("Failed to load configuration:\n%s", strings.Join(descriptions, "\n"))
	}
}

// mustCreateBackgroundRoutines runs the Routines function of each of the given jobs concurrently.
// If an error occurs from any of them, a fatal log message will be emitted. Otherwise, the set
// of background routines from each job will be returned.
func mustCreateBackgroundRoutines(jobs map[string]Job) []goroutine.BackgroundRoutine {
	var (
		allRoutines  []goroutine.BackgroundRoutine
		descriptions []string
	)

	for result := range runRoutinesConcurrently(jobs) {
		if result.err == nil {
			allRoutines = append(allRoutines, result.routines...)
		} else {
			descriptions = append(descriptions, fmt.Sprintf("  - %s: %s", result.name, result.err))
		}
	}
	sort.Strings(descriptions)

	if len(descriptions) != 0 {
		log.Fatalf("Failed to initialize worker:\n%s", strings.Join(descriptions, "\n"))
	}

	return allRoutines
}

type routinesResult struct {
	name     string
	routines []goroutine.BackgroundRoutine
	err      error
}

// runRoutinesConcurrently returns a channel that will be populated with the return value of
// the Routines function from each given job. Each function is called concurrently. If an
// error occurs in one function, the context passed to all its siblings will be canceled.
func runRoutinesConcurrently(jobs map[string]Job) chan routinesResult {
	results := make(chan routinesResult, len(jobs))
	defer close(results)

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, name := range jobNames(jobs) {
		jobLogger := log15.New("name", name)

		if !shouldRunJob(name) {
			jobLogger.Info("Skipping job")
			continue
		}

		wg.Add(1)
		jobLogger.Info("Running job")

		go func(name string) {
			defer wg.Done()

			routines, err := jobs[name].Routines(ctx)
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
func jobNames(jobs map[string]Job) []string {
	names := make([]string, 0, len(jobs))
	for name := range jobs {
		names = append(names, name)
	}
	sort.Strings(names)

	return names
}
