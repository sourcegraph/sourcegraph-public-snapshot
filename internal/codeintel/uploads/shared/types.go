package shared

import (
	"database/sql"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

// Upload is a subset of the lsif_uploads table and stores both processed and unprocessed
// records.
type Upload struct {
	ID                int
	Commit            string
	Root              string
	VisibleAtTip      bool
	UploadedAt        time.Time
	State             string
	FailureMessage    *string
	StartedAt         *time.Time
	FinishedAt        *time.Time
	ProcessAfter      *time.Time
	NumResets         int
	NumFailures       int
	RepositoryID      int
	RepositoryName    string
	Indexer           string
	IndexerVersion    string
	NumParts          int
	UploadedParts     []int
	UploadSize        *int64
	UncompressedSize  *int64
	Rank              *int
	AssociatedIndexID *int
}

// Dump is a subset of the lsif_uploads table (queried via the lsif_dumps_with_repository_name view)
// and stores only processed records.
type Dump struct {
	ID                int
	Commit            string
	Root              string
	VisibleAtTip      bool
	UploadedAt        time.Time
	State             string
	FailureMessage    *string
	StartedAt         *time.Time
	FinishedAt        *time.Time
	ProcessAfter      *time.Time
	NumResets         int
	NumFailures       int
	RepositoryID      int
	RepositoryName    string
	Indexer           string
	IndexerVersion    string
	AssociatedIndexID *int
}

type SourcedCommits struct {
	RepositoryID   int
	RepositoryName string
	Commits        []string
}

type GetUploadsOptions struct {
	RepositoryID            int
	State                   string
	Term                    string
	VisibleAtTip            bool
	DependencyOf            int
	DependentOf             int
	UploadedBefore          *time.Time
	UploadedAfter           *time.Time
	LastRetentionScanBefore *time.Time
	AllowExpired            bool
	AllowDeletedRepo        bool
	AllowDeletedUpload      bool
	OldestFirst             bool
	Limit                   int
	Offset                  int

	// InCommitGraph ensures that the repository commit graph was updated strictly
	// after this upload was processed. This condition helps us filter out new uploads
	// that we might later mistake for unreachable.
	InCommitGraph bool
}

type DependencyReferenceCountUpdateType int

const (
	DependencyReferenceCountUpdateTypeNone DependencyReferenceCountUpdateType = iota
	DependencyReferenceCountUpdateTypeAdd
	DependencyReferenceCountUpdateTypeRemove
)

type CursorAdjustedUpload struct {
	DumpID               int      `json:"dumpID"`
	AdjustedPath         string   `json:"adjustedPath"`
	AdjustedPosition     Position `json:"adjustedPosition"`
	AdjustedPathInBundle string   `json:"adjustedPathInBundle"`
}

// AdjustedUpload pairs an upload visible from the current target commit with the
// current target path and position adjusted so that it matches the data within the
// underlying index.
type AdjustedUpload struct {
	Upload               Dump
	AdjustedPath         string
	AdjustedPosition     lsifstore.Position
	AdjustedPathInBundle string
}

// Position is a unique position within a file.
type Position struct {
	Line      int
	Character int
}

// Package pairs a package schem+name+version with the dump that provides it.
type Package struct {
	DumpID  int
	Scheme  string
	Name    string
	Version string
}

// PackageReference is a package scheme+name+version
type PackageReference struct {
	Package
}

// PackageReferenceScanner allows for on-demand scanning of PackageReference values.
//
// A scanner for this type was introduced as a memory optimization. Instead of reading a
// large number of large byte arrays into memory at once, we allow the user to request
// the next filter value when they are ready to process it. This allows us to hold only
// a single bloom filter in memory at any give time during reference requests.
type PackageReferenceScanner interface {
	// Next reads the next package reference value from the database cursor.
	Next() (PackageReference, bool, error)

	// Close the underlying row object.
	Close() error
}

type rowScanner struct {
	rows *sql.Rows
}

// packageReferenceScannerFromRows creates a PackageReferenceScanner that feeds the given values.
func PackageReferenceScannerFromRows(rows *sql.Rows) PackageReferenceScanner {
	return &rowScanner{
		rows: rows,
	}
}

// Next reads the next package reference value from the database cursor.
func (s *rowScanner) Next() (reference PackageReference, _ bool, _ error) {
	if !s.rows.Next() {
		return PackageReference{}, false, nil
	}

	if err := s.rows.Scan(
		&reference.DumpID,
		&reference.Scheme,
		&reference.Name,
		&reference.Version,
	); err != nil {
		return PackageReference{}, false, err
	}

	return reference, true, nil
}

// Close the underlying row object.
func (s *rowScanner) Close() error {
	return basestore.CloseRows(s.rows, nil)
}

type sliceScanner struct {
	references []PackageReference
}

// PackageReferenceScannerFromSlice creates a PackageReferenceScanner that feeds the given values.
func PackageReferenceScannerFromSlice(references ...PackageReference) PackageReferenceScanner {
	return &sliceScanner{
		references: references,
	}
}

func (s *sliceScanner) Next() (PackageReference, bool, error) {
	if len(s.references) == 0 {
		return PackageReference{}, false, nil
	}

	next := s.references[0]
	s.references = s.references[1:]
	return next, true, nil
}

func (s *sliceScanner) Close() error {
	return nil
}

type UploadsWithRepositoryNamespace struct {
	Root    string
	Indexer string
	Uploads []Upload
}

type UploadLog struct {
	LogTimestamp      time.Time
	RecordDeletedAt   *time.Time
	UploadID          int
	Commit            string
	Root              string
	RepositoryID      int
	UploadedAt        time.Time
	Indexer           string
	IndexerVersion    *string
	UploadSize        *int
	AssociatedIndexID *int
	TransitionColumns []map[string]*string
	Reason            *string
	Operation         string
}
