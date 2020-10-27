package main

import (
	"flag"
	"log"
	"os"

	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
)

var client *gqltestutil.Client

var (
	baseURL  = flag.String("base-url", "http://127.0.0.1:7080", "The base URL of the Sourcegraph instance")
	email    = flag.String("email", "test@sourcegraph.com", "The email of the admin user")
	username = flag.String("username", "admin", "The username of the admin user")
	password = flag.String("password", "supersecurepassword", "The password of the admin user")
)

func main() {
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

	siteConfig, err := client.SiteConfiguration()
	if err != nil {
		log.Fatal(err)
	}

	siteConfig.ExternalURL = "http://127.0.0.1:7080"

	err = client.UpdateSiteConfiguration(siteConfig)
	if err != nil {
		log.Fatal(err)
	}

	envvar := "export SOURCEGRAPH_SUDO_TOKEN=" + token

	file, err := os.OpenFile("/root/.profile", os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()
	if _, err := file.WriteString(envvar); err != nil {
		log.Fatal(err)
	}

}
