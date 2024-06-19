package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/src-cli/internal/api"
	"github.com/sourcegraph/src-cli/internal/version"
)

func init() {
	usage := `
Examples:

  Get the src-cli version and the Sourcegraph instance's recommended version:

    	$ src version
`

	flagSet := flag.NewFlagSet("version", flag.ExitOnError)

	var (
		clientOnly = flagSet.Bool("client-only", false, "If true, only the client version will be printed.")
		apiFlags   = api.NewFlags(flagSet)
	)

	handler := func(args []string) error {
		fmt.Printf("Current version: %s\n", version.BuildTag)
		if clientOnly != nil && *clientOnly {
			return nil
		}

		client := cfg.apiClient(apiFlags, flagSet.Output())
		recommendedVersion, err := getRecommendedVersion(context.Background(), client)
		if err != nil {
			return errors.Wrap(err, "failed to get recommended version for Sourcegraph deployment")
		}
		if recommendedVersion == "" {
			fmt.Println("Recommended version: <unknown>")
			fmt.Println("This Sourcegraph instance does not support this feature.")
			return nil
		}
		fmt.Printf("Recommended version: %s or later\n", recommendedVersion)
		return nil
	}

	// Register the command.
	commands = append(commands, &command{
		flagSet: flagSet,
		handler: handler,
		usageFunc: func() {
			fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src %s':\n", flagSet.Name())
			flagSet.PrintDefaults()
			fmt.Println(usage)
		},
	})
}

func getRecommendedVersion(ctx context.Context, client api.Client) (string, error) {
	req, err := client.NewHTTPRequest(ctx, "GET", ".api/src-cli/version", nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return "", nil
		}

		return "", errors.Newf("error: %s\n\n%s", resp.Status, body)
	}

	payload := struct {
		Version string `json:"version"`
	}{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", err
	}

	return payload.Version, nil
}
