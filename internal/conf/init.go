package conf

import (
	"reflect"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

// Init function completes the initialization process of the conf package, starting the configuration continuous changes polling
// if in client mode. The conf.Watch function can safely be called before calling Init to register callbacks reacting to the changes.
//
// The Init function must be called early in an application initialization process, but tests do not need to call it.
func Init() {
	// The default client is started in InitConfigurationServerFrontendOnly in
	// the case of server mode.
	if getMode() == modeClient {
		go DefaultClient().continuouslyUpdate(nil)
		close(configurationServerFrontendOnlyInitialized)
	}

	EnsureHTTPClientIsConfigured()
}

var ensureHTTPClientIsConfiguredOnce sync.Once

// EnsureHTTPClientIsConfigured configures the httpcli package settings. We have to do this
// in this package as conf itself uses httpcli's internal client.
func EnsureHTTPClientIsConfigured() {
	ensureHTTPClientIsConfiguredOnce.Do(func() {
		go Watch(func() {
			// TLS external config
			tlsBefore := httpcli.TLSExternalConfig()
			tlsAfter := Get().ExperimentalFeatures.TlsExternal
			if !reflect.DeepEqual(tlsBefore, tlsAfter) {
				httpcli.SetTLSExternalConfig(tlsAfter)
			}

			// Outbound request log limit and redact headers
			outboundRequestLogLimitBefore := httpcli.OutboundRequestLogLimit()
			outboundRequestLogLimitAfter := int32(Get().OutboundRequestLogLimit)
			if outboundRequestLogLimitBefore != outboundRequestLogLimitAfter {
				httpcli.SetOutboundRequestLogLimit(outboundRequestLogLimitAfter)
			}
			redactOutboundRequestHeadersBefore := httpcli.RedactOutboundRequestHeaders()
			redactOutboundRequestHeadersAfter := true
			if Get().RedactOutboundRequestHeaders != nil {
				redactOutboundRequestHeadersAfter = *Get().RedactOutboundRequestHeaders
			}
			if redactOutboundRequestHeadersBefore != redactOutboundRequestHeadersAfter {
				httpcli.SetRedactOutboundRequestHeaders(redactOutboundRequestHeadersAfter)
			}
		})
	})
}
