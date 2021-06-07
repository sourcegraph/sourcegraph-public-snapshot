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
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/tracer"
)

const addr = ":3189"

// Start runs the worker. This method does not return.
func Start(additionalTasks map[string]Task) {
	tasks := map[string]Task{}
	for name, task := range bultins {
		tasks[name] = task
	}
	for name, task := range additionalTasks {
		tasks[name] = task
	}

	// Setup environment variables
	loadConfigs(tasks)

	env.Lock()
	env.HandleHelpFlag()
	logging.Init()
	tracer.Init()
	trace.Init(true)

	// Start debug server
	ready := make(chan struct{})
	go debugserver.NewServerRoutine(ready).Start()

	// Validate environment variables
	mustValidateConfigs(tasks)

	// Create the background routines that the worker will monitor for its
	// lifetime. There may be a non-trivial startup time on this step as we
	// connect to external databases, wait for migrations, etc.
	allRoutines := mustCreateBackgroundRoutines(tasks)

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

// loadConfigs calls Load on the configs of each of the tasks registered in this binary.
// All configs will be loaded regardless if they would later be validated - this is the
// best place we have to manipulate the environment before the call to env.Lock.
func loadConfigs(tasks map[string]Task) {
	// Load the worker config
	config.names = taskNames(tasks)
	config.Load()

	// Load all other registered configs
	for _, t := range tasks {
		for _, c := range t.Config() {
			c.Load()
		}
	}
}

// mustValidateConfigs calls Validate on the configs of each of the tasks that will be run
// by this instance of the worker. If any config has a validation error, a fatal log message
// will be emitted.
func mustValidateConfigs(tasks map[string]Task) {
	validationErrors := map[string][]error{}
	if err := config.Validate(); err != nil {
		log.Fatalf("Failed to load configuration: %s", err)
	}

	if len(validationErrors) == 0 {
		// If the worker config is valid, validate the children configs. We guard this
		// in the case of worker config errors because we don't want to spew validation
		// errors for things that should be disabled.
		for name, task := range tasks {
			if !shouldRunTask(name) {
				continue
			}

			for _, c := range task.Config() {
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

// mustCreateBackgroundRoutines runs the Routines function of each of the given tasks concurrently.
// If an error occurs from any of them, a fatal log message will be emitted. Otherwise, the set
// of background routines from each task will be returned.
func mustCreateBackgroundRoutines(tasks map[string]Task) []goroutine.BackgroundRoutine {
	var (
		allRoutines  []goroutine.BackgroundRoutine
		descriptions []string
	)

	for result := range runRoutinesConcurrently(tasks) {
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
// the Routines function from each given task. Each function is called concurrently. If an
// error occurs in one function, the context passed to all its siblings will be canceled.
func runRoutinesConcurrently(tasks map[string]Task) chan routinesResult {
	results := make(chan routinesResult, len(tasks))
	defer close(results)

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, name := range taskNames(tasks) {
		taskLoggr := log15.New("name", name)

		if !shouldRunTask(name) {
			taskLoggr.Info("Skipping task")
			continue
		}

		wg.Add(1)
		taskLoggr.Info("Running task")

		go func(name string) {
			defer wg.Done()

			routines, err := tasks[name].Routines(ctx)
			results <- routinesResult{name, routines, err}

			if err == nil {
				taskLoggr.Info("Finished initializing task")
			} else {
				cancel()
			}
		}(name)
	}

	wg.Wait()
	return results
}

// taskNames returns an ordered slice of keys from the given map.
func taskNames(tasks map[string]Task) []string {
	names := make([]string, 0, len(tasks))
	for name := range tasks {
		names = append(names, name)
	}
	sort.Strings(names)

	return names
}
