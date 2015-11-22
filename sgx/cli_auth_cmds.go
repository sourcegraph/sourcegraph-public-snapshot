package sgx

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/howeyc/gopass"
	"golang.org/x/oauth2"

	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/env"
	"src.sourcegraph.com/sourcegraph/fed"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/sgx/cli"
	"src.sourcegraph.com/sourcegraph/sgx/sgxcmd"
)

// userAuth holds user auth credentials keyed on API endpoint
// URL. It's typically saved in a file named by userAuthFile.
type userAuth map[string]*userEndpointAuth

func (ua userAuth) setDefault(endpoint string) {
	for k, v := range ua {
		if k == endpoint {
			v.Default = true
		} else {
			v.Default = false
		}
	}
}

// getDefault returns the user-endpoint auth entry that is marked as
// the default, if any exists.
func (ua userAuth) getDefault() (endpoint string, a *userEndpointAuth) {
	for k, v := range ua {
		if v.Default {
			return k, v
		}
	}
	return "", nil
}

// userEndpointAuth holds a user's authentication credentials for a
// sourcegraph endpoint.
type userEndpointAuth struct {
	AccessToken string

	// Default is whether this endpoint and access token should be
	// used as the defaults if none are specified.
	Default bool `json:",omitempty"`
}

// readAuth attempts to read a userAuth struct from the
// userAuthFile. It is not considered an error if the userAuthFile
// doesn't exist; in that case, an empty userAuth and a nil error is
// returned.
func readUserAuth() (userAuth, error) {
	if Credentials.AuthFile == "/dev/null" {
		return userAuth{}, nil
	}
	f, err := os.Open(userAuthFileName())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var ua userAuth
	if err := json.NewDecoder(f).Decode(&ua); err != nil {
		return nil, err
	}
	return ua, nil
}

// writeUserAuth writes ua to the userAuthFile.
func writeUserAuth(a userAuth) error {
	f, err := os.Create(userAuthFileName())
	if err != nil {
		return err
	}
	defer f.Close()
	if err := os.Chmod(f.Name(), 0600); err != nil {
		return err
	}
	b, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return err
	}
	_, err = f.Write(b)
	return err
}

// Resolves user auth file name platform-independent way
func userAuthFileName() string {
	ret := Credentials.AuthFile
	if runtime.GOOS == "windows" {
		// on Windows there is no HOME
		ret = strings.Replace(ret, "$HOME", env.CurrentUserHomeDir(), -1)
	}
	return filepath.FromSlash(os.ExpandEnv(ret))
}

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
	Username string `long:"username" short:"u" description:"username of user to log in"`
	Password string `long:"password" short:"p" description:"password of user to log in"`
	Args     struct {
		EndpointURL string `name:"endpoint" description:"Optionally specify the endpoint to authenticate against."`
	} `positional-args:"yes" count:"1"`
}

// getSavedToken checks if we already have a token for an endpoint, and
// validates that it still works.
func getSavedToken(endpointURL *url.URL) string {
	a, err := readUserAuth()
	if err != nil || a == nil {
		return ""
	}
	var accessToken string
	e, ok := a[endpointURL.String()]
	if !ok {
		return ""
	}
	accessToken = e.AccessToken

	ctx := sourcegraph.WithCredentials(cliCtx,
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

	unauthedCtx := sourcegraph.WithCredentials(cliCtx, nil)
	unauthedCtx = fed.NewRemoteContext(unauthedCtx, endpointURL)
	cl := sourcegraph.NewClientFromContext(unauthedCtx)

	// Get client ID of server.
	conf, err := cl.Meta.Config(unauthedCtx, &pbtypes.Void{})
	if err != nil {
		return "", fmt.Errorf("Could not fetch config for %s: %v", endpointURL, err)
	}
	isLocalAuth := conf.IsFederationRoot || conf.AuthSource == "local" || conf.AuthSource == "ldap"

	///
	// Generate the token URL. OAuth2 auth is performed via the
	// federation root (which is also the OAuth2 Authorization
	// Server).
	///
	var rootURL *url.URL
	if isLocalAuth {
		rootURL = endpointURL
	} else {
		var err error
		rootURL, err = url.Parse(conf.FederationRootURL)
		if err != nil {
			return "", fmt.Errorf("%s has a malformed FederationRootURL: %s", endpointURL, err)
		}
	}
	tokenURL, err := router.Rel.URLToOrError(router.OAuth2ServerToken)
	if err != nil {
		return "", err
	}
	tokenURL = rootURL.ResolveReference(tokenURL)

	// Using the tokenURL, login either via username/password (local) or
	// get an oauth token from the fedRoot
	var accessTok string
	if isLocalAuth {
		var username, password string
		if c.Username != "" {
			username, password = c.Username, c.Password
		} else {
			fmt.Printf("Enter credentials for %s\n", rootURL)
			fmt.Print("Username: ")
			var err error
			username, err = getLine()
			if err != nil {
				return "", err
			}
			fmt.Print("Password: ")
			password = string(gopass.GetPasswd())
		}

		// Create a context for communicating with the fed root directly
		// (to avoid leaking username/password to the leaf server).
		rootCtx := fed.NewRemoteContext(unauthedCtx, rootURL)
		rootCl := sourcegraph.NewClientFromContext(rootCtx)

		// First, get a user access token to the root.
		//
		// We could do this via HTTP (as defined in the OAuth2 spec), but
		// using gRPC is a bit simpler and is consistent with how we do it
		// below (where using gRPC makes it much simpler, since we don't
		// have to mimic a browser's cookies and CSRF tokens).
		rootTok, err := rootCl.Auth.GetAccessToken(rootCtx, &sourcegraph.AccessTokenRequest{
			AuthorizationGrant: &sourcegraph.AccessTokenRequest_ResourceOwnerPassword{
				ResourceOwnerPassword: &sourcegraph.LoginCredentials{Login: username, Password: password},
			},
		})
		if err != nil {
			return "", fmt.Errorf("authenticating to %s: %s", rootURL, err)
		}

		accessTok = rootTok.AccessToken
	} else {
		// Create a context for communicating with the fed root directly
		// (to avoid leaking username/password to the leaf server).
		rootCtx := fed.NewRemoteContext(unauthedCtx, rootURL)
		rootCl := sourcegraph.NewClientFromContext(rootCtx)

		// Now, use the root access token to issue an auth code that
		// the leaf server can then exchange for an access token.
		rootAccessToken, err := c.getAccessToken(rootURL)
		if err != nil {
			return "", err
		}
		rootAuthedCtx := sourcegraph.WithCredentials(rootCtx,
			oauth2.StaticTokenSource(&oauth2.Token{TokenType: "Bearer", AccessToken: rootAccessToken}),
		)
		authInfo, err := rootCl.Auth.Identify(rootAuthedCtx, &pbtypes.Void{})
		if err != nil {
			return "", fmt.Errorf("identifying user on %s: %s", rootURL, err)
		}

		code, err := rootCl.Auth.GetAuthorizationCode(rootAuthedCtx, &sourcegraph.AuthorizationCodeRequest{
			ResponseType: "code",
			ClientID:     conf.IDKey,
			UID:          authInfo.UID,
		})
		if err != nil {
			return "", fmt.Errorf("getting auth code from %s: %s", rootURL, err)
		}

		// Exchange the auth code (from the root) for an access token.
		tok, err := cl.Auth.GetAccessToken(unauthedCtx, &sourcegraph.AccessTokenRequest{
			AuthorizationGrant: &sourcegraph.AccessTokenRequest_AuthorizationCode{
				AuthorizationCode: code,
			},
			TokenURL: tokenURL.String(),
		})
		if err != nil {
			return "", fmt.Errorf("exchanging auth code from %s for access token on %s: %s", rootURL, endpointURL, err)
		}

		accessTok = tok.AccessToken
	}

	saveCredentials(endpointURL, accessTok, false)
	return accessTok, nil
}

func saveCredentials(endpointURL *url.URL, accessTok string, makeDefault bool) error {
	a, err := readUserAuth()
	if err != nil {
		return err
	}
	if a == nil {
		a = userAuth{}
	}

	var updatedDefault, updatedCredentials bool
	ua, ok := a[endpointURL.String()]
	if ok {
		if ua.AccessToken != accessTok {
			updatedCredentials = true
			ua.AccessToken = accessTok
		}
	} else {
		updatedCredentials = true
		ua = &userEndpointAuth{AccessToken: accessTok}
		a[endpointURL.String()] = ua
	}
	if makeDefault && !ua.Default {
		updatedDefault = true
		a.setDefault(endpointURL.String())
	}

	if err := writeUserAuth(a); err != nil {
		return err
	}
	if updatedCredentials {
		log.Printf("# Credentials for %s saved to %s.", endpointURL, userAuthFileName())
	}
	if updatedDefault {
		log.Printf("# Default endpoint set to %s.", endpointURL)
	}
	return nil
}

func (c *loginCmd) Execute(args []string) error {
	// Check if parseable, before attempting authentication
	_, err := readUserAuth()
	if err != nil {
		return err
	}

	// We allow the endpoint URL to be passed in as an argument as a
	// convenience to --endpoint
	if c.Args.EndpointURL != "" {
		Endpoint.URL = c.Args.EndpointURL
	}

	endpointURL := Endpoint.URLOrDefault()
	accessTok, err := c.getAccessToken(endpointURL)
	if err != nil {
		return err
	}

	err = saveCredentials(endpointURL, accessTok, true)
	if err != nil {
		return err
	}
	return nil
}

type whoamiCmd struct {
	PrintToken bool `long:"print-token" description:"print the token used for authenticating requests"`
}

func (c *whoamiCmd) Execute(args []string) error {
	a, err := readUserAuth()
	if err != nil {
		return err
	}
	endpointURL := Endpoint.URLOrDefault()
	ua := a[endpointURL.String()]
	if ua == nil {
		log.Fatalf("# No authentication info set for %s (use `%s login` to authenticate)", endpointURL, sgxcmd.Name)
	}

	cl := Client()

	authInfo, err := cl.Auth.Identify(cliCtx, &pbtypes.Void{})
	if err != nil {
		log.Fatalf("Error verifying auth credentials with endpoint %s: %s.", endpointURL, err)
	}
	log.Printf("%s (UID %d) on %s", authInfo.Login, authInfo.UID, endpointURL)

	if c.PrintToken {
		log.Printf(" Auth token: %s", ua.AccessToken)
	}

	return nil
}
