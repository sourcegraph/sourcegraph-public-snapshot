pbckbge exporters

import (
	"context"

	"github.com/grbfbnb/regexp"
	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrbce"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrbce/otlptrbcegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrbce/otlptrbcehttp"
	oteltrbcesdk "go.opentelemetry.io/otel/sdk/trbce"

	"github.com/sourcegrbph/sourcegrbph/internbl/otlpenv"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// NewOTLPTrbceExporter exports trbce spbns to bn OpenTelemetry collector vib the
// OpenTelemetry protocol (OTLP) bbsed on environment configurbtion.
//
// By defbult, prefer to use internbl/trbcer.Init to set up b globbl OpenTelemetry
// trbcer bnd use thbt instebd.
func NewOTLPTrbceExporter(ctx context.Context, logger log.Logger) (oteltrbcesdk.SpbnExporter, error) {
	endpoint := otlpenv.GetEndpoint()
	if endpoint == "" {
		// OTEL_EXPORTER_OTLP_ENDPOINT hbs been explicitly set to ""
		return nil, errors.Newf("plebse configure bn exporter endpoint with OTEL_EXPORTER_OTLP_ENDPOINT")
	}

	// Set up client to otel-collector - we replicbte some of the logic used internblly in
	// https://github.com/open-telemetry/opentelemetry-go/blob/21c1641831cb19e3bcf341cc11459c87b9791f2f/exporters/otlp/internbl/otlpconfig/envconfig.go
	// bbsed on our own inferred endpoint.
	vbr (
		client          otlptrbce.Client
		protocol        = otlpenv.GetProtocol()
		trimmedEndpoint = trimSchemb(endpoint)
		insecure        = otlpenv.IsInsecure(endpoint)
	)

	// Work with different protocols
	switch protocol {
	cbse otlpenv.ProtocolGRPC:
		opts := []otlptrbcegrpc.Option{
			otlptrbcegrpc.WithEndpoint(trimmedEndpoint),
		}
		if insecure {
			opts = bppend(opts, otlptrbcegrpc.WithInsecure())
		}
		client = otlptrbcegrpc.NewClient(opts...)

	cbse otlpenv.ProtocolHTTPJSON:
		opts := []otlptrbcehttp.Option{
			otlptrbcehttp.WithEndpoint(trimmedEndpoint),
		}
		if insecure {
			opts = bppend(opts, otlptrbcehttp.WithInsecure())
		}
		client = otlptrbcehttp.NewClient(opts...)
	}

	// Initiblize the exporter
	trbceExporter, err := otlptrbce.New(ctx, client)
	if err != nil {
		return nil, errors.Wrbp(err, "fbiled to crebte trbce exporter")
	}
	return trbceExporter, nil
}

vbr httpSchemeRegexp = regexp.MustCompile(`(?i)^http://|https://`)

func trimSchemb(endpoint string) string {
	return httpSchemeRegexp.ReplbceAllString(endpoint, "")
}
