package cloudsql

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.GetTracerProvider().Tracer("msp/cloudsql/pgx")

type pgxTracer struct{}

// Select tracing hooks we want to implement.
var (
	_ pgx.QueryTracer   = pgxTracer{}
	_ pgx.ConnectTracer = pgxTracer{}
	// Future:
	// _ pgx.BatchTracer    = pgxTracer{}
	// _ pgx.CopyFromTracer = pgxTracer{}
	// _ pgx.PrepareTracer  = pgxTracer{}
)

// TraceQueryStart is called at the beginning of Query, QueryRow, and Exec calls. The returned context is used for the
// rest of the call and will be passed to TraceQueryEnd.
func (pgxTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	ctx, _ = tracer.Start(ctx, "pgx.Query",
		trace.WithAttributes(
			attribute.String("query", data.SQL),
			attribute.Int("args.len", len(data.Args)),
		),
		trace.WithAttributes(argsAsAttributes(data.Args)...))
	return ctx
}

func (pgxTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	span := trace.SpanFromContext(ctx)
	defer span.End()

	span.SetAttributes(
		attribute.String("command_tag", data.CommandTag.String()),
		attribute.Int64("rows_affected", data.CommandTag.RowsAffected()),
	)
	switch {
	case data.CommandTag.Insert():
		span.SetName("pgx.Query: INSERT")
	case data.CommandTag.Update():
		span.SetName("pgx.Query: UPDATE")
	case data.CommandTag.Delete():
		span.SetName("pgx.Query: DELETE")
	case data.CommandTag.Select():
		span.SetName("pgx.Query: SELECT")
	}

	if data.Err != nil {
		span.SetStatus(codes.Error, data.Err.Error())
	}
}

func (pgxTracer) TraceConnectStart(ctx context.Context, data pgx.TraceConnectStartData) context.Context {
	ctx, _ = tracer.Start(ctx, "pgx.Connect", trace.WithAttributes(
		attribute.String("database", data.ConnConfig.Database),
		attribute.String("instance", fmt.Sprintf("%s:%d", data.ConnConfig.Host, data.ConnConfig.Port)),
		attribute.String("user", data.ConnConfig.User)))
	return ctx
}

func (pgxTracer) TraceConnectEnd(ctx context.Context, data pgx.TraceConnectEndData) {
	span := trace.SpanFromContext(ctx)
	defer span.End()

	if data.Err != nil {
		span.SetStatus(codes.Error, data.Err.Error())
	}
}

func argsAsAttributes(args []any) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, len(args))
	for i, arg := range args {
		key := "db.args.$" + strconv.Itoa(i)
		switch v := arg.(type) {
		case nil:
			attrs[i] = attribute.String(key, "nil")
		case int:
			attrs[i] = attribute.Int(key, v)
		case int32:
			attrs[i] = attribute.Int(key, int(v))
		case int64:
			attrs[i] = attribute.Int64(key, v)
		case float32:
			attrs[i] = attribute.Float64(key, float64(v))
		case float64:
			attrs[i] = attribute.Float64(key, v)
		case bool:
			attrs[i] = attribute.Bool(key, v)
		case []byte:
			attrs[i] = attribute.String(key, truncateStringValue(string(v)))
		case string:
			attrs[i] = attribute.String(key, truncateStringValue(v))
		case time.Time:
			attrs[i] = attribute.String(key, v.String())

		default: // in case we miss anything
			attrs[i] = attribute.String(key, fmt.Sprintf("unhandled type %T", v))
		}
	}
	return attrs
}

// utf8Replace is the same value as used in other strings.ToValidUTF8 callsites
// in sourcegraph/sourcegraph.
const utf8Replace = "�"

// Maximum size to use in heuristics of SQL arguments size in argument
// instrumentation before they are truncated. This is NOT a hard cap on bytes
// size of values.
const argsAttributesValueLimit = 240

// truncateStringValue should be used on all string attributes in the otelsql
// instrumentation. It ensures the length of v is within argsAttributesValueLimit,
// and ensures v is valid UTF8.
func truncateStringValue(v string) string {
	if len(v) > argsAttributesValueLimit {
		return strings.ToValidUTF8(v[:argsAttributesValueLimit], utf8Replace)
	}
	return strings.ToValidUTF8(v, utf8Replace)
}
