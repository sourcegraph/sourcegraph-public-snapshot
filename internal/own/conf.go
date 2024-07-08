package own

import (
	"os"
	"strconv"

	"github.com/sourcegraph/sourcegraph/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
)

func IsEnabled() bool {
	if dotcom.SourcegraphDotComMode() {
		return false
	}
	if v, _ := strconv.ParseBool(os.Getenv("DISABLE_OWN")); v {
		return false
	}
	return licensing.Check(licensing.FeatureCodeSearch) == nil
}
