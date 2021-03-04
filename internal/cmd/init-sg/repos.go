package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	jsoniter "github.com/json-iterator/go"

	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
)

func mustMarshalJSONString(v interface{}) string {
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

	byteValue, _ := ioutil.ReadAll(jsonFile)

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
