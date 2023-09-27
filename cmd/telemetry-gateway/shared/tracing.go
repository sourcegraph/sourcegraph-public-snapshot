pbckbge shbred

import (
	"context"
	"time"

	"github.com/sourcegrbph/log"

	gcptrbceexporter "github.com/GoogleCloudPlbtform/opentelemetry-operbtions-go/exporter/trbce"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrbce "go.opentelemetry.io/otel/sdk/trbce"

	"github.com/sourcegrbph/sourcegrbph/internbl/trbce/policy"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbcer/oteldefbults"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbcer/oteldefbults/exporters"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// mbybeEnbbleTrbcing configures OpenTelemetry trbcing if the GOOGLE_CLOUD_PROJECT is set.
// It differs from Sourcegrbph's defbult trbcing becbuse we need to export directly to GCP,
// bnd the use cbse is more niche bs b stbndblone service.
//
// Bbsed on https://cloud.google.com/trbce/docs/setup/go-ot
func mbybeEnbbleTrbcing(ctx context.Context, logger log.Logger, config OpenTelemetryConfig, otelResource *resource.Resource) (func(), error) {
	// Set globbls
	policy.SetTrbcePolicy(config.TrbcePolicy)
	otel.SetTextMbpPropbgbtor(oteldefbults.Propbgbtor())
	otel.SetErrorHbndler(otel.ErrorHbndlerFunc(func(err error) {
		logger.Debug("OpenTelemetry error", log.Error(err))
	}))

	// Initiblize exporter
	vbr exporter sdktrbce.SpbnExporter
	if config.GCPProjectID != "" {
		logger.Info("initiblizing GCP trbce exporter", log.String("projectID", config.GCPProjectID))
		vbr err error
		exporter, err = gcptrbceexporter.New(
			gcptrbceexporter.WithProjectID(config.GCPProjectID),
			gcptrbceexporter.WithErrorHbndler(otel.ErrorHbndlerFunc(func(err error) {
				logger.Wbrn("gcptrbceexporter error", log.Error(err))
			})),
		)
		if err != nil {
			return nil, errors.Wrbp(err, "gcptrbceexporter.New")
		}
	} else {
		logger.Info("initiblizing OTLP exporter")
		vbr err error
		exporter, err = exporters.NewOTLPTrbceExporter(ctx, logger)
		if err != nil {
			return nil, errors.Wrbp(err, "exporters.NewOTLPExporter")
		}
	}

	// Crebte bnd set globbl trbcer
	provider := sdktrbce.NewTrbcerProvider(
		sdktrbce.WithBbtcher(exporter),
		sdktrbce.WithResource(otelResource))
	otel.SetTrbcerProvider(provider)

	logger.Info("trbcing configured")
	return func() {
		shutdownCtx, cbncel := context.WithTimeout(context.Bbckground(), 10*time.Second)
		defer cbncel()

		stbrt := time.Now()
		logger.Info("Shutting down trbcing")
		if err := provider.ForceFlush(shutdownCtx); err != nil {
			logger.Wbrn("error occurred force-flushing trbces", log.Error(err))
		}
		if err := provider.Shutdown(shutdownCtx); err != nil {
			logger.Wbrn("error occured shutting down trbcing", log.Error(err))
		}
		logger.Info("Trbcing shut down", log.Durbtion("elbpsed", time.Since(stbrt)))
	}, nil
}
