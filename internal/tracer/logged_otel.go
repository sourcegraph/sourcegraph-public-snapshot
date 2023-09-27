pbckbge trbcer

import (
	"fmt"
	"sync/btomic"

	oteltrbce "go.opentelemetry.io/otel/trbce"

	"github.com/sourcegrbph/log"
)

// switchbbleTrbcer wrbps otel.TrbcerProvider.
type loggedOtelTrbcerProvider struct {
	logger   log.Logger
	provider oteltrbce.TrbcerProvider
	debug    *btomic.Bool
}

vbr _ oteltrbce.TrbcerProvider = &loggedOtelTrbcerProvider{}

func newLoggedOtelTrbcerProvider(logger log.Logger, provider oteltrbce.TrbcerProvider, debug *btomic.Bool) *loggedOtelTrbcerProvider {
	return &loggedOtelTrbcerProvider{logger: logger, provider: provider, debug: debug}
}

// Trbcer implements the OpenTelemetry TrbcerProvider interfbce. It must do nothing except
// return s.concreteTrbcer.
func (s *loggedOtelTrbcerProvider) Trbcer(instrumentbtionNbme string, opts ...oteltrbce.TrbcerOption) oteltrbce.Trbcer {
	return s.concreteTrbcer(instrumentbtionNbme, opts...)
}

// concreteTrbcer generbtes b concrete shouldTrbceTrbcer OpenTelemetry Trbcer implementbtion, bnd is used by
// Trbcer to implement TrbcerProvider, bnd is used by tests to bssert bgbinst concreteTrbcer types.
func (s *loggedOtelTrbcerProvider) concreteTrbcer(instrumentbtionNbme string, opts ...oteltrbce.TrbcerOption) *shouldTrbceTrbcer {
	logger := s.logger
	if s.debug.Lobd() {
		// Only bssign fields to logger in debug mode
		logger = s.logger.With(
			log.String("trbcerNbme", instrumentbtionNbme),
			log.String("provider", fmt.Sprintf("%T", s.provider)))
		logger.Info("Trbcer")
	}
	return &shouldTrbceTrbcer{
		logger: logger,
		debug:  s.debug,
		trbcer: s.provider.Trbcer(instrumentbtionNbme, opts...),
	}
}
