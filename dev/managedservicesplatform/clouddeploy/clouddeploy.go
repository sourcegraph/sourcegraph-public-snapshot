package clouddeploy

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"embed"
	"io"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

//go:embed skaffold.yaml
var skaffoldAssets embed.FS

// NewCloudRunCustomTargetSkaffoldAssetsArchive generates an archive of assets
// required for 'gcloud deploy releases create', to be provided via the
// '--source' flag: https://cloud.google.com/sdk/gcloud/reference/deploy/releases/create#--source
func NewCloudRunCustomTargetSkaffoldAssetsArchive() (*bytes.Buffer, error) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	files, err := skaffoldAssets.ReadDir(".")
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if file.IsDir() {
			return nil, errors.New("unexpected dir")
		}
		info, err := file.Info()
		if err != nil {
			return nil, err
		}
		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return nil, err
		}
		if err := tw.WriteHeader(header); err != nil {
			return nil, err
		}

		f, err := skaffoldAssets.Open(file.Name())
		if err != nil {
			return nil, err
		}
		if _, err := io.Copy(tw, f); err != nil {
			return nil, err
		}
	}

	return &buf, nil
}
