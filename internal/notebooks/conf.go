package notebooks

import (
	"os"
	"strconv"

	"github.com/sourcegraph/sourcegraph/internal/licensing"
)

func IsEnabled() bool {
	if v, _ := strconv.ParseBool(os.Getenv("DISABLE_NOTEBOOKS")); v {
		return false
	}
	return licensing.Check(licensing.FeatureCodeSearch) == nil
}
