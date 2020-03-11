package lsifserver

import (
	"errors"
	"strings"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	"github.com/sourcegraph/sourcegraph/internal/env"
)

var (
	lsifServerURL = env.Get("LSIF_SERVER_URL", "k8s+http://lsif-server:3186", "lsif-server URL")

	lsifServerURLsOnce sync.Once
	lsifServerURLs     *endpoint.Map
)

func LSIFServerURLs() *endpoint.Map {
	lsifServerURLsOnce.Do(func() {
		if len(strings.Fields(lsifServerURL)) == 0 {
			lsifServerURLs = endpoint.Empty(errors.New("an lsif-server has not been configured"))
		} else {
			lsifServerURLs = endpoint.New(lsifServerURL)
		}
	})
	return lsifServerURLs
}
