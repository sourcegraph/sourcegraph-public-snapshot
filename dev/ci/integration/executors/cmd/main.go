package main

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
)

const q = `query { currentUser { username } }`

type Config struct {
	URL   string   `json:"url"`
	Repos []string `json:"repos"`
}

type ExternalSvc struct {
	Kind        string `json:"Kind"`
	DisplayName string `json:"DisplayName"`
	Config      Config `json:"Config"`
}

type configWithToken struct {
	URL   string   `json:"url"`
	Repos []string `json:"repos"`
	Token string   `json:"token"`
}

func main() {
	ctx := context.Background()
	logfuncs := log.Init(log.Resource{
		Name: "executors test runner",
	})
	defer logfuncs.Sync()

	logger := log.Scoped("init", "runner initialization process")

	SourcegraphAccessToken = createSudoToken()

	if err := InitializeGraphQLClient(); err != nil {
		logger.Fatal("cannot initialize graphql client", log.Error(err))
	}

	res := map[string]any{}
	err := queryGraphQL(ctx, logger.Scoped("graphql", ""), "", q, nil, &res)
	if err != nil {
		logger.Fatal("graphql failed with", log.Error(err))
	}

	b, _ := json.MarshalIndent(res, "", "  ")
	println(string(b))

	// --- adding repos ---
	client, err = gqltestutil.SignIn(SourcegraphEndpoint, adminEmail, adminPassword)
	if err != nil {
		logger.Fatal("Failed to sign in:", log.Error(err))
	}

	f, err := os.Open("config/repos.json")
	if err != nil {
		logger.Fatal("Failed to open config/repos.json:", log.Error(err))
	}
	defer f.Close()

	svcs := []ExternalSvc{}
	dec := json.NewDecoder(f)
	if err := dec.Decode(&svcs); err != nil {
		logger.Fatal("cannot parse repos.json", log.Error(err))
	}

	githubToken := os.Getenv("GITHUB_TOKEN")
	for _, svc := range svcs {
		b, _ := json.Marshal(configWithToken{
			Repos: svc.Config.Repos,
			URL:   svc.Config.URL,
			Token: githubToken,
		})

		_, err := GraphQLClient().AddExternalService(gqltestutil.AddExternalServiceInput{
			Kind:        svc.Kind,
			DisplayName: svc.DisplayName,
			Config:      string(b),
		})
		if err != nil {
			logger.Fatal("failed to add external service", log.String("name", svc.DisplayName), log.Error(err))
		}

		for _, repo := range svc.Config.Repos {
			split := strings.Split(svc.Config.URL, "https://")
			r := split[1] + "/" + repo
			if err = client.WaitForReposToBeCloned(r); err != nil {
				logger.Fatal("failed to wait for repo to be cloned", log.String("repo", r))
			}
		}
	}
}
