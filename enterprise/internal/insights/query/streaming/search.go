package streaming

import (
	"context"
	"io"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/compute/client"
	"github.com/sourcegraph/sourcegraph/internal/api/internalapi"
	streamhttp "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

// Opts contains the search options supported by Search.
type Opts struct {
	Display int
	Trace   bool
	Json    bool
}

// Search calls the streaming search endpoint and uses decoder to decode the
// response body.
func Search(ctx context.Context, query string, patternType *string, decoder streamhttp.FrontendStreamDecoder) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "InsightsStreamSearch")
	defer func() {
		span.LogFields(
			log.Error(err),
		)
		span.Finish()
	}()
	req, err := streamhttp.NewRequest(internalapi.Client.URL+"/.internal", query)
	if err != nil {
		return err
	}
	if patternType != nil {
		req.URL.Query().Add("t", *patternType)
	}

	req = req.WithContext(ctx)
	req.Header.Set("User-Agent", "code-insights-backend")

	if span != nil {
		carrier := opentracing.HTTPHeadersCarrier(req.Header)
		span.Tracer().Inject(
			span.Context(),
			opentracing.HTTPHeaders,
			carrier)
	}

	resp, err := httpcli.InternalClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	decErr := decoder.ReadAll(resp.Body)
	if decErr != nil {
		return decErr
	}
	return err
}

func genericComputeStream(ctx context.Context, handler func(io.Reader) error, query, operation string) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, operation)
	defer func() {
		span.LogFields(
			log.Error(err),
		)
		span.Finish()
	}()

	req, err := client.NewComputeStreamRequest(internalapi.Client.URL+"/.internal", query)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	req.Header.Set("User-Agent", "code-insights-backend")

	if span != nil {
		carrier := opentracing.HTTPHeadersCarrier(req.Header)
		span.Tracer().Inject(
			span.Context(),
			opentracing.HTTPHeaders,
			carrier)
	}

	resp, err := httpcli.InternalClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return handler(resp.Body)
}

func ComputeMatchContextStream(ctx context.Context, query string, decoder client.ComputeMatchContextStreamDecoder) (err error) {
	return genericComputeStream(ctx, decoder.ReadAll, query, "InsightsComputeStreamSearch")
}

func ComputeTextExtraStream(ctx context.Context, query string, decoder client.ComputeTextExtraStreamDecoder) (err error) {
	return genericComputeStream(ctx, decoder.ReadAll, query, "InsightsComputeTextSearch")
}
