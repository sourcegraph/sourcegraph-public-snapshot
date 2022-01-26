package runner

import (
	"context"
	"fmt"
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
)

type schemaContext struct {
	schema        *schemas.Schema
	store         Store
	schemaVersion schemaVersion
}

type schemaVersion struct {
	version int
	dirty   bool
}

type visitFunc func(ctx context.Context, schemaName string, schemaContext schemaContext) error

// forEachSchema invokes the given function once for each schema in the given list, with
// store instances initialized for each given schema name. Each function invocation occurs
// concurrently. Errors from each invocation are collected and returned. An error from one
// goroutine will not cancel the progress of another.
func (r *Runner) forEachSchema(ctx context.Context, schemaNames []string, visitor visitFunc) error {
	// Create map of relevant schemas keyed by name
	schemaMap, err := r.prepareSchemas(schemaNames)
	if err != nil {
		return err
	}

	// Create map of migration stores keyed by name
	storeMap, err := r.prepareStores(ctx, schemaNames)
	if err != nil {
		return err
	}

	// Create map of versions keyed by name
	versionMap, err := r.fetchVersions(ctx, storeMap)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	errorCh := make(chan error, len(schemaNames))

	for _, schemaName := range schemaNames {
		wg.Add(1)

		go func(schemaName string) {
			defer wg.Done()

			errorCh <- visitor(ctx, schemaName, schemaContext{
				schema:        schemaMap[schemaName],
				store:         storeMap[schemaName],
				schemaVersion: versionMap[schemaName],
			})
		}(schemaName)
	}

	wg.Wait()
	close(errorCh)

	var errs *multierror.Error
	for err := range errorCh {
		if err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	return errs.ErrorOrNil()
}

func (r *Runner) prepareSchemas(schemaNames []string) (map[string]*schemas.Schema, error) {
	schemaMap := make(map[string]*schemas.Schema, len(schemaNames))

	for _, targetSchemaName := range schemaNames {
		for _, schema := range schemas.Schemas {
			if schema.Name == targetSchemaName {
				schemaMap[schema.Name] = schema
				break
			}
		}
	}

	// Ensure that all supplied schema names are valid
	for _, schemaName := range schemaNames {
		if _, ok := schemaMap[schemaName]; !ok {
			return nil, fmt.Errorf("unknown schema %q", schemaName)
		}
	}

	return schemaMap, nil
}

func (r *Runner) prepareStores(ctx context.Context, schemaNames []string) (map[string]Store, error) {
	storeMap := make(map[string]Store, len(schemaNames))

	for _, schemaName := range schemaNames {
		storeFactory, ok := r.storeFactories[schemaName]
		if !ok {
			return nil, fmt.Errorf("unknown schema %q", schemaName)
		}

		store, err := storeFactory(ctx)
		if err != nil {
			return nil, err
		}

		storeMap[schemaName] = store
	}

	return storeMap, nil
}

func (r *Runner) fetchVersions(ctx context.Context, storeMap map[string]Store) (map[string]schemaVersion, error) {
	versions := make(map[string]schemaVersion, len(storeMap))

	for schemaName, store := range storeMap {
		version, dirty, err := r.fetchVersion(ctx, schemaName, store)
		if err != nil {
			return nil, err
		}

		versions[schemaName] = schemaVersion{version, dirty}
	}

	return versions, nil
}

func (r *Runner) fetchVersion(ctx context.Context, schemaName string, store Store) (int, bool, error) {
	version, dirty, _, err := store.Version(ctx)
	if err != nil {
		return 0, false, err
	}

	log15.Info("Checked current version", "schema", schemaName, "version", version, "dirty", dirty)
	return version, dirty, nil
}
