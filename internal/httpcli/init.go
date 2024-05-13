package httpcli

import (
	"reflect"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
)

// Configure configures the httpcli package settings with TLS and outbound request
// logging settings from the site configuration.
func Configure(cfg conftypes.WatchableSiteConfig) {
	cfg.Watch(func() {
		siteConfig := cfg.SiteConfig()
		// TLS external config
		tlsBefore := tlsExternalConfig()
		tlsAfter := siteConfig.ExperimentalFeatures.TlsExternal
		if !reflect.DeepEqual(tlsBefore, tlsAfter) {
			setTLSExternalConfig(tlsAfter)
		}

		// Outbound request log limit and redact headers
		outboundRequestLogLimitBefore := outboundRequestLogLimit()
		outboundRequestLogLimitAfter := int32(siteConfig.OutboundRequestLogLimit)
		if outboundRequestLogLimitBefore != outboundRequestLogLimitAfter {
			setOutboundRequestLogLimit(outboundRequestLogLimitAfter)
		}
		redactOutboundRequestHeadersBefore := redactOutboundRequestHeaders()
		redactOutboundRequestHeadersAfter := true
		if siteConfig.RedactOutboundRequestHeaders != nil {
			redactOutboundRequestHeadersAfter = *siteConfig.RedactOutboundRequestHeaders
		}
		if redactOutboundRequestHeadersBefore != redactOutboundRequestHeadersAfter {
			setRedactOutboundRequestHeaders(redactOutboundRequestHeadersAfter)
		}
	})
}
