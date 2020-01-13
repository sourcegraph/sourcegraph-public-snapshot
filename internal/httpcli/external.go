package httpcli

import (
	"crypto/tls"
	"net/http"
	"reflect"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

type externalTransport struct {
	base *http.Transport

	mu        sync.RWMutex
	config    *schema.TlsExternal
	effective *http.Transport
}

func (t *externalTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	t.mu.RLock()
	config, effective := t.config, t.effective
	t.mu.RUnlock()

	if current := conf.Get().ExperimentalFeatures.TlsExternal; current == nil {
		return t.base.RoundTrip(r)
	} else if reflect.DeepEqual(config, current) {
		return effective.RoundTrip(r)
	}

	effective = t.update()
	return effective.RoundTrip(r)
}

func (t *externalTransport) update() *http.Transport {
	t.mu.Lock()
	defer t.mu.Unlock()

	config := conf.Get().ExperimentalFeatures.TlsExternal
	effective := t.base.Clone()

	if effective.TLSClientConfig == nil {
		effective.TLSClientConfig = new(tls.Config)
	}

	effective.TLSClientConfig.InsecureSkipVerify = config.InsecureSkipVerify

	t.config = config
	t.effective = effective
	return effective
}
