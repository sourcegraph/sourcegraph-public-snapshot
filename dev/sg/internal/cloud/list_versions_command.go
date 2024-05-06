package cloud

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	artifactregistry "cloud.google.com/go/artifactregistry/apiv1"
	artifactregistrypb "cloud.google.com/go/artifactregistry/apiv1/artifactregistrypb"
	"github.com/grafana/regexp"
	"github.com/urfave/cli/v2"
	"google.golang.org/api/iterator"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

const DefaultArtifactRegistryPageSize = 10000

var ListVersionsEphemeralCommand = cli.Command{
	Name:        "list-versions",
	Usage:       "sg could list-versions",
	Description: "list ephemeral cloud instances attached to your GCP account",
	Action:      listTagsCloudEphemeral,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "format",
			Usage: "format to print out the list of versions out - can be one json, raw or formatted",
			Value: "formatted",
		},
		&cli.IntFlag{
			Name:  "limit",
			Usage: "limit the number of versions to list - to list everything set limt to a negative value",
			Value: 100,
		},
		&cli.StringFlag{
			Name:  "filter",
			Usage: "filter versions by regex",
		},
	},
}

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

func listTagsCloudEphemeral(ctx *cli.Context) error {
	var filterRegex *regexp.Regexp
	if ctx.String("filter") != "" {
		filterRegex = regexp.MustCompile(ctx.String("filter"))
	}
	ar, err := NewArtifactRegistry(ctx.Context, "sourcegraph-ci", "us-central1", "cloud-ephemeral")
	if err != nil {
		return err
	}
	pending := std.Out.Pending(output.Linef(CloudEmoji, output.StylePending, "Retrieving docker images from registry %q", ar.RepositoryName))
	images, err := ar.ListDockerImages(ctx.Context)
	if err != nil {
		pending.Complete(output.Linef(output.EmojiFailure, output.StyleFailure, "failed to retreive images from registry %q", ar.RepositoryName))
		return err
	}
	pending.Complete(output.Linef(CloudEmoji, output.StyleSuccess, "Retrieved %d docker images from registry %q", len(images), ar.RepositoryName))

	imagesByTag := map[string][]*DockerImage{}
	for _, image := range images {
		for _, tag := range image.Tags {
			if filterRegex == nil || filterRegex.MatchString(tag) {
				imagesByTag[tag] = append(imagesByTag[tag], image)
			}
		}
	}

	switch ctx.String("format") {
	case "json":
		{
			return json.NewEncoder(os.Stdout).Encode(imagesByTag)
		}
	case "raw":
		{
			count := 0
			limit := ctx.Int("limit")
			for tag, images := range imagesByTag {
				image := images[0]
				std.Out.Writef("Tag: %s\n", tag)
				std.Out.Writef(" %-50s %-20s %s", "Name", "Upload time", "URI")
				std.Out.Writef("- %-50s %-20s %s", image.Name, image.UploadTime.AsTime().Format(time.DateTime), image.Uri)
				count++
				if limit >= 1 && count >= limit {
					break
				}
			}
		}
	default:
		{
			count := 0
			limit := ctx.Int("limit")
			std.Out.Writef("%-50s %-20s %-5s", "Tag", "Upload time", "Image count")
			for tag, images := range imagesByTag {
				// we use the first image to get the upload time
				image := images[0]
				tag := tag[:min(50, len(tag))]
				std.Out.Writef("%-50s %-20s %-5d", tag, image.UploadTime.AsTime().Format(time.DateTime), len(images))
				count++
				if limit >= 1 && count >= limit {
					break
				}
			}
		}
	}
	return nil
}
