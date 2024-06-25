package release

import (
	"encoding/json"
	"os"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/category"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/internal/releaseregistry"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var listReleaseCommand = &cli.Command{
	Name:      "list",
	Usage:     "list versions from the release registry",
	Category:  category.Util,
	UsageText: "sg release list <flags>",
	Action:    listRegistryVersions,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "format",
			Usage: "output the list of versions in 'json' or 'terminal' format",
			Value: "terminal",
		},
	},
}

func listRegistryVersions(cmd *cli.Context) error {
	client := releaseregistry.NewClient(releaseregistry.Endpoint)

	out := std.NewOutput(os.Stderr, false)
	pending := out.Pending(output.Line("", output.StylePending, "Fetching versions from the release registry..."))
	versions, err := client.ListVersions(cmd.Context, "")
	if err != nil {
		pending.Complete(output.Linef(output.EmojiFailure, output.StyleFailure, "Failed to fetch versions from release registry"))
		return err
	}
	pending.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Fetched %d versions from release registry", len(versions)))

	switch cmd.String("format") {
	case "json":
		json.NewEncoder(os.Stdout).Encode(versions)
	default:
		std.Out.WriteLine(output.Linef("", output.StyleBold, "%-4s%-20s%-12s%-23s%s", "ID", "Product", "Version", "Created", "SHA"))
		for _, v := range versions {
			productName := v.Name
			if len(productName) > 20 {
				productName = productName[:17] + "..."
			}
			std.Out.WriteLine(output.Linef("", output.StyleGrey, "%-4d%-20s%-12s%-23s%s", v.ID, productName, v.Version, v.CreatedAt.Format(time.RFC3339), v.GitSHA))
		}
	}

	return nil
}
