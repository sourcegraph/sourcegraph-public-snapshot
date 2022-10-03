package dependencies

import (
	"github.com/sourcegraph/sourcegraph/internal/conf"
)

// For mocking in tests
var autoIndexingEnabled = conf.CodeIntelAutoIndexingEnabled
