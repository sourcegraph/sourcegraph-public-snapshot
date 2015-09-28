package sgx

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"

	"github.com/howeyc/gopass"
	"golang.org/x/oauth2"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/fed/discover"
	"src.sourcegraph.com/sourcegraph/sgx/cli"
	"src.sourcegraph.com/sourcegraph/sgx/sgxcmd"
)

var userAuthFile = filepath.Join(os.Getenv("HOME"), ".src-auth")

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
// srclib endpoint.
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
	f, err := os.Open(userAuthFile)
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
	f, err := os.Create(userAuthFile)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := f.Chmod(0600); err != nil {
		return err
	}
	b, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return err
	}
	_, err = f.Write(b)
	return err
}

func init() {
	_, err := cli.CLI.AddCommand("login",
		"log in to Sourcegraph.com",
		"The login command logs into Sourcegraph.com using an access token. To obtain these values, sign up or log into Sourcegraph.com, then go to the 'Integrations' page in your user settings.",
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
	Root string `long:"root" description:"URL to federation root server"`
}

func (c *loginCmd) Execute(args []string) error {
	a, err := readUserAuth()
	if err != nil {
		return err
	}
	if a == nil {
		a = userAuth{}
	}

	endpointURL := getEndpointURL()
	cl := Client()
	unauthedCtx := sourcegraph.WithCredentials(cliCtx, nil)

	// Get client ID of server.
	conf, err := cl.Meta.Config(unauthedCtx, &pbtypes.Void{})
	if err != nil {
		return err
	}

	///
	// Generate the token URL. OAuth2 auth is performed via the
	// federation root (which is also the OAuth2 Authorization
	// Server).
	///
	var rootURL *url.URL
	// TODO(sqs): can remove second expression below
	// (`conf.FederationRootURL == ""`) when sourcegraph.com is
	// updated to include this commit (since then it will set the
	// IsFederationRoot bool value).
	isFedRoot := conf.IsFederationRoot || conf.FederationRootURL == ""
	if isFedRoot {
		rootURL = endpointURL
	} else {
		var err error
		rootURL, err = url.Parse(conf.FederationRootURL)
		if err != nil {
			return err
		}
	}
	tokenURL, err := router.Rel.URLToOrError(router.OAuth2ServerToken)
	if err != nil {
		return err
	}
	tokenURL = rootURL.ResolveReference(tokenURL)

	if isFedRoot {
		fmt.Printf("Enter credentials for %s\n", rootURL)
	} else {
		fmt.Printf("Enter credentials for %s to log into %s\n", rootURL, endpointURL)
	}
	fmt.Print("Username: ")
	username, err := getLine()
	if err != nil {
		return err
	}
	fmt.Print("Password: ")
	password := string(gopass.GetPasswd())

	// Create a context for communicating with the fed root directly
	// (to avoid leaking username/password to the leaf server).
	rootInfo, err := discover.SiteURL(cliCtx, rootURL.String())
	if err != nil {
		return err
	}
	rootCtx, err := rootInfo.NewContext(unauthedCtx)
	if err != nil {
		return err
	}
	rootCl := sourcegraph.NewClientFromContext(rootCtx)

	// First, get a user access token to the root.
	//
	// We could do this via HTTP (as defined in the OAuth2 spec), but
	// using gRPC is a bit simpler and is consistent with how we do it
	// below (where using gRPC makes it much simpler, since we don't
	// have to mimic a browser's cookies and CSRF tokens).
	rootTok, err := rootCl.Auth.GetAccessToken(rootCtx, &sourcegraph.AccessTokenRequest{
		ResourceOwnerPassword: &sourcegraph.LoginCredentials{Login: username, Password: password},
	})
	if err != nil {
		return fmt.Errorf("authenticating to root: %s", err)
	}

	var accessTok string
	if isFedRoot {
		accessTok = rootTok.AccessToken
	} else {
		// Now, use the root access token to issue an auth code that
		// the leaf server can then exchange for an access token.
		rootAuthedCtx := sourcegraph.WithCredentials(rootCtx,
			oauth2.StaticTokenSource(&oauth2.Token{TokenType: "Bearer", AccessToken: rootTok.AccessToken}),
		)
		authInfo, err := rootCl.Auth.Identify(rootAuthedCtx, &pbtypes.Void{})
		if err != nil {
			return fmt.Errorf("identifying root user: %s", err)
		}

		code, err := rootCl.Auth.GetAuthorizationCode(rootAuthedCtx, &sourcegraph.AuthorizationCodeRequest{
			ResponseType: "code",
			ClientID:     conf.IDKey,
			UID:          authInfo.UID,
		})
		if err != nil {
			return fmt.Errorf("getting auth code from root: %s", err)
		}

		// Exchange the auth code (from the root) for an access token.
		tok, err := cl.Auth.GetAccessToken(unauthedCtx, &sourcegraph.AccessTokenRequest{
			AuthorizationCode: code,
			TokenURL:          tokenURL.String(),
		})
		if err != nil {
			return fmt.Errorf("exchanging auth code for access token: %s", err)
		}

		accessTok = tok.AccessToken
	}

	ua := userEndpointAuth{AccessToken: accessTok}
	a[endpointURL.String()] = &ua
	a.setDefault(endpointURL.String())
	if err := writeUserAuth(a); err != nil {
		return err
	}
	log.Printf("# Credentials saved to %s.", userAuthFile)
	log.Printf("# Default endpoint set to %s.", endpointURL)
	return nil
}

type whoamiCmd struct {
}

func (c *whoamiCmd) Execute(args []string) error {
	a, err := readUserAuth()
	if err != nil {
		return err
	}
	endpointURL := getEndpointURL()
	ua := a[endpointURL.String()]
	if ua == nil {
		log.Fatalf("# No authentication info set for %s (use `%s login` to authenticate)", endpointURL, sgxcmd.Name)
	}

	cl := Client()

	authInfo, err := cl.Auth.Identify(cliCtx, &pbtypes.Void{})
	if err != nil {
		log.Fatalf("Error verifying auth credentials with endpoint %s: %s.", endpointURL, err)
	}
	user, err := cl.Users.Get(cliCtx, &sourcegraph.UserSpec{UID: authInfo.UID})
	if err != nil {
		log.Fatalf("Error getting user with UID %d: %s.", authInfo.UID, err)
	}
	log.Printf("%s (UID %d) on %s (verified remotely)", user.Login, user.UID, endpointURL)

	return nil
}
