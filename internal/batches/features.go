package batches

import (
	"github.com/pkg/errors"
	"github.com/sourcegraph/src-cli/internal/api"
)

// FeatureFlags represent features that are only available on certain
// Sourcegraph versions and we therefore have to detect at runtime.
type FeatureFlags struct {
	AllowArrayEnvironments   bool
	IncludeAutoAuthorDetails bool
	UseGzipCompression       bool
	AllowTransformChanges    bool
	AllowWorkspaces          bool
	BatchChanges             bool
}

func (ff *FeatureFlags) SetFromVersion(version string) error {
	for _, feature := range []struct {
		flag       *bool
		constraint string
		minDate    string
	}{
		{&ff.AllowArrayEnvironments, ">= 3.23.0", "2020-11-24"},
		{&ff.IncludeAutoAuthorDetails, ">= 3.20.0", "2020-09-10"},
		{&ff.UseGzipCompression, ">= 3.21.0", "2020-10-12"},
		{&ff.AllowTransformChanges, ">= 3.23.0", "2020-12-11"},
		{&ff.AllowWorkspaces, ">= 3.25.0", "2021-01-29"},
		{&ff.BatchChanges, ">= 3.26.0", "2021-03-07"},
	} {
		value, err := api.CheckSourcegraphVersion(version, feature.constraint, feature.minDate)
		if err != nil {
			return errors.Wrap(err, "failed to check version returned by Sourcegraph")
		}
		*feature.flag = value
	}

	return nil
}
