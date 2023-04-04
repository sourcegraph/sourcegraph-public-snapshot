package otlpadapter

import (
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/collector/component"
)

// otelHost is a mostly no-op implementation of OTEL component.Host
type otelHost struct {
	logger log.Logger
}

var _ component.Host = (*otelHost)(nil)

func (h *otelHost) ReportFatalError(err error) {
	h.logger.Fatal("OTLP receiver error", log.Error(err))
}

func (*otelHost) GetFactory(_ component.Kind, _ component.Type) component.Factory {
	return nil
}

func (*otelHost) GetExtensions() map[component.ID]component.Component {
	return nil
}

func (*otelHost) GetExporters() map[component.DataType]map[component.ID]component.Component {
	return nil
}
