package main

import (
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
)

const currentUsernameQuery = `query { currentUser { id username } }`

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

var SourcegraphEndpoint = env.Get("SOURCEGRAPH_BASE_URL", "http://127.0.0.1:7080", "Sourcegraph frontend endpoint")
var SourcegraphAccessToken string

func main() {
	logfuncs := log.Init(log.Resource{
		Name: "executors test runner",
	})
	defer logfuncs.Sync()

	logger := log.Scoped("init", "runner initialization process")

	var client *gqltestutil.Client
	client, SourcegraphAccessToken = createSudoToken()

	f, err := os.Open("config/repos.json")
	if err != nil {
		logger.Fatal("Failed to open config/repos.json:", log.Error(err))
	}

	svcs := []ExternalSvc{}
	dec := json.NewDecoder(f)
	if err := dec.Decode(&svcs); err != nil {
		f.Close()
		logger.Fatal("cannot parse repos.json", log.Error(err))
	}
	f.Close()

	githubToken := os.Getenv("GITHUB_TOKEN")
	for _, svc := range svcs {
		b, _ := json.Marshal(configWithToken{
			Repos: svc.Config.Repos,
			URL:   svc.Config.URL,
			Token: githubToken,
		})

		_, err := client.AddExternalService(gqltestutil.AddExternalServiceInput{
			Kind:        svc.Kind,
			DisplayName: svc.DisplayName,
			Config:      string(b),
		})
		if err != nil {
			logger.Fatal("failed to add external service", log.String("name", svc.DisplayName), log.Error(err))
		}

		for _, repo := range svc.Config.Repos {
			split := strings.Split(svc.Config.URL, "https://")
			r := split[1] + "/" + repo
			if err = client.WaitForReposToBeCloned(r); err != nil {
				logger.Fatal("failed to wait for repo to be cloned", log.String("repo", r))
			}
		}
	}

	// Now that we have our repositories synced and cloned into the instance, we
	// can start triggering an execution.
	batchSpecID, err := createBatchSpecForExecution(logger, client)
	if err != nil {
		logger.Fatal("failed to create batch spec for execution", log.Error(err))
	}

	// Now an execution has been enqueued. We wait for it to complete now.
	if err := awaitBatchSpecExecution(logger, client, batchSpecID); err != nil {
		logger.Fatal("failed to await batch spec execution", log.Error(err))
	}

	// Finally, we assert that the execution is in the correct shape.
	if err := assertBatchSpecExecution(logger, client, batchSpecID); err != nil {
		logger.Fatal("failed to assert batch spec execution", log.Error(err))
	}
}

const batchSpec = `
name: e2e-test-batch-change
description: Add Hello World to READMEs

on:
  - repository: github.com/sourcegraph/automation-testing
    branch: executors-e2e

steps:
  - run: IFS=$'\n'; echo Hello World | tee -a $(find -name README.md)
    container: ubuntu:18.04

changesetTemplate:
  title: Hello World
  body: My first batch change!
  branch: hello-world # Push the commit to this branch.
  commit:
    message: Append Hello World to all README.md files
`

func createBatchSpecForExecution(logger log.Logger, client *gqltestutil.Client) (string, error) {
	id, err := client.CurrentUserID("")
	if err != nil {
		return "", err
	}
	batchChangeID, err := client.CreateEmptyBatchChange(id, "e2e-test-batch-change")
	if err != nil {
		return "", err
	}
	batchSpecID, err := client.CreateBatchSpecFromRaw(batchChangeID, id, batchSpec)
	if err != nil {
		return "", err
	}
	start := time.Now()
	for {
		if time.Now().Sub(start) > 60*time.Second {
			logger.Fatal("Waiting for batch spec workspace resolution to complete timed out")
		}
		state, err := client.GetBatchSpecWorkspaceResolutionStatus(batchSpecID)
		if err != nil {
			return "", err
		}
		if state == "COMPLETE" {
			break
		}
	}
	// Execute with cache disabled.
	return batchSpecID, client.ExecuteBatchSpec(batchSpecID, true)
}

func awaitBatchSpecExecution(logger log.Logger, client *gqltestutil.Client, batchSpecID string) error {
	start := time.Now()
	for {
		// Wait for at most 3 minutes.
		if time.Now().Sub(start) > 3*60*time.Second {
			logger.Fatal("Waiting for batch spec execution to complete timed out")
		}
		state, err := client.GetBatchSpecState(batchSpecID)
		if err != nil {
			return err
		}
		if state == "FAILED" {
			logger.Fatal("Batch spec ended in failed state")
		}
		if state == "COMPLETE" {
			break
		}
	}
	return nil
}

func assertBatchSpecExecution(logger log.Logger, client *gqltestutil.Client, batchSpecID string) error {
	batchSpec, err := client.GetBatchSpecDeep(batchSpecID)
	if err != nil {
		return err
	}
	if batchSpec.State != "COMPLETE" {
		logger.Fatal("batch spec is not complete")
	}
	return nil
}
