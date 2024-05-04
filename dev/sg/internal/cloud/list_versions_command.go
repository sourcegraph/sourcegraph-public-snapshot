package cloud

import (
	"encoding/json"
	"os"
	"time"

	"github.com/grafana/regexp"
	"github.com/urfave/cli/v2"

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

func listTagsCloudEphemeral(ctx *cli.Context) error {
	var filterRegex *regexp.Regexp
	if ctx.String("filter") != "" {
		filterRegex = regexp.MustCompile(ctx.String("filter"))
	}
	ar, err := NewDefaultCloudEphemeralRegistry(ctx.Context)
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
