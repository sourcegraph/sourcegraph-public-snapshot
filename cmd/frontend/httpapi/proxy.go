package httpapi

import "github.com/sourcegraph/sourcegraph/enterprise/pkg/codeintelligence/lsifserver/proxy"

// Set by enterprise frontend
var NewLSIFServerProxy func() (*proxy.LSIFServerProxy, error)
