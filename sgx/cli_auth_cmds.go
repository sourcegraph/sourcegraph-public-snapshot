package sgx

import (
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/howeyc/gopass"
	"src.sourcegraph.com/sourcegraph/sgx/cli"
	"golang.org/x/oauth2"

	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/auth/userauth"
	"src.sourcegraph.com/sourcegraph/fed"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/sgx/client"
	"src.sourcegraph.com/sourcegraph/sgx/sgxcmd"
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
	a, err := userauth.Read(client.Credentials.AuthFile)
	if err != nil || a == nil {
		return ""
	}
	var accessToken string
	e, ok := a[endpointURL.String()]
	if !ok {
		return ""
	}
	accessToken = e.AccessToken

	ctx := sourcegraph.WithCredentials(client.Ctx,
		oauth2.StaticTokenSource(&oauth2.Token{TokenType: "Bearer", AccessToken: accessToken}),
	)
	ctx = fed.NewRemoteContext(ctx, endpointURL)
	cl := sourcegraph.NewClientFromContext(ctx)
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

	unauthedCtx := sourcegraph.WithCredentials(client.Ctx, nil)
	cl := sourcegraph.NewClientFromContext(unauthedCtx)

	var username, password string
	if c.Username != "" {
		username, password = c.Username, c.Password
	} else {
		fmt.Printf("Enter credentials for %s\n", endpointURL)
		fmt.Print("Username: ")
		var err error
		username, err = getLine()
		if err != nil {
			return "", err
		}
		fmt.Print("Password: ")
		password = string(gopass.GetPasswd())
	}

	// Get a user access token.
	//
	// We could do this via HTTP (as defined in the OAuth2 spec), but
	// using gRPC is a bit simpler and is consistent with how we do it
	// below (where using gRPC makes it much simpler, since we don't
	// have to mimic a browser's cookies and CSRF tokens).
	tok, err := cl.Auth.GetAccessToken(unauthedCtx, &sourcegraph.AccessTokenRequest{
		AuthorizationGrant: &sourcegraph.AccessTokenRequest_ResourceOwnerPassword{
			ResourceOwnerPassword: &sourcegraph.LoginCredentials{Login: username, Password: password},
		},
	})
	if err != nil {
		return "", fmt.Errorf("authenticating to %s: %s", endpointURL, err)
	}

	if err := userauth.SaveCredentials(client.Credentials.AuthFile, endpointURL, tok.AccessToken, false); err != nil {
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
	_, err := userauth.Read(client.Credentials.AuthFile)
	if err != nil {
		return err
	}

	// We allow the endpoint URL to be passed in as an argument as a
	// convenience to --endpoint
	if c.Args.EndpointURL != "" {
		client.Endpoint.URL = c.Args.EndpointURL
	}

	endpointURL := client.Endpoint.URLOrDefault()
	accessTok, err := c.getAccessToken(endpointURL)
	if err != nil {
		return err
	}

	err = userauth.SaveCredentials(client.Credentials.AuthFile, endpointURL, accessTok, true)
	if err != nil {
		return err
	}
	return nil
}

type whoamiCmd struct {
	PrintToken bool `long:"print-token" description:"print the token used for authenticating requests"`
}

func (c *whoamiCmd) Execute(args []string) error {
	a, err := userauth.Read(client.Credentials.AuthFile)
	if err != nil {
		return err
	}
	endpointURL := client.Endpoint.URLOrDefault()
	ua := a[endpointURL.String()]
	if ua == nil {
		log.Fatalf("# No authentication info set for %s (use `%s login` to authenticate)", endpointURL, sgxcmd.Name)
	}

	cl := client.Client()

	authInfo, err := cl.Auth.Identify(client.Ctx, &pbtypes.Void{})
	if err != nil {
		log.Fatalf("Error verifying auth credentials with endpoint %s: %s.", endpointURL, err)
	}
	log.Printf("%s (UID %d) on %s (write: %v, admin: %v)", authInfo.Login, authInfo.UID, endpointURL, authInfo.Write, authInfo.Admin)

	if c.PrintToken {
		log.Printf(" Auth token: %s", ua.AccessToken)
	}

	return nil
}
