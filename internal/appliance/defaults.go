package appliance

import (
	"k8s.io/utils/ptr"

	"github.com/sourcegraph/sourcegraph/internal/appliance/config"
)

// Default config.
//
// Warning: never extract `ptr.To(thing)` into a package-level variable! If you
// do this, reconciling a config that overrides a default value for that
// pointer, will affect the subsequent _default_ for all future resources
// reconciled.
func newDefaultConfig() Sourcegraph {
	return Sourcegraph{
		Spec: SourcegraphSpec{
			Blobstore: BlobstoreSpec{
				StorageSize: "100Gi",
			},
			RepoUpdater: RepoUpdaterSpec{
				StandardConfig: config.StandardConfig{
					PrometheusPort: ptr.To(6060),
				},
			},
			StorageClass: StorageClassSpec{
				Name: "sourcegraph",
			},
			Symbols: SymbolsSpec{
				StandardConfig: config.StandardConfig{
					PrometheusPort: ptr.To(6060),
				},
				Replicas:    1,
				StorageSize: "12Gi",
			},
			GitServer: GitServerSpec{
				StandardConfig: config.StandardConfig{
					PrometheusPort: ptr.To(6060),
				},
				Replicas:    1,
				StorageSize: "200Gi",
			},
		},
	}
}
