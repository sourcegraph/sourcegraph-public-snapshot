package client

import (
	"context"
	"database/sql"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence/postgres"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// BundleClient is the interface to the precise-code-intel-bundle-manager service scoped to a particular dump.
type BundleManagerClient interface {
	// Exists determines if the given path exists in the dump.
	Exists(ctx context.Context, bundleID int, path string) (bool, error)

	// Ranges returns definition, reference, and hover data for each range within the given span of lines.
	Ranges(ctx context.Context, bundleID int, path string, startLine, endLine int) ([]CodeIntelligenceRange, error)

	// Definitions retrieves a list of definition locations for the symbol under the given location.
	Definitions(ctx context.Context, bundleID int, path string, line, character int) ([]Location, error)

	// Definitions retrieves a list of reference locations for the symbol under the given location.
	References(ctx context.Context, bundleID int, path string, line, character int) ([]Location, error)

	// Hover retrieves the hover text for the symbol under the given location.
	Hover(ctx context.Context, bundleID int, path string, line, character int) (string, Range, bool, error)

	// Diagnostics retrieves the diagnostics and total count of diagnostics for the documents that have the given path prefix.
	Diagnostics(ctx context.Context, bundleID int, prefix string, skip, take int) ([]Diagnostic, int, error)

	// MonikersByPosition retrieves a list of monikers attached to the symbol under the given location. There may
	// be multiple ranges enclosing this point. The returned monikers are partitioned such that inner ranges occur
	// first in the result, and outer ranges occur later.
	MonikersByPosition(ctx context.Context, bundleID int, path string, line, character int) ([][]MonikerData, error)

	// MonikerResults retrieves a page of locations attached to a moniker and a total count of such locations.
	MonikerResults(ctx context.Context, bundleID int, modelType, scheme, identifier string, skip, take int) ([]Location, int, error)

	// PackageInformation retrieves package information data by its identifier.
	PackageInformation(ctx context.Context, bundleID int, path, packageInformationID string) (PackageInformationData, error)
}

type bundleManagerClientImpl struct {
	codeIntelDB        *sql.DB
	observationContext *observation.Context
}

var _ BundleManagerClient = &bundleManagerClientImpl{}

// Exists determines if the given path exists in the dump.
func (c *bundleManagerClientImpl) Exists(ctx context.Context, bundleID int, path string) (bool, error) {
	db, err := c.openDatabase(ctx, bundleID)
	if err != nil {
		return false, err
	}

	return db.Exists(ctx, bundleID, path)
}

// Ranges returns definition, reference, and hover data for each range within the given span of lines.
func (c *bundleManagerClientImpl) Ranges(ctx context.Context, bundleID int, path string, startLine, endLine int) ([]CodeIntelligenceRange, error) {
	db, err := c.openDatabase(ctx, bundleID)
	if err != nil {
		return nil, err
	}

	return db.Ranges(ctx, bundleID, path, startLine, endLine)
}

// Definitions retrieves a list of definition locations for the symbol under the given location.
func (c *bundleManagerClientImpl) Definitions(ctx context.Context, bundleID int, path string, line, character int) ([]Location, error) {
	db, err := c.openDatabase(ctx, bundleID)
	if err != nil {
		return nil, err
	}

	locations, err := db.Definitions(ctx, bundleID, path, line, character)
	c.addBundleIDToLocations(locations, bundleID)
	return locations, err
}

// Definitions retrieves a list of reference locations for the symbol under the given location.
func (c *bundleManagerClientImpl) References(ctx context.Context, bundleID int, path string, line, character int) ([]Location, error) {
	db, err := c.openDatabase(ctx, bundleID)
	if err != nil {
		return nil, err
	}

	locations, err := db.References(ctx, bundleID, path, line, character)
	c.addBundleIDToLocations(locations, bundleID)
	return locations, err
}

// Hover retrieves the hover text for the symbol under the given location.
func (c *bundleManagerClientImpl) Hover(ctx context.Context, bundleID int, path string, line, character int) (string, Range, bool, error) {
	db, err := c.openDatabase(ctx, bundleID)
	if err != nil {
		return "", Range{}, false, err
	}

	return db.Hover(ctx, bundleID, path, line, character)
}

// Diagnostics retrieves the diagnostics and total count of diagnostics for the documents that have the given path prefix.
func (c *bundleManagerClientImpl) Diagnostics(ctx context.Context, bundleID int, prefix string, skip, take int) ([]Diagnostic, int, error) {
	db, err := c.openDatabase(ctx, bundleID)
	if err != nil {
		return nil, 0, err
	}

	diagnostics, count, err := db.Diagnostics(ctx, bundleID, prefix, skip, take)
	if err != nil {
		return nil, 0, err
	}

	c.addBundleIDToDiagnostics(diagnostics, bundleID)
	return diagnostics, count, err
}

// MonikersByPosition retrieves a list of monikers attached to the symbol under the given location. There may
// be multiple ranges enclosing this point. The returned monikers are partitioned such that inner ranges occur
// first in the result, and outer ranges occur later.
func (c *bundleManagerClientImpl) MonikersByPosition(ctx context.Context, bundleID int, path string, line, character int) ([][]MonikerData, error) {
	db, err := c.openDatabase(ctx, bundleID)
	if err != nil {
		return nil, err
	}

	return db.MonikersByPosition(ctx, bundleID, path, line, character)
}

// MonikerResults retrieves a page of locations attached to a moniker and a total count of such locations.
func (c *bundleManagerClientImpl) MonikerResults(ctx context.Context, bundleID int, modelType, scheme, identifier string, skip, take int) ([]Location, int, error) {
	db, err := c.openDatabase(ctx, bundleID)
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

	locations, count, err := db.MonikerResults(ctx, bundleID, tableName, scheme, identifier, skip, take)
	if err != nil {
		return nil, 0, err
	}

	c.addBundleIDToLocations(locations, bundleID)
	return locations, count, err
}

// PackageInformation retrieves package information data by its identifier.
func (c *bundleManagerClientImpl) PackageInformation(ctx context.Context, bundleID int, path, packageInformationID string) (PackageInformationData, error) {
	db, err := c.openDatabase(ctx, bundleID)
	if err != nil {
		return PackageInformationData{}, err
	}

	data, _, err := db.PackageInformation(ctx, bundleID, path, packageInformationID)
	return data, err
}

func (c *bundleManagerClientImpl) openDatabase(ctx context.Context, bundleID int) (database.Database, error) {
	return database.NewObserved(database.OpenDatabase(persistence.NewObserved(postgres.NewStore(c.codeIntelDB), c.observationContext)), c.observationContext), nil
}

func (c *bundleManagerClientImpl) addBundleIDToLocations(locations []Location, bundleID int) {
	for i := range locations {
		locations[i].DumpID = bundleID
	}
}

func (c *bundleManagerClientImpl) addBundleIDToDiagnostics(diagnostics []Diagnostic, bundleID int) {
	for i := range diagnostics {
		diagnostics[i].DumpID = bundleID
	}
}
