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

	githubToken           = flag.String("github-token", os.Getenv("GITHUB_TOKEN"), "The GitHub personal access token that will be used to authenticate a GitHub external service")
	awsAccessKeyID        = flag.String("aws-access-key-id", os.Getenv("AWS_ACCESS_KEY_ID"), "The AWS access key ID")
	awsSecretAccessKey    = flag.String("aws-secret-access-key", os.Getenv("AWS_SECRET_ACCESS_KEY"), "The AWS secret access key")
	awsCodeCommitUsername = flag.String("aws-code-commit-username", os.Getenv("AWS_CODE_COMMIT_USERNAME"), "The AWS code commit username")
	awsCodeCommitPassword = flag.String("aws-code-commit-password", os.Getenv("AWS_CODE_COMMIT_PASSWORD"), "The AWS code commit password")
	bbsURL                = flag.String("bbs-url", os.Getenv("BITBUCKET_SERVER_URL"), "The Bitbucket Server URL")
	bbsToken              = flag.String("bbs-token", os.Getenv("BITBUCKET_SERVER_TOKEN"), "The Bitbucket Server token")
	bbsUsername           = flag.String("bbs-username", os.Getenv("BITBUCKET_SERVER_USERNAME"), "The Bitbucket Server username")
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
