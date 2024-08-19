// Copyright 2023 Google LLC
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
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"math/big"

	"cloud.google.com/go/civil"
	"github.com/apache/arrow/go/v14/arrow"
	"github.com/apache/arrow/go/v14/arrow/array"
	"github.com/apache/arrow/go/v14/arrow/ipc"
	"github.com/apache/arrow/go/v14/arrow/memory"
	"google.golang.org/api/iterator"
)

// ArrowRecordBatch represents an Arrow RecordBatch with the source PartitionID
type ArrowRecordBatch struct {
	reader io.Reader
	// Serialized Arrow Record Batch.
	Data []byte
	// Serialized Arrow Schema.
	Schema []byte
	// Source partition ID. In the Storage API world, it represents the ReadStream.
	PartitionID string
}

// Read makes ArrowRecordBatch implements io.Reader
func (r *ArrowRecordBatch) Read(p []byte) (int, error) {
	if r.reader == nil {
		buf := bytes.NewBuffer(r.Schema)
		buf.Write(r.Data)
		r.reader = buf
	}
	return r.reader.Read(p)
}

// ArrowIterator represents a way to iterate through a stream of arrow records.
// Experimental: this interface is experimental and may be modified or removed in future versions,
// regardless of any other documented package stability guarantees.
type ArrowIterator interface {
	Next() (*ArrowRecordBatch, error)
	Schema() Schema
	SerializedArrowSchema() []byte
}

// NewArrowIteratorReader allows to consume an ArrowIterator as an io.Reader.
// Experimental: this interface is experimental and may be modified or removed in future versions,
// regardless of any other documented package stability guarantees.
func NewArrowIteratorReader(it ArrowIterator) io.Reader {
	return &arrowIteratorReader{
		it: it,
	}
}

type arrowIteratorReader struct {
	buf *bytes.Buffer
	it  ArrowIterator
}

// Read makes ArrowIteratorReader implement io.Reader
func (r *arrowIteratorReader) Read(p []byte) (int, error) {
	if r.it == nil {
		return -1, errors.New("bigquery: nil ArrowIterator")
	}
	if r.buf == nil { // init with schema
		buf := bytes.NewBuffer(r.it.SerializedArrowSchema())
		r.buf = buf
	}
	n, err := r.buf.Read(p)
	if err == io.EOF {
		batch, err := r.it.Next()
		if err == iterator.Done {
			return 0, io.EOF
		}
		r.buf.Write(batch.Data)
		return r.Read(p)
	}
	return n, err
}

type arrowDecoder struct {
	allocator   memory.Allocator
	tableSchema Schema
	arrowSchema *arrow.Schema
}

func newArrowDecoder(arrowSerializedSchema []byte, schema Schema) (*arrowDecoder, error) {
	buf := bytes.NewBuffer(arrowSerializedSchema)
	r, err := ipc.NewReader(buf)
	if err != nil {
		return nil, err
	}
	defer r.Release()
	p := &arrowDecoder{
		tableSchema: schema,
		arrowSchema: r.Schema(),
		allocator:   memory.DefaultAllocator,
	}
	return p, nil
}

func (ap *arrowDecoder) createIPCReaderForBatch(arrowRecordBatch *ArrowRecordBatch) (*ipc.Reader, error) {
	return ipc.NewReader(
		arrowRecordBatch,
		ipc.WithSchema(ap.arrowSchema),
		ipc.WithAllocator(ap.allocator),
	)
}

// decodeArrowRecords decodes BQ ArrowRecordBatch into rows of []Value.
func (ap *arrowDecoder) decodeArrowRecords(arrowRecordBatch *ArrowRecordBatch) ([][]Value, error) {
	r, err := ap.createIPCReaderForBatch(arrowRecordBatch)
	if err != nil {
		return nil, err
	}
	defer r.Release()
	rs := make([][]Value, 0)
	for r.Next() {
		rec := r.Record()
		values, err := ap.convertArrowRecordValue(rec)
		if err != nil {
			return nil, err
		}
		rs = append(rs, values...)
	}
	return rs, nil
}

// decodeRetainedArrowRecords decodes BQ ArrowRecordBatch into a list of retained arrow.Record.
func (ap *arrowDecoder) decodeRetainedArrowRecords(arrowRecordBatch *ArrowRecordBatch) ([]arrow.Record, error) {
	r, err := ap.createIPCReaderForBatch(arrowRecordBatch)
	if err != nil {
		return nil, err
	}
	defer r.Release()
	records := []arrow.Record{}
	for r.Next() {
		rec := r.Record()
		rec.Retain()
		records = append(records, rec)
	}
	return records, nil
}

// convertArrowRows converts an arrow.Record into a series of Value slices.
func (ap *arrowDecoder) convertArrowRecordValue(record arrow.Record) ([][]Value, error) {
	rs := make([][]Value, record.NumRows())
	for i := range rs {
		rs[i] = make([]Value, record.NumCols())
	}
	for j, col := range record.Columns() {
		fs := ap.tableSchema[j]
		ft := ap.arrowSchema.Field(j).Type
		for i := 0; i < col.Len(); i++ {
			v, err := convertArrowValue(col, i, ft, fs)
			if err != nil {
				return nil, fmt.Errorf("found arrow type %s, but could not convert value: %v", ap.arrowSchema.Field(j).Type, err)
			}
			rs[i][j] = v
		}
	}
	return rs, nil
}

// convertArrow gets row value in the given column and converts to a Value.
// Arrow is a colunar storage, so we navigate first by column and get the row value.
// More details on conversions can be seen here: https://cloud.google.com/bigquery/docs/reference/storage#arrow_schema_details
func convertArrowValue(col arrow.Array, i int, ft arrow.DataType, fs *FieldSchema) (Value, error) {
	if !col.IsValid(i) {
		return nil, nil
	}
	switch ft.(type) {
	case *arrow.BooleanType:
		v := col.(*array.Boolean).Value(i)
		return convertBasicType(fmt.Sprintf("%v", v), fs.Type)
	case *arrow.Int8Type:
		v := col.(*array.Int8).Value(i)
		return convertBasicType(fmt.Sprintf("%v", v), fs.Type)
	case *arrow.Int16Type:
		v := col.(*array.Int16).Value(i)
		return convertBasicType(fmt.Sprintf("%v", v), fs.Type)
	case *arrow.Int32Type:
		v := col.(*array.Int32).Value(i)
		return convertBasicType(fmt.Sprintf("%v", v), fs.Type)
	case *arrow.Int64Type:
		v := col.(*array.Int64).Value(i)
		return convertBasicType(fmt.Sprintf("%v", v), fs.Type)
	case *arrow.Float16Type:
		v := col.(*array.Float16).Value(i)
		return convertBasicType(fmt.Sprintf("%v", v.Float32()), fs.Type)
	case *arrow.Float32Type:
		v := col.(*array.Float32).Value(i)
		return convertBasicType(fmt.Sprintf("%v", v), fs.Type)
	case *arrow.Float64Type:
		v := col.(*array.Float64).Value(i)
		return convertBasicType(fmt.Sprintf("%v", v), fs.Type)
	case *arrow.BinaryType:
		v := col.(*array.Binary).Value(i)
		encoded := base64.StdEncoding.EncodeToString(v)
		return convertBasicType(encoded, fs.Type)
	case *arrow.StringType:
		v := col.(*array.String).Value(i)
		return convertBasicType(v, fs.Type)
	case *arrow.Date32Type:
		v := col.(*array.Date32).Value(i)
		return convertBasicType(v.FormattedString(), fs.Type)
	case *arrow.Date64Type:
		v := col.(*array.Date64).Value(i)
		return convertBasicType(v.FormattedString(), fs.Type)
	case *arrow.TimestampType:
		v := col.(*array.Timestamp).Value(i)
		dft := ft.(*arrow.TimestampType)
		t := v.ToTime(dft.Unit)
		if dft.TimeZone == "" { // Datetime
			return Value(civil.DateTimeOf(t)), nil
		}
		return Value(t.UTC()), nil // Timestamp
	case *arrow.Time32Type:
		v := col.(*array.Time32).Value(i)
		return convertBasicType(v.FormattedString(arrow.Microsecond), fs.Type)
	case *arrow.Time64Type:
		v := col.(*array.Time64).Value(i)
		return convertBasicType(v.FormattedString(arrow.Microsecond), fs.Type)
	case *arrow.Decimal128Type:
		dft := ft.(*arrow.Decimal128Type)
		v := col.(*array.Decimal128).Value(i)
		rat := big.NewRat(1, 1)
		rat.Num().SetBytes(v.BigInt().Bytes())
		d := rat.Denom()
		d.Exp(big.NewInt(10), big.NewInt(int64(dft.Scale)), nil)
		return Value(rat), nil
	case *arrow.Decimal256Type:
		dft := ft.(*arrow.Decimal256Type)
		v := col.(*array.Decimal256).Value(i)
		rat := big.NewRat(1, 1)
		rat.Num().SetBytes(v.BigInt().Bytes())
		d := rat.Denom()
		d.Exp(big.NewInt(10), big.NewInt(int64(dft.Scale)), nil)
		return Value(rat), nil
	case *arrow.ListType:
		arr := col.(*array.List)
		dft := ft.(*arrow.ListType)
		values := []Value{}
		start, end := arr.ValueOffsets(i)
		slice := array.NewSlice(arr.ListValues(), start, end)
		for j := 0; j < slice.Len(); j++ {
			v, err := convertArrowValue(slice, j, dft.Elem(), fs)
			if err != nil {
				return nil, err
			}
			values = append(values, v)
		}
		return values, nil
	case *arrow.StructType:
		arr := col.(*array.Struct)
		nestedValues := []Value{}
		fields := ft.(*arrow.StructType).Fields()
		for fIndex, f := range fields {
			v, err := convertArrowValue(arr.Field(fIndex), i, f.Type, fs.Schema[fIndex])
			if err != nil {
				return nil, err
			}
			nestedValues = append(nestedValues, v)
		}
		return nestedValues, nil
	default:
		return nil, fmt.Errorf("unknown arrow type: %v", ft)
	}
}
