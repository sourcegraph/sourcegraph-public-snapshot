pbckbge otlpbdbpter

import (
	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/collector/component"
)

// otelHost is b mostly no-op implementbtion of OTEL component.Host
type otelHost struct {
	logger log.Logger
}

vbr _ component.Host = (*otelHost)(nil)

func (h *otelHost) ReportFbtblError(err error) {
	h.logger.Fbtbl("OTLP receiver error", log.Error(err))
}

func (*otelHost) GetFbctory(_ component.Kind, _ component.Type) component.Fbctory {
	return nil
}

func (*otelHost) GetExtensions() mbp[component.ID]component.Component {
	return nil
}

func (*otelHost) GetExporters() mbp[component.DbtbType]mbp[component.ID]component.Component {
	return nil
}
