package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log" //nolint:logging // TODO move all logging to sourcegraph/log
	"os"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"

	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
)

var (
	client *gqltestutil.Client

	initSG       = flag.NewFlagSet("initserver", flag.ExitOnError)
	addRepos     = flag.NewFlagSet("addrepos", flag.ExitOnError)
	oobmigration = flag.NewFlagSet("oobmigration", flag.ExitOnError)

	baseURL  = initSG.String("baseurl", os.Getenv("SOURCEGRAPH_BASE_URL"), "The base URL of the Sourcegraph instance. (Required)")
	email    = initSG.String("email", os.Getenv("TEST_USER_EMAIL"), "The email of the admin user. (Required)")
	username = initSG.String("username", os.Getenv("SOURCEGRAPH_SUDO_USER"), "The username of the admin user. (Required)")
	password = initSG.String("password", os.Getenv("TEST_USER_PASSWORD"), "The password of the admin user. (Required)")
	sgenvrc  = initSG.String("sg_envrc", os.Getenv("SG_ENVRC"), "Location of the sg_envrc file to write down the sudo token to")

	githubToken    = addRepos.String("githubtoken", os.Getenv("GITHUB_TOKEN"), "The github access token that will be used to authenticate an external service. (Required)")
	addReposConfig = addRepos.String("config", "", "Path to the external service config. (Required)")

	migrationID       = oobmigration.String("id", "", "The target oobmigration identifier. (Required)")
	migrationDownFlag = oobmigration.Bool("down", false, "Supply to change the migration from up (default) to down.")

	home    = os.Getenv("HOME")
	profile = home + "/.sg_envrc"
)

func main() {
	flag.Parse()

	if len(os.Args) < 2 {
		fmt.Println("initSG, addRepos, or oobmigration subcommand is required")
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
	case "oobmigration":
		oobmigration.Parse(os.Args[2:])
		oobmigrationCommand()
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

	Days60 := int(86400 * 60)
	token, err := client.CreateAccessToken("TestAccessToken", []string{"user:all", "site-admin:sudo"}, &Days60) // default to a 60 day token
	if err != nil {
		log.Fatal("Failed to create token: ", err)
	}
	if token == "" {
		log.Fatal("Failed to create token")
	}

	// Ensure site configuration is set up correctly
	siteConfig, lastID, err := client.SiteConfiguration()
	if err != nil {
		log.Fatal(err)
	}
	if siteConfig.ExternalURL != *baseURL {
		siteConfig.ExternalURL = *baseURL
		err = client.UpdateSiteConfiguration(siteConfig, lastID)
		if err != nil {
			log.Fatal(err)
		}
	}

	envvar := "export SOURCEGRAPH_SUDO_TOKEN=" + token
	if *sgenvrc != "" {
		profile = *sgenvrc
	}
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

	client, err := gqltestutil.NewClient(*baseURL)
	if err != nil {
		log.Fatal("Failed to create gql client: ", err)
	}
	if err := client.SignIn(*email, *password); err != nil {
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

const MigrationTimeout = time.Minute * 5

func oobmigrationCommand() {
	if *migrationID == "" {
		log.Fatal("migration identifier (-id) is not supplied")
	}
	id := *migrationID
	up := !*migrationDownFlag

	client, err := gqltestutil.NewClient(*baseURL)
	if err != nil {
		log.Fatal("Failed to create gql client: ", err)
	}
	if err := client.SignIn(*email, *password); err != nil {
		log.Fatal("Failed to sign in:", err)
	}
	log.Println("Site admin authenticated:", *username)

	if err := client.SetMigrationDirection(id, up); err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), MigrationTimeout)
	defer cancel()

	if err := client.PollMigration(ctx, id, func(progress float64) bool {
		if up {
			log.Printf("Waiting for migration %s to complete (%.2f%% done).", id, progress*100)
		} else {
			log.Printf("Waiting for migration %s to rollback (%.2f%% done).", id, (1-progress)*100)
		}

		return (up && progress == 1) || (!up && progress == 0)
	}); err != nil {
		log.Fatal(err)
	}
}
