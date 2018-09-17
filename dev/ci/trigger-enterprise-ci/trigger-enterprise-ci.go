package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {
	var build struct {
		Number int64
	}
	commit := os.Getenv("BUILDKITE_COMMIT")
	if commit == "" {
		panic("BUILDKITE_COMMIT env var not set")
	}
	buildCreator := os.Getenv("BUILDKITE_BUILD_CREATOR")
	if buildCreator == "" {
		panic("BUILDKITE_BUILD_CREATOR env var not set")
	}
	buildCreatorEmail := os.Getenv("BUILDKITE_BUILD_CREATOR_EMAIL")
	if buildCreatorEmail == "" {
		panic("BUILDKITE_BUILD_CREATOR_EMAIL env var not set")
	}
	buildURL := os.Getenv("BUILDKITE_BUILD_URL")
	if buildURL == "" {
		panic("BUILDKITE_BUILD_URL env var not set")
	}
	buildNumber := os.Getenv("BUILDKITE_BUILD_NUMBER")
	if buildURL == "" {
		panic("BUILDKITE_BUILD_NUMBER env var not set")
	}
	webappVersion, err := exec.Command("buildkite-agent", "meta-data", "get", "oss-webapp-version").Output()
	if err != nil {
		panic(err)
	}
	webappVersionStr := strings.TrimSpace(string(webappVersion))
	if webappVersionStr == "" {
		panic("no webapp version was set")
	}
	{
		body, err := json.Marshal(map[string]interface{}{
			"commit":  "HEAD",
			"branch":  "master",
			"message": "Open source repository commit " + commit[0:7],
			"author": map[string]interface{}{
				// Forward creator data so that build appears in "my builds" list etc.
				"name":  buildCreator,
				"email": buildCreatorEmail,
			},
			"meta_data": map[string]interface{}{
				"oss-repo-revision":  commit,
				"oss-webapp-version": webappVersionStr,
				"oss-build-url":      buildURL,
				"oss-build-number":   buildNumber,
			},
		})
		if err != nil {
			panic(err)
		}
		req, err := http.NewRequest("POST", os.ExpandEnv("https://api.buildkite.com/v2/organizations/sourcegraph/pipelines/enterprise/builds?access_token=$BUILDKITE_TOKEN"), bytes.NewReader(body))
		if err != nil {
			panic(err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		err = json.NewDecoder(resp.Body).Decode(&build)
		if err != nil {
			panic(err)
		}
	}
	fmt.Println("Waiting for enterprise build to finish")
	for {
		req, err := http.NewRequest("GET", os.ExpandEnv(fmt.Sprintf("https://api.buildkite.com/v2/organizations/sourcegraph/pipelines/enterprise/builds/%v?access_token=$BUILDKITE_TOKEN", build.Number)), nil)
		if err != nil {
			panic(err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			panic(err)
		}
		var build struct {
			State string
		}
		err = json.NewDecoder(resp.Body).Decode(&build)
		if err != nil {
			panic(err)
		}
		resp.Body.Close()
		bs := build.State
		fmt.Println("State: " + bs)
		switch bs {
		case "passed":
			os.Exit(0)
		case "running", "scheduled":
			time.Sleep(1 * time.Second)
			continue
		default:
			fmt.Println("enterprise build ended with status:", bs)
			os.Exit(1)
		}
	}
}
