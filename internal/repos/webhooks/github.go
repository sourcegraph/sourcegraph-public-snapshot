package webhookapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

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

func CreateSyncWebhook(repoURL string, secret string, token string) error { // will need secret, token, client
	fmt.Println("Creating webhook:", repoURL)

	// HOW TO GENERATE THE SECRET
	// EXTRACTING THE TOKEN FROM THE USER

	// u := "https://api.github.com/repos/susantoscott/Task-Tracker/hooks"
	parts := strings.Split(repoURL, "/")
	serviceID := parts[0]
	owner := parts[1]
	repoName := parts[2]
	url := fmt.Sprintf("https://api.%s/repos/%s/%s/hooks", serviceID, owner, repoName)
	// fmt.Println("Url:", url)
	payload := Payload{
		Name:   "web",
		Active: true,
		Config: Config{
			Url:          "https://test01/webhooks", // the url will be to /enqueue-repo-update?
			Content_type: "json",
			Secret:       secret,
			Insecure_ssl: "0",
			Token:        token,
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
	req.Header.Add("Authorization", fmt.Sprintf("token %s", token))

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

func ListWebhooks(reponame string) []Payload {
	// fmt.Println("Listing webhooks...")

	// url := "https://api.github.com/repos/susantoscott/Task-Tracker/hooks"
	url := fmt.Sprintf("https://api.github.com/repos/susantoscott/%s/hooks", reponame)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer([]byte("")))
	if err != nil {
		fmt.Println("making new request error:", err)
	}
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	// req.Header.Add("Authorization", fmt.Sprintf("token %s", token))
	req.Header.Add("Authorization", "token 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")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("client do error:", err)
	}
	// fmt.Println("Status Code:", resp.StatusCode)

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("readall error:", err)
	}

	var obj []Payload
	if err := json.Unmarshal(respBody, &obj); err != nil {
		fmt.Println("unmarshal error:", err)
	}

	return obj
}

func DeleteWebhook(reponame string, hookID int) {
	url := fmt.Sprintf("https://api.github.com/repos/susantoscott/%s/hooks/%d", reponame, hookID)
	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer([]byte("")))
	if err != nil {
		fmt.Println("making new request error:", err)
	}
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	req.Header.Add("Authorization", "token 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")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("client do error:", err)
	}
	fmt.Println("Status Code:", resp.StatusCode)
}
