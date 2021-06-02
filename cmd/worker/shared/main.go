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

func Main(enterpriseSetupHooks map[string]SetupHook) {
	allHooks := map[string]SetupHook{}
	for name, setupHook := range setupHooks {
		allHooks[name] = setupHook
	}
	for name, setupHook := range enterpriseSetupHooks {
		allHooks[name] = setupHook
	}

	// Load worker service config which will determine which of
	// the registered setup hooks to invoke on application startup.

	names := make([]string, 0, len(allHooks))
	for name := range allHooks {
		names = append(names, name)
	}
	config.names = names
	config.Load()

	// Load all other registered configs. We load all configs
	// regardless if their attached hook will run to encourage
	// consistency in deployments.

	for _, setupHook := range allHooks {
		for _, config := range setupHook.Config() {
			config.Load()
		}
	}

	env.Lock()
	env.HandleHelpFlag()
	logging.Init()
	tracer.Init()
	trace.Init(true)

	// Validate the loaded configs. If there are errors at this
	// point, collect all that are relevant, log, then exit.

	validationErrors := map[string][]error{}
	if err := config.Validate(); err != nil {
		log.Fatalf("Failed to load configuration: %s", err)
	}
	for name, setupHook := range allHooks {
		for _, config := range setupHook.Config() {
			if err := config.Validate(); err != nil {
				validationErrors[name] = append(validationErrors[name], err)
			}
		}
	}
	if len(validationErrors) > 0 {
		var descriptions []string
		for name, errs := range validationErrors {
			for _, err := range errs {
				descriptions = append(descriptions, fmt.Sprintf("  - %s: %s ", name, err))
			}
		}
		sort.Strings(descriptions)

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
	results := make(chan routinesResult, len(allHooks))
	ctx, cancel := context.WithCancel(context.Background())

	for name, setupHook := range allHooks {
		if shouldRunSetupHook(name) {
			log15.Info("Running setup hook", "name", name)

			wg.Add(1)

			go func(name string, setupHook SetupHook) {
				defer wg.Done()

				routines, err := setupHook.Routines(ctx)
				results <- routinesResult{name, routines, err}
				if err != nil {
					cancel()
				}
			}(name, setupHook)
		} else {
			log15.Info("Skipping setup hook", "name", name)
		}
	}

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
		sort.Strings(descriptions)
		log.Fatalf("Failed to initialize worker:\n%s", strings.Join(descriptions, "\n"))
	}

	// Initialize health server
	server := httpserver.NewFromAddr(addr, &http.Server{
		ReadTimeout:  75 * time.Second,
		WriteTimeout: 10 * time.Minute,
		Handler:      httpserver.NewHandler(nil),
	})
	allRoutines = append(allRoutines, server)

	close(ready)
	goroutine.MonitorBackgroundRoutines(context.Background(), allRoutines...)
}

type routinesResult struct {
	name     string
	routines []goroutine.BackgroundRoutine
	err      error
}
