package cli

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"

	"gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/cli"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/internal/userauth"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/internal/usercreds"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/srccmd"
	"sourcegraph.com/sqs/pbtypes"
)

func init() {
	_, err := cli.CLI.AddCommand("login",
		"log in to Sourcegraph.com",
		"The login command logs into a Sourcegraph instance. It requires an account on a root Sourcegraph instance (usually sourcegraph.com).",
		&loginCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = cli.CLI.AddCommand("whoami",
		"show logged-in user login and info",
		"The whoami command prints the username and other information about the user authenticated by a previous call to `src login`.",
		&whoamiCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
}

type loginCmd struct {
	// Optional username and password (set via environment variables)
	Username string
	Password string

	Args struct {
		EndpointURL string `name:"endpoint" description:"Optionally specify the endpoint to authenticate against."`
	} `positional-args:"yes" count:"1"`
}

// getSavedToken checks if we already have a token for an endpoint, and
// validates that it still works.
func getSavedToken(endpointURL *url.URL) string {
	a, err := userauth.Read(authFile())
	if err != nil || a == nil {
		return ""
	}
	var accessToken string
	e, ok := a[endpointURL.String()]
	if !ok {
		return ""
	}
	accessToken = e.AccessToken

	ctx := sourcegraph.WithAccessToken(cliContext, accessToken)
	cl, err := sourcegraph.NewClientFromContext(sourcegraph.WithGRPCEndpoint(ctx, endpointURL))
	if err != nil {
		log15.Error("Failed to verify saved auth credentials", "endpointURL", endpointURL, "error", err)
		return ""
	}
	_, err = cl.Auth.Identify(ctx, &pbtypes.Void{})
	if err != nil {
		log.Printf("# Failed to verify saved auth credentials for %s", endpointURL)
		return ""
	}
	return accessToken
}

func (c *loginCmd) getAccessToken(endpointURL *url.URL) (string, error) {
	if savedToken := getSavedToken(endpointURL); savedToken != "" {
		log.Printf("Using saved auth token for %s", endpointURL)
		return savedToken, nil
	}

	unauthedCtx := sourcegraph.WithAccessToken(cliContext, "")
	cl, err := sourcegraph.NewClientFromContext(unauthedCtx)
	if err != nil {
		return "", err
	}

	var creds *usercreds.LoginCredentials
	if c.Username != "" {
		creds = &usercreds.LoginCredentials{Login: c.Username, Password: c.Password}
	} else {
		creds = usercreds.FromNetRC(endpointURL)
		if creds != nil {
			fmt.Println("# Using credentials from netrc")
		} else if creds == nil {
			fmt.Printf("Enter credentials for %s\n", endpointURL)
			creds = usercreds.FromTTY()
		}
		if creds == nil {
			return "", errors.New("Failed to get credentials from user")
		}
	}

	// Get a user access token.
	//
	// We could do this via HTTP (as defined in the OAuth2 spec), but
	// using gRPC is a bit simpler and is consistent with how we do it
	// below (where using gRPC makes it much simpler, since we don't
	// have to mimic a browser's cookies and CSRF tokens).
	tok, err := cl.Auth.GetAccessToken(unauthedCtx, &sourcegraph.AccessTokenRequest{
		AuthorizationGrant: &sourcegraph.AccessTokenRequest_ResourceOwnerPassword{
			ResourceOwnerPassword: &sourcegraph.LoginCredentials{Login: creds.Login, Password: creds.Password},
		},
	})
	if err != nil {
		return "", fmt.Errorf("authenticating to %s: %s", endpointURL, err)
	}

	if err := userauth.SaveCredentials(authFile(), endpointURL, tok.AccessToken, false); err != nil {
		log.Printf("warning: failed to save credentials: %s.", err)
	}
	return tok.AccessToken, nil
}

func (c *loginCmd) Execute(args []string) error {
	if username := os.Getenv("SG_USERNAME"); username != "" {
		c.Username = username
		c.Password = os.Getenv("SG_PASSWORD")
	}

	// Check if parseable, before attempting authentication
	_, err := userauth.Read(authFile())
	if err != nil {
		return err
	}

	// We allow the endpoint URL to be passed in as an argument as a
	// convenience to --endpoint
	if c.Args.EndpointURL != "" {
		endpoint.URL = c.Args.EndpointURL
	}

	endpointURL := endpoint.URLOrDefault()
	accessTok, err := c.getAccessToken(endpointURL)
	if err != nil {
		return err
	}

	err = userauth.SaveCredentials(authFile(), endpointURL, accessTok, true)
	if err != nil {
		return err
	}
	return nil
}

type whoamiCmd struct {
	PrintToken bool `long:"print-token" description:"print the token used for authenticating requests"`
}

func (c *whoamiCmd) Execute(args []string) error {
	a, err := userauth.Read(authFile())
	if err != nil {
		return err
	}
	endpointURL := endpoint.URLOrDefault()
	ua := a[endpointURL.String()]
	if ua == nil {
		log.Fatalf("# No authentication info set for %s (use `%s login` to authenticate)", endpointURL, srccmd.Name)
	}

	cl := cliClient

	authInfo, err := cl.Auth.Identify(cliContext, &pbtypes.Void{})
	if err != nil {
		log.Fatalf("Error verifying auth credentials with endpoint %s: %s.", endpointURL, err)
	}
	log.Printf("%s (UID %d) on %s (write: %v, admin: %v)", authInfo.Login, authInfo.UID, endpointURL, authInfo.Write, authInfo.Admin)

	if c.PrintToken {
		log.Printf(" Auth token: %s", ua.AccessToken)
	}

	return nil
}
