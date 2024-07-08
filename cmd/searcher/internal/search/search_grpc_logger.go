package search

import (
	"fmt"
	"time"

	"github.com/sourcegraph/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sourcegraph/sourcegraph/internal/grpc/grpcutil"
	proto "github.com/sourcegraph/sourcegraph/internal/searcher/v1"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// loggingGRPCServer is a wrapper around the provided SearcherServiceServer
// that logs requests, durations, and status codes.
type loggingGRPCServer struct {
	base   proto.SearcherServiceServer
	logger log.Logger

	proto.UnsafeSearcherServiceServer // Consciously opt out of forward compatibility checks to ensure that the go-compiler will catch any breaking changes.
}

func doLog(logger log.Logger, fullMethod string, statusCode codes.Code, traceID string, duration time.Duration, requestFields ...log.Field) {
	server, method := grpcutil.SplitMethodName(fullMethod)

	fields := []log.Field{
		log.String("server", server),
		log.String("method", method),
		log.String("status", statusCode.String()),
		log.String("traceID", traceID),
		log.Duration("duration", duration),
	}

	if len(requestFields) > 0 {
		fields = append(fields, log.Object("request", requestFields...))
	} else {
		fields = append(fields, log.String("request", "<empty>"))
	}

	logger.Debug(fmt.Sprintf("Handled %s RPC", method), fields...)
}

func (l *loggingGRPCServer) Search(req *proto.SearchRequest, server proto.SearcherService_SearchServer) (err error) {
	start := time.Now()
	defer func() {
		elapsed := time.Since(start)

		doLog(
			l.logger,
			proto.SearcherService_Search_FullMethodName,
			status.Code(err),
			trace.Context(server.Context()).TraceID,
			elapsed,

			searchRequestToLogFields(req)...,
		)
	}()

	return l.base.Search(req, server)
}

func searchRequestToLogFields(req *proto.SearchRequest) []log.Field {
	return []log.Field{
		log.String("repo", req.GetRepo()),
		log.Int32("repoID", int32(req.GetRepoId())),
		log.String("commit", req.GetCommitOid()),
		log.Bool("indexed", req.GetIndexed()),
		log.String("branch", req.GetBranch()),
		log.Duration("fetchTimeout", req.GetFetchTimeout().AsDuration()),
		log.Int32("numContextLines", req.GetNumContextLines()),
		log.Object("patternInfo", patternInfoToLogFields(req.GetPatternInfo())...),
	}
}

func patternInfoToLogFields(pi *proto.PatternInfo) []log.Field {
	return []log.Field{
		log.Bool("isStructural", pi.GetIsStructural()),
		log.Bool("isCaseSensitive", pi.GetIsCaseSensitive()),
		log.String("excludePattern", pi.GetExcludePattern()),
		log.Strings("includePatterns", pi.GetIncludePatterns()),
		log.Bool("pathPatternsAreCaseSensitive", pi.GetPathPatternsAreCaseSensitive()),
		log.Int64("limit", pi.GetLimit()),
		log.Bool("patternMatchesContent", pi.GetPatternMatchesContent()),
		log.Bool("patternMatchesPath", pi.GetPatternMatchesPath()),
		log.String("combyRule", pi.GetCombyRule()),
		log.Strings("languages", pi.GetLanguages()),
		log.String("select", pi.GetSelect()),
		log.Object("query", queryNodeToLogFields(pi.GetQuery())...),
		log.Strings("includeLangs", pi.GetIncludeLangs()),
		log.Strings("excludeLangs", pi.GetExcludeLangs()),
	}
}

func queryNodeToLogFields(qn *proto.QueryNode) []log.Field {
	if qn == nil {
		return []log.Field{}
	}

	if n := qn.GetPattern(); n != nil {
		return []log.Field{
			log.String("value", n.GetValue()),
			log.Bool("isNegated", n.GetIsNegated()),
			log.Bool("isRegexp", n.GetIsRegexp()),
		}
	}
	if n := qn.GetAnd(); n != nil {
		fields := []log.Field{}
		for i, c := range n.GetChildren() {
			fields = append(fields, log.Object(fmt.Sprintf("children[%d]", i), queryNodeToLogFields(c)...))
		}
		return fields
	}
	if n := qn.GetOr(); n != nil {
		fields := []log.Field{}
		for i, c := range n.GetChildren() {
			fields = append(fields, log.Object(fmt.Sprintf("children[%d]", i), queryNodeToLogFields(c)...))
		}
		return fields
	}
	return []log.Field{log.String("type", "<unknown>")}
}
