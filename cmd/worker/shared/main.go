package shared

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/tracer"
)

// SetupHook creates configuration struct and background routine instances
// to be run as part of the worker process.
type SetupHook interface {
	// Config returns a set of configuration struct pointers that should
	// be loaded and validated as part of application startup.
	Config() []env.Config

	// Routines constructs and returns the set of background routines that
	// should run as part of the worker process. Service initialization should
	// be shared between setup hooks when possible (e.g. sync.Once initialization).
	Routines(ctx context.Context) ([]goroutine.BackgroundRoutine, error)
}

var setupHooks = map[string]SetupHook{
	// Empty for now
}

func Main(enterpriseSetupHooks map[string]SetupHook) {
	for _, setupHook := range enterpriseSetupHooks {
		for _, config := range setupHook.Config() {
			config.Load()
		}
	}

	env.Lock()
	env.HandleHelpFlag()
	logging.Init()
	tracer.Init()
	trace.Init(true)

	var validationErrors []error
	for _, setupHook := range enterpriseSetupHooks {
		for _, config := range setupHook.Config() {
			if err := config.Validate(); err != nil {
				validationErrors = append(validationErrors, err)
			}
		}
	}
	if len(validationErrors) > 0 {
		descriptions := make([]string, 0, len(validationErrors))
		for _, err := range validationErrors {
			descriptions = append(descriptions, "  - "+err.Error())
		}

		log.Fatalf("Failed to load configuration:\n%s", strings.Join(descriptions, "\n"))
	}

	// Start debug server
	ready := make(chan struct{})
	go debugserver.NewServerRoutine(ready).Start()

	// Run each startup hook concurrently. If an error occurs in one startup hook, it
	// will cancel the context that is running all sibling setup hooks. The routines
	// created by the setup hooks will be started at the same time, regardless of how
	// quickly one set of routines initializes.

	var wg sync.WaitGroup
	results := make(chan routinesResult, len(setupHooks)+len(enterpriseSetupHooks))
	ctx, cancel := context.WithCancel(context.Background())

	queue := func(setupHooks map[string]SetupHook) {
		for name, setupHook := range setupHooks {
			wg.Add(1)

			go func(name string, setupHook SetupHook) {
				defer wg.Done()

				routines, err := setupHook.Routines(ctx)
				results <- routinesResult{name, routines, err}
				if err != nil {
					cancel()
				}
			}(name, setupHook)
		}
	}
	queue(setupHooks)
	queue(enterpriseSetupHooks)

	wg.Wait()
	cancel()
	close(results)

	// Collect the results from the concurrent execution above. If there is an error
	// in any of the startup hooks, we'll report and exit. Otherwise, collect all of
	// the routines and monitor them. Only after the long setup and immediately before
	// monitoring running background routines do we mark the health server as ready.

	var descriptions []string
	var allRoutines []goroutine.BackgroundRoutine
	for result := range results {
		if result.err == nil {
			allRoutines = append(allRoutines, result.routines...)
		} else {
			descriptions = append(descriptions, fmt.Sprintf("  - %s: %s", result.name, result.err))
		}
	}
	if len(descriptions) > 0 {
		log.Fatalf("Failed to initialize worker:\n%s", strings.Join(descriptions, "\n"))
	}

	close(ready)
	goroutine.MonitorBackgroundRoutines(context.Background(), allRoutines...)
}

type routinesResult struct {
	name     string
	routines []goroutine.BackgroundRoutine
	err      error
}
