package cloud

import (
	"context"
	"fmt"
	"slices"
	"strings"

	artifactregistry "cloud.google.com/go/artifactregistry/apiv1"
	artifactregistrypb "cloud.google.com/go/artifactregistry/apiv1/artifactregistrypb"
	"google.golang.org/api/iterator"
)

// DockerImage is a type alias around the Google artifact registry Docker Image type
type DockerImage artifactregistrypb.DockerImage

// ArtifactRegistry is wrapper around the Google Artifact Registry client
type ArtifactRegistry struct {
	Project        string
	Location       string
	RepositoryName string
	PageSize       int32

	client *artifactregistry.Client
}

// NewDefaultCloudEphemeralRegistry creates an Artifact Registry with all the details set to point to the cloud
// ephemeral registry
func NewDefaultCloudEphemeralRegistry(ctx context.Context) (*ArtifactRegistry, error) {
	return NewArtifactRegistry(ctx, "sourcegraph-ci", "us-central1", "cloud-ephemeral")
}

func NewArtifactRegistry(ctx context.Context, project, location, repositoryName string) (*ArtifactRegistry, error) {
	client, err := artifactregistry.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	return &ArtifactRegistry{
		Project:        project,
		Location:       location,
		RepositoryName: repositoryName,
		PageSize:       DefaultArtifactRegistryPageSize,

		client: client,
	}, nil
}

// Parent returns the parent string in the format of projects/<project>/locations/<location>/repositories/<repositoryName>
func (ar *ArtifactRegistry) Parent() string {
	return fmt.Sprintf("projects/%s/locations/%s/repositories/%s", ar.Project, ar.Location, ar.RepositoryName)
}

func (ar *ArtifactRegistry) String() string {
	return ar.Parent()
}

// ListVersions lists all Docker Images present in the Artifact Registry.
func (ar *ArtifactRegistry) ListDockerImages(ctx context.Context) ([]*DockerImage, error) {
	req := &artifactregistrypb.ListDockerImagesRequest{
		Parent:   ar.Parent(),
		PageSize: ar.PageSize,
	}

	images := []*DockerImage{}
	resp := ar.client.ListDockerImages(ctx, req)
	for {
		image, err := resp.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return nil, err
		}
		images = append(images, (*DockerImage)(image))
	}

	return images, nil
}

func (ar *ArtifactRegistry) HasDockerImageTag(ctx context.Context, name, tag string) (bool, error) {
	image, err := ar.GetDockerImage(ctx, name)
	if err != nil {
		return false, err
	}

	idx := slices.IndexFunc(image.Tags, func(t string) bool {
		return strings.EqualFold(t, tag)
	})

	return idx != -1, nil
}

// GetDockerImage gets a Docker Image by name from the Artifact Registry.
func (ar *ArtifactRegistry) GetDockerImage(ctx context.Context, name string) (*DockerImage, error) {
	req := &artifactregistrypb.GetDockerImageRequest{
		Name: name,
	}

	image, err := ar.client.GetDockerImage(ctx, req)
	if err != nil {
		return nil, err
	}

	return (*DockerImage)(image), nil
}
