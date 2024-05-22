package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/sourcegraph/sourcegraph/dev/ci/integration/executors/tester/config"
	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const (
	adminEmail    = "sourcegraph@sourcegraph.com"
	adminUsername = "sourcegraph"
	adminPassword = "sourcegraphsourcegraph"
)

func initAndAuthenticate() (*gqltestutil.Client, error) {
	needsSiteInit, resp, err := gqltestutil.NeedsSiteInit(SourcegraphEndpoint)
	if resp != "" && os.Getenv("BUILDKITE") == "true" {
		log.Println("server response: ", resp)
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to check if site needs init")
	}

	client, err := gqltestutil.NewClient(SourcegraphEndpoint)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create gql client")
	}
	if needsSiteInit {
		if err := client.SiteAdminInit(adminEmail, adminUsername, adminPassword); err != nil {
			return nil, errors.Wrap(err, "failed to create site admin")
		}
		log.Println("Site admin has been created:", adminUsername)
	} else {
		if err = client.SignIn(adminEmail, adminPassword); err != nil {
			return nil, errors.Wrap(err, "failed to sign in")
		}
		log.Println("Site admin authenticated:", adminUsername)
	}

	return client, nil
}

func ensureRepos(client *gqltestutil.Client) error {

	var svcs []ExternalSvc
	if err := json.Unmarshal([]byte(config.Repos), &svcs); err != nil {
		return errors.Wrap(err, "cannot parse repos.json")
	}

	for _, svc := range svcs {
		b, err := json.Marshal(configWithToken{
			Repos: svc.Config.Repos,
			URL:   svc.Config.URL,
			Token: githubToken,
		})
		if err != nil {
			return err
		}

		_, err = client.AddExternalService(gqltestutil.AddExternalServiceInput{
			Kind:        svc.Kind,
			DisplayName: svc.DisplayName,
			Config:      string(b),
		})
		if err != nil {
			return errors.Wrapf(err, "failed to add external service %s", svc.DisplayName)
		}

		u, err := url.Parse(svc.Config.URL)
		if err != nil {
			return err
		}
		repos := []string{}
		for _, repo := range svc.Config.Repos {
			repos = append(repos, fmt.Sprintf("%s/%s", u.Host, repo))
		}

		log.Printf("waiting for repos to be cloned %v\n", repos)

		if err = client.WaitForReposToBeCloned(repos...); err != nil {
			return errors.Wrap(err, "failed to wait for repos to be cloned")
		}
	}

	return nil
}

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
