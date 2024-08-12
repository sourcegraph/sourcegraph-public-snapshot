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

type contextKey int

const contextKeyWithoutTrace contextKey = iota

// WithoutTrace disables CloudSQL connection tracing for child contexts.
func WithoutTrace(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKeyWithoutTrace, true)
}

func shouldNotTrace(ctx context.Context) bool {
	withoutTrace, ok := ctx.Value(contextKeyWithoutTrace).(bool)
	return ok && withoutTrace
}

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
	if shouldNotTrace(ctx) {
		return ctx // do nothing
	}

	ctx, _ = tracer.Start(ctx, "pgx.Query",
		trace.WithAttributes(
			attribute.String("query", data.SQL),
			attribute.Int("args.len", len(data.Args)),
		),
		trace.WithAttributes(argsAsAttributes(data.Args)...))
	return ctx
}

func (pgxTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	if shouldNotTrace(ctx) {
		return // do nothing
	}

	span := trace.SpanFromContext(ctx)
	defer span.End()

	span.SetAttributes(
		attribute.String("command_tag", data.CommandTag.String()),
		attribute.Int64("rows_affected", data.CommandTag.RowsAffected()),
	)
	span.SetName("pgx.Query: " + data.CommandTag.String())

	if data.Err != nil {
		span.SetStatus(codes.Error, data.Err.Error())
	}
}

func (pgxTracer) TraceConnectStart(ctx context.Context, data pgx.TraceConnectStartData) context.Context {
	if shouldNotTrace(ctx) {
		return ctx // do nothing
	}

	ctx, _ = tracer.Start(ctx, "pgx.Connect", trace.WithAttributes(
		attribute.String("database", data.ConnConfig.Database),
		attribute.String("instance", fmt.Sprintf("%s:%d", data.ConnConfig.Host, data.ConnConfig.Port)),
		attribute.String("user", data.ConnConfig.User)))
	return ctx
}

func (pgxTracer) TraceConnectEnd(ctx context.Context, data pgx.TraceConnectEndData) {
	if shouldNotTrace(ctx) {
		return // do nothing
	}

	span := trace.SpanFromContext(ctx)
	defer span.End()

	if data.Err != nil {
		span.SetStatus(codes.Error, data.Err.Error())
	}
}

// Number of SQL arguments allowed to enable argument instrumentation
const argsAttributesCountLimit = 24

// Maximum size to use in heuristics of SQL arguments size in argument
// instrumentation before they are truncated. This is NOT a hard cap on bytes
// size of values.
const argsAttributesValueLimit = 240

func argsAsAttributes(args []any) []attribute.KeyValue {
	if len(args) > argsAttributesCountLimit {
		return []attribute.KeyValue{
			attribute.String("db.args", "too many args"),
		}
	}

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

		case pgx.NamedArgs:
			attrs[i] = attribute.String(key, truncateStringValue(fmt.Sprintf("%+v", v)))

		default: // in case we miss anything
			attrs[i] = attribute.String(key, fmt.Sprintf("unhandled type %T", v))
		}
	}
	return attrs
}

// utf8Replace is the same value as used in other strings.ToValidUTF8 callsites
// in sourcegraph/sourcegraph.
const utf8Replace = "�"

// truncateStringValue should be used on all string attributes in the otelsql
// instrumentation. It ensures the length of v is within argsAttributesValueLimit,
// and ensures v is valid UTF8.
func truncateStringValue(v string) string {
	if len(v) > argsAttributesValueLimit {
		return strings.ToValidUTF8(v[:argsAttributesValueLimit], utf8Replace)
	}
	return strings.ToValidUTF8(v, utf8Replace)
}
