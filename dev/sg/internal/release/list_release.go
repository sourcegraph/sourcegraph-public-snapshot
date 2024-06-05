package release

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/category"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

const releaseRegistryEndpoint = "https://releaseregistry.sourcegraph.com/v1/"

type RegistryVersion struct {
	ID            int32      `json:"id"`
	Name          string     `json:"name"`
	Public        bool       `json:"public"`
	CreatedAt     time.Time  `json:"created_at"`
	PromotedAt    *time.Time `json:"promoted_at"`
	Version       string     `json:"version"`
	GitSHA        string     `json:"git_sha"`
	IsDevelopment bool       `json:"is_development"`
}

type releaseRegistryClient struct {
	endpoint string
	client   http.Client
}

func newReleaseRegistryClient(endpoint string) *releaseRegistryClient {
	return &releaseRegistryClient{
		endpoint: endpoint,
		client:   http.Client{},
	}
}

func (r *releaseRegistryClient) newRequest(method, path string) (*http.Request, error) {
	urlPath, err := url.JoinPath(r.endpoint, path)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, urlPath, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	return req, nil
}

func (r *releaseRegistryClient) ListVersions(ctx context.Context) ([]RegistryVersion, error) {
	req, err := r.newRequest(http.MethodGet, "releases")
	if err != nil {
		return nil, err
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	results := []RegistryVersion{}

	err = json.NewDecoder(resp.Body).Decode(&results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

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
	client := newReleaseRegistryClient(releaseRegistryEndpoint)

	out := std.NewOutput(os.Stderr, false)
	pending := out.Pending(output.Line("", output.StylePending, "Fetching versions from the release registry..."))
	versions, err := client.ListVersions(cmd.Context)
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
