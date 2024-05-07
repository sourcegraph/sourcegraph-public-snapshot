package appliance

import (
	"fmt"

	"k8s.io/utils/ptr"

	"github.com/sourcegraph/sourcegraph/internal/appliance/config"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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
			// Global config
			ImageRepository: "index.docker.io/sourcegraph",

			// Service-specific config
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

// Images

// Map of version to map of service to image tag
var defaultImages = map[string]map[string]string{
	"5.3.9104": defaultImagesForVersion_5_3_9104,
}

var defaultImagesForVersion_5_3_9104 = map[string]string{
	"blobstore":    "blobstore:5.3.2@sha256:d625be1eefe61cc42f94498e3c588bf212c4159c8b20c519db84eae4ff715efa",
	"gitserver":    "gitserver:5.3.2@sha256:6c6042cf3e5f3f16de9b82e3d4ab1647f8bb924cd315245bd7a3162f5489e8c4",
	"repo-updater": "repo-updater:5.3.2@sha256:5a414aa030c7e0922700664a43b449ee5f3fafa68834abef93988c5992c747c6",
	"symbols":      "symbols:5.3.2@sha256:dd7f923bdbd5dbd231b749a7483110d40d59159084477b9fff84afaf58aad98e",
}

func getDefaultImage(sg *Sourcegraph, component string) (string, error) {
	images, ok := defaultImages[sg.Spec.RequestedVersion]
	if !ok {
		return "", errors.Newf("no default images found for version %s", sg.Spec.RequestedVersion)
	}
	image, ok := images[component]
	if !ok {
		return "", errors.Newf("no default image found for service %s", component)
	}
	return fmt.Sprintf("%s/%s", sg.Spec.ImageRepository, image), nil
}
