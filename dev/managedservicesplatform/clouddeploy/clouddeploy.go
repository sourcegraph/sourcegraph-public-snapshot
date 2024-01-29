package clouddeploy

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"embed"
	"fmt"
	"html/template"
	"io"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
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
			continue
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

//go:embed customtarget.yaml
var cloudDeployCustomTarget []byte

//go:embed target.template.yaml
var cloudDeployTargetTemplateRaw string
var cloudDeployTargetTemplate = template.Must(template.New("cloudDeployTargetTemplate").
	Parse(cloudDeployTargetTemplateRaw))

func RenderSpec(
	service spec.ServiceSpec,
	build spec.BuildSpec,
	config spec.RolloutPipelineConfiguration,
	region string,
) (*bytes.Buffer, error) {
	var targetsSpec bytes.Buffer
	if _, err := targetsSpec.Write(cloudDeployCustomTarget); err != nil {
		return nil, err
	}

	for _, stage := range config.Stages {
		if _, err := targetsSpec.WriteString("\n---\n"); err != nil {
			return nil, err
		}

		var b bytes.Buffer
		if err := cloudDeployTargetTemplate.Execute(&b, map[string]any{
			"Stage":   stage,
			"Service": service,
			"Build":   build,
			"Region":  region,

			// Stable naming: always a SA in the last stage's project.
			"CloudDeployServiceAccount": fmt.Sprintf("clouddeploy-executor@%s.iam.gserviceaccount.com",
				config.Stages[len(config.Stages)-1].ProjectID),
		}); err != nil {
			return nil, err
		}
		if _, err := targetsSpec.Write(b.Bytes()); err != nil {
			return nil, err
		}
	}

	return &targetsSpec, nil
}
