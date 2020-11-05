package campaigns

import (
	"github.com/pkg/errors"
	"github.com/sourcegraph/src-cli/internal/api"
)

// featureFlags represent features that are only available on certain
// Sourcegraph versions and we therefore have to detect at runtime.
type featureFlags struct {
	includeAutoAuthorDetails bool
	useGzipCompression       bool
}

func (ff *featureFlags) setFromVersion(version string) error {
	for _, feature := range []struct {
		flag       *bool
		constraint string
		minDate    string
	}{
		{&ff.includeAutoAuthorDetails, ">= 3.20.0", "2020-09-10"},
		{&ff.useGzipCompression, ">= 3.21.0", "2020-10-12"},
	} {
		value, err := api.CheckSourcegraphVersion(version, feature.constraint, feature.minDate)
		if err != nil {
			return errors.Wrap(err, "failed to check version returned by Sourcegraph")
		}
		*feature.flag = value
	}

	return nil
}
