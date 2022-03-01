package types

import "github.com/sourcegraph/sourcegraph/internal/api"

// RevSpecSet is a utility type for a set of RevSpecs.
type RevSpecSet map[api.RevSpec]struct{}
