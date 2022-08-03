package otlpadapter

import (
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
)

// otelHost is a mostly no-op implementation of OTEL component.Host
type otelHost struct {
	logger log.Logger
}

func (h *otelHost) ReportFatalError(err error) {
	h.logger.Fatal("OTLP receiver error", log.Error(err))
}

func (*otelHost) GetFactory(_ component.Kind, _ config.Type) component.Factory {
	return nil
}

func (*otelHost) GetExtensions() map[config.ComponentID]component.Extension {
	return nil
}

func (*otelHost) GetExporters() map[config.DataType]map[config.ComponentID]component.Exporter {
	return nil
}
