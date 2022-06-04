package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	jsoniter "github.com/json-iterator/go"

	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
)

var (
	client *gqltestutil.Client

	initSG   = flag.NewFlagSet("initserver", flag.ExitOnError)
	addRepos = flag.NewFlagSet("addrepos", flag.ExitOnError)

	baseURL  = initSG.String("baseurl", os.Getenv("SOURCEGRAPH_BASE_URL"), "The base URL of the Sourcegraph instance. (Required)")
	email    = initSG.String("email", os.Getenv("TEST_USER_EMAIL"), "The email of the admin user. (Required)")
	username = initSG.String("username", os.Getenv("SOURCEGRAPH_SUDO_USER"), "The username of the admin user. (Required)")
	password = initSG.String("password", os.Getenv("TEST_USER_PASSWORD"), "The password of the admin user. (Required)")

	githubToken    = addRepos.String("githubtoken", os.Getenv("GITHUB_TOKEN"), "The github access token that will be used to authenticate an external service. (Required)")
	addReposConfig = addRepos.String("config", "", "Path to the external service config. (Required)")

	home    = os.Getenv("HOME")
	profile = home + "/.sg_envrc"
)

func main() {
	flag.Parse()

	if len(os.Args) < 2 {
		fmt.Println("initSG or addRepos subcommand is required")
		flag.PrintDefaults()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "initSG":
		initSG.Parse(os.Args[2:])
		initSourcegraph()
	case "addRepos":
		addRepos.Parse(os.Args[2:])
		addReposCommand()
	case "default":
		flag.PrintDefaults()
		os.Exit(1)
	}

}

func initSourcegraph() {
	log.Println("Running initializer")

	needsSiteInit, resp, err := gqltestutil.NeedsSiteInit(*baseURL)
	if resp != "" && os.Getenv("BUILDKITE") == "true" {
		log.Println("server response: ", resp)
	}
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
	file, err := os.Create(profile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	if _, err := file.WriteString(envvar); err != nil {
		log.Fatal(err)
	}

	log.Println("Instance initialized, SOURCEGRAPH_SUDO_TOKEN set in", profile)
}
func mustMarshalJSONString(v any) string {
	str, err := jsoniter.MarshalToString(v)
	if err != nil {
		panic(err)
	}
	fmt.Print(str)
	return str
}

func addReposCommand() {
	if len(*githubToken) == 0 {
		log.Fatal("Environment variable GITHUB_TOKEN is not set")
	}

	client, err := gqltestutil.SignIn(*baseURL, *email, *password)
	if err != nil {
		log.Fatal("Failed to sign in:", err)
	}
	log.Println("Site admin authenticated:", *username)

	// Open our jsonFile
	jsonFile, err := os.Open(*addReposConfig)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}

	defer jsonFile.Close()

	type Config struct {
		URL   string   `json:"url"`
		Repos []string `json:"repos"`
	}

	type ExternalSvc struct {
		Kind        string `json:"Kind"`
		DisplayName string `json:"DisplayName"`
		Config      Config `json:"Config"`
	}

	byteValue, _ := io.ReadAll(jsonFile)

	var externalsvcs []ExternalSvc

	jsoniter.Unmarshal(byteValue, &externalsvcs)

	for i := range externalsvcs {

		// Set up external service
		esID, err := client.AddExternalService(gqltestutil.AddExternalServiceInput{
			Kind:        externalsvcs[i].Kind,
			DisplayName: externalsvcs[i].DisplayName,
			Config: mustMarshalJSONString(struct {
				URL   string   `json:"url"`
				Token string   `json:"token"`
				Repos []string `json:"repos"`
			}{
				URL:   externalsvcs[i].Config.URL,
				Token: *githubToken,
				Repos: externalsvcs[i].Config.Repos,
			}),
		})

		if err != nil {
			log.Fatal(err)
		}
		for _, r := range externalsvcs[i].Config.Repos {
			split := strings.Split(externalsvcs[i].Config.URL, "https://")
			repo := split[1] + "/" + r
			log.Print(repo)
			err = client.WaitForReposToBeCloned(repo)
		}
		if err != nil {
			log.Fatal(err)
		} else {
			log.Print(esID)
		}
	}
}
