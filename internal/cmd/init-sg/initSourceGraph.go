package main

import (
	"log"
	"os"

	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
)

func initSourceGraph() {
	log.Println("Running initializer")

	needsSiteInit, err := gqltestutil.NeedsSiteInit(*baseURL)
	if err != nil {
		log.Fatal("Failed to check if site needs init: ", err)
	}

	if needsSiteInit {
		client, err = gqltestutil.SiteAdminInit(*baseURL, *email, *username, *password)
		if err != nil {
			log.Fatal("Failed to create site admin: ", err)
		}
		log.Println("Site admin has been created:", *username)
	} else {
		client, err = gqltestutil.SignIn(*baseURL, *email, *password)
		if err != nil {
			log.Fatal("Failed to sign in:", err)
		}
		log.Println("Site admin authenticated:", *username)
	}

	token, err := client.CreateAccessToken("TestAccessToken", []string{"user:all", "site-admin:sudo"})
	if err != nil {
		log.Fatal("Failed to create token: ", err)
	}
	if token == "" {
		log.Fatal("Failed to create token")
	}

	// Ensure site configuration is set up correctly
	siteConfig, err := client.SiteConfiguration()
	if err != nil {
		log.Fatal(err)
	}
	if siteConfig.ExternalURL != *baseURL {
		siteConfig.ExternalURL = *baseURL
		err = client.UpdateSiteConfiguration(siteConfig)
		if err != nil {
			log.Fatal(err)
		}
	}

	envvar := "export SOURCEGRAPH_SUDO_TOKEN=" + token
	file, err := os.OpenFile(profile, os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	if _, err := file.WriteString(envvar); err != nil {
		log.Fatal(err)
	}

	log.Println("Instance initialized, SOURCEGRAPH_SUDO_TOKEN set in", profile)
}
