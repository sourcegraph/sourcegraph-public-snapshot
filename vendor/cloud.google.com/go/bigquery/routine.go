// Copyright 2019 Google LLC
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
)

// Routine represents a reference to a BigQuery routine.  There are multiple
// types of routines including stored procedures and scalar user-defined functions (UDFs).
// For more information, see the BigQuery documentation at https://cloud.google.com/bigquery/docs/
type Routine struct {
	ProjectID string
	DatasetID string
	RoutineID string

	c *Client
}

func (r *Routine) toBQ() *bq.RoutineReference {
	return &bq.RoutineReference{
		ProjectId: r.ProjectID,
		DatasetId: r.DatasetID,
		RoutineId: r.RoutineID,
	}
}

// Identifier returns the ID of the routine in the requested format.
//
// For Standard SQL format, the identifier will be quoted if the
// ProjectID contains dash (-) characters.
func (r *Routine) Identifier(f IdentifierFormat) (string, error) {
	switch f {
	case StandardSQLID:
		if strings.Contains(r.ProjectID, "-") {
			return fmt.Sprintf("`%s`.%s.%s", r.ProjectID, r.DatasetID, r.RoutineID), nil
		}
		return fmt.Sprintf("%s.%s.%s", r.ProjectID, r.DatasetID, r.RoutineID), nil
	default:
		return "", ErrUnknownIdentifierFormat
	}
}

// FullyQualifiedName returns an identifer for the routine in project.dataset.routine format.
func (r *Routine) FullyQualifiedName() string {
	s, _ := r.Identifier(StandardSQLID)
	return s
}

// Create creates a Routine in the BigQuery service.
// Pass in a RoutineMetadata to define the routine.
func (r *Routine) Create(ctx context.Context, rm *RoutineMetadata) (err error) {
	ctx = trace.StartSpan(ctx, "cloud.google.com/go/bigquery.Routine.Create")
	defer func() { trace.EndSpan(ctx, err) }()

	routine, err := rm.toBQ()
	if err != nil {
		return err
	}
	routine.RoutineReference = &bq.RoutineReference{
		ProjectId: r.ProjectID,
		DatasetId: r.DatasetID,
		RoutineId: r.RoutineID,
	}
	req := r.c.bqs.Routines.Insert(r.ProjectID, r.DatasetID, routine).Context(ctx)
	setClientHeader(req.Header())
	_, err = req.Do()
	return err
}

// Metadata fetches the metadata for a given Routine.
func (r *Routine) Metadata(ctx context.Context) (rm *RoutineMetadata, err error) {
	ctx = trace.StartSpan(ctx, "cloud.google.com/go/bigquery.Routine.Metadata")
	defer func() { trace.EndSpan(ctx, err) }()

	req := r.c.bqs.Routines.Get(r.ProjectID, r.DatasetID, r.RoutineID).Context(ctx)
	setClientHeader(req.Header())
	var routine *bq.Routine
	err = runWithRetry(ctx, func() (err error) {
		ctx = trace.StartSpan(ctx, "bigquery.routines.get")
		routine, err = req.Do()
		trace.EndSpan(ctx, err)
		return err
	})
	if err != nil {
		return nil, err
	}
	return bqToRoutineMetadata(routine)
}

// Update modifies properties of a Routine using the API.
func (r *Routine) Update(ctx context.Context, upd *RoutineMetadataToUpdate, etag string) (rm *RoutineMetadata, err error) {
	ctx = trace.StartSpan(ctx, "cloud.google.com/go/bigquery.Routine.Update")
	defer func() { trace.EndSpan(ctx, err) }()

	bqr, err := upd.toBQ()
	if err != nil {
		return nil, err
	}
	//TODO: remove when routines update supports partial requests.
	bqr.RoutineReference = &bq.RoutineReference{
		ProjectId: r.ProjectID,
		DatasetId: r.DatasetID,
		RoutineId: r.RoutineID,
	}

	call := r.c.bqs.Routines.Update(r.ProjectID, r.DatasetID, r.RoutineID, bqr).Context(ctx)
	setClientHeader(call.Header())
	if etag != "" {
		call.Header().Set("If-Match", etag)
	}
	var res *bq.Routine
	if err := runWithRetry(ctx, func() (err error) {
		ctx = trace.StartSpan(ctx, "bigquery.routines.update")
		res, err = call.Do()
		trace.EndSpan(ctx, err)
		return err
	}); err != nil {
		return nil, err
	}
	return bqToRoutineMetadata(res)
}

// Delete removes a Routine from a dataset.
func (r *Routine) Delete(ctx context.Context) (err error) {
	ctx = trace.StartSpan(ctx, "cloud.google.com/go/bigquery.Model.Delete")
	defer func() { trace.EndSpan(ctx, err) }()

	req := r.c.bqs.Routines.Delete(r.ProjectID, r.DatasetID, r.RoutineID).Context(ctx)
	setClientHeader(req.Header())
	return req.Do()
}

// RoutineDeterminism specifies the level of determinism that javascript User Defined Functions
// exhibit.
type RoutineDeterminism string

const (
	// Deterministic indicates that two calls with the same input to a UDF yield the same output.
	Deterministic RoutineDeterminism = "DETERMINISTIC"
	// NotDeterministic indicates that the output of the UDF is not guaranteed to yield the same
	// output each time for a given set of inputs.
	NotDeterministic RoutineDeterminism = "NOT_DETERMINISTIC"
)

const (
	// ScalarFunctionRoutine scalar function routine type
	ScalarFunctionRoutine = "SCALAR_FUNCTION"
	// ProcedureRoutine procedure routine type
	ProcedureRoutine = "PROCEDURE"
	// TableValuedFunctionRoutine routine type for table valued functions
	TableValuedFunctionRoutine = "TABLE_VALUED_FUNCTION"
)

// RoutineMetadata represents details of a given BigQuery Routine.
type RoutineMetadata struct {
	ETag string
	// Type indicates the type of routine, such as SCALAR_FUNCTION, PROCEDURE,
	// or TABLE_VALUED_FUNCTION.
	Type         string
	CreationTime time.Time
	Description  string
	// DeterminismLevel is only applicable to Javascript UDFs.
	DeterminismLevel RoutineDeterminism
	LastModifiedTime time.Time
	// Language of the routine, such as SQL or JAVASCRIPT.
	Language string
	// The list of arguments for the the routine.
	Arguments []*RoutineArgument

	// Information for a remote user-defined function.
	RemoteFunctionOptions *RemoteFunctionOptions

	ReturnType *StandardSQLDataType

	// Set only if the routine type is TABLE_VALUED_FUNCTION.
	ReturnTableType *StandardSQLTableType
	// For javascript routines, this indicates the paths for imported libraries.
	ImportedLibraries []string
	// Body contains the routine's body.
	// For functions, Body is the expression in the AS clause.
	//
	// For SQL functions, it is the substring inside the parentheses of a CREATE
	// FUNCTION statement.
	//
	// For JAVASCRIPT function, it is the evaluated string in the AS clause of
	// a CREATE FUNCTION statement.
	Body string

	// For data governance use cases.  If set to "DATA_MASKING", the function
	// is validated and made available as a masking function. For more information,
	// see: https://cloud.google.com/bigquery/docs/user-defined-functions#custom-mask
	DataGovernanceType string
}

// RemoteFunctionOptions contains information for a remote user-defined function.
type RemoteFunctionOptions struct {

	// Fully qualified name of the user-provided connection object which holds
	// the authentication information to send requests to the remote service.
	// Format:
	// projects/{projectId}/locations/{locationId}/connections/{connectionId}
	Connection string

	// Endpoint of the user-provided remote service (e.g. a function url in
	// Google Cloud Function or Cloud Run )
	Endpoint string

	// Max number of rows in each batch sent to the remote service.
	// If absent or if 0, it means no limit.
	MaxBatchingRows int64

	// User-defined context as a set of key/value pairs,
	// which will be sent as function invocation context together with
	// batched arguments in the requests to the remote service. The total
	// number of bytes of keys and values must be less than 8KB.
	UserDefinedContext map[string]string
}

func bqToRemoteFunctionOptions(in *bq.RemoteFunctionOptions) (*RemoteFunctionOptions, error) {
	if in == nil {
		return nil, nil
	}
	rfo := &RemoteFunctionOptions{
		Connection:      in.Connection,
		Endpoint:        in.Endpoint,
		MaxBatchingRows: in.MaxBatchingRows,
	}
	if in.UserDefinedContext != nil {
		rfo.UserDefinedContext = make(map[string]string)
		for k, v := range in.UserDefinedContext {
			rfo.UserDefinedContext[k] = v
		}
	}
	return rfo, nil
}

func (rfo *RemoteFunctionOptions) toBQ() (*bq.RemoteFunctionOptions, error) {
	if rfo == nil {
		return nil, nil
	}
	r := &bq.RemoteFunctionOptions{
		Connection:      rfo.Connection,
		Endpoint:        rfo.Endpoint,
		MaxBatchingRows: rfo.MaxBatchingRows,
	}
	if rfo.UserDefinedContext != nil {
		r.UserDefinedContext = make(map[string]string)
		for k, v := range rfo.UserDefinedContext {
			r.UserDefinedContext[k] = v
		}
	}
	return r, nil
}

func (rm *RoutineMetadata) toBQ() (*bq.Routine, error) {
	r := &bq.Routine{}
	if rm == nil {
		return r, nil
	}
	r.Description = rm.Description
	r.DeterminismLevel = string(rm.DeterminismLevel)
	r.Language = rm.Language
	r.RoutineType = rm.Type
	r.DefinitionBody = rm.Body
	r.DataGovernanceType = rm.DataGovernanceType
	rt, err := rm.ReturnType.toBQ()
	if err != nil {
		return nil, err
	}
	r.ReturnType = rt
	if rm.ReturnTableType != nil {
		tt, err := rm.ReturnTableType.toBQ()
		if err != nil {
			return nil, fmt.Errorf("couldn't convert return table type: %w", err)
		}
		r.ReturnTableType = tt
	}
	var args []*bq.Argument
	for _, v := range rm.Arguments {
		bqa, err := v.toBQ()
		if err != nil {
			return nil, err
		}
		args = append(args, bqa)
	}
	r.Arguments = args
	r.ImportedLibraries = rm.ImportedLibraries
	if rm.RemoteFunctionOptions != nil {
		rfo, err := rm.RemoteFunctionOptions.toBQ()
		if err != nil {
			return nil, err
		}
		r.RemoteFunctionOptions = rfo
	}
	if !rm.CreationTime.IsZero() {
		return nil, errors.New("cannot set CreationTime on create")
	}
	if !rm.LastModifiedTime.IsZero() {
		return nil, errors.New("cannot set LastModifiedTime on create")
	}
	if rm.ETag != "" {
		return nil, errors.New("cannot set ETag on create")
	}
	return r, nil
}

// RoutineArgument represents an argument supplied to a routine such as a UDF or
// stored procedured.
type RoutineArgument struct {
	// The name of this argument.  Can be absent for function return argument.
	Name string
	// Kind indicates the kind of argument represented.
	// Possible values:
	//   ARGUMENT_KIND_UNSPECIFIED
	//   FIXED_TYPE - The argument is a variable with fully specified
	//     type, which can be a struct or an array, but not a table.
	//   ANY_TYPE - The argument is any type, including struct or array,
	//     but not a table.
	Kind string
	// Mode is optional, and indicates whether an argument is input or output.
	// Mode can only be set for procedures.
	//
	// Possible values:
	//   MODE_UNSPECIFIED
	//   IN - The argument is input-only.
	//   OUT - The argument is output-only.
	//   INOUT - The argument is both an input and an output.
	Mode string
	// DataType provides typing information.  Unnecessary for ANY_TYPE Kind
	// arguments.
	DataType *StandardSQLDataType
}

func (ra *RoutineArgument) toBQ() (*bq.Argument, error) {
	if ra == nil {
		return nil, nil
	}
	a := &bq.Argument{
		Name:         ra.Name,
		ArgumentKind: ra.Kind,
		Mode:         ra.Mode,
	}
	if ra.DataType != nil {
		dt, err := ra.DataType.toBQ()
		if err != nil {
			return nil, err
		}
		a.DataType = dt
	}
	return a, nil
}

func bqToRoutineArgument(bqa *bq.Argument) (*RoutineArgument, error) {
	arg := &RoutineArgument{
		Name: bqa.Name,
		Kind: bqa.ArgumentKind,
		Mode: bqa.Mode,
	}
	dt, err := bqToStandardSQLDataType(bqa.DataType)
	if err != nil {
		return nil, err
	}
	arg.DataType = dt
	return arg, nil
}

func bqToArgs(in []*bq.Argument) ([]*RoutineArgument, error) {
	var out []*RoutineArgument
	for _, a := range in {
		arg, err := bqToRoutineArgument(a)
		if err != nil {
			return nil, err
		}
		out = append(out, arg)
	}
	return out, nil
}

func routineArgumentsToBQ(in []*RoutineArgument) ([]*bq.Argument, error) {
	var out []*bq.Argument
	for _, inarg := range in {
		arg, err := inarg.toBQ()
		if err != nil {
			return nil, err
		}
		out = append(out, arg)
	}
	return out, nil
}

// RoutineMetadataToUpdate governs updating a routine.
type RoutineMetadataToUpdate struct {
	Arguments          []*RoutineArgument
	Description        optional.String
	DeterminismLevel   optional.String
	Type               optional.String
	Language           optional.String
	Body               optional.String
	ImportedLibraries  []string
	ReturnType         *StandardSQLDataType
	ReturnTableType    *StandardSQLTableType
	DataGovernanceType optional.String
}

func (rm *RoutineMetadataToUpdate) toBQ() (*bq.Routine, error) {
	r := &bq.Routine{}
	forceSend := func(field string) {
		r.ForceSendFields = append(r.ForceSendFields, field)
	}
	nullField := func(field string) {
		r.NullFields = append(r.NullFields, field)
	}
	if rm.Description != nil {
		r.Description = optional.ToString(rm.Description)
		forceSend("Description")
	}
	if rm.DeterminismLevel != nil {
		processed := false
		// Allow either string or RoutineDeterminism, a type based on string.
		if x, ok := rm.DeterminismLevel.(RoutineDeterminism); ok {
			r.DeterminismLevel = string(x)
			processed = true
		}
		if x, ok := rm.DeterminismLevel.(string); ok {
			r.DeterminismLevel = x
			processed = true
		}
		if !processed {
			panic(fmt.Sprintf("DeterminismLevel should be either type string or RoutineDetermism in update, got %T", rm.DeterminismLevel))
		}
	}
	if rm.Arguments != nil {
		if len(rm.Arguments) == 0 {
			nullField("Arguments")
		} else {
			args, err := routineArgumentsToBQ(rm.Arguments)
			if err != nil {
				return nil, err
			}
			r.Arguments = args
			forceSend("Arguments")
		}
	}
	if rm.Type != nil {
		r.RoutineType = optional.ToString(rm.Type)
		forceSend("RoutineType")
	}
	if rm.Language != nil {
		r.Language = optional.ToString(rm.Language)
		forceSend("Language")
	}
	if rm.Body != nil {
		r.DefinitionBody = optional.ToString(rm.Body)
		forceSend("DefinitionBody")
	}
	if rm.ImportedLibraries != nil {
		if len(rm.ImportedLibraries) == 0 {
			nullField("ImportedLibraries")
		} else {
			r.ImportedLibraries = rm.ImportedLibraries
			forceSend("ImportedLibraries")
		}
	}
	if rm.ReturnType != nil {
		dt, err := rm.ReturnType.toBQ()
		if err != nil {
			return nil, err
		}
		r.ReturnType = dt
		forceSend("ReturnType")
	}
	if rm.ReturnTableType != nil {
		tt, err := rm.ReturnTableType.toBQ()
		if err != nil {
			return nil, err
		}
		r.ReturnTableType = tt
		forceSend("ReturnTableType")
	}
	if rm.DataGovernanceType != nil {
		r.DataGovernanceType = optional.ToString(rm.DataGovernanceType)
		forceSend("DataGovernanceType")
	}
	return r, nil
}

func bqToRoutineMetadata(r *bq.Routine) (*RoutineMetadata, error) {
	meta := &RoutineMetadata{
		ETag:               r.Etag,
		Type:               r.RoutineType,
		CreationTime:       unixMillisToTime(r.CreationTime),
		Description:        r.Description,
		DeterminismLevel:   RoutineDeterminism(r.DeterminismLevel),
		LastModifiedTime:   unixMillisToTime(r.LastModifiedTime),
		Language:           r.Language,
		ImportedLibraries:  r.ImportedLibraries,
		Body:               r.DefinitionBody,
		DataGovernanceType: r.DataGovernanceType,
	}
	args, err := bqToArgs(r.Arguments)
	if err != nil {
		return nil, err
	}
	meta.Arguments = args
	ret, err := bqToStandardSQLDataType(r.ReturnType)
	if err != nil {
		return nil, err
	}
	meta.ReturnType = ret
	rfo, err := bqToRemoteFunctionOptions(r.RemoteFunctionOptions)
	if err != nil {
		return nil, err
	}
	meta.RemoteFunctionOptions = rfo
	tt, err := bqToStandardSQLTableType(r.ReturnTableType)
	if err != nil {
		return nil, err
	}
	meta.ReturnTableType = tt
	return meta, nil
}
