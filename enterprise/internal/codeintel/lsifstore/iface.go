package lsifstore

import (
	"context"
)

// KeyedDocumentData pairs a document with its path.
type KeyedDocumentData struct {
	Path     string
	Document DocumentData
}

// IndexedResultChunkData pairs a result chunk with its index.
type IndexedResultChunkData struct {
	Index       int
	ResultChunk ResultChunkData
}

type Store interface {
	Transact(ctx context.Context) (Store, error)
	Done(err error) error

	// Exists determines if the path exists in the database.
	Exists(ctx context.Context, bundleID int, path string) (bool, error)

	// Ranges returns definition, reference, and hover data for each range within the given span of lines.
	Ranges(ctx context.Context, bundleID int, path string, startLine, endLine int) ([]CodeIntelligenceRange, error)

	// Definitions returns the set of locations defining the symbol at the given position.
	Definitions(ctx context.Context, bundleID int, path string, line, character int) ([]Location, error)

	// References returns the set of locations referencing the symbol at the given position.
	References(ctx context.Context, bundleID int, path string, line, character int) ([]Location, error)

	// Hover returns the hover text of the symbol at the given position.
	Hover(ctx context.Context, bundleID int, path string, line, character int) (string, Range, bool, error)

	// Diagnostics returns the diagnostics for the documents that have the given path prefix. This method
	// also returns the size of the complete result set to aid in pagination (along with skip and take).
	Diagnostics(ctx context.Context, bundleID int, prefix string, skip, take int) ([]Diagnostic, int, error)

	// MonikersByPosition returns all monikers attached ranges containing the given position. If multiple
	// ranges contain the position, then this method will return multiple sets of monikers. Each slice
	// of monikers are attached to a single range. The order of the output slice is "outside-in", so that
	// the range attached to earlier monikers enclose the range attached to later monikers.
	MonikersByPosition(ctx context.Context, bundleID int, path string, line, character int) ([][]MonikerData, error)

	// MonikerResults returns the locations that define or reference the given moniker. This method
	// also returns the size of the complete result set to aid in pagination (along with skip and take).
	MonikerResults(ctx context.Context, bundleID int, tableName, scheme, identifier string, skip, take int) ([]Location, int, error)

	// PackageInformation looks up package information data by identifier.
	PackageInformation(ctx context.Context, bundleID int, path string, packageInformationID string) (PackageInformationData, bool, error)

	Clear(ctx context.Context, bundleIDs ...int) error

	ReadMeta(ctx context.Context, bundleID int) (MetaData, error)
	PathsWithPrefix(ctx context.Context, bundleID int, prefix string) ([]string, error)
	ReadDocument(ctx context.Context, bundleID int, path string) (DocumentData, bool, error)
	ReadResultChunk(ctx context.Context, bundleID int, id int) (ResultChunkData, bool, error)
	ReadDefinitions(ctx context.Context, bundleID int, scheme, identifier string, skip, take int) ([]LocationData, int, error)
	ReadReferences(ctx context.Context, bundleID int, scheme, identifier string, skip, take int) ([]LocationData, int, error)

	WriteMeta(ctx context.Context, bundleID int, meta MetaData) error
	WriteDocuments(ctx context.Context, bundleID int, documents chan KeyedDocumentData) error
	WriteResultChunks(ctx context.Context, bundleID int, resultChunks chan IndexedResultChunkData) error
	WriteDefinitions(ctx context.Context, bundleID int, monikerLocations chan MonikerLocations) error
	WriteReferences(ctx context.Context, bundleID int, monikerLocations chan MonikerLocations) error
}
