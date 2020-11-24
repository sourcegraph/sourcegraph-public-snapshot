package main

import (
	"log"

	jsoniter "github.com/json-iterator/go"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
)

func mustMarshalJSONString(v interface{}) string {
	str, err := jsoniter.MarshalToString(v)
	if err != nil {
		panic(err)
	}
	return str
}

func addReposCommand() error {

	if len(githubToken) == 0 {
		log.Fatal("Environment variable GITHUB_TOKEN is not set")
	}

	client, err := gqltestutil.SignIn(endpoint, email, password)
	if err != nil {
		log.Fatal("Failed to sign in:", err)
	}
	log.Println("Site admin authenticated:", username)

	// Set up external service
	esID, err := client.AddExternalService(gqltestutil.AddExternalServiceInput{
		Kind:        extsvc.KindGitHub,
		DisplayName: "code-intel-repos",
		Config: mustMarshalJSONString(struct {
			URL   string   `json:"url"`
			Token string   `json:"token"`
			Repos []string `json:"repos"`
		}{
			URL:   "http://github.com",
			Token: githubToken,
			Repos: []string{
				"sourcegraph-testing/etcd",
				"sourcegraph-testing/tidb",
				"sourcegraph-testing/titan",
				"sourcegraph-testing/zap",
			},
		}),
	})
	if err != nil {
		log.Fatal(err)
	}

	err = client.WaitForReposToBeCloned(
		"github.com/sourcegraph-testing/etcd",
		"github.com/sourcegraph-testing/tidb",
		"github.com/sourcegraph-testing/titan",
		"github.com/sourcegraph-testing/zap",
	)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Print(esID)
	}

	return err
}
