package appdash

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	influxDBClient "github.com/influxdata/influxdb/client"
	influxDBServer "github.com/influxdata/influxdb/cmd/influxd/run"
	influxDBModels "github.com/influxdata/influxdb/models"
	influxDBErrors "github.com/influxdata/influxdb/services/meta"
	"github.com/influxdata/influxdb/toml"
)

const (
	defaultTracesPerPage  int    = 10             // Default number of traces per page.
	releaseDBName         string = "appdash"      // InfluxDB release DB name.
	schemasFieldName      string = "schemas"      // Span's measurement field name for schemas field.
	schemasFieldSeparator string = ","            // Span's measurement character separator for schemas field.
	spanMeasurementName   string = "spans"        // InfluxDB container name for trace spans.
	testDBName            string = "appdash_test" // InfluxDB test DB name (will be deleted entirely in test mode).
)

type mode int

const (
	releaseMode mode = iota // Default InfluxDBStore mode.
	testMode                // Used to setup InfluxDBStore for tests.
)

// Compile-time "implements" check.
var _ interface {
	Store
	Queryer
} = (*InfluxDBStore)(nil)

var (
	errMultipleSeries        = errors.New("unexpected multiple series")
	zeroID            string = ID(0).String() // zeroID is ID's zero value as string.
)

// pointFields -> influxDBClient.Point.Fields
type pointFields map[string]interface{}

type InfluxDBStore struct {
	config       *InfluxDBConfig
	con          *influxDBClient.Client // InfluxDB client connection.
	dbName       string                 // InfluxDB database name for this store.
	clientTarget *url.URL               // HTTP URL that the client should connect to,

	// When set to `testMode` - `testDBName` will be dropped and created, so newly database is ready for tests.
	server        *influxDBServer.Server // InfluxDB API server.
	tracesPerPage int                    // Number of traces per page.

	batchMu         sync.Mutex
	batchSizeBytes  int
	batch           []influxDBClient.Point
	flusherStopChan chan struct{}
	log             *log.Logger
}

func (in *InfluxDBStore) Collect(id SpanID, anns ...Annotation) error {
	// trace_id, span_id & parent_id are mostly used as part of the "where" part on queries so
	// to have performant queries these are set as tags(InfluxDB indexes tags).
	tags := map[string]string{
		"trace_id":  id.Trace.String(),
		"span_id":   id.Span.String(),
		"parent_id": id.Parent.String(),
	}

	// Find the start and end time of the span.
	var events []Event
	if err := UnmarshalEvents(anns, &events); err != nil {
		return err
	}
	var (
		foundItems int
		name       string
		duration   time.Duration
	)
	for _, ev := range events {
		switch v := ev.(type) {
		case spanName:
			foundItems++
			name = v.Name
		case TimespanEvent:
			foundItems++
			duration = v.End().Sub(v.Start())
		}
	}

	// Annotations `anns` are set as fields(InfluxDB does not index fields).
	fields := make(map[string]interface{}, len(anns))
	for _, ann := range anns {
		fields[ann.Key] = string(ann.Value)
	}

	// If we have span name and duration, set them as a tag and field.
	if foundItems == 2 {
		tags["name"] = name
		fields["duration"] = float64(duration) / float64(time.Second)
	}

	// `schemasFieldName` field contains all the schemas found on `anns`.
	// Eg. fields[schemasFieldName] = "HTTPClient,HTTPServer"
	fields[schemasFieldName] = schemasFromAnnotations(anns)

	// Tag values cannot contain newlines or they mess up the WRITE and cause
	// errors.
	//
	// TODO: investigate why this is; possibly related to https://github.com/influxdata/influxdb/issues/3545
	//       which was only for fields (not tags).
	for k, v := range tags {
		tags[k] = strings.Replace(v, "\n", " ", -1)
	}

	p := &influxDBClient.Point{
		Measurement: spanMeasurementName,
		Tags:        tags,
		Fields:      fields,
		Time:        time.Now().UTC(),
	}
	pointSizeBytes := pointMemoryUsage(reflect.ValueOf(p))

	in.batchMu.Lock()
	defer in.batchMu.Unlock()

	if in.batchSizeBytes+pointSizeBytes > in.config.MaxBatchSizeBytes {
		in.log.Println("InfluxDBStore: point batch entirely dropped (this should never happen, trace data will be missing)")
		in.log.Printf("InfluxDBStore: batchSize:%v batchSizeBytes:%v + pointSize:%v\n", len(in.batch), in.batchSizeBytes, pointSizeBytes)
		in.batch = nil
		in.batchSizeBytes = 0
		return ErrQueueDropped
	}
	in.batchSizeBytes += pointSizeBytes
	in.batch = append(in.batch, *p)
	return nil
}

// flush immediately sends all pending spans in the underlying batch to
// InfluxDB.
func (in *InfluxDBStore) flush() error {
	// Grab what information we need and unlock quickly to avoid contention.
	in.batchMu.Lock()
	batch := in.batch
	in.batch = nil
	in.batchSizeBytes = 0
	in.batchMu.Unlock()

	if len(batch) == 0 {
		return nil
	}

	// Write the batch to InfluxDB.
	bps := influxDBClient.BatchPoints{
		Points:   batch,
		Database: in.dbName,
	}
	_, writeErr := in.con.Write(bps)
	return writeErr
}

// flusher constantly flushes batches to InfluxDB at an interval.
func (in *InfluxDBStore) flusher() {
	in.flusherStopChan = make(chan struct{}, 1)
	go func() {
		for {
			t := time.After(in.config.BatchFlushInterval)
			select {
			case <-t:
				if err := in.flush(); err != nil {
					in.log.Println("Flush:", err)
				}
			case <-in.flusherStopChan:
				return // stop
			}
		}
	}()
}

func (in *InfluxDBStore) Trace(id ID) (*Trace, error) {
	trace := &Trace{}
	q := fmt.Sprintf("SELECT * FROM spans WHERE trace_id='%s'", id)
	result, err := in.executeOneQuery(q)
	if err != nil {
		return nil, err
	}
	if result.Err != nil {
		return nil, result.Err
	}
	if len(result.Series) == 0 {
		return nil, ErrTraceNotFound
	}

	// Expecting only one element, which contains the queried spans information.
	if len(result.Series) > 1 {
		return nil, errMultipleSeries
	}

	var (
		rootSpanSet bool
		children    []*Trace
	)

	spans, err := spansFromRow(result.Series[0])
	if err != nil {
		return nil, err
	}

	// Iterate over spans to find and set `trace`'s root span & append children spans as sub-traces to `children` for later usage.
	for _, span := range spans {
		var isRootSpan bool
		if span.ID.IsRoot() && rootSpanSet {
			return nil, errors.New("unexpected multiple root spans")
		}
		if span.ID.IsRoot() && !rootSpanSet {
			isRootSpan = true
		}
		if isRootSpan { // root span.
			trace.Span = *span
			rootSpanSet = true
		} else { // children span.
			children = append(children, &Trace{Span: *span})
		}
	}
	if err := addChildren(trace, children); err != nil {
		return nil, err
	}
	return trace, nil
}

func mustJSONFloat64(x interface{}) float64 {
	n := x.(json.Number)
	v, err := n.Float64()
	if err != nil {
		panic(err)
	}
	return v
}

func mustJSONInt64(x interface{}) int64 {
	n := x.(json.Number)
	v, err := n.Int64()
	if err != nil {
		panic(err)
	}
	return v
}

// Aggregate implements the Aggregator interface.
func (in *InfluxDBStore) Aggregate(start, end time.Duration) ([]*AggregatedResult, error) {
	// Find the mean (average), minimum, maximum, std. deviation, and count of
	// all spans. For details on how this works see the createContinuousQueries
	// method.
	q := `SELECT MEAN("mean"),MIN("min"),MAX("max"),MEAN("stddev"),SUM("count") from downsampled_spans`
	q += fmt.Sprintf(
		" WHERE time >= '%s' AND time <= '%s'",
		time.Now().Add(start).UTC().Format(time.RFC3339Nano),
		time.Now().Add(end).UTC().Format(time.RFC3339Nano),
	)
	q += ` GROUP BY "name"`
	result, err := in.executeOneQuery(q)
	if err != nil {
		return nil, err
	}

	// Populate the results.
	results := make([]*AggregatedResult, len(result.Series))
	for i, row := range result.Series {
		v := row.Values[0]
		mean, min, max, stddev, count := v[1], v[2], v[3], v[4], v[5]
		if stddev == nil {
			// stddev will be nil when there were not enough items to be able
			// to calculate a standard deviation.
			stddev = json.Number("0")
		}

		results[i] = &AggregatedResult{
			RootSpanName: row.Tags["name"],
			Average:      time.Duration(mustJSONFloat64(mean) * float64(time.Second)),
			Min:          time.Duration(mustJSONFloat64(min) * float64(time.Second)),
			Max:          time.Duration(mustJSONFloat64(max) * float64(time.Second)),
			StdDev:       time.Duration(mustJSONFloat64(stddev) * float64(time.Second)),
			Samples:      mustJSONInt64(count),
		}
	}
	if len(result.Series) == 0 {
		return nil, nil
	}

	n := 5
	if n > len(result.Series) {
		n = len(result.Series)
	}

	// Add in the N-slowest trace IDs for each span.
	//
	// TODO(slimsag): make N a pagination parameter instead.
	q = fmt.Sprintf(`SELECT TOP("top",%d),trace_id FROM nslowest_spans GROUP BY "name"`, n)
	result, err = in.executeOneQuery(q)
	if err != nil {
		return nil, err
	}
	for rowIndex, row := range result.Series {
		if row.Tags["name"] != results[rowIndex].RootSpanName {
			panic("expectation violated") // never happens, just for sanity.
		}
		for _, fields := range row.Values {
			for i, field := range fields {
				switch row.Columns[i] {
				case "trace_id":
					id, err := ParseID(field.(string))
					if err != nil {
						panic(err) // never happens, just for sanity.
					}
					results[rowIndex].Slowest = append(results[rowIndex].Slowest, id)
				}
			}
		}
	}
	return results, nil
}

func (in *InfluxDBStore) Traces(opts TracesOpts) ([]*Trace, error) {
	traces := make([]*Trace, 0)
	rootSpansQuery := fmt.Sprintf("SELECT * FROM spans WHERE parent_id='%s'", zeroID)

	// Extends `rootSpansQuery` to add time range filter using the start/end values from `opts.Timespan`.
	if opts.Timespan != (Timespan{}) {
		start := opts.Timespan.S.UTC().Format(time.RFC3339Nano)
		end := opts.Timespan.E.UTC().Format(time.RFC3339Nano)
		rootSpansQuery += fmt.Sprintf(" AND time >= '%s' AND time <= '%s'", start, end)
	}

	// Extends `rootSpansQuery` to add a filter to include only those traces present in `opts.TraceIDs`.
	traceIDsLen := len(opts.TraceIDs)
	if traceIDsLen > 0 {
		rootSpansQuery += " AND ("
		for i, traceID := range opts.TraceIDs {
			rootSpansQuery += fmt.Sprintf("trace_id = '%s'", traceID)
			lastIteration := (i+1 == traceIDsLen)
			if !lastIteration {
				rootSpansQuery += " OR "
			}
		}
		rootSpansQuery += ")"
	} else { // Otherwise continue limiting the number of traces to be returned.
		rootSpansQuery += fmt.Sprintf(" LIMIT %d", in.tracesPerPage)
	}

	rootSpansResult, err := in.executeOneQuery(rootSpansQuery)
	if err != nil {
		return nil, err
	}
	if rootSpansResult.Err != nil {
		return nil, rootSpansResult.Err
	}
	if len(rootSpansResult.Series) == 0 {
		return traces, nil
	}

	// Expecting only one element, which contains the queried spans information.
	if len(rootSpansResult.Series) > 1 {
		return nil, errMultipleSeries
	}

	// Cache to keep track of traces to be returned.
	tracesCache := make(map[ID]*Trace, 0)

	rootSpans, err := spansFromRow(rootSpansResult.Series[0])
	if err != nil {
		return nil, err
	}

	for _, span := range rootSpans {
		_, present := tracesCache[span.ID.Trace]
		if !present {
			tracesCache[span.ID.Trace] = &Trace{Span: *span}
		} else {
			return nil, errors.New("duplicated root span")
		}
	}

	// Using 'OR' since 'IN' not supported yet.
	where := `WHERE `
	var i int = 1
	for _, trace := range tracesCache {
		where += fmt.Sprintf("(trace_id='%s' AND parent_id!='%s')", trace.Span.ID.Trace, zeroID)

		// Adds 'OR' except for last iteration.
		if i != len(tracesCache) && len(tracesCache) > 1 {
			where += " OR "
		}
		i += 1
	}

	// Queries for all children spans of the root traces.
	childrenSpansQuery := fmt.Sprintf("SELECT * FROM spans %s", where)
	childrenSpansResult, err := in.executeOneQuery(childrenSpansQuery)
	if err != nil {
		return nil, err
	}

	// Cache to keep track all trace children of root traces to be returned.
	children := make(map[ID][]*Trace, 0) // Span.ID.Trace -> []*Trace

	childrenSpans, err := spansFromRow(childrenSpansResult.Series[0])
	if err != nil {
		return nil, err
	}

	// Iterates over `childrenSpans` to fill `children` cache.
	for _, span := range childrenSpans {
		trace, present := tracesCache[span.ID.Trace]
		if !present { // Root trace not added.
			return nil, errors.New("parent not found")
		} else { // Root trace already added, append `child` to `children` for later usage.
			child := &Trace{Span: *span}
			t, found := children[trace.ID.Trace]
			if !found {
				children[trace.ID.Trace] = []*Trace{child}
			} else {
				children[trace.ID.Trace] = append(t, child)
			}
		}
	}

	// Iterates over `tracesCache` to find and set their trace children.
	for _, trace := range tracesCache {
		traceChildren, present := children[trace.ID.Trace]
		if present {
			if err := addChildren(trace, traceChildren); err != nil {
				return nil, err
			}
		}
		traces = append(traces, trace)
	}
	return traces, nil
}

// Close flushes the last batch to InfluxDB and shuts down the InfluxDBStore.
func (in *InfluxDBStore) Close() error {
	close(in.flusherStopChan)
	if err := in.flush(); err != nil {
		in.log.Println("Flush:", err)
	}
	return in.server.Close()
}

func (in *InfluxDBStore) createDBIfNotExists() error {
	q := fmt.Sprintf("CREATE DATABASE %s", in.dbName)

	// If a default retention policy is provided, it's used to extend the query in order to create the database with
	// a default retention policy.
	rp := in.config.DefaultRP
	if rp.Duration != "" {
		q = fmt.Sprintf("%s WITH DURATION %s", q, rp.Duration)

		// Retention policy name must be placed to the end of the query or it will be syntactically incorrect.
		if rp.Name != "" {
			q = fmt.Sprintf("%s NAME %s", q, rp.Name)
		}
	}

	// If there are no errors, query execution was successfully - either DB was created or already exists.
	return in.executeQueryNoResults(q)
}

// createContinuousQueries creates the continuous queries used by Appdash. If
// they already exist, no error is returned.
func (in *InfluxDBStore) createContinuousQueries() error {
	// The 'GROUP BY' (or 'bucket', if you will) size is calculated as the
	// maximum number of samples we want over 72hr. The more samples we have,
	// the more precise the stats on the Dashboard are and the slower query
	// time is. 5k data points on a 2015 MBP takes about 500ms to calculate
	// our statistics currently, thus:
	//
	// 	5000 points / (72h * 60m) == 1.157 (points per minute) == 1 point in 69s == GROUP BY 1m
	//
	// Because we're downsampling we can downsample different methods (mean,
	// min, max, etc) for the most accurate query later on. Thus we have a
	// mapping of aggregated -> real query likeso:
	//
	// 	| Dashboard      | Aggregated         | Real Query     | Description                     |
	// 	|----------------|--------------------|----------------|---------------------------------|
	// 	| Average        | MEAN("duration")   | MEAN("mean")   | average of all averages         |
	// 	| Min            | MIN("duration")    | MIN("min")     | minimum of all minimums         |
	// 	| Max            | MAX("duration")    | MAX("max")     | maximum of all maximums         |
	// 	| Std. Deviation | STDDEV("duration") | MEAN("stddev") | average of all std. deviations  |
	// 	| Timespans      | COUNT("duration")  | SUM("count")   | total counted durations (spans) |
	//
	q := fmt.Sprintf(`CREATE CONTINUOUS QUERY cq_downsampled_spans_1m ON %s RESAMPLE EVERY 1m BEGIN SELECT MEAN("duration"),MIN("duration"),MAX("duration"),STDDEV("duration"),COUNT("duration") INTO downsampled_spans FROM spans GROUP BY time(1m), "name" END`, in.dbName)
	if err := in.executeQueryNoResults(q); err != nil {
		return err
	}

	// The above query doesn't give us a reference to the trace IDs because
	// InfluxDB doesn't yet support mixing aggregate and non-aggregate queries
	// together. To workaround this, we create a secondary CQ. This CQ
	// downsamples the N-slowest span IDs by name so that we can efficiently
	// link to them from the dashboard.
	//
	// Note: This actually only gives us 1 N-slowest trace over the GROUP BY
	// timeframe (1 N-slowest trace over 1m), but this is good enough in
	// general.
	q = fmt.Sprintf(`CREATE CONTINUOUS QUERY cq_nslowest_spans_1m ON %s RESAMPLE EVERY 1m BEGIN SELECT TOP("duration",1),trace_id,span_id,parent_id INTO nslowest_spans FROM spans GROUP BY time(1m), "name" END`, in.dbName)
	return in.executeQueryNoResults(q)
}

// createAdminUserIfNotExists finds admin user(`in.adminUser`) if not found it's created.
func (in *InfluxDBStore) createAdminUserIfNotExists() error {
	admin := in.config.AdminUser
	userInfo, err := in.server.MetaClient.Authenticate(admin.Username, admin.Password)
	if err == influxDBErrors.ErrUserNotFound {
		if _, createUserErr := in.server.MetaClient.CreateUser(admin.Username, admin.Password, true); createUserErr != nil {
			return createUserErr
		}
		return nil
	} else {
		return err
	}
	if !userInfo.Admin { // must be admin user.
		return errors.New("failed to validate InfluxDB user type, found non-admin user")
	}
	return nil
}

func (in *InfluxDBStore) executeOneQuery(command string) (*influxDBClient.Result, error) {
	response, err := in.con.Query(influxDBClient.Query{
		Command:  command,
		Database: in.dbName,
	})
	if err != nil {
		return nil, err
	}
	if err := response.Error(); err != nil {
		return nil, err
	}

	// Expecting one result, since a single query is executed.
	if len(response.Results) != 1 {
		return nil, errors.New("unexpected number of results for an influxdb single query")
	}
	return &response.Results[0], nil
}

// executeQueryNoResults is a helper function which executes a single query and
// expects no results. If any error occurs, it is returned.
func (in *InfluxDBStore) executeQueryNoResults(command string) error {
	response, err := in.con.Query(influxDBClient.Query{
		Command:  command,
		Database: in.dbName,
	})
	if err != nil {
		return err
	}
	if err := response.Error(); err != nil {
		return err
	}
	return nil
}

func (in *InfluxDBStore) init(server *influxDBServer.Server) error {
	in.server = server
	// TODO: Upgrade to client v2, see: github.com/influxdata/influxdb/blob/master/client/v2/client.go
	// We're currently using v1.
	con, err := influxDBClient.NewClient(influxDBClient.Config{
		URL:      *in.clientTarget,
		Username: in.config.AdminUser.Username,
		Password: in.config.AdminUser.Password,
	})
	if err != nil {
		return err
	}
	in.con = con
	if err := in.createAdminUserIfNotExists(); err != nil {
		return err
	}
	switch in.config.Mode {
	case testMode:
		if err := in.setUpTestMode(); err != nil {
			return err
		}
	default:
		if err := in.setUpReleaseMode(); err != nil {
			return err
		}
	}
	if err := in.createDBIfNotExists(); err != nil {
		return err
	}
	if err := in.createContinuousQueries(); err != nil {
		return err
	}

	// TODO: let lib users decide `in.tracesPerPage` through InfluxDBConfig.
	in.tracesPerPage = defaultTracesPerPage

	go in.flusher()
	return nil
}

func (in *InfluxDBStore) setUpReleaseMode() error {
	in.dbName = releaseDBName
	return nil
}

func (in *InfluxDBStore) setUpTestMode() error {
	in.dbName = testDBName
	return in.executeQueryNoResults(fmt.Sprintf("DROP DATABASE IF EXISTS %s", in.dbName))
}

func annotationsFromEvents(a Annotations) (Annotations, error) {
	var (
		annotations Annotations
		events      []Event
	)
	if err := UnmarshalEvents(a, &events); err != nil {
		return nil, err
	}
	for _, e := range events {
		anns, err := MarshalEvent(e)
		if err != nil {
			return nil, err
		}
		annotations = append(annotations, anns...)
	}
	return annotations, nil
}

// fieldToSpanID converts given field to span ID.
func fieldToSpanID(field interface{}, errFieldType error) (*ID, error) {
	f, ok := field.(string)
	if !ok {
		return nil, errFieldType
	}
	id, err := ParseID(f)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

// filterSchemas returns `Annotations` which contains items taken from `anns`.
// Some items from `anns` won't be included(those which were not saved by `InfluxDBStore.Collect(...)`).
func filterSchemas(anns []Annotation) Annotations {
	var annotations Annotations

	// Finds an annotation which: `Annotation.Key` is equal to `schemasFieldName`.
	schemasAnn := findSchemasAnnotation(anns)

	// Converts `schemasAnn.Value` into slice of strings, each item is a schema.
	// Eg. schemas := []string{"HTTPClient", "HTTPServer"}
	schemas := strings.Split(string(schemasAnn.Value), schemasFieldSeparator)

	// Iterates over `anns` to check if each annotation should be included or not to the `annotations` be returned.
	for _, a := range anns {
		if strings.HasPrefix(a.Key, schemaPrefix) { // Check if current annotation is schema related one.
			schema := a.Key[len(schemaPrefix):] // Excludes the schema prefix part.

			// Checks if `schema` exists in `schemas`, if so means current annotation was saved by `InfluxDBStore.Collect(...)`.
			// If does not exist it means current annotation is empty on `InfluxDBStore.dbName` but still included within a query result.
			// Eg. If point "f" with a field "foo" & point "b" with a field "bar" are written to the same InfluxDB measurement
			// and later queried, the result will include two fields: "foo" & "bar" for both points, even though each was written with one field.
			if schemaExists(schema, schemas) { // Saved by `InfluxDBStore.Collect(...)` so should be added.
				annotations = append(annotations, a)
			} else { // Do not add current annotation, is empty & not saved by `InfluxDBStore.Collect(...)`.
				continue
			}
		} else {
			// Not a schema related annotation so just add it.
			annotations = append(annotations, a)
		}
	}
	return annotations
}

// schemaExists checks if `schema` is present on `schemas`.
func schemaExists(schema string, schemas []string) bool {
	for _, s := range schemas {
		if schema == s {
			return true
		}
	}
	return false
}

// findSchemasAnnotation finds & returns an annotation which: `Annotation.Key` is equal to `schemasFieldName`.
func findSchemasAnnotation(anns []Annotation) *Annotation {
	for _, a := range anns {
		if a.Key == schemasFieldName {
			return &a
		}
	}
	return nil
}

// findTraceParent walks through `rootTrace` to look for `child`; once found — it's trace parent is returned.
func findTraceParent(root, child *Trace) *Trace {
	var walkToParent func(root, child *Trace) *Trace
	walkToParent = func(root, child *Trace) *Trace {
		if root.ID.Span == child.ID.Parent {
			return root
		}
		for _, sub := range root.Sub {
			if sub.ID.Span == child.ID.Parent {
				return sub
			}
			if r := walkToParent(sub, child); r != nil {
				return r
			}
		}
		return nil
	}
	return walkToParent(root, child)
}

// schemasFromAnnotations returns a string(a set of schemas(strings) separated by `schemasFieldSeparator`) - eg. "HTTPClient,HTTPServer,name".
// Each schema is extracted from each `Annotation.Key` from `anns`.
func schemasFromAnnotations(anns []Annotation) string {
	var schemas []string
	for _, ann := range anns {

		// Checks if current annotation is schema related.
		if strings.HasPrefix(ann.Key, schemaPrefix) {
			schemas = append(schemas, ann.Key[len(schemaPrefix):])
		}
	}
	return strings.Join(schemas, schemasFieldSeparator)
}

// addChildren adds `children` to `root`; each child is appended to it's trace parent.
func addChildren(root *Trace, children []*Trace) error {
	var (
		addFn   func()                 // Handles children appending to it's trace parent.
		retries int    = len(children) // Maximum number of retries to add `children` elements to `root`.
		try     int                    // Current number of try to add `children` elements to `root`.
	)
	addFn = func() {
		for i := len(children) - 1; i >= 0; i-- {
			child := children[i]
			t := findTraceParent(root, child)
			if t != nil { // Trace found.
				if t.Sub == nil { // Empty sub-traces slice.
					t.Sub = []*Trace{child}
				} else { // Non-empty sub-traces slice.
					t.Sub = append(t.Sub, child)
				}

				// Removes current child since was added to it's parent.
				children = append(children[:i], children[i+1:]...)
			}
		}
	}

	// Loops until all `children` elements were added to it's trace parent or when maximum number of retries reached.
	for {
		if len(children) == 0 {
			break
		}

		// At this point, all children were added to their parent spans. Any children
		// left over in the children slice do not have parents. This could happen if,
		// for example, a parent service fails to record its span information to the
		// collection server but its downstream services do send their span information
		// properly. In this case, we gracefully degrade by adding these orphan spans to
		// the root span.
		if try == retries {

			// Iterates over children(without parent found on `root`) and appends them as sub-traces to `root`.
			for _, child := range children {
				if child.ID.Trace == root.ID.Trace {
					root.Sub = append(root.Sub, child)
				}
			}
			return nil
		}
		addFn()
		try++
	}
	return nil
}

// pointMemoryUsage returns an estimate of the memory usage of the given *influxDBClient.Point
// object in bytes.
func pointMemoryUsage(v reflect.Value) int {
	s := int(v.Type().Size())
	switch v.Kind() {
	case reflect.String:
		s += v.Len()
	case reflect.Ptr:
		if !v.IsNil() {
			s += pointMemoryUsage(reflect.Indirect(v))
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			s += pointMemoryUsage(v.MapIndex(key))
		}
	}
	return s
}

// withoutEmptyFields filters `pf` and returns `pointFields` excluding those that have empty values.
func withoutEmptyFields(pf pointFields) pointFields {
	r := make(pointFields, 0)
	for k, v := range pf {
		switch v.(type) {
		case string:
			if v.(string) == "" {
				continue
			}
			r[k] = v
		case nil:
			continue
		default:
			r[k] = v
		}
	}
	return r
}

// spansFromRow rebuilds spans from given InfluxDB row.
func spansFromRow(row influxDBModels.Row) ([]*Span, error) {
	var spans []*Span

	// Iterates through all `row.Values`, each represents a set of single span information(ids and annotations).
	for _, fields := range row.Values {
		span := &Span{
			Annotations: make(Annotations, 0),
		}

		// Iterates over fields, each field might be a span's annotation value or span's ID(either a TraceID, SpanID or ParentID).
		for i, field := range fields {
			var (
				// column is current field column's name (Eg. field ='Server.Request.Method', columns[i] = 'GET').
				column string = row.Columns[i]

				// errFieldType is an error for unexpected field type.
				errFieldType error = fmt.Errorf("unexpected field type: %v", reflect.TypeOf(field))
			)

			// Checks if current column is some span's ID, if so set to the span & continue with next field.
			switch column {
			case "name", "duration":
				continue // aggregation
			case "trace_id":
				traceID, err := fieldToSpanID(field, errFieldType)
				if err != nil {
					return spans, err
				}
				span.ID.Trace = *traceID
				continue
			case "span_id":
				spanID, err := fieldToSpanID(field, errFieldType)
				if err != nil {
					return spans, err
				}
				span.ID.Span = *spanID
				continue
			case "parent_id":
				parentID, err := fieldToSpanID(field, errFieldType)
				if err != nil {
					return spans, err
				}
				span.ID.Parent = *parentID
				continue
			}

			// At this point the current field is a span's annotation value.
			var value []byte // Annotation value.
			switch field.(type) {
			case string:
				value = []byte(field.(string))
			case nil:
			default:
				return nil, errFieldType
			}
			span.Annotations = append(span.Annotations, Annotation{
				Key:   column,
				Value: value,
			})
		}
		anns, err := annotationsFromEvents(filterSchemas(span.Annotations))
		if err != nil {
			return nil, err
		}
		span.Annotations = anns
		spans = append(spans, span)
	}
	return spans, nil
}

// NewInfluxDBStore returns a new InfluxDB-backed store. It starts an
// in-process / embedded InfluxDB server.
func NewInfluxDBStore(config *InfluxDBConfig) (*InfluxDBStore, error) {
	s, err := influxDBServer.NewServer(config.Server, config.BuildInfo)
	if err != nil {
		return nil, err
	}

	// If the user specified a log output location, configure everything to use that.
	if config.LogOutput != nil {
		s.SetLogOutput(config.LogOutput)
	}
	if err := s.Open(); err != nil {
		return nil, err
	}

	httpdAddr := config.Server.HTTPD.BindAddress
	if httpdAddr == "" {
		httpdAddr = fmt.Sprintf("%s:%d", influxDBClient.DefaultHost, influxDBClient.DefaultPort)
	}
	clientTarget, err := url.Parse("http://" + httpdAddr)
	if err != nil {
		return nil, err
	}
	in := InfluxDBStore{
		config:       config,
		clientTarget: clientTarget,
		log:          log.New(os.Stderr, "appdash: InfluxDBStore: ", log.LstdFlags),
	}
	if err := in.init(s); err != nil {
		return nil, err
	}
	return &in, nil
}

// NewInfluxDBConfig returns a new InfluxDBConfig with the default values.
func NewInfluxDBConfig() (*InfluxDBConfig, error) {
	// Create the default InfluxDB server configuration.
	server, err := influxDBServer.NewDemoConfig()
	if err != nil {
		return nil, err
	}

	// Enables retention policies which will be executed within an interval of 30 minutes.
	server.Retention.Enabled = true
	server.Retention.CheckInterval = toml.Duration(30 * time.Minute)

	return &InfluxDBConfig{
		Server: server,

		// Specify the branch as "appdash" just for identification purposes.
		BuildInfo: &influxDBServer.BuildInfo{
			Branch: "appdash",
		},

		// Create a retention policy which keeps data for only three days, this is
		// because the Dashboard is hard-coded to displaying a 72hr timeline.
		//
		// Minimum duration time is 1 hour ("1h") - See: github.com/influxdata/influxdb/issues/5198
		DefaultRP: InfluxDBRetentionPolicy{
			Name:     "three_days_only",
			Duration: "3d",
		},

		MaxBatchSizeBytes:  128 * 1024 * 1024, // 128 MB
		BatchFlushInterval: 500 * time.Millisecond,
	}, nil
}

type InfluxDBConfig struct {
	AdminUser InfluxDBAdminUser
	BuildInfo *influxDBServer.BuildInfo
	DefaultRP InfluxDBRetentionPolicy
	Mode      mode
	Server    *influxDBServer.Config

	// LogOutput, if specified, controls where all InfluxDB logs are written to.
	LogOutput io.Writer

	// MaxBatchSizeBytes specifies the maximum size (estimated memory usage) in
	// bytes that a batch may grow to become before being entirely dropped (and
	// inherently, trace data lost). This prevents any potential memory leak in
	// the case of an unresponsive or too slow InfluxDB server / pending flush
	// operation.
	//
	// The default value used by NewInfluxDBConfig is 128*1024*1024 (128 MB).
	MaxBatchSizeBytes int

	// BatchFlushInterval specifies the minimum interval between flush calls by
	// the background goroutine in order to flush point batches out to
	// InfluxDB. That is, after each batch flush the goroutine will sleep for
	// this amount of time to prevent CPU overutilization.
	//
	// The default value used by NewInfluxDBConfig is 500 * time.Millisecond.
	BatchFlushInterval time.Duration
}

type InfluxDBRetentionPolicy struct {
	Name     string // Name used to indentify this retention policy.
	Duration string // How long InfluxDB keeps the data. Eg: "1h", "1d", "1w".
}

type InfluxDBAdminUser struct {
	Username string
	Password string
}
