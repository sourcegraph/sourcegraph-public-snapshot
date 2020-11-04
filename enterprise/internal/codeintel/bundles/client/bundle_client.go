package client

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence"
)

// BundleClient is the interface to the precise-code-intel-bundle-manager service scoped to a particular dump.
type BundleClient interface {
	// ID gets the identifier of the target bundle.
	ID() int

	// Exists determines if the given path exists in the dump.
	Exists(ctx context.Context, path string) (bool, error)

	// Ranges returns definition, reference, and hover data for each range within the given span of lines.
	Ranges(ctx context.Context, path string, startLine, endLine int) ([]CodeIntelligenceRange, error)

	// Definitions retrieves a list of definition locations for the symbol under the given location.
	Definitions(ctx context.Context, path string, line, character int) ([]Location, error)

	// Definitions retrieves a list of reference locations for the symbol under the given location.
	References(ctx context.Context, path string, line, character int) ([]Location, error)

	// Hover retrieves the hover text for the symbol under the given location.
	Hover(ctx context.Context, path string, line, character int) (string, Range, bool, error)

	// Diagnostics retrieves the diagnostics and total count of diagnostics for the documents that have the given path prefix.
	Diagnostics(ctx context.Context, prefix string, skip, take int) ([]Diagnostic, int, error)

	// MonikersByPosition retrieves a list of monikers attached to the symbol under the given location. There may
	// be multiple ranges enclosing this point. The returned monikers are partitioned such that inner ranges occur
	// first in the result, and outer ranges occur later.
	MonikersByPosition(ctx context.Context, path string, line, character int) ([][]MonikerData, error)

	// MonikerResults retrieves a page of locations attached to a moniker and a total count of such locations.
	MonikerResults(ctx context.Context, modelType, scheme, identifier string, skip, take int) ([]Location, int, error)

	// PackageInformation retrieves package information data by its identifier.
	PackageInformation(ctx context.Context, path, packageInformationID string) (PackageInformationData, error)
}

type bundleClientImpl struct {
	bundleID       int
	store          persistence.Store
	databaseOpener func(ctx context.Context, filename string, store persistence.Store) (database.Database, error)
}

var _ BundleClient = &bundleClientImpl{}

// ID gets the identifier of the target bundle.
func (c *bundleClientImpl) ID() int {
	return c.bundleID
}

// Exists determines if the given path exists in the dump.
func (c *bundleClientImpl) Exists(ctx context.Context, path string) (bool, error) {
	db, err := c.openDatabase(ctx)
	if err != nil {
		return false, err
	}

	return db.Exists(ctx, path)
}

// Ranges returns definition, reference, and hover data for each range within the given span of lines.
func (c *bundleClientImpl) Ranges(ctx context.Context, path string, startLine, endLine int) ([]CodeIntelligenceRange, error) {
	db, err := c.openDatabase(ctx)
	if err != nil {
		return nil, err
	}

	return db.Ranges(ctx, path, startLine, endLine)
}

// Definitions retrieves a list of definition locations for the symbol under the given location.
func (c *bundleClientImpl) Definitions(ctx context.Context, path string, line, character int) ([]Location, error) {
	db, err := c.openDatabase(ctx)
	if err != nil {
		return nil, err
	}

	locations, err := db.Definitions(ctx, path, line, character)
	c.addBundleIDToLocations(locations)
	return locations, err
}

// Definitions retrieves a list of reference locations for the symbol under the given location.
func (c *bundleClientImpl) References(ctx context.Context, path string, line, character int) ([]Location, error) {
	db, err := c.openDatabase(ctx)
	if err != nil {
		return nil, err
	}

	locations, err := db.References(ctx, path, line, character)
	c.addBundleIDToLocations(locations)
	return locations, err
}

// Hover retrieves the hover text for the symbol under the given location.
func (c *bundleClientImpl) Hover(ctx context.Context, path string, line, character int) (string, Range, bool, error) {
	db, err := c.openDatabase(ctx)
	if err != nil {
		return "", Range{}, false, err
	}

	return db.Hover(ctx, path, line, character)
}

// Diagnostics retrieves the diagnostics and total count of diagnostics for the documents that have the given path prefix.
func (c *bundleClientImpl) Diagnostics(ctx context.Context, prefix string, skip, take int) ([]Diagnostic, int, error) {
	db, err := c.openDatabase(ctx)
	if err != nil {
		return nil, 0, err
	}

	diagnostics, count, err := db.Diagnostics(ctx, prefix, skip, take)
	if err != nil {
		return nil, 0, err
	}

	c.addBundleIDToDiagnostics(diagnostics)
	return diagnostics, count, err
}

// MonikersByPosition retrieves a list of monikers attached to the symbol under the given location. There may
// be multiple ranges enclosing this point. The returned monikers are partitioned such that inner ranges occur
// first in the result, and outer ranges occur later.
func (c *bundleClientImpl) MonikersByPosition(ctx context.Context, path string, line, character int) ([][]MonikerData, error) {
	db, err := c.openDatabase(ctx)
	if err != nil {
		return nil, err
	}

	return db.MonikersByPosition(ctx, path, line, character)
}

// MonikerResults retrieves a page of locations attached to a moniker and a total count of such locations.
func (c *bundleClientImpl) MonikerResults(ctx context.Context, modelType, scheme, identifier string, skip, take int) ([]Location, int, error) {
	db, err := c.openDatabase(ctx)
	if err != nil {
		return nil, 0, err
	}

	var tableName string
	switch modelType {
	case "definition":
		tableName = "definitions"
	case "reference":
		tableName = "references"
	}

	locations, count, err := db.MonikerResults(ctx, tableName, scheme, identifier, skip, take)
	if err != nil {
		return nil, 0, err
	}

	c.addBundleIDToLocations(locations)
	return locations, count, err
}

// PackageInformation retrieves package information data by its identifier.
func (c *bundleClientImpl) PackageInformation(ctx context.Context, path, packageInformationID string) (PackageInformationData, error) {
	db, err := c.openDatabase(ctx)
	if err != nil {
		return PackageInformationData{}, err
	}

	data, _, err := db.PackageInformation(ctx, path, packageInformationID)
	return data, err
}

func (c *bundleClientImpl) openDatabase(ctx context.Context) (database.Database, error) {
	if _, err := c.store.ReadMeta(ctx); err != nil {
		return nil, err
	}

	return c.databaseOpener(ctx, fmt.Sprintf("%d", c.bundleID), c.store)
}

func (c *bundleClientImpl) addBundleIDToLocations(locations []Location) {
	for i := range locations {
		locations[i].DumpID = c.bundleID
	}
}

func (c *bundleClientImpl) addBundleIDToDiagnostics(diagnostics []Diagnostic) {
	for i := range diagnostics {
		diagnostics[i].DumpID = c.bundleID
	}
}
