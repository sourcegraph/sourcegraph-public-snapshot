package internal

import (
	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
)

var (
	client         *gqltestutil.Client
	requestWriter  = &requestResponseWriter{}
	responseWriter = &requestResponseWriter{}
)

func InitializeGraphQLClient() (err error) {
	client, err = gqltestutil.NewClient(SourcegraphEndpoint, gqltestutil.ClientOption{
		GraphQLRequestLogger:  requestWriter.Write,
		GraphQLResponseLogger: responseWriter.Write,
	})
	return err
}

func GraphQLClient() *gqltestutil.Client {
	return client
}

func LastRequestResponsePair() (string, string) {
	return requestWriter.Last(), responseWriter.Last()
}

type requestResponseWriter struct {
	payloads []string
}

func (w *requestResponseWriter) Write(payload []byte) {
	w.payloads = append(w.payloads, string(payload))
}

func (w *requestResponseWriter) Last() string {
	if len(w.payloads) == 0 {
		return ""
	}

	return w.payloads[len(w.payloads)-1]
}
