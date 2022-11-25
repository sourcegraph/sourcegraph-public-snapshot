package main

import (
	"log"
	"os"

	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
)

const (
	adminEmail    = "sourcegraph@sourcegraph.com"
	adminUsername = "sourcegraph"
	adminPassword = "sourcegraphsourcegraph"
)

// TODO @jhchabran
// crudely hacked based on cmd/internal/init-sg
// doesn't edit siteconfig so we can use the siteconfig file
func createSudoToken() (*gqltestutil.Client, string) {
	needsSiteInit, resp, err := gqltestutil.NeedsSiteInit(SourcegraphEndpoint)
	if resp != "" && os.Getenv("BUILDKITE") == "true" {
		log.Println("server response: ", resp)
	}
	if err != nil {
		log.Fatal("Failed to check if site needs init: ", err)
	}

	var client *gqltestutil.Client
	if needsSiteInit {
		client, err = gqltestutil.SiteAdminInit(SourcegraphEndpoint, adminEmail, adminUsername, adminPassword)
		if err != nil {
			log.Fatal("Failed to create site admin: ", err)
		}
		log.Println("Site admin has been created:", adminUsername)
	} else {
		client, err = gqltestutil.SignIn(SourcegraphEndpoint, adminEmail, adminPassword)
		if err != nil {
			log.Fatal("Failed to sign in:", err)
		}
		log.Println("Site admin authenticated:", adminUsername)
	}

	token, err := client.CreateAccessToken("TestAccessToken", []string{"user:all", "site-admin:sudo"})
	if err != nil {
		log.Fatal("Failed to create token: ", err)
	}
	if token == "" {
		log.Fatal("Failed to create token")
	}

	return client, token
}
