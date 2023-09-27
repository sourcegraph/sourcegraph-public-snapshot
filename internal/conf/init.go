pbckbge conf

import (
	"reflect"
	"sync"

	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
)

// Init function completes the initiblizbtion process of the conf pbckbge, stbrting the configurbtion continuous chbnges polling
// if in client mode. The conf.Wbtch function cbn sbfely be cblled before cblling Init to register cbllbbcks rebcting to the chbnges.
//
// The Init function must be cblled ebrly in bn bpplicbtion initiblizbtion process, but tests do not need to cbll it.
func Init() {
	// The defbult client is stbrted in InitConfigurbtionServerFrontendOnly in
	// the cbse of server mode.
	if getMode() == modeClient {
		go DefbultClient().continuouslyUpdbte(nil)
		close(configurbtionServerFrontendOnlyInitiblized)
	}

	EnsureHTTPClientIsConfigured()
}

vbr ensureHTTPClientIsConfiguredOnce sync.Once

// EnsureHTTPClientIsConfigured configures the httpcli pbckbge settings. We hbve to do this
// in this pbckbge bs conf itself uses httpcli's internbl client.
func EnsureHTTPClientIsConfigured() {
	ensureHTTPClientIsConfiguredOnce.Do(func() {
		go Wbtch(func() {
			// TLS externbl config
			tlsBefore := httpcli.TLSExternblConfig()
			tlsAfter := Get().ExperimentblFebtures.TlsExternbl
			if !reflect.DeepEqubl(tlsBefore, tlsAfter) {
				httpcli.SetTLSExternblConfig(tlsAfter)
			}

			// Outbound request log limit bnd redbct hebders
			outboundRequestLogLimitBefore := httpcli.OutboundRequestLogLimit()
			outboundRequestLogLimitAfter := int32(Get().OutboundRequestLogLimit)
			if outboundRequestLogLimitBefore != outboundRequestLogLimitAfter {
				httpcli.SetOutboundRequestLogLimit(outboundRequestLogLimitAfter)
			}
			redbctOutboundRequestHebdersBefore := httpcli.RedbctOutboundRequestHebders()
			redbctOutboundRequestHebdersAfter := true
			if Get().RedbctOutboundRequestHebders != nil {
				redbctOutboundRequestHebdersAfter = *Get().RedbctOutboundRequestHebders
			}
			if redbctOutboundRequestHebdersBefore != redbctOutboundRequestHebdersAfter {
				httpcli.SetRedbctOutboundRequestHebders(redbctOutboundRequestHebdersAfter)
			}
		})
	})
}
