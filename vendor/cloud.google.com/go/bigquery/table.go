// Copyright 2015 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bigquery

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cloud.google.com/go/internal/optional"
	"cloud.google.com/go/internal/trace"
	bq "google.golang.org/api/bigquery/v2"
)

// A Table is a reference to a BigQuery table.
type Table struct {
	// ProjectID, DatasetID and TableID may be omitted if the Table is the destination for a query.
	// In this case the result will be stored in an ephemeral table.
	ProjectID string
	DatasetID string
	// TableID must contain only letters (a-z, A-Z), numbers (0-9), or underscores (_).
	// The maximum length is 1,024 characters.
	TableID string

	c *Client
}

// TableMetadata contains information about a BigQuery table.
type TableMetadata struct {
	// The following fields can be set when creating a table.

	// The user-friendly name for the table.
	Name string

	// Output-only location of the table, based on the encapsulating dataset.
	Location string

	// The user-friendly description of the table.
	Description string

	// The table schema. If provided on create, ViewQuery must be empty.
	Schema Schema

	// If non-nil, this table is a materialized view.
	MaterializedView *MaterializedViewDefinition

	// The query to use for a logical view. If provided on create, Schema must be nil.
	ViewQuery string

	// Use Legacy SQL for the view query.
	// At most one of UseLegacySQL and UseStandardSQL can be true.
	UseLegacySQL bool

	// Use Standard SQL for the view query. The default.
	// At most one of UseLegacySQL and UseStandardSQL can be true.
	// Deprecated: use UseLegacySQL.
	UseStandardSQL bool

	// If non-nil, the table is partitioned by time. Only one of
	// time partitioning or range partitioning can be specified.
	TimePartitioning *TimePartitioning

	// If non-nil, the table is partitioned by integer range.  Only one of
	// time partitioning or range partitioning can be specified.
	RangePartitioning *RangePartitioning

	// If set to true, queries that reference this table must specify a
	// partition filter (e.g. a WHERE clause) that can be used to eliminate
	// partitions. Used to prevent unintentional full data scans on large
	// partitioned tables.
	RequirePartitionFilter bool

	// Clustering specifies the data clustering configuration for the table.
	Clustering *Clustering

	// The time when this table expires. If set, this table will expire at the
	// specified time. Expired tables will be deleted and their storage
	// reclaimed. The zero value is ignored.
	ExpirationTime time.Time

	// User-provided labels.
	Labels map[string]string

	// Information about a table stored outside of BigQuery.
	ExternalDataConfig *ExternalDataConfig

	// Custom encryption configuration (e.g., Cloud KMS keys).
	EncryptionConfig *EncryptionConfig

	// All the fields below are read-only.

	FullID           string // An opaque ID uniquely identifying the table.
	Type             TableType
	CreationTime     time.Time
	LastModifiedTime time.Time

	// The size of the table in bytes.
	// This does not include data that is being buffered during a streaming insert.
	NumBytes int64

	// The number of bytes in the table considered "long-term storage" for reduced
	// billing purposes.  See https://cloud.google.com/bigquery/pricing#long-term-storage
	// for more information.
	NumLongTermBytes int64

	// The number of rows of data in this table.
	// This does not include data that is being buffered during a streaming insert.
	NumRows uint64

	// SnapshotDefinition contains additional information about the provenance of a
	// given snapshot table.
	SnapshotDefinition *SnapshotDefinition

	// CloneDefinition contains additional information about the provenance of a
	// given cloned table.
	CloneDefinition *CloneDefinition

	// Contains information regarding this table's streaming buffer, if one is
	// present. This field will be nil if the table is not being streamed to or if
	// there is no data in the streaming buffer.
	StreamingBuffer *StreamingBuffer

	// ETag is the ETag obtained when reading metadata. Pass it to Table.Update to
	// ensure that the metadata hasn't changed since it was read.
	ETag string

	// Defines the default collation specification of new STRING fields
	// in the table. During table creation or update, if a STRING field is added
	// to this table without explicit collation specified, then the table inherits
	// the table default collation. A change to this field affects only fields
	// added afterwards, and does not alter the existing fields.
	// The following values are supported:
	//   - 'und:ci': undetermined locale, case insensitive.
	//   - '': empty string. Default to case-sensitive behavior.
	// More information: https://cloud.google.com/bigquery/docs/reference/standard-sql/collation-concepts
	DefaultCollation string

	// TableConstraints contains table primary and foreign keys constraints.
	// Present only if the table has primary or foreign keys.
	TableConstraints *TableConstraints

	// The tags associated with this table. Tag
	// keys are globally unique. See additional information on tags
	// (https://cloud.google.com/iam/docs/tags-access-control#definitions).
	// An object containing a list of "key": value pairs. The key is the
	// namespaced friendly name of the tag key, e.g. "12345/environment"
	// where 12345 is parent id. The value is the friendly short name of the
	// tag value, e.g. "production".
	ResourceTags map[string]string
}

// TableConstraints defines the primary key and foreign key of a table.
type TableConstraints struct {
	// PrimaryKey constraint on a table's columns.
	// Present only if the table has a primary key.
	// The primary key is not enforced.
	PrimaryKey *PrimaryKey

	// ForeignKeys represent a list of foreign keys constraints.
	// Foreign keys are not enforced.
	ForeignKeys []*ForeignKey
}

// PrimaryKey represents the primary key constraint on a table's columns.
type PrimaryKey struct {
	// Columns that compose the primary key constraint.
	Columns []string
}

func (pk *PrimaryKey) toBQ() *bq.TableConstraintsPrimaryKey {
	return &bq.TableConstraintsPrimaryKey{
		Columns: pk.Columns,
	}
}

func bqToPrimaryKey(tc *bq.TableConstraints) *PrimaryKey {
	if tc.PrimaryKey == nil {
		return nil
	}
	return &PrimaryKey{
		Columns: tc.PrimaryKey.Columns,
	}
}

// ForeignKey represents a foreign key constraint on a table's columns.
type ForeignKey struct {
	// Foreign key constraint name.
	Name string

	// Table that holds the primary key and is referenced by this foreign key.
	ReferencedTable *Table

	// Columns that compose the foreign key.
	ColumnReferences []*ColumnReference
}

func (fk *ForeignKey) toBQ() *bq.TableConstraintsForeignKeys {
	colRefs := []*bq.TableConstraintsForeignKeysColumnReferences{}
	for _, colRef := range fk.ColumnReferences {
		colRefs = append(colRefs, colRef.toBQ())
	}
	return &bq.TableConstraintsForeignKeys{
		Name: fk.Name,
		ReferencedTable: &bq.TableConstraintsForeignKeysReferencedTable{
			DatasetId: fk.ReferencedTable.DatasetID,
			ProjectId: fk.ReferencedTable.ProjectID,
			TableId:   fk.ReferencedTable.TableID,
		},
		ColumnReferences: colRefs,
	}
}

func bqToForeignKeys(tc *bq.TableConstraints, c *Client) []*ForeignKey {
	fks := []*ForeignKey{}
	for _, fk := range tc.ForeignKeys {
		colRefs := []*ColumnReference{}
		for _, colRef := range fk.ColumnReferences {
			colRefs = append(colRefs, &ColumnReference{
				ReferencedColumn:  colRef.ReferencedColumn,
				ReferencingColumn: colRef.ReferencingColumn,
			})
		}
		fks = append(fks, &ForeignKey{
			Name:             fk.Name,
			ReferencedTable:  c.DatasetInProject(fk.ReferencedTable.ProjectId, fk.ReferencedTable.DatasetId).Table(fk.ReferencedTable.TableId),
			ColumnReferences: colRefs,
		})
	}
	return fks
}

// ColumnReference represents the pair of the foreign key column and primary key column.
type ColumnReference struct {
	// ReferencingColumn is the column in the current table that composes the foreign key.
	ReferencingColumn string
	// ReferencedColumn is the column in the primary key of the foreign table that
	// is referenced by the ReferencingColumn.
	ReferencedColumn string
}

func (colRef *ColumnReference) toBQ() *bq.TableConstraintsForeignKeysColumnReferences {
	return &bq.TableConstraintsForeignKeysColumnReferences{
		ReferencedColumn:  colRef.ReferencedColumn,
		ReferencingColumn: colRef.ReferencingColumn,
	}
}

// TableCreateDisposition specifies the circumstances under which destination table will be created.
// Default is CreateIfNeeded.
type TableCreateDisposition string

const (
	// CreateIfNeeded will create the table if it does not already exist.
	// Tables are created atomically on successful completion of a job.
	CreateIfNeeded TableCreateDisposition = "CREATE_IF_NEEDED"

	// CreateNever ensures the table must already exist and will not be
	// automatically created.
	CreateNever TableCreateDisposition = "CREATE_NEVER"
)

// TableWriteDisposition specifies how existing data in a destination table is treated.
// Default is WriteAppend.
type TableWriteDisposition string

const (
	// WriteAppend will append to any existing data in the destination table.
	// Data is appended atomically on successful completion of a job.
	WriteAppend TableWriteDisposition = "WRITE_APPEND"

	// WriteTruncate overrides the existing data in the destination table.
	// Data is overwritten atomically on successful completion of a job.
	WriteTruncate TableWriteDisposition = "WRITE_TRUNCATE"

	// WriteEmpty fails writes if the destination table already contains data.
	WriteEmpty TableWriteDisposition = "WRITE_EMPTY"
)

// TableType is the type of table.
type TableType string

const (
	// RegularTable is a regular table.
	RegularTable TableType = "TABLE"
	// ViewTable is a table type describing that the table is a logical view.
	// See more information at https://cloud.google.com/bigquery/docs/views.
	ViewTable TableType = "VIEW"
	// ExternalTable is a table type describing that the table is an external
	// table (also known as a federated data source). See more information at
	// https://cloud.google.com/bigquery/external-data-sources.
	ExternalTable TableType = "EXTERNAL"
	// MaterializedView represents a managed storage table that's derived from
	// a base table.
	MaterializedView TableType = "MATERIALIZED_VIEW"
	// Snapshot represents an immutable point in time snapshot of some other
	// table.
	Snapshot TableType = "SNAPSHOT"
)

// MaterializedViewDefinition contains information for materialized views.
type MaterializedViewDefinition struct {
	// EnableRefresh governs whether the derived view is updated to reflect
	// changes in the base table.
	EnableRefresh bool

	// LastRefreshTime reports the time, in millisecond precision, that the
	// materialized view was last updated.
	LastRefreshTime time.Time

	// Query contains the SQL query used to define the materialized view.
	Query string

	// RefreshInterval defines the maximum frequency, in millisecond precision,
	// at which this this materialized view will be refreshed.
	RefreshInterval time.Duration

	// AllowNonIncrementalDefinition for materialized view definition.
	// The default value is false.
	AllowNonIncrementalDefinition bool

	// MaxStaleness of data that could be returned when materialized
	// view is queried.
	MaxStaleness *IntervalValue
}

func (mvd *MaterializedViewDefinition) toBQ() *bq.MaterializedViewDefinition {
	if mvd == nil {
		return nil
	}
	maxStaleness := ""
	if mvd.MaxStaleness != nil {
		maxStaleness = mvd.MaxStaleness.String()
	}
	return &bq.MaterializedViewDefinition{
		EnableRefresh:                 mvd.EnableRefresh,
		Query:                         mvd.Query,
		LastRefreshTime:               mvd.LastRefreshTime.UnixNano() / 1e6,
		RefreshIntervalMs:             int64(mvd.RefreshInterval) / 1e6,
		AllowNonIncrementalDefinition: mvd.AllowNonIncrementalDefinition,
		MaxStaleness:                  maxStaleness,
		// force sending the bool in all cases due to how Go handles false.
		ForceSendFields: []string{"EnableRefresh", "AllowNonIncrementalDefinition"},
	}
}

func bqToMaterializedViewDefinition(q *bq.MaterializedViewDefinition) *MaterializedViewDefinition {
	if q == nil {
		return nil
	}
	var maxStaleness *IntervalValue
	if q.MaxStaleness != "" {
		maxStaleness, _ = ParseInterval(q.MaxStaleness)
	}
	return &MaterializedViewDefinition{
		EnableRefresh:                 q.EnableRefresh,
		Query:                         q.Query,
		LastRefreshTime:               unixMillisToTime(q.LastRefreshTime),
		RefreshInterval:               time.Duration(q.RefreshIntervalMs) * time.Millisecond,
		AllowNonIncrementalDefinition: q.AllowNonIncrementalDefinition,
		MaxStaleness:                  maxStaleness,
	}
}

// SnapshotDefinition provides metadata related to the origin of a snapshot.
type SnapshotDefinition struct {

	// BaseTableReference describes the ID of the table that this snapshot
	// came from.
	BaseTableReference *Table

	// SnapshotTime indicates when the base table was snapshot.
	SnapshotTime time.Time
}

func (sd *SnapshotDefinition) toBQ() *bq.SnapshotDefinition {
	if sd == nil {
		return nil
	}
	return &bq.SnapshotDefinition{
		BaseTableReference: sd.BaseTableReference.toBQ(),
		SnapshotTime:       sd.SnapshotTime.Format(time.RFC3339),
	}
}

func bqToSnapshotDefinition(q *bq.SnapshotDefinition, c *Client) *SnapshotDefinition {
	if q == nil {
		return nil
	}
	sd := &SnapshotDefinition{
		BaseTableReference: bqToTable(q.BaseTableReference, c),
	}
	// It's possible we could fail to populate SnapshotTime if we fail to parse
	// the backend representation.
	if t, err := time.Parse(time.RFC3339, q.SnapshotTime); err == nil {
		sd.SnapshotTime = t
	}
	return sd
}

// CloneDefinition provides metadata related to the origin of a clone.
type CloneDefinition struct {

	// BaseTableReference describes the ID of the table that this clone
	// came from.
	BaseTableReference *Table

	// CloneTime indicates when the base table was cloned.
	CloneTime time.Time
}

func (cd *CloneDefinition) toBQ() *bq.CloneDefinition {
	if cd == nil {
		return nil
	}
	return &bq.CloneDefinition{
		BaseTableReference: cd.BaseTableReference.toBQ(),
		CloneTime:          cd.CloneTime.Format(time.RFC3339),
	}
}

func bqToCloneDefinition(q *bq.CloneDefinition, c *Client) *CloneDefinition {
	if q == nil {
		return nil
	}
	cd := &CloneDefinition{
		BaseTableReference: bqToTable(q.BaseTableReference, c),
	}
	// It's possible we could fail to populate CloneTime if we fail to parse
	// the backend representation.
	if t, err := time.Parse(time.RFC3339, q.CloneTime); err == nil {
		cd.CloneTime = t
	}
	return cd
}

// TimePartitioningType defines the interval used to partition managed data.
type TimePartitioningType string

const (
	// DayPartitioningType uses a day-based interval for time partitioning.
	DayPartitioningType TimePartitioningType = "DAY"

	// HourPartitioningType uses an hour-based interval for time partitioning.
	HourPartitioningType TimePartitioningType = "HOUR"

	// MonthPartitioningType uses a month-based interval for time partitioning.
	MonthPartitioningType TimePartitioningType = "MONTH"

	// YearPartitioningType uses a year-based interval for time partitioning.
	YearPartitioningType TimePartitioningType = "YEAR"
)

// TimePartitioning describes the time-based date partitioning on a table.
// For more information see: https://cloud.google.com/bigquery/docs/creating-partitioned-tables.
type TimePartitioning struct {
	// Defines the partition interval type.  Supported values are "HOUR", "DAY", "MONTH", and "YEAR".
	// When the interval type is not specified, default behavior is DAY.
	Type TimePartitioningType

	// The amount of time to keep the storage for a partition.
	// If the duration is empty (0), the data in the partitions do not expire.
	Expiration time.Duration

	// If empty, the table is partitioned by pseudo column '_PARTITIONTIME'; if set, the
	// table is partitioned by this field. The field must be a top-level TIMESTAMP or
	// DATE field. Its mode must be NULLABLE or REQUIRED.
	Field string

	// If set to true, queries that reference this table must specify a
	// partition filter (e.g. a WHERE clause) that can be used to eliminate
	// partitions. Used to prevent unintentional full data scans on large
	// partitioned tables.
	// DEPRECATED: use the top-level RequirePartitionFilter in TableMetadata.
	RequirePartitionFilter bool
}

func (p *TimePartitioning) toBQ() *bq.TimePartitioning {
	if p == nil {
		return nil
	}
	// Treat unspecified values as DAY-based partitioning.
	intervalType := DayPartitioningType
	if p.Type != "" {
		intervalType = p.Type
	}
	return &bq.TimePartitioning{
		Type:                   string(intervalType),
		ExpirationMs:           int64(p.Expiration / time.Millisecond),
		Field:                  p.Field,
		RequirePartitionFilter: p.RequirePartitionFilter,
	}
}

func bqToTimePartitioning(q *bq.TimePartitioning) *TimePartitioning {
	if q == nil {
		return nil
	}
	return &TimePartitioning{
		Type:                   TimePartitioningType(q.Type),
		Expiration:             time.Duration(q.ExpirationMs) * time.Millisecond,
		Field:                  q.Field,
		RequirePartitionFilter: q.RequirePartitionFilter,
	}
}

// RangePartitioning indicates an integer-range based storage organization strategy.
type RangePartitioning struct {
	// The field by which the table is partitioned.
	// This field must be a top-level field, and must be typed as an
	// INTEGER/INT64.
	Field string
	// The details of how partitions are mapped onto the integer range.
	Range *RangePartitioningRange
}

// RangePartitioningRange defines the boundaries and width of partitioned values.
type RangePartitioningRange struct {
	// The start value of defined range of values, inclusive of the specified value.
	Start int64
	// The end of the defined range of values, exclusive of the defined value.
	End int64
	// The width of each interval range.
	Interval int64
}

func (rp *RangePartitioning) toBQ() *bq.RangePartitioning {
	if rp == nil {
		return nil
	}
	return &bq.RangePartitioning{
		Field: rp.Field,
		Range: rp.Range.toBQ(),
	}
}

func bqToRangePartitioning(q *bq.RangePartitioning) *RangePartitioning {
	if q == nil {
		return nil
	}
	return &RangePartitioning{
		Field: q.Field,
		Range: bqToRangePartitioningRange(q.Range),
	}
}

func bqToRangePartitioningRange(br *bq.RangePartitioningRange) *RangePartitioningRange {
	if br == nil {
		return nil
	}
	return &RangePartitioningRange{
		Start:    br.Start,
		End:      br.End,
		Interval: br.Interval,
	}
}

func (rpr *RangePartitioningRange) toBQ() *bq.RangePartitioningRange {
	if rpr == nil {
		return nil
	}
	return &bq.RangePartitioningRange{
		Start:           rpr.Start,
		End:             rpr.End,
		Interval:        rpr.Interval,
		ForceSendFields: []string{"Start", "End", "Interval"},
	}
}

// Clustering governs the organization of data within a managed table.
// For more information, see https://cloud.google.com/bigquery/docs/clustered-tables
type Clustering struct {
	Fields []string
}

func (c *Clustering) toBQ() *bq.Clustering {
	if c == nil {
		return nil
	}
	return &bq.Clustering{
		Fields: c.Fields,
	}
}

func bqToClustering(q *bq.Clustering) *Clustering {
	if q == nil {
		return nil
	}
	return &Clustering{
		Fields: q.Fields,
	}
}

// EncryptionConfig configures customer-managed encryption on tables and ML models.
type EncryptionConfig struct {
	// Describes the Cloud KMS encryption key that will be used to protect
	// destination BigQuery table. The BigQuery Service Account associated with your
	// project requires access to this encryption key.
	KMSKeyName string
}

func (e *EncryptionConfig) toBQ() *bq.EncryptionConfiguration {
	if e == nil {
		return nil
	}
	return &bq.EncryptionConfiguration{
		KmsKeyName: e.KMSKeyName,
	}
}

func bqToEncryptionConfig(q *bq.EncryptionConfiguration) *EncryptionConfig {
	if q == nil {
		return nil
	}
	return &EncryptionConfig{
		KMSKeyName: q.KmsKeyName,
	}
}

// StreamingBuffer holds information about the streaming buffer.
type StreamingBuffer struct {
	// A lower-bound estimate of the number of bytes currently in the streaming
	// buffer.
	EstimatedBytes uint64

	// A lower-bound estimate of the number of rows currently in the streaming
	// buffer.
	EstimatedRows uint64

	// The time of the oldest entry in the streaming buffer.
	OldestEntryTime time.Time
}

func (t *Table) toBQ() *bq.TableReference {
	return &bq.TableReference{
		ProjectId: t.ProjectID,
		DatasetId: t.DatasetID,
		TableId:   t.TableID,
	}
}

// IdentifierFormat represents a how certain resource identifiers such as table references
// are formatted.
type IdentifierFormat string

var (
	// StandardSQLID returns an identifier suitable for use with Standard SQL.
	StandardSQLID IdentifierFormat = "SQL"

	// LegacySQLID returns an identifier suitable for use with Legacy SQL.
	LegacySQLID IdentifierFormat = "LEGACY_SQL"

	// StorageAPIResourceID returns an identifier suitable for use with the Storage API.  Namely, it's for formatting
	// a table resource for invoking read and write functionality.
	StorageAPIResourceID IdentifierFormat = "STORAGE_API_RESOURCE"

	// ErrUnknownIdentifierFormat is indicative of requesting an identifier in a format that is
	// not supported.
	ErrUnknownIdentifierFormat = errors.New("unknown identifier format")
)

// Identifier returns the ID of the table in the requested format.
func (t *Table) Identifier(f IdentifierFormat) (string, error) {
	switch f {
	case LegacySQLID:
		return fmt.Sprintf("%s:%s.%s", t.ProjectID, t.DatasetID, t.TableID), nil
	case StorageAPIResourceID:
		return fmt.Sprintf("projects/%s/datasets/%s/tables/%s", t.ProjectID, t.DatasetID, t.TableID), nil
	case StandardSQLID:
		// Note we don't need to quote the project ID here, as StandardSQL has special rules to allow
		// dash identifiers for projects without issue in table identifiers.
		return fmt.Sprintf("%s.%s.%s", t.ProjectID, t.DatasetID, t.TableID), nil
	default:
		return "", ErrUnknownIdentifierFormat
	}
}

// FullyQualifiedName returns the ID of the table in projectID:datasetID.tableID format.
func (t *Table) FullyQualifiedName() string {
	s, _ := t.Identifier(LegacySQLID)
	return s
}

// implicitTable reports whether Table is an empty placeholder, which signifies that a new table should be created with an auto-generated Table ID.
func (t *Table) implicitTable() bool {
	return t.ProjectID == "" && t.DatasetID == "" && t.TableID == ""
}

// Create creates a table in the BigQuery service.
// Pass in a TableMetadata value to configure the table.
// If tm.View.Query is non-empty, the created table will be of type VIEW.
// If no ExpirationTime is specified, the table will never expire.
// After table creation, a view can be modified only if its table was initially created
// with a view.
func (t *Table) Create(ctx context.Context, tm *TableMetadata) (err error) {
	ctx = trace.StartSpan(ctx, "cloud.google.com/go/bigquery.Table.Create")
	defer func() { trace.EndSpan(ctx, err) }()

	table, err := tm.toBQ()
	if err != nil {
		return err
	}
	table.TableReference = &bq.TableReference{
		ProjectId: t.ProjectID,
		DatasetId: t.DatasetID,
		TableId:   t.TableID,
	}

	req := t.c.bqs.Tables.Insert(t.ProjectID, t.DatasetID, table).Context(ctx)
	setClientHeader(req.Header())
	return runWithRetry(ctx, func() (err error) {
		ctx = trace.StartSpan(ctx, "bigquery.tables.insert")
		_, err = req.Do()
		trace.EndSpan(ctx, err)
		return err
	})
}

func (tm *TableMetadata) toBQ() (*bq.Table, error) {
	t := &bq.Table{}
	if tm == nil {
		return t, nil
	}
	if tm.Schema != nil && tm.ViewQuery != "" {
		return nil, errors.New("bigquery: provide Schema or ViewQuery, not both")
	}
	t.FriendlyName = tm.Name
	t.Description = tm.Description
	t.Labels = tm.Labels
	if tm.Schema != nil {
		t.Schema = tm.Schema.toBQ()
	}
	if tm.ViewQuery != "" {
		if tm.UseStandardSQL && tm.UseLegacySQL {
			return nil, errors.New("bigquery: cannot provide both UseStandardSQL and UseLegacySQL")
		}
		t.View = &bq.ViewDefinition{Query: tm.ViewQuery}
		if tm.UseLegacySQL {
			t.View.UseLegacySql = true
		} else {
			t.View.UseLegacySql = false
			t.View.ForceSendFields = append(t.View.ForceSendFields, "UseLegacySql")
		}
	} else if tm.UseLegacySQL || tm.UseStandardSQL {
		return nil, errors.New("bigquery: UseLegacy/StandardSQL requires ViewQuery")
	}
	t.MaterializedView = tm.MaterializedView.toBQ()
	t.TimePartitioning = tm.TimePartitioning.toBQ()
	t.RangePartitioning = tm.RangePartitioning.toBQ()
	t.Clustering = tm.Clustering.toBQ()
	t.RequirePartitionFilter = tm.RequirePartitionFilter
	t.SnapshotDefinition = tm.SnapshotDefinition.toBQ()
	t.CloneDefinition = tm.CloneDefinition.toBQ()

	if !validExpiration(tm.ExpirationTime) {
		return nil, fmt.Errorf("invalid expiration time: %v.\n"+
			"Valid expiration times are after 1678 and before 2262", tm.ExpirationTime)
	}
	if !tm.ExpirationTime.IsZero() && tm.ExpirationTime != NeverExpire {
		t.ExpirationTime = tm.ExpirationTime.UnixNano() / 1e6
	}
	if tm.ExternalDataConfig != nil {
		edc := tm.ExternalDataConfig.toBQ()
		t.ExternalDataConfiguration = &edc
	}
	t.EncryptionConfiguration = tm.EncryptionConfig.toBQ()
	if tm.FullID != "" {
		return nil, errors.New("cannot set FullID on create")
	}
	if tm.Type != "" {
		return nil, errors.New("cannot set Type on create")
	}
	if !tm.CreationTime.IsZero() {
		return nil, errors.New("cannot set CreationTime on create")
	}
	if !tm.LastModifiedTime.IsZero() {
		return nil, errors.New("cannot set LastModifiedTime on create")
	}
	if tm.NumBytes != 0 {
		return nil, errors.New("cannot set NumBytes on create")
	}
	if tm.NumLongTermBytes != 0 {
		return nil, errors.New("cannot set NumLongTermBytes on create")
	}
	if tm.NumRows != 0 {
		return nil, errors.New("cannot set NumRows on create")
	}
	if tm.StreamingBuffer != nil {
		return nil, errors.New("cannot set StreamingBuffer on create")
	}
	if tm.ETag != "" {
		return nil, errors.New("cannot set ETag on create")
	}
	t.DefaultCollation = string(tm.DefaultCollation)

	if tm.TableConstraints != nil {
		t.TableConstraints = &bq.TableConstraints{}
		if tm.TableConstraints.PrimaryKey != nil {
			t.TableConstraints.PrimaryKey = tm.TableConstraints.PrimaryKey.toBQ()
		}
		if len(tm.TableConstraints.ForeignKeys) > 0 {
			t.TableConstraints.ForeignKeys = make([]*bq.TableConstraintsForeignKeys, len(tm.TableConstraints.ForeignKeys))
			for i, fk := range tm.TableConstraints.ForeignKeys {
				t.TableConstraints.ForeignKeys[i] = fk.toBQ()
			}
		}
	}
	if tm.ResourceTags != nil {
		t.ResourceTags = make(map[string]string)
		for k, v := range tm.ResourceTags {
			t.ResourceTags[k] = v
		}
	}
	return t, nil
}

// We use this for the option pattern rather than exposing the underlying
// discovery type directly.
type tableGetCall struct {
	call *bq.TablesGetCall
}

// TableMetadataOption allow requests to alter requests for table metadata.
type TableMetadataOption func(*tableGetCall)

// TableMetadataView specifies which details about a table are desired.
type TableMetadataView string

const (
	// BasicMetadataView populates basic table information including schema partitioning,
	// but does not contain storage statistics like number or rows or bytes.  This is a more
	// efficient view to use for large tables or higher metadata query rates.
	BasicMetadataView TableMetadataView = "BASIC"

	// FullMetadataView returns all table information, including storage statistics.  It currently
	// returns the same information as StorageStatsMetadataView, but may include additional information
	// in the future.
	FullMetadataView TableMetadataView = "FULL"

	// StorageStatsMetadataView includes all information from the basic view, and includes storage statistics.  It currently
	StorageStatsMetadataView TableMetadataView = "STORAGE_STATS"
)

// WithMetadataView is used to customize what details are returned when interrogating a
// table via the Metadata() call.  Generally this is used to limit data returned for performance
// reasons (such as large tables that take time computing storage statistics).
func WithMetadataView(tmv TableMetadataView) TableMetadataOption {
	return func(tgc *tableGetCall) {
		tgc.call.View(string(tmv))
	}
}

// Metadata fetches the metadata for the table.
func (t *Table) Metadata(ctx context.Context, opts ...TableMetadataOption) (md *TableMetadata, err error) {
	ctx = trace.StartSpan(ctx, "cloud.google.com/go/bigquery.Table.Metadata")
	defer func() { trace.EndSpan(ctx, err) }()

	tgc := &tableGetCall{
		call: t.c.bqs.Tables.Get(t.ProjectID, t.DatasetID, t.TableID).Context(ctx),
	}

	for _, o := range opts {
		o(tgc)
	}

	setClientHeader(tgc.call.Header())
	var res *bq.Table
	if err := runWithRetry(ctx, func() (err error) {
		sCtx := trace.StartSpan(ctx, "bigquery.tables.get")
		res, err = tgc.call.Do()
		trace.EndSpan(sCtx, err)
		return err
	}); err != nil {
		return nil, err
	}
	return bqToTableMetadata(res, t.c)
}

func bqToTableMetadata(t *bq.Table, c *Client) (*TableMetadata, error) {
	md := &TableMetadata{
		Description:            t.Description,
		Name:                   t.FriendlyName,
		Location:               t.Location,
		Type:                   TableType(t.Type),
		FullID:                 t.Id,
		Labels:                 t.Labels,
		NumBytes:               t.NumBytes,
		NumLongTermBytes:       t.NumLongTermBytes,
		NumRows:                t.NumRows,
		ExpirationTime:         unixMillisToTime(t.ExpirationTime),
		CreationTime:           unixMillisToTime(t.CreationTime),
		LastModifiedTime:       unixMillisToTime(int64(t.LastModifiedTime)),
		ETag:                   t.Etag,
		DefaultCollation:       t.DefaultCollation,
		EncryptionConfig:       bqToEncryptionConfig(t.EncryptionConfiguration),
		RequirePartitionFilter: t.RequirePartitionFilter,
		SnapshotDefinition:     bqToSnapshotDefinition(t.SnapshotDefinition, c),
		CloneDefinition:        bqToCloneDefinition(t.CloneDefinition, c),
	}
	if t.MaterializedView != nil {
		md.MaterializedView = bqToMaterializedViewDefinition(t.MaterializedView)
	}
	if t.Schema != nil {
		md.Schema = bqToSchema(t.Schema)
	}
	if t.View != nil {
		md.ViewQuery = t.View.Query
		md.UseLegacySQL = t.View.UseLegacySql
	}
	md.TimePartitioning = bqToTimePartitioning(t.TimePartitioning)
	md.RangePartitioning = bqToRangePartitioning(t.RangePartitioning)
	md.Clustering = bqToClustering(t.Clustering)
	if t.StreamingBuffer != nil {
		md.StreamingBuffer = &StreamingBuffer{
			EstimatedBytes:  t.StreamingBuffer.EstimatedBytes,
			EstimatedRows:   t.StreamingBuffer.EstimatedRows,
			OldestEntryTime: unixMillisToTime(int64(t.StreamingBuffer.OldestEntryTime)),
		}
	}
	if t.ExternalDataConfiguration != nil {
		edc, err := bqToExternalDataConfig(t.ExternalDataConfiguration)
		if err != nil {
			return nil, err
		}
		md.ExternalDataConfig = edc
	}
	if t.TableConstraints != nil {
		md.TableConstraints = &TableConstraints{
			PrimaryKey:  bqToPrimaryKey(t.TableConstraints),
			ForeignKeys: bqToForeignKeys(t.TableConstraints, c),
		}
	}
	if t.ResourceTags != nil {
		md.ResourceTags = make(map[string]string)
		for k, v := range t.ResourceTags {
			md.ResourceTags[k] = v
		}
	}
	return md, nil
}

// Delete deletes the table.
func (t *Table) Delete(ctx context.Context) (err error) {
	ctx = trace.StartSpan(ctx, "cloud.google.com/go/bigquery.Table.Delete")
	defer func() { trace.EndSpan(ctx, err) }()

	call := t.c.bqs.Tables.Delete(t.ProjectID, t.DatasetID, t.TableID).Context(ctx)
	setClientHeader(call.Header())

	return runWithRetry(ctx, func() (err error) {
		ctx = trace.StartSpan(ctx, "bigquery.tables.delete")
		err = call.Do()
		trace.EndSpan(ctx, err)
		return err
	})
}

// Read fetches the contents of the table.
func (t *Table) Read(ctx context.Context) *RowIterator {
	return t.read(ctx, fetchPage)
}

func (t *Table) read(ctx context.Context, pf pageFetcher) *RowIterator {
	if t.c.isStorageReadAvailable() {
		it, err := newStorageRowIteratorFromTable(ctx, t, false)
		if err == nil {
			return it
		}
	}
	return newRowIterator(ctx, &rowSource{t: t}, pf)
}

// NeverExpire is a sentinel value used to remove a table'e expiration time.
var NeverExpire = time.Time{}.Add(-1)

// We use this for the option pattern rather than exposing the underlying
// discovery type directly.
type tablePatchCall struct {
	call *bq.TablesPatchCall
}

// TableUpdateOption allow requests to update table metadata.
type TableUpdateOption func(*tablePatchCall)

// WithAutoDetectSchema governs whether the schema autodetection occurs as part of the table update.
// This is relevant in cases like external tables where schema is detected from the source data.
func WithAutoDetectSchema(b bool) TableUpdateOption {
	return func(tpc *tablePatchCall) {
		tpc.call.AutodetectSchema(b)
	}
}

// Update modifies specific Table metadata fields.
func (t *Table) Update(ctx context.Context, tm TableMetadataToUpdate, etag string, opts ...TableUpdateOption) (md *TableMetadata, err error) {
	ctx = trace.StartSpan(ctx, "cloud.google.com/go/bigquery.Table.Update")
	defer func() { trace.EndSpan(ctx, err) }()

	bqt, err := tm.toBQ()
	if err != nil {
		return nil, err
	}

	tpc := &tablePatchCall{
		call: t.c.bqs.Tables.Patch(t.ProjectID, t.DatasetID, t.TableID, bqt).Context(ctx),
	}

	for _, o := range opts {
		o(tpc)
	}

	setClientHeader(tpc.call.Header())
	if etag != "" {
		tpc.call.Header().Set("If-Match", etag)
	}
	var res *bq.Table
	if err := runWithRetry(ctx, func() (err error) {
		ctx = trace.StartSpan(ctx, "bigquery.tables.patch")
		res, err = tpc.call.Do()
		trace.EndSpan(ctx, err)
		return err
	}); err != nil {
		return nil, err
	}
	return bqToTableMetadata(res, t.c)
}

func (tm *TableMetadataToUpdate) toBQ() (*bq.Table, error) {
	t := &bq.Table{}
	forceSend := func(field string) {
		t.ForceSendFields = append(t.ForceSendFields, field)
	}

	if tm.Description != nil {
		t.Description = optional.ToString(tm.Description)
		forceSend("Description")
	}
	if tm.Name != nil {
		t.FriendlyName = optional.ToString(tm.Name)
		forceSend("FriendlyName")
	}
	if tm.MaterializedView != nil {
		t.MaterializedView = tm.MaterializedView.toBQ()
		forceSend("MaterializedView")
	}
	if tm.Schema != nil {
		t.Schema = tm.Schema.toBQ()
		forceSend("Schema")
	}
	if tm.EncryptionConfig != nil {
		t.EncryptionConfiguration = tm.EncryptionConfig.toBQ()
	}
	if tm.ExternalDataConfig != nil {
		cfg := tm.ExternalDataConfig.toBQ()
		t.ExternalDataConfiguration = &cfg
	}

	if tm.Clustering != nil {
		t.Clustering = tm.Clustering.toBQ()
	}

	if !validExpiration(tm.ExpirationTime) {
		return nil, invalidTimeError(tm.ExpirationTime)
	}
	if tm.ExpirationTime == NeverExpire {
		t.NullFields = append(t.NullFields, "ExpirationTime")
	} else if !tm.ExpirationTime.IsZero() {
		t.ExpirationTime = tm.ExpirationTime.UnixNano() / 1e6
		forceSend("ExpirationTime")
	}
	if tm.TimePartitioning != nil {
		t.TimePartitioning = tm.TimePartitioning.toBQ()
		t.TimePartitioning.ForceSendFields = []string{"RequirePartitionFilter"}
		if tm.TimePartitioning.Expiration == 0 {
			t.TimePartitioning.NullFields = []string{"ExpirationMs"}
		}
	}
	if tm.RequirePartitionFilter != nil {
		t.RequirePartitionFilter = optional.ToBool(tm.RequirePartitionFilter)
		forceSend("RequirePartitionFilter")
	}
	if tm.ViewQuery != nil {
		t.View = &bq.ViewDefinition{
			Query:           optional.ToString(tm.ViewQuery),
			ForceSendFields: []string{"Query"},
		}
	}
	if tm.UseLegacySQL != nil {
		if t.View == nil {
			t.View = &bq.ViewDefinition{}
		}
		t.View.UseLegacySql = optional.ToBool(tm.UseLegacySQL)
		t.View.ForceSendFields = append(t.View.ForceSendFields, "UseLegacySql")
	}
	if tm.DefaultCollation != nil {
		t.DefaultCollation = optional.ToString(tm.DefaultCollation)
		forceSend("DefaultCollation")
	}
	if tm.TableConstraints != nil {
		t.TableConstraints = &bq.TableConstraints{}
		if tm.TableConstraints.PrimaryKey != nil {
			t.TableConstraints.PrimaryKey = tm.TableConstraints.PrimaryKey.toBQ()
			t.TableConstraints.PrimaryKey.ForceSendFields = append(t.TableConstraints.PrimaryKey.ForceSendFields, "Columns")
			t.TableConstraints.ForceSendFields = append(t.TableConstraints.ForceSendFields, "PrimaryKey")
		}
		if tm.TableConstraints.ForeignKeys != nil {
			t.TableConstraints.ForeignKeys = make([]*bq.TableConstraintsForeignKeys, len(tm.TableConstraints.ForeignKeys))
			for i, fk := range tm.TableConstraints.ForeignKeys {
				t.TableConstraints.ForeignKeys[i] = fk.toBQ()
			}
			t.TableConstraints.ForceSendFields = append(t.TableConstraints.ForceSendFields, "ForeignKeys")
		}
	}
	if tm.ResourceTags != nil {
		t.ResourceTags = make(map[string]string)
		for k, v := range tm.ResourceTags {
			t.ResourceTags[k] = v
		}
		forceSend("ResourceTags")
	}
	labels, forces, nulls := tm.update()
	t.Labels = labels
	t.ForceSendFields = append(t.ForceSendFields, forces...)
	t.NullFields = append(t.NullFields, nulls...)
	return t, nil
}

// validExpiration ensures a specified time is either the sentinel NeverExpire,
// the zero value, or within the defined range of UnixNano. Internal
// represetations of expiration times are based upon Time.UnixNano. Any time
// before 1678 or after 2262 cannot be represented by an int64 and is therefore
// undefined and invalid. See https://godoc.org/time#Time.UnixNano.
func validExpiration(t time.Time) bool {
	return t == NeverExpire || t.IsZero() || time.Unix(0, t.UnixNano()).Equal(t)
}

// invalidTimeError emits a consistent error message for failures of the
// validExpiration function.
func invalidTimeError(t time.Time) error {
	return fmt.Errorf("invalid expiration time %v. "+
		"Valid expiration times are after 1678 and before 2262", t)
}

// TableMetadataToUpdate is used when updating a table's metadata.
// Only non-nil fields will be updated.
type TableMetadataToUpdate struct {
	// The user-friendly description of this table.
	Description optional.String

	// The user-friendly name for this table.
	Name optional.String

	// The table's schema.
	// When updating a schema, you can add columns but not remove them.
	Schema Schema

	// The table's clustering configuration.
	// For more information on how modifying clustering affects the table, see:
	// https://cloud.google.com/bigquery/docs/creating-clustered-tables#modifying-cluster-spec
	Clustering *Clustering

	// The table's encryption configuration.
	EncryptionConfig *EncryptionConfig

	// The time when this table expires. To remove a table's expiration,
	// set ExpirationTime to NeverExpire. The zero value is ignored.
	ExpirationTime time.Time

	// ExternalDataConfig controls the definition of a table defined against
	// an external source, such as one based on files in Google Cloud Storage.
	ExternalDataConfig *ExternalDataConfig

	// The query to use for a view.
	ViewQuery optional.String

	// Use Legacy SQL for the view query.
	UseLegacySQL optional.Bool

	// MaterializedView allows changes to the underlying materialized view
	// definition. When calling Update, ensure that all mutable fields of
	// MaterializedViewDefinition are populated.
	MaterializedView *MaterializedViewDefinition

	// TimePartitioning allows modification of certain aspects of partition
	// configuration such as partition expiration and whether partition
	// filtration is required at query time.  When calling Update, ensure
	// that all mutable fields of TimePartitioning are populated.
	TimePartitioning *TimePartitioning

	// RequirePartitionFilter governs whether the table enforces partition
	// elimination when referenced in a query.
	RequirePartitionFilter optional.Bool

	// Defines the default collation specification of new STRING fields
	// in the table.
	DefaultCollation optional.String

	// TableConstraints allows modification of table constraints
	// such as primary and foreign keys.
	TableConstraints *TableConstraints

	// The tags associated with this table. Tag
	// keys are globally unique. See additional information on tags
	// (https://cloud.google.com/iam/docs/tags-access-control#definitions).
	// An object containing a list of "key": value pairs. The key is the
	// namespaced friendly name of the tag key, e.g. "12345/environment"
	// where 12345 is parent id. The value is the friendly short name of the
	// tag value, e.g. "production".
	ResourceTags map[string]string

	labelUpdater
}

// labelUpdater contains common code for updating labels.
type labelUpdater struct {
	setLabels    map[string]string
	deleteLabels map[string]bool
}

// SetLabel causes a label to be added or modified on a call to Update.
func (u *labelUpdater) SetLabel(name, value string) {
	if u.setLabels == nil {
		u.setLabels = map[string]string{}
	}
	u.setLabels[name] = value
}

// DeleteLabel causes a label to be deleted on a call to Update.
func (u *labelUpdater) DeleteLabel(name string) {
	if u.deleteLabels == nil {
		u.deleteLabels = map[string]bool{}
	}
	u.deleteLabels[name] = true
}

func (u *labelUpdater) update() (labels map[string]string, forces, nulls []string) {
	if u.setLabels == nil && u.deleteLabels == nil {
		return nil, nil, nil
	}
	labels = map[string]string{}
	for k, v := range u.setLabels {
		labels[k] = v
	}
	if len(labels) == 0 && len(u.deleteLabels) > 0 {
		forces = []string{"Labels"}
	}
	for l := range u.deleteLabels {
		nulls = append(nulls, "Labels."+l)
	}
	return labels, forces, nulls
}
