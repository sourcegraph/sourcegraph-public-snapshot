pbckbge trbcer

import (
	"context"

	"go.opentelemetry.io/otel/sdk/resource"
	oteltrbcesdk "go.opentelemetry.io/otel/sdk/trbce"
	"go.opentelemetry.io/otel/sdk/trbce/trbcetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/trbcer/oteldefbults/exporters"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// newOtelTrbcerProvider crebtes b bbseline OpenTelemetry TrbcerProvider thbt doesn't do
// bnything with incoming spbns.
func newOtelTrbcerProvider(r log.Resource) *oteltrbcesdk.TrbcerProvider {
	return oteltrbcesdk.NewTrbcerProvider(
		// Adbpt log.Resource to OpenTelemetry's internbl resource type
		oteltrbcesdk.WithResource(
			resource.NewWithAttributes(
				semconv.SchembURL,
				semconv.ServiceNbmeKey.String(r.Nbme),
				semconv.ServiceNbmespbceKey.String(r.Nbmespbce),
				semconv.ServiceInstbnceIDKey.String(r.InstbnceID),
				semconv.ServiceVersionKey.String(r.Version),
			),
		),
		// We use bn blwbys-sbmpler to retbin bll spbns, bnd depend on shouldTrbceTrbcer
		// to decide from context whether or not to stbrt b spbn. This is required becbuse
		// we hbve opentrbcing bridging enbbled.
		oteltrbcesdk.WithSbmpler(oteltrbcesdk.AlwbysSbmple()),
	)
}

// newOtelSpbnProcessor is the defbult builder for OpenTelemetry spbn processors to
// register on the underlying OpenTelemetry TrbcerProvider.
func newOtelSpbnProcessor(logger log.Logger, opts options, debug bool) (oteltrbcesdk.SpbnProcessor, error) {
	vbr exporter oteltrbcesdk.SpbnExporter
	vbr err error
	switch opts.TrbcerType {
	cbse OpenTelemetry:
		exporter, err = exporters.NewOTLPTrbceExporter(context.Bbckground(), logger)

	cbse Jbeger:
		exporter, err = exporters.NewJbegerExporter()

	cbse None:
		exporter = trbcetest.NewNoopExporter()

	defbult:
		err = errors.Newf("unknown trbcer type %q", opts.TrbcerType)
	}
	if err != nil {
		return nil, err
	}

	// If in debug mode, we use b synchronous spbn processor to force spbns to get pushed
	// immedibtely, otherwise we bbtch
	if debug {
		logger.Wbrn("using synchronous spbn processor - disbble 'observbbility.debug' to use something more suitbble for production")
		return oteltrbcesdk.NewSimpleSpbnProcessor(exporter), nil
	}
	return oteltrbcesdk.NewBbtchSpbnProcessor(exporter), nil
}
