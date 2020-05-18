package main

import (
	"flag"
	"os"
	"strings"
	"time"

	"github.com/inconshreveable/log15"
	jsoniter "github.com/json-iterator/go"
	"github.com/sourcegraph/sourcegraph/internal/e2eutil"
)

/*
	NOTE: For easier testing, run Sourcegraph instance without volume:
			docker run --publish 7080:7080 --rm sourcegraph/server:insiders
*/

func main() {
	baseURL := flag.String("base-url", "http://127.0.0.1:7080", "The base URL of the Sourcegraph instance")
	email := flag.String("email", "e2e@sourcegraph.com", "The email of the admin user")
	username := flag.String("username", "e2e-admin", "The username of the admin user")
	password := flag.String("password", "supersecurepassword", "The password of the admin user")

	githubToken := flag.String("github-token", os.Getenv("GITHUB_TOKEN"), "The GitHub personal access token that will be used to authenticate a GitHub external service")
	flag.Parse()

	*baseURL = strings.TrimSuffix(*baseURL, "/")

	// TODO(jchen): Find an easy and fast way to determine if the instance has done "Site admin init".
	client, err := e2eutil.SiteAdminInit(*baseURL, *email, *username, *password)
	if err != nil {
		log15.Error("Failed to create site admin", "error", err)
		os.Exit(1)
	}
	log15.Info("Site admin has been created", "username", *username)
	//client, err := e2eutil.SignIn(*baseURL, *email, *password)
	//if err != nil {
	//	log15.Error("Failed to sign in", "err", err)
	//	os.Exit(1)
	//}
	//log15.Info("Site admin authenticated", "username", *username)

	// Set up external service
	err = client.AddExternalService(e2eutil.AddExternalServiceInput{
		Kind:        "GITHUB",
		DisplayName: "e2e-test-github",
		Config: mustMarshalJSONString(struct {
			URL   string   `json:"url"`
			Token string   `json:"token"`
			Repos []string `json:"repos"`
		}{
			"http://github.com",
			*githubToken,
			[]string{
				"sourcegraph/java-langserver",
				"gorilla/mux",
				"gorilla/securecookie",
				"sourcegraph/jsonrpc2",
				"sourcegraph/go-diff",
				"sourcegraph/appdash",
				"sourcegraph/sourcegraph-typescript",
				"sourcegraph-testing/automation-e2e-test",
				"sourcegraph/e2e-test-private-repository",
			},
		}),
	})
	if err != nil {
		log15.Error("Failed to add external service", "error", err)
		os.Exit(1)
	}

	time.Sleep(10 * time.Second) // TODO
	resutls, err := client.SearchRepositories("type:repo visibility:private")
	if err != nil {
		log15.Error("Failed to search", "error", err)
		os.Exit(1)
	}
	found := false
	for _, r := range resutls {
		if r.Name == "github.com/sourcegraph/e2e-test-private-repository" {
			found = true
			break
		}
	}
	if !found {
		log15.Error("Visibility filter", "error", "private repository not found")
		os.Exit(1)
	}
}

func mustMarshalJSONString(v interface{}) string {
	str, err := jsoniter.MarshalToString(v)
	if err != nil {
		panic(err)
	}
	return str
}
