package httpapi

import "github.com/sourcegraph/sourcegraph/enterprise/pkg/codeintelligence/proxy"

// Set by enterprise frontend
var NewLSIFProxy func() (*proxy.LSIFProxy, error)
