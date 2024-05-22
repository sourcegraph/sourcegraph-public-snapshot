package cloud

import (
	"context"
	"fmt"

	"github.com/grafana/regexp"

	artifactregistry "cloud.google.com/go/artifactregistry/apiv1"
	artifactregistrypb "cloud.google.com/go/artifactregistry/apiv1/artifactregistrypb"
	"google.golang.org/api/iterator"
)

const DefaultArtifactRegistryPageSize = 10000

// DockerImage is a type alias around the Google artifact registry Docker Image type
type DockerImage artifactregistrypb.DockerImage

type DockerImageFilterOpt func(image *DockerImage) bool

// FilterTagByRegex filters the Docker Image tags by the given regular expression
func FilterTagByRegex(regex *regexp.Regexp) DockerImageFilterOpt {
	if regex == nil {
		return func(image *DockerImage) bool {
			return true
		}
	}

	return func(image *DockerImage) bool {
		for _, tag := range image.Tags {
			if regex.MatchString(tag) {
				return true
			}
		}
		return false
	}
}

// FilterNameByRegex filters the Docker Image names by the given regular expression
func FilterNameByRegex(regex *regexp.Regexp) DockerImageFilterOpt {
	return func(image *DockerImage) bool {
		return regex.MatchString(image.Name)
	}
}

// FilterByName filters the Docker Images by name
func FilterByName(name string) DockerImageFilterOpt {
	// the name format is projects/*/locations/*/repositories/*/dockerImages/<name>@sha256:digest
	regexp := regexp.MustCompile(fmt.Sprintf("/%s@", name))
	return FilterNameByRegex(regexp)
}

// FilterByTag filters the Docker Images by tag
func FilterByTag(tag string) DockerImageFilterOpt {
	return func(image *DockerImage) bool {
		for _, imageTag := range image.Tags {
			if imageTag == tag {
				return true
			}
		}
		return false
	}
}

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

// ListVersions lists all Docker Images present in the Artifact Registry and optionally filter the images as the images are iterated upon.
func (ar *ArtifactRegistry) ListDockerImages(ctx context.Context, filterOpts ...DockerImageFilterOpt) ([]*DockerImage, error) {
	req := &artifactregistrypb.ListDockerImagesRequest{
		Parent:   ar.Parent(),
		PageSize: ar.PageSize,
		OrderBy:  "upload_time",
	}

	// if we have no any filter options, we just accept all images
	if len(filterOpts) == 0 {
		acceptAll := func(image *DockerImage) bool {
			return true
		}
		filterOpts = append(filterOpts, acceptAll)
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

		dockerImage := (*DockerImage)(image)
		var shouldAdd bool
		for _, filterOpt := range filterOpts {
			shouldAdd = filterOpt(dockerImage)
			if !shouldAdd {
				break
			}
		}
		if shouldAdd {
			images = append(images, dockerImage)
		}
	}

	return images, nil
}

// FindDockerImageExact finds Docker Images that match the name and tag exactly in the Artifact Registry.
func (ar *ArtifactRegistry) FindDockerImageExact(ctx context.Context, name string, tag string) ([]*DockerImage, error) {
	return ar.ListDockerImages(ctx, FilterByName(name), FilterByTag(tag))
}

// FindDockerImageByTagPattern finds all Docker Images that have a tag that matches the given tag pattern.
func (ar *ArtifactRegistry) FindDockerImageByTagPattern(ctx context.Context, tagPattern string) ([]*DockerImage, error) {
	tagRegex := regexp.MustCompile(tagPattern)
	return ar.ListDockerImages(ctx, FilterTagByRegex(tagRegex))
}

// GetDockerImage gets a Docker Image by name and digest from the Artifact Registry.
func (ar *ArtifactRegistry) GetDockerImage(ctx context.Context, name, digest string) (*DockerImage, error) {
	name = fmt.Sprintf("%s/dockerImages/%s@sha256:%s", ar.Parent(), name, digest)
	req := &artifactregistrypb.GetDockerImageRequest{
		Name: name,
	}

	image, err := ar.client.GetDockerImage(ctx, req)
	if err != nil {
		return nil, err
	}

	return (*DockerImage)(image), nil
}
