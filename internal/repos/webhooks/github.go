package githubwebhook

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/google/go-github/v43/github"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

var Url = ""

func handleSyncWebhook(ctx context.Context, extSvc *types.ExternalService, payload any) error {
	fmt.Println("handleSyncWebhook...")
	repo := payload.(*github.PushEvent).GetRepo()
	name := api.RepoName(*repo.Name)

	var cli *repoupdater.Client
	if Url == "" {
		cli = repoupdater.DefaultClient
	} else {
		cli = repoupdater.NewClient(Url)
	}

	res, err := cli.EnqueueRepoUpdate(ctx, name)
	if err != nil {
		return errors.New(fmt.Sprint("error enqueuing repo", err))
	}
	fmt.Printf("Enqueued:%+v\n", res)

	return nil
}

type Config struct {
	Url          string `json:"url"`
	Content_type string `json:"content_type"`
	Secret       string `json:"secret"`
	Insecure_ssl string `json:"insecure_ssl"`
	Token        string `json:"token"`
	Digest       string `json:"digest,omitempty"`
}

type Payload struct {
	Name   string   `json:"name"`
	ID     int      `json:"id,omitempty"`
	Config Config   `json:"config"`
	Events []string `json:"events"`
	Active bool     `json:"active"`
}

var TOKEN = ""

func CreateSyncWebhook(repoName string, secret string, token string) error {
	fmt.Println("Creating webhook:", repoName)

	// HOW TO GENERATE THE SECRET
	// EXTRACTING THE TOKEN FROM THE USER

	// u := "https://api.github.com/repos/susantoscott/Task-Tracker/hooks"
	url := fmt.Sprintf("https://api.github.com/repos/%s/hooks", repoName)
	// fmt.Println("Url:", url)
	payload := Payload{
		Name:   "web",
		Active: true,
		Config: Config{
			Url:          fmt.Sprintf("%s/enqueue-repo-update", repoupdater.DefaultClient.URL),
			Content_type: "json",
			Secret:       secret,
			Insecure_ssl: "0",
			Token:        TOKEN,
			Digest:       "",
		},
		Events: []string{
			"push",
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	req.Header.Add("Authorization", fmt.Sprintf("token %s", TOKEN))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	// fmt.Println("RespBody:", string(respBody))

	if resp.StatusCode >= 300 {
		// fmt.Println("STATUS CODE:", resp.StatusCode)
		// return errors.Newf("non-200 status code, %s", err)
	}

	var obj Payload
	if err := json.Unmarshal(respBody, &obj); err != nil {
		return err
	}

	return nil
}

func ListSyncWebhooks(repoName string, secret string, token string) []Payload {
	fmt.Println("Listing webhooks...")

	// url := "https://api.github.com/repos/susantoscott/Task-Tracker/hooks"
	url := fmt.Sprintf("https://api.github.com/repos/%s/hooks", repoName)
	fmt.Println("url:", url)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer([]byte("")))
	if err != nil {
		fmt.Println("making new request error:", err)
	}
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	req.Header.Add("Authorization", fmt.Sprintf("token %s", TOKEN))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("client do error:", err)
	}
	// fmt.Println("Status Code:", resp.StatusCode)

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("readall error:", err)
	}
	// fmt.Println("resp:", string(respBody))

	var obj []Payload
	if err := json.Unmarshal(respBody, &obj); err != nil {
		fmt.Println("unmarshal error:", err)
	}

	return obj
}

func FindSyncWebhook(repoName string, secret string, token string) string {
	payloads := ListSyncWebhooks(repoName, secret, token)
	for _, payload := range payloads {
		if payload.Name == "web" {
			return "web"
		}
	}
	return ""
}

func DeleteSyncWebhook(repoName string, hookID int) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/hooks/%d", repoName, hookID)
	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer([]byte("")))
	if err != nil {
		fmt.Println("making new request error:", err)
	}
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	req.Header.Add("Authorization", fmt.Sprintf("token %s", TOKEN))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("client do error:", err)
	}
	fmt.Println("Status Code:", resp.StatusCode)
}
