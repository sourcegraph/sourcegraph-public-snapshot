package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sourcegraph/src-cli/internal/api"
	"github.com/sourcegraph/src-cli/internal/cmderrors"
)

func init() {
	usage := `'src login' helps you authenticate 'src' to access a Sourcegraph instance with your user credentials.

Usage:

    src login SOURCEGRAPH_URL

Examples:

  Authenticate to a Sourcegraph instance at https://sourcegraph.example.com:

    $ src login https://sourcegraph.example.com

  Authenticate to Sourcegraph.com:

    $ src login https://sourcegraph.com
`

	flagSet := flag.NewFlagSet("login", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintln(flag.CommandLine.Output(), usage)
		flagSet.PrintDefaults()
	}

	var (
		apiFlags = api.NewFlags(flagSet)
	)

	handler := func(args []string) error {
		if err := flagSet.Parse(args); err != nil {
			return err
		}
		endpoint := cfg.Endpoint
		if flagSet.NArg() >= 1 {
			endpoint = flagSet.Arg(0)
		}
		if endpoint == "" {
			return cmderrors.Usage("expected exactly one argument: the Sourcegraph URL, or SRC_ENDPOINT to be set")
		}

		client := cfg.apiClient(apiFlags, io.Discard)

		return loginCmd(context.Background(), cfg, client, endpoint, os.Stdout)
	}

	commands = append(commands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}

func loginCmd(ctx context.Context, cfg *config, client api.Client, endpointArg string, out io.Writer) error {
	endpointArg = cleanEndpoint(endpointArg)

	printProblem := func(problem string) {
		fmt.Fprintf(out, "❌ Problem: %s\n", problem)
	}

	createAccessTokenMessage := fmt.Sprintf("\n"+`🛠  To fix: Create an access token at %s/user/settings/tokens, then set the following environment variables:

   SRC_ENDPOINT=%s
   SRC_ACCESS_TOKEN=(the access token you just created)

   To verify that it's working, run this command again.
`, endpointArg, endpointArg)

	if cfg.ConfigFilePath != "" {
		fmt.Fprintln(out)
		fmt.Fprintf(out, "⚠️  Warning: Configuring src with a JSON file is deprecated. Please migrate to using the env vars SRC_ENDPOINT and SRC_ACCESS_TOKEN instead, and then remove %s. See https://github.com/sourcegraph/src-cli#readme for more information.\n", cfg.ConfigFilePath)
	}

	noToken := cfg.AccessToken == ""
	endpointConflict := endpointArg != cfg.Endpoint
	if noToken || endpointConflict {
		fmt.Fprintln(out)
		switch {
		case noToken:
			printProblem("No access token is configured.")
		case endpointConflict:
			printProblem(fmt.Sprintf("The configured endpoint is %s, not %s.", cfg.Endpoint, endpointArg))
		}
		fmt.Fprintln(out, createAccessTokenMessage)
		return cmderrors.ExitCode1
	}

	// See if the user is already authenticated.
	query := `query CurrentUser { currentUser { username } }`
	var result struct {
		CurrentUser *struct{ Username string }
	}
	if _, err := client.NewRequest(query, nil).Do(ctx, &result); err != nil {
		if strings.HasPrefix(err.Error(), "error: 401 Unauthorized") || strings.HasPrefix(err.Error(), "error: 403 Forbidden") {
			printProblem("Invalid access token.")
		} else {
			printProblem(fmt.Sprintf("Error communicating with %s: %s", endpointArg, err))
		}
		fmt.Fprintln(out, createAccessTokenMessage)
		fmt.Fprintln(out, "   (If you need to supply custom HTTP request headers, see information about SRC_HEADER_* and SRC_HEADERS env vars at https://github.com/sourcegraph/src-cli/blob/main/AUTH_PROXY.md.)")
		return cmderrors.ExitCode1
	}

	if result.CurrentUser == nil {
		// This should never happen; we verified there is an access token, so there should always be
		// a user.
		printProblem(fmt.Sprintf("Unable to determine user on %s.", endpointArg))
		return cmderrors.ExitCode1
	}
	fmt.Fprintln(out)
	fmt.Fprintf(out, "✔️  Authenticated as %s on %s\n", result.CurrentUser.Username, endpointArg)
	fmt.Fprintln(out)
	return nil
}
