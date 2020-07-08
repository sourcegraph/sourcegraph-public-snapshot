package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// buildTag is the git tag at the time of build and is used to
// denote the binary's current version. This value is supplied
// as an ldflag at compile time in travis.
var buildTag = "dev"

func init() {
	usage := `
Examples:

  Get the src-cli version and the Sourcegraph instance's recommended version:

    	$ src version
`

	flagSet := flag.NewFlagSet("version", flag.ExitOnError)

	handler := func(args []string) error {
		fmt.Printf("Current version: %s\n", buildTag)

		recommendedVersion, err := getRecommendedVersion()
		if err != nil {
			return err
		}
		if recommendedVersion == "" {
			fmt.Println("Recommended Version: <unknown>")
			fmt.Println("This Sourcegraph instance does not support this feature.")
			return nil
		}
		fmt.Printf("Recommended Version: %s\n", recommendedVersion)
		return nil
	}

	// Register the command.
	commands = append(commands, &command{
		flagSet: flagSet,
		handler: handler,
		usageFunc: func() {
			fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src %s':\n", flagSet.Name())
			flagSet.PrintDefaults()
			fmt.Println(usage)
		},
	})
}

func getRecommendedVersion() (string, error) {
	url, err := url.Parse(cfg.Endpoint + "/.api/src-cli/version")
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return "", err
	}
	for k, v := range cfg.AdditionalHeaders {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return "", nil
		}

		return "", fmt.Errorf("error: %s\n\n%s", resp.Status, body)
	}

	payload := struct {
		Version string `json:"version"`
	}{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", err
	}

	return payload.Version, nil
}
