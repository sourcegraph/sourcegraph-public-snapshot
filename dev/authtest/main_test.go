package authtest

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/inconshreveable/log15"
	jsoniter "github.com/json-iterator/go"

	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
)

var client *gqltestutil.Client

var (
	long = flag.Bool("long", false, "Enable the auth tests to run. Required flag, otherwise tests are skipped.")

	baseURL  = flag.String("base-url", "http://127.0.0.1:7080", "The base URL of the Sourcegraph instance")
	email    = flag.String("email", "authtest@sourcegraph.com", "The email of the admin user")
	username = flag.String("username", "authtest-admin", "The username of the admin user")
	password = flag.String("password", "supersecurepassword", "The password of the admin user")

	githubToken = flag.String("github-token", os.Getenv("GITHUB_TOKEN"), "The GitHub personal access token that will be used to authenticate a GitHub external service")

	dotcom = flag.Bool("dotcom", false, "Whether to test dotcom specific constraints")
)

func TestMain(m *testing.M) {
	flag.Parse()

	if !*long {
		fmt.Println("SKIP: skipping authtest since -long is not specified.")
		return
	}

	*baseURL = strings.TrimSuffix(*baseURL, "/")

	// Make it possible to use a different token on non-default branches
	// so we don't break builds on the default branch.
	mockGitHubToken := os.Getenv("MOCK_GITHUB_TOKEN")
	if mockGitHubToken != "" {
		*githubToken = mockGitHubToken
	}

	needsSiteInit, resp, err := gqltestutil.NeedsSiteInit(*baseURL)
	if resp != "" && os.Getenv("BUILDKITE") == "true" {
		log.Println("server response: ", resp)
	}
	if err != nil {
		log.Fatal("Failed to check if site needs init: ", err)
	}

	client, err = gqltestutil.NewClient(*baseURL)
	if err != nil {
		log.Fatal("Failed to create gql client: ", err)
	}
	if needsSiteInit {
		if err := client.SiteAdminInit(*email, *username, *password); err != nil {
			log.Fatal("Failed to create site admin: ", err)
		}
		log.Println("Site admin has been created:", *username)
	} else {
		if err := client.SignIn(*email, *password); err != nil {
			log.Fatal("Failed to sign in:", err)
		}
		log.Println("Site admin authenticated:", *username)
	}

	if !testing.Verbose() {
		log15.Root().SetHandler(log15.DiscardHandler())
	}
	os.Exit(m.Run())
}

func mustMarshalJSONString(v any) string {
	str, err := jsoniter.MarshalToString(v)
	if err != nil {
		panic(err)
	}
	return str
}
