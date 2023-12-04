package dbconn

import (
	"context"
	"fmt"
	"hash/fnv"
	"path/filepath"
	"runtime"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// instrumentQuery modifies the query text to include front-loaded metadata that is
// useful when looking at global queries in a Postgres instance such as with Cloud SQL
// Query Insights.
//
// Metadata added includes:
//   - the query text's hash (correlates traces + query insights)
//   - the query length and number of arguments
//   - the calling function name and source location (inferred by stack trace)
//
// This method returns both a modified context and SQL query text. The context is
// used to add the query hash into the trace so that particular hash can be searched
// when query text is available.
func instrumentQuery(ctx context.Context, query string, numArguments int) (context.Context, string) {
	hash := hash(query)

	hashPrefix := fmt.Sprintf("-- query hash: %d", hash)
	lengthPrefix := fmt.Sprintf("-- query length: %d (%d args)", len(query), numArguments)
	metadataLines := []string{hashPrefix, lengthPrefix}

	callerPrefix, ok := getSourceMetadata(ctx)
	if ok {
		metadataLines = append(metadataLines, callerPrefix)
	} else {
		metadataLines = append(metadataLines, "-- (could not infer source)")
	}

	// Set the hash on the span.
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.Int64("db.statement.checksum", int64(hash)))

	return ctx, strings.Join(append(metadataLines, query), "\n")
}

// hash returns the 32-bit FNV-1a hash of the given query text.
func hash(query string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(query))
	return h.Sum32()
}

type functionsSkippedForQuerySourceType struct{}

var functionsSkippedForQuerySource = functionsSkippedForQuerySourceType{}

func getFunctionsSkippedForQuerySource(ctx context.Context) []string {
	skips, _ := ctx.Value(functionsSkippedForQuerySource).([]string)
	return skips
}

// SkipFrameForQuerySource adds the function in which this method was called to a list
// of functions to be skipped when inferring the relevant source location executing a
// given query.
//
// This should be applied to contexts in helper functions, or shim layers that only
// proxy calls to the underlying handle(s).
func SkipFrameForQuerySource(ctx context.Context) context.Context {
	frame, ok := getFrames().Next()
	if !ok {
		return ctx
	}

	current := getFunctionsSkippedForQuerySource(ctx)
	updated := append(current, frame.Function)
	return context.WithValue(ctx, functionsSkippedForQuerySource, updated)
}

const sourcegraphPrefix = "github.com/sourcegraph/sourcegraph/"

var dropFramesFromPackages = []string{
	sourcegraphPrefix + "internal/database/basestore",
	sourcegraphPrefix + "internal/database/batch",
	sourcegraphPrefix + "internal/database/connections",
	sourcegraphPrefix + "internal/database/dbconn",
	sourcegraphPrefix + "internal/database/dbtest",
	sourcegraphPrefix + "internal/database/dbutil",
	sourcegraphPrefix + "internal/database/locker",
	sourcegraphPrefix + "internal/database/migration",
}

// getSourceMetadata returns the metadata line indicating the inferred source location
// of the caller.
func getSourceMetadata(ctx context.Context) (string, bool) {
	frames := getFrames()

frameLoop:
	for {
		frame, ok := frames.Next()
		if !ok {
			break
		}

		// If we're in a third-party package, skip
		if !strings.HasPrefix(frame.Function, sourcegraphPrefix) {
			continue
		}

		// If we're in a package that deals with connections and SQL machinery
		// rather than performing queries for application data, skip
		for _, prefix := range dropFramesFromPackages {
			if strings.HasPrefix(frame.Function, prefix) {
				continue frameLoop
			}
		}

		// If we match a function that was explicitly tagged as not the true
		// source of the query, skip
		for _, function := range getFunctionsSkippedForQuerySource(ctx) {
			if frame.Function == function {
				continue frameLoop
			}
		}

		// Trim the frame function to exclude the common prefix
		functionName := frame.Function[len(sourcegraphPrefix):]

		// Reconstruct the frame file path so that we don't include the local
		// path on the machine that built this instance
		pathPrefix := strings.Split(functionName, ".")[0]
		file := filepath.Join(pathPrefix, filepath.Base(frame.File))

		// Construct metadata values
		callerLine := fmt.Sprintf("-- caller: %s", functionName)
		sourceLine := fmt.Sprintf("-- source: %s:%d", file, frame.Line)
		return callerLine + "\n" + sourceLine, true
	}

	return "", false
}

const pcLen = 1024

func getFrames() *runtime.Frames {
	skip := 3 // caller of caller
	pc := make([]uintptr, pcLen)
	n := runtime.Callers(skip, pc)
	return runtime.CallersFrames(pc[:n])
}
