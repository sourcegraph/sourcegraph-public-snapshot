package dbconn

import (
	"context"
	"fmt"
	"strconv"

	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/qustavo/sqlhooks/v2"

	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

var (
	// postgresBulkInsertRowsPattern matches `($1, $2, $3), ($4, $5, $6), ...` which
	// we use to cut out the row payloads from bulk insertion tracing data. We don't
	// need all the parameter data for such requests, which are too big to fit into
	// Jaeger spans. Note that we don't just capture `($1.*`, as we want queries with
	// a trailing ON CONFLICT clause not to be semantically mangled in the log output.
	postgresBulkInsertRowsPattern = lazyregexp.New(`(\([$\d,\s]+\)[,\s]*)+`)

	// postgresBulkInsertRowsReplacement replaces the all-placeholder rows matched
	// by the pattern defined above.
	postgresBulkInsertRowsReplacement = []byte("(...) ")
)

type tracingHooks struct{}

var _ sqlhooks.Hooks = &tracingHooks{}
var _ sqlhooks.OnErrorer = &tracingHooks{}

func (h *tracingHooks) Before(ctx context.Context, query string, args ...any) (context.Context, error) {
	if bulkInsertion(ctx) {
		query = string(postgresBulkInsertRowsPattern.ReplaceAll([]byte(query), postgresBulkInsertRowsReplacement))
	}

	tr, ctx := trace.New(ctx, "sql", query,
		trace.Tag{Key: "span.kind", Value: "client"},
		trace.Tag{Key: "database.type", Value: "sql"},
	)

	if !bulkInsertion(ctx) {
		tr.LogFields(otlog.Lazy(func(fv otlog.Encoder) {
			emittedChars := 0
			for i, arg := range args {
				k := strconv.Itoa(i + 1)
				v := fmt.Sprintf("%v", arg)
				emittedChars += len(k) + len(v)
				// Limit the amount of characters reported in a span because
				// a Jaeger span may not exceed 65k. Usually, the arguments are
				// not super helpful if it's so many of them anyways.
				if emittedChars > 32768 {
					fv.EmitString("more omitted", strconv.Itoa(len(args)-i))
					break
				}
				fv.EmitString(k, v)
			}
		}))
	} else {
		tr.LogFields(otlog.Bool("bulk_insert", true), otlog.Int("num_args", len(args)))
	}

	return ctx, nil
}

func (h *tracingHooks) After(ctx context.Context, query string, args ...any) (context.Context, error) {
	if tr := trace.TraceFromContext(ctx); tr != nil {
		tr.Finish()
	}

	return ctx, nil
}

func (h *tracingHooks) OnError(ctx context.Context, err error, query string, args ...any) error {
	if tr := trace.TraceFromContext(ctx); tr != nil {
		tr.SetError(err)
		tr.Finish()
	}

	return err
}

type key int

const bulkInsertionKey key = iota

// bulkInsertion returns true if the bulkInsertionKey context value is true.
func bulkInsertion(ctx context.Context) bool {
	v, ok := ctx.Value(bulkInsertionKey).(bool)
	if !ok {
		return false
	}
	return v
}

// WithBulkInsertion sets the bulkInsertionKey context value.
func WithBulkInsertion(ctx context.Context, bulkInsertion bool) context.Context {
	return context.WithValue(ctx, bulkInsertionKey, bulkInsertion)
}
