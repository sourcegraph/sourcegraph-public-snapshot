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
	"strings"
	"time"

	"cloud.google.com/go/internal/optional"
	"cloud.google.com/go/internal/trace"
	bq "google.golang.org/api/bigquery/v2"
	"google.golang.org/api/iterator"
)

// Dataset is a reference to a BigQuery dataset.
type Dataset struct {
	ProjectID string
	DatasetID string
	c         *Client
}

// DatasetMetadata contains information about a BigQuery dataset.
type DatasetMetadata struct {
	// These fields can be set when creating a dataset.
	Name                    string            // The user-friendly name for this dataset.
	Description             string            // The user-friendly description of this dataset.
	Location                string            // The geo location of the dataset.
	DefaultTableExpiration  time.Duration     // The default expiration time for new tables.
	Labels                  map[string]string // User-provided labels.
	Access                  []*AccessEntry    // Access permissions.
	DefaultEncryptionConfig *EncryptionConfig

	// DefaultPartitionExpiration is the default expiration time for
	// all newly created partitioned tables in the dataset.
	DefaultPartitionExpiration time.Duration

	// Defines the default collation specification of future tables
	// created in the dataset. If a table is created in this dataset without
	// table-level default collation, then the table inherits the dataset default
	// collation, which is applied to the string fields that do not have explicit
	// collation specified. A change to this field affects only tables created
	// afterwards, and does not alter the existing tables.
	// More information: https://cloud.google.com/bigquery/docs/reference/standard-sql/collation-concepts
	DefaultCollation string

	// For externally defined datasets, contains information about the configuration.
	ExternalDatasetReference *ExternalDatasetReference

	// MaxTimeTravel represents the number of hours for the max time travel for all tables
	// in the dataset.  Durations are rounded towards zero for the nearest hourly value.
	MaxTimeTravel time.Duration

	// Storage billing model to be used for all tables in the dataset.
	// Can be set to PHYSICAL. Default is LOGICAL.
	// Once you create a dataset with storage billing model set to physical bytes, you can't change it back to using logical bytes again.
	// More details: https://cloud.google.com/bigquery/docs/datasets-intro#dataset_storage_billing_models
	StorageBillingModel string

	// These fields are read-only.
	CreationTime     time.Time
	LastModifiedTime time.Time // When the dataset or any of its tables were modified.
	FullID           string    // The full dataset ID in the form projectID:datasetID.

	// The tags associated with this dataset. Tag keys are
	// globally unique, and managed via the resource manager API.
	// More information: https://cloud.google.com/resource-manager/docs/tags/tags-overview
	Tags []*DatasetTag

	// ETag is the ETag obtained when reading metadata. Pass it to Dataset.Update to
	// ensure that the metadata hasn't changed since it was read.
	ETag string
}

// DatasetTag is a representation of a single tag key/value.
type DatasetTag struct {
	// TagKey is the namespaced friendly name of the tag key, e.g.
	// "12345/environment" where 12345 is org id.
	TagKey string

	// TagValue is the friendly short name of the tag value, e.g.
	// "production".
	TagValue string
}

const (
	// LogicalStorageBillingModel indicates billing for logical bytes.
	LogicalStorageBillingModel = ""

	// PhysicalStorageBillingModel indicates billing for physical bytes.
	PhysicalStorageBillingModel = "PHYSICAL"
)

func bqToDatasetTag(in *bq.DatasetTags) *DatasetTag {
	if in == nil {
		return nil
	}
	return &DatasetTag{
		TagKey:   in.TagKey,
		TagValue: in.TagValue,
	}
}

// DatasetMetadataToUpdate is used when updating a dataset's metadata.
// Only non-nil fields will be updated.
type DatasetMetadataToUpdate struct {
	Description optional.String // The user-friendly description of this table.
	Name        optional.String // The user-friendly name for this dataset.

	// DefaultTableExpiration is the default expiration time for new tables.
	// If set to time.Duration(0), new tables never expire.
	DefaultTableExpiration optional.Duration

	// DefaultTableExpiration is the default expiration time for
	// all newly created partitioned tables.
	// If set to time.Duration(0), new table partitions never expire.
	DefaultPartitionExpiration optional.Duration

	// DefaultEncryptionConfig defines CMEK settings for new resources created
	// in the dataset.
	DefaultEncryptionConfig *EncryptionConfig

	// Defines the default collation specification of future tables
	// created in the dataset.
	DefaultCollation optional.String

	// For externally defined datasets, contains information about the configuration.
	ExternalDatasetReference *ExternalDatasetReference

	// MaxTimeTravel represents the number of hours for the max time travel for all tables
	// in the dataset.  Durations are rounded towards zero for the nearest hourly value.
	MaxTimeTravel optional.Duration

	// Storage billing model to be used for all tables in the dataset.
	// Can be set to PHYSICAL. Default is LOGICAL.
	// Once you change a dataset's storage billing model to use physical bytes, you can't change it back to using logical bytes again.
	// More details: https://cloud.google.com/bigquery/docs/datasets-intro#dataset_storage_billing_models
	StorageBillingModel optional.String

	// The entire access list. It is not possible to replace individual entries.
	Access []*AccessEntry

	labelUpdater
}

// Dataset creates a handle to a BigQuery dataset in the client's project.
func (c *Client) Dataset(id string) *Dataset {
	return c.DatasetInProject(c.projectID, id)
}

// DatasetInProject creates a handle to a BigQuery dataset in the specified project.
func (c *Client) DatasetInProject(projectID, datasetID string) *Dataset {
	return &Dataset{
		ProjectID: projectID,
		DatasetID: datasetID,
		c:         c,
	}
}

// Identifier returns the ID of the dataset in the requested format.
//
// For Standard SQL format, the identifier will be quoted if the
// ProjectID contains dash (-) characters.
func (d *Dataset) Identifier(f IdentifierFormat) (string, error) {
	switch f {
	case LegacySQLID:
		return fmt.Sprintf("%s:%s", d.ProjectID, d.DatasetID), nil
	case StandardSQLID:
		// Quote project identifiers if they have a dash character.
		if strings.Contains(d.ProjectID, "-") {
			return fmt.Sprintf("`%s`.%s", d.ProjectID, d.DatasetID), nil
		}
		return fmt.Sprintf("%s.%s", d.ProjectID, d.DatasetID), nil
	default:
		return "", ErrUnknownIdentifierFormat
	}
}

// Create creates a dataset in the BigQuery service. An error will be returned if the
// dataset already exists. Pass in a DatasetMetadata value to configure the dataset.
func (d *Dataset) Create(ctx context.Context, md *DatasetMetadata) (err error) {
	ctx = trace.StartSpan(ctx, "cloud.google.com/go/bigquery.Dataset.Create")
	defer func() { trace.EndSpan(ctx, err) }()

	ds, err := md.toBQ()
	if err != nil {
		return err
	}
	ds.DatasetReference = &bq.DatasetReference{DatasetId: d.DatasetID}
	// Use Client.Location as a default.
	if ds.Location == "" {
		ds.Location = d.c.Location
	}
	call := d.c.bqs.Datasets.Insert(d.ProjectID, ds).Context(ctx)
	setClientHeader(call.Header())
	_, err = call.Do()
	return err
}

func (dm *DatasetMetadata) toBQ() (*bq.Dataset, error) {
	ds := &bq.Dataset{}
	if dm == nil {
		return ds, nil
	}
	ds.FriendlyName = dm.Name
	ds.Description = dm.Description
	ds.Location = dm.Location
	ds.DefaultTableExpirationMs = int64(dm.DefaultTableExpiration / time.Millisecond)
	ds.DefaultPartitionExpirationMs = int64(dm.DefaultPartitionExpiration / time.Millisecond)
	ds.DefaultCollation = dm.DefaultCollation
	ds.MaxTimeTravelHours = int64(dm.MaxTimeTravel / time.Hour)
	ds.StorageBillingModel = string(dm.StorageBillingModel)
	ds.Labels = dm.Labels
	var err error
	ds.Access, err = accessListToBQ(dm.Access)
	if err != nil {
		return nil, err
	}
	if !dm.CreationTime.IsZero() {
		return nil, errors.New("bigquery: Dataset.CreationTime is not writable")
	}
	if !dm.LastModifiedTime.IsZero() {
		return nil, errors.New("bigquery: Dataset.LastModifiedTime is not writable")
	}
	if dm.FullID != "" {
		return nil, errors.New("bigquery: Dataset.FullID is not writable")
	}
	if dm.ETag != "" {
		return nil, errors.New("bigquery: Dataset.ETag is not writable")
	}
	if dm.DefaultEncryptionConfig != nil {
		ds.DefaultEncryptionConfiguration = dm.DefaultEncryptionConfig.toBQ()
	}
	if dm.ExternalDatasetReference != nil {
		ds.ExternalDatasetReference = dm.ExternalDatasetReference.toBQ()
	}
	return ds, nil
}

func accessListToBQ(a []*AccessEntry) ([]*bq.DatasetAccess, error) {
	var q []*bq.DatasetAccess
	for _, e := range a {
		a, err := e.toBQ()
		if err != nil {
			return nil, err
		}
		q = append(q, a)
	}
	return q, nil
}

// Delete deletes the dataset.  Delete will fail if the dataset is not empty.
func (d *Dataset) Delete(ctx context.Context) (err error) {
	return d.deleteInternal(ctx, false)
}

// DeleteWithContents deletes the dataset, as well as contained resources.
func (d *Dataset) DeleteWithContents(ctx context.Context) (err error) {
	return d.deleteInternal(ctx, true)
}

func (d *Dataset) deleteInternal(ctx context.Context, deleteContents bool) (err error) {
	ctx = trace.StartSpan(ctx, "cloud.google.com/go/bigquery.Dataset.Delete")
	defer func() { trace.EndSpan(ctx, err) }()

	call := d.c.bqs.Datasets.Delete(d.ProjectID, d.DatasetID).Context(ctx).DeleteContents(deleteContents)
	setClientHeader(call.Header())
	return runWithRetry(ctx, func() (err error) {
		sCtx := trace.StartSpan(ctx, "bigquery.datasets.delete")
		err = call.Do()
		trace.EndSpan(sCtx, err)
		return err
	})
}

// Metadata fetches the metadata for the dataset.
func (d *Dataset) Metadata(ctx context.Context) (md *DatasetMetadata, err error) {
	ctx = trace.StartSpan(ctx, "cloud.google.com/go/bigquery.Dataset.Metadata")
	defer func() { trace.EndSpan(ctx, err) }()

	call := d.c.bqs.Datasets.Get(d.ProjectID, d.DatasetID).Context(ctx)
	setClientHeader(call.Header())
	var ds *bq.Dataset
	if err := runWithRetry(ctx, func() (err error) {
		sCtx := trace.StartSpan(ctx, "bigquery.datasets.get")
		ds, err = call.Do()
		trace.EndSpan(sCtx, err)
		return err
	}); err != nil {
		return nil, err
	}
	return bqToDatasetMetadata(ds, d.c)
}

func bqToDatasetMetadata(d *bq.Dataset, c *Client) (*DatasetMetadata, error) {
	dm := &DatasetMetadata{
		CreationTime:               unixMillisToTime(d.CreationTime),
		LastModifiedTime:           unixMillisToTime(d.LastModifiedTime),
		DefaultTableExpiration:     time.Duration(d.DefaultTableExpirationMs) * time.Millisecond,
		DefaultPartitionExpiration: time.Duration(d.DefaultPartitionExpirationMs) * time.Millisecond,
		DefaultCollation:           d.DefaultCollation,
		ExternalDatasetReference:   bqToExternalDatasetReference(d.ExternalDatasetReference),
		MaxTimeTravel:              time.Duration(d.MaxTimeTravelHours) * time.Hour,
		StorageBillingModel:        d.StorageBillingModel,
		DefaultEncryptionConfig:    bqToEncryptionConfig(d.DefaultEncryptionConfiguration),
		Description:                d.Description,
		Name:                       d.FriendlyName,
		FullID:                     d.Id,
		Location:                   d.Location,
		Labels:                     d.Labels,
		ETag:                       d.Etag,
	}
	for _, a := range d.Access {
		e, err := bqToAccessEntry(a, c)
		if err != nil {
			return nil, err
		}
		dm.Access = append(dm.Access, e)
	}
	for _, bqTag := range d.Tags {
		tag := bqToDatasetTag(bqTag)
		if tag != nil {
			dm.Tags = append(dm.Tags, tag)
		}
	}
	return dm, nil
}

// Update modifies specific Dataset metadata fields.
// To perform a read-modify-write that protects against intervening reads,
// set the etag argument to the DatasetMetadata.ETag field from the read.
// Pass the empty string for etag for a "blind write" that will always succeed.
func (d *Dataset) Update(ctx context.Context, dm DatasetMetadataToUpdate, etag string) (md *DatasetMetadata, err error) {
	ctx = trace.StartSpan(ctx, "cloud.google.com/go/bigquery.Dataset.Update")
	defer func() { trace.EndSpan(ctx, err) }()

	ds, err := dm.toBQ()
	if err != nil {
		return nil, err
	}
	call := d.c.bqs.Datasets.Patch(d.ProjectID, d.DatasetID, ds).Context(ctx)
	setClientHeader(call.Header())
	if etag != "" {
		call.Header().Set("If-Match", etag)
	}
	var ds2 *bq.Dataset
	if err := runWithRetry(ctx, func() (err error) {
		sCtx := trace.StartSpan(ctx, "bigquery.datasets.patch")
		ds2, err = call.Do()
		trace.EndSpan(sCtx, err)
		return err
	}); err != nil {
		return nil, err
	}
	return bqToDatasetMetadata(ds2, d.c)
}

func (dm *DatasetMetadataToUpdate) toBQ() (*bq.Dataset, error) {
	ds := &bq.Dataset{}
	forceSend := func(field string) {
		ds.ForceSendFields = append(ds.ForceSendFields, field)
	}

	if dm.Description != nil {
		ds.Description = optional.ToString(dm.Description)
		forceSend("Description")
	}
	if dm.Name != nil {
		ds.FriendlyName = optional.ToString(dm.Name)
		forceSend("FriendlyName")
	}
	if dm.DefaultTableExpiration != nil {
		dur := optional.ToDuration(dm.DefaultTableExpiration)
		if dur == 0 {
			// Send a null to delete the field.
			ds.NullFields = append(ds.NullFields, "DefaultTableExpirationMs")
		} else {
			ds.DefaultTableExpirationMs = int64(dur / time.Millisecond)
		}
	}
	if dm.DefaultPartitionExpiration != nil {
		dur := optional.ToDuration(dm.DefaultPartitionExpiration)
		if dur == 0 {
			// Send a null to delete the field.
			ds.NullFields = append(ds.NullFields, "DefaultPartitionExpirationMs")
		} else {
			ds.DefaultPartitionExpirationMs = int64(dur / time.Millisecond)
		}
	}
	if dm.DefaultCollation != nil {
		ds.DefaultCollation = optional.ToString(dm.DefaultCollation)
		forceSend("DefaultCollation")
	}
	if dm.ExternalDatasetReference != nil {
		ds.ExternalDatasetReference = dm.ExternalDatasetReference.toBQ()
		forceSend("ExternalDatasetReference")
	}
	if dm.MaxTimeTravel != nil {
		dur := optional.ToDuration(dm.MaxTimeTravel)
		if dur == 0 {
			// Send a null to delete the field.
			ds.NullFields = append(ds.NullFields, "MaxTimeTravelHours")
		} else {
			ds.MaxTimeTravelHours = int64(dur / time.Hour)
		}
	}
	if dm.StorageBillingModel != nil {
		ds.StorageBillingModel = optional.ToString(dm.StorageBillingModel)
		forceSend("StorageBillingModel")
	}
	if dm.DefaultEncryptionConfig != nil {
		ds.DefaultEncryptionConfiguration = dm.DefaultEncryptionConfig.toBQ()
		ds.DefaultEncryptionConfiguration.ForceSendFields = []string{"KmsKeyName"}
	}
	if dm.Access != nil {
		var err error
		ds.Access, err = accessListToBQ(dm.Access)
		if err != nil {
			return nil, err
		}
		if len(ds.Access) == 0 {
			ds.NullFields = append(ds.NullFields, "Access")
		}
	}
	labels, forces, nulls := dm.update()
	ds.Labels = labels
	ds.ForceSendFields = append(ds.ForceSendFields, forces...)
	ds.NullFields = append(ds.NullFields, nulls...)
	return ds, nil
}

// Table creates a handle to a BigQuery table in the dataset.
// To determine if a table exists, call Table.Metadata.
// If the table does not already exist, use Table.Create to create it.
func (d *Dataset) Table(tableID string) *Table {
	return &Table{ProjectID: d.ProjectID, DatasetID: d.DatasetID, TableID: tableID, c: d.c}
}

// Tables returns an iterator over the tables in the Dataset.
func (d *Dataset) Tables(ctx context.Context) *TableIterator {
	it := &TableIterator{
		ctx:     ctx,
		dataset: d,
	}
	it.pageInfo, it.nextFunc = iterator.NewPageInfo(
		it.fetch,
		func() int { return len(it.tables) },
		func() interface{} { b := it.tables; it.tables = nil; return b })
	return it
}

// A TableIterator is an iterator over Tables.
type TableIterator struct {
	ctx      context.Context
	dataset  *Dataset
	tables   []*Table
	pageInfo *iterator.PageInfo
	nextFunc func() error
}

// Next returns the next result. Its second return value is Done if there are
// no more results. Once Next returns Done, all subsequent calls will return
// Done.
func (it *TableIterator) Next() (*Table, error) {
	if err := it.nextFunc(); err != nil {
		return nil, err
	}
	t := it.tables[0]
	it.tables = it.tables[1:]
	return t, nil
}

// PageInfo supports pagination. See the google.golang.org/api/iterator package for details.
func (it *TableIterator) PageInfo() *iterator.PageInfo { return it.pageInfo }

// listTables exists to aid testing.
var listTables = func(it *TableIterator, pageSize int, pageToken string) (*bq.TableList, error) {
	call := it.dataset.c.bqs.Tables.List(it.dataset.ProjectID, it.dataset.DatasetID).
		PageToken(pageToken).
		Context(it.ctx)
	setClientHeader(call.Header())
	if pageSize > 0 {
		call.MaxResults(int64(pageSize))
	}
	var res *bq.TableList
	err := runWithRetry(it.ctx, func() (err error) {
		sCtx := trace.StartSpan(it.ctx, "bigquery.tables.list")
		res, err = call.Do()
		trace.EndSpan(sCtx, err)
		return err
	})
	return res, err
}

func (it *TableIterator) fetch(pageSize int, pageToken string) (string, error) {
	res, err := listTables(it, pageSize, pageToken)
	if err != nil {
		return "", err
	}
	for _, t := range res.Tables {
		it.tables = append(it.tables, bqToTable(t.TableReference, it.dataset.c))
	}
	return res.NextPageToken, nil
}

func bqToTable(tr *bq.TableReference, c *Client) *Table {
	if tr == nil {
		return nil
	}
	return &Table{
		ProjectID: tr.ProjectId,
		DatasetID: tr.DatasetId,
		TableID:   tr.TableId,
		c:         c,
	}
}

// Model creates a handle to a BigQuery model in the dataset.
// To determine if a model exists, call Model.Metadata.
// If the model does not already exist, you can create it via execution
// of a CREATE MODEL query.
func (d *Dataset) Model(modelID string) *Model {
	return &Model{ProjectID: d.ProjectID, DatasetID: d.DatasetID, ModelID: modelID, c: d.c}
}

// Models returns an iterator over the models in the Dataset.
func (d *Dataset) Models(ctx context.Context) *ModelIterator {
	it := &ModelIterator{
		ctx:     ctx,
		dataset: d,
	}
	it.pageInfo, it.nextFunc = iterator.NewPageInfo(
		it.fetch,
		func() int { return len(it.models) },
		func() interface{} { b := it.models; it.models = nil; return b })
	return it
}

// A ModelIterator is an iterator over Models.
type ModelIterator struct {
	ctx      context.Context
	dataset  *Dataset
	models   []*Model
	pageInfo *iterator.PageInfo
	nextFunc func() error
}

// Next returns the next result. Its second return value is Done if there are
// no more results. Once Next returns Done, all subsequent calls will return
// Done.
func (it *ModelIterator) Next() (*Model, error) {
	if err := it.nextFunc(); err != nil {
		return nil, err
	}
	t := it.models[0]
	it.models = it.models[1:]
	return t, nil
}

// PageInfo supports pagination. See the google.golang.org/api/iterator package for details.
func (it *ModelIterator) PageInfo() *iterator.PageInfo { return it.pageInfo }

// listTables exists to aid testing.
var listModels = func(it *ModelIterator, pageSize int, pageToken string) (*bq.ListModelsResponse, error) {
	call := it.dataset.c.bqs.Models.List(it.dataset.ProjectID, it.dataset.DatasetID).
		PageToken(pageToken).
		Context(it.ctx)
	setClientHeader(call.Header())
	if pageSize > 0 {
		call.MaxResults(int64(pageSize))
	}
	var res *bq.ListModelsResponse
	err := runWithRetry(it.ctx, func() (err error) {
		sCtx := trace.StartSpan(it.ctx, "bigquery.models.list")
		res, err = call.Do()
		trace.EndSpan(sCtx, err)
		return err
	})
	return res, err
}

func (it *ModelIterator) fetch(pageSize int, pageToken string) (string, error) {
	res, err := listModels(it, pageSize, pageToken)
	if err != nil {
		return "", err
	}
	for _, t := range res.Models {
		it.models = append(it.models, bqToModel(t.ModelReference, it.dataset.c))
	}
	return res.NextPageToken, nil
}

func bqToModel(mr *bq.ModelReference, c *Client) *Model {
	if mr == nil {
		return nil
	}
	return &Model{
		ProjectID: mr.ProjectId,
		DatasetID: mr.DatasetId,
		ModelID:   mr.ModelId,
		c:         c,
	}
}

// Routine creates a handle to a BigQuery routine in the dataset.
// To determine if a routine exists, call Routine.Metadata.
func (d *Dataset) Routine(routineID string) *Routine {
	return &Routine{
		ProjectID: d.ProjectID,
		DatasetID: d.DatasetID,
		RoutineID: routineID,
		c:         d.c}
}

// Routines returns an iterator over the routines in the Dataset.
func (d *Dataset) Routines(ctx context.Context) *RoutineIterator {
	it := &RoutineIterator{
		ctx:     ctx,
		dataset: d,
	}
	it.pageInfo, it.nextFunc = iterator.NewPageInfo(
		it.fetch,
		func() int { return len(it.routines) },
		func() interface{} { b := it.routines; it.routines = nil; return b })
	return it
}

// A RoutineIterator is an iterator over Routines.
type RoutineIterator struct {
	ctx      context.Context
	dataset  *Dataset
	routines []*Routine
	pageInfo *iterator.PageInfo
	nextFunc func() error
}

// Next returns the next result. Its second return value is Done if there are
// no more results. Once Next returns Done, all subsequent calls will return
// Done.
func (it *RoutineIterator) Next() (*Routine, error) {
	if err := it.nextFunc(); err != nil {
		return nil, err
	}
	t := it.routines[0]
	it.routines = it.routines[1:]
	return t, nil
}

// PageInfo supports pagination. See the google.golang.org/api/iterator package for details.
func (it *RoutineIterator) PageInfo() *iterator.PageInfo { return it.pageInfo }

// listRoutines exists to aid testing.
var listRoutines = func(it *RoutineIterator, pageSize int, pageToken string) (*bq.ListRoutinesResponse, error) {
	call := it.dataset.c.bqs.Routines.List(it.dataset.ProjectID, it.dataset.DatasetID).
		PageToken(pageToken).
		Context(it.ctx)
	setClientHeader(call.Header())
	if pageSize > 0 {
		call.MaxResults(int64(pageSize))
	}
	var res *bq.ListRoutinesResponse
	err := runWithRetry(it.ctx, func() (err error) {
		sCtx := trace.StartSpan(it.ctx, "bigquery.routines.list")
		res, err = call.Do()
		trace.EndSpan(sCtx, err)
		return err
	})
	return res, err
}

func (it *RoutineIterator) fetch(pageSize int, pageToken string) (string, error) {
	res, err := listRoutines(it, pageSize, pageToken)
	if err != nil {
		return "", err
	}
	for _, t := range res.Routines {
		it.routines = append(it.routines, bqToRoutine(t.RoutineReference, it.dataset.c))
	}
	return res.NextPageToken, nil
}

func bqToRoutine(mr *bq.RoutineReference, c *Client) *Routine {
	if mr == nil {
		return nil
	}
	return &Routine{
		ProjectID: mr.ProjectId,
		DatasetID: mr.DatasetId,
		RoutineID: mr.RoutineId,
		c:         c,
	}
}

// Datasets returns an iterator over the datasets in a project.
// The Client's project is used by default, but that can be
// changed by setting ProjectID on the returned iterator before calling Next.
func (c *Client) Datasets(ctx context.Context) *DatasetIterator {
	return c.DatasetsInProject(ctx, c.projectID)
}

// DatasetsInProject returns an iterator over the datasets in the provided project.
//
// Deprecated: call Client.Datasets, then set ProjectID on the returned iterator.
func (c *Client) DatasetsInProject(ctx context.Context, projectID string) *DatasetIterator {
	it := &DatasetIterator{
		ctx:       ctx,
		c:         c,
		ProjectID: projectID,
	}
	it.pageInfo, it.nextFunc = iterator.NewPageInfo(
		it.fetch,
		func() int { return len(it.items) },
		func() interface{} { b := it.items; it.items = nil; return b })
	return it
}

// DatasetIterator iterates over the datasets in a project.
type DatasetIterator struct {
	// ListHidden causes hidden datasets to be listed when set to true.
	// Set before the first call to Next.
	ListHidden bool

	// Filter restricts the datasets returned by label. The filter syntax is described in
	// https://cloud.google.com/bigquery/docs/labeling-datasets#filtering_datasets_using_labels
	// Set before the first call to Next.
	Filter string

	// The project ID of the listed datasets.
	// Set before the first call to Next.
	ProjectID string

	ctx      context.Context
	c        *Client
	pageInfo *iterator.PageInfo
	nextFunc func() error
	items    []*Dataset
}

// PageInfo supports pagination. See the google.golang.org/api/iterator package for details.
func (it *DatasetIterator) PageInfo() *iterator.PageInfo { return it.pageInfo }

// Next returns the next Dataset. Its second return value is iterator.Done if
// there are no more results. Once Next returns Done, all subsequent calls will
// return Done.
func (it *DatasetIterator) Next() (*Dataset, error) {
	if err := it.nextFunc(); err != nil {
		return nil, err
	}
	item := it.items[0]
	it.items = it.items[1:]
	return item, nil
}

// for testing
var listDatasets = func(it *DatasetIterator, pageSize int, pageToken string) (*bq.DatasetList, error) {
	call := it.c.bqs.Datasets.List(it.ProjectID).
		Context(it.ctx).
		PageToken(pageToken).
		All(it.ListHidden)
	setClientHeader(call.Header())
	if pageSize > 0 {
		call.MaxResults(int64(pageSize))
	}
	if it.Filter != "" {
		call.Filter(it.Filter)
	}
	var res *bq.DatasetList
	err := runWithRetry(it.ctx, func() (err error) {
		sCtx := trace.StartSpan(it.ctx, "bigquery.datasets.list")
		res, err = call.Do()
		trace.EndSpan(sCtx, err)
		return err
	})
	return res, err
}

func (it *DatasetIterator) fetch(pageSize int, pageToken string) (string, error) {
	res, err := listDatasets(it, pageSize, pageToken)
	if err != nil {
		return "", err
	}
	for _, d := range res.Datasets {
		it.items = append(it.items, &Dataset{
			ProjectID: d.DatasetReference.ProjectId,
			DatasetID: d.DatasetReference.DatasetId,
			c:         it.c,
		})
	}
	return res.NextPageToken, nil
}

// An AccessEntry describes the permissions that an entity has on a dataset.
type AccessEntry struct {
	Role       AccessRole          // The role of the entity
	EntityType EntityType          // The type of entity
	Entity     string              // The entity (individual or group) granted access
	View       *Table              // The view granted access (EntityType must be ViewEntity)
	Routine    *Routine            // The routine granted access (only UDF currently supported)
	Dataset    *DatasetAccessEntry // The resources within a dataset granted access.
}

// AccessRole is the level of access to grant to a dataset.
type AccessRole string

const (
	// OwnerRole is the OWNER AccessRole.
	OwnerRole AccessRole = "OWNER"
	// ReaderRole is the READER AccessRole.
	ReaderRole AccessRole = "READER"
	// WriterRole is the WRITER AccessRole.
	WriterRole AccessRole = "WRITER"
)

// EntityType is the type of entity in an AccessEntry.
type EntityType int

const (
	// DomainEntity is a domain (e.g. "example.com").
	DomainEntity EntityType = iota + 1

	// GroupEmailEntity is an email address of a Google Group.
	GroupEmailEntity

	// UserEmailEntity is an email address of an individual user.
	UserEmailEntity

	// SpecialGroupEntity is a special group: one of projectOwners, projectReaders, projectWriters or
	// allAuthenticatedUsers.
	SpecialGroupEntity

	// ViewEntity is a BigQuery logical view.
	ViewEntity

	// IAMMemberEntity represents entities present in IAM but not represented using
	// the other entity types.
	IAMMemberEntity

	// RoutineEntity is a BigQuery routine, referencing a User Defined Function (UDF).
	RoutineEntity

	// DatasetEntity is BigQuery dataset, present in the access list.
	DatasetEntity
)

func (e *AccessEntry) toBQ() (*bq.DatasetAccess, error) {
	q := &bq.DatasetAccess{Role: string(e.Role)}
	switch e.EntityType {
	case DomainEntity:
		q.Domain = e.Entity
	case GroupEmailEntity:
		q.GroupByEmail = e.Entity
	case UserEmailEntity:
		q.UserByEmail = e.Entity
	case SpecialGroupEntity:
		q.SpecialGroup = e.Entity
	case ViewEntity:
		q.View = e.View.toBQ()
	case IAMMemberEntity:
		q.IamMember = e.Entity
	case RoutineEntity:
		q.Routine = e.Routine.toBQ()
	case DatasetEntity:
		q.Dataset = e.Dataset.toBQ()
	default:
		return nil, fmt.Errorf("bigquery: unknown entity type %d", e.EntityType)
	}
	return q, nil
}

func bqToAccessEntry(q *bq.DatasetAccess, c *Client) (*AccessEntry, error) {
	e := &AccessEntry{Role: AccessRole(q.Role)}
	switch {
	case q.Domain != "":
		e.Entity = q.Domain
		e.EntityType = DomainEntity
	case q.GroupByEmail != "":
		e.Entity = q.GroupByEmail
		e.EntityType = GroupEmailEntity
	case q.UserByEmail != "":
		e.Entity = q.UserByEmail
		e.EntityType = UserEmailEntity
	case q.SpecialGroup != "":
		e.Entity = q.SpecialGroup
		e.EntityType = SpecialGroupEntity
	case q.View != nil:
		e.View = c.DatasetInProject(q.View.ProjectId, q.View.DatasetId).Table(q.View.TableId)
		e.EntityType = ViewEntity
	case q.IamMember != "":
		e.Entity = q.IamMember
		e.EntityType = IAMMemberEntity
	case q.Routine != nil:
		e.Routine = c.DatasetInProject(q.Routine.ProjectId, q.Routine.DatasetId).Routine(q.Routine.RoutineId)
		e.EntityType = RoutineEntity
	case q.Dataset != nil:
		e.Dataset = bqToDatasetAccessEntry(q.Dataset, c)
		e.EntityType = DatasetEntity
	default:
		return nil, errors.New("bigquery: invalid access value")
	}
	return e, nil
}

// DatasetAccessEntry is an access entry that refers to resources within
// another dataset.
type DatasetAccessEntry struct {
	// The dataset to which this entry applies.
	Dataset *Dataset
	// The list of target types within the dataset
	// to which this entry applies.
	//
	// Current supported values:
	//
	// VIEWS - This entry applies to views in the dataset.
	TargetTypes []string
}

func (dae *DatasetAccessEntry) toBQ() *bq.DatasetAccessEntry {
	if dae == nil {
		return nil
	}
	return &bq.DatasetAccessEntry{
		Dataset: &bq.DatasetReference{
			ProjectId: dae.Dataset.ProjectID,
			DatasetId: dae.Dataset.DatasetID,
		},
		TargetTypes: dae.TargetTypes,
	}
}

func bqToDatasetAccessEntry(entry *bq.DatasetAccessEntry, c *Client) *DatasetAccessEntry {
	if entry == nil {
		return nil
	}
	return &DatasetAccessEntry{
		Dataset:     c.DatasetInProject(entry.Dataset.ProjectId, entry.Dataset.DatasetId),
		TargetTypes: entry.TargetTypes,
	}
}

// ExternalDatasetReference provides information about external dataset metadata.
type ExternalDatasetReference struct {
	//The connection id that is used to access the external_source.
	// Format: projects/{project_id}/locations/{location_id}/connections/{connection_id}
	Connection string

	// External source that backs this dataset.
	ExternalSource string
}

func bqToExternalDatasetReference(bq *bq.ExternalDatasetReference) *ExternalDatasetReference {
	if bq == nil {
		return nil
	}
	return &ExternalDatasetReference{
		Connection:     bq.Connection,
		ExternalSource: bq.ExternalSource,
	}
}

func (edr *ExternalDatasetReference) toBQ() *bq.ExternalDatasetReference {
	if edr == nil {
		return nil
	}
	return &bq.ExternalDatasetReference{
		Connection:     edr.Connection,
		ExternalSource: edr.ExternalSource,
	}
}
