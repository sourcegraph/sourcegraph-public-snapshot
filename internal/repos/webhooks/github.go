package githubwebhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Payload struct {
	Name   string   `json:"name"`
	ID     int      `json:"id,omitempty"`
	Config Config   `json:"config"`
	Events []string `json:"events"`
	Active bool     `json:"active"`
	URL    string   `json:"url"`
}

type Config struct {
	Url          string `json:"url"`
	Content_type string `json:"content_type"`
	Secret       string `json:"secret"`
	Insecure_ssl string `json:"insecure_ssl"`
	Token        string `json:"token"`
	Digest       string `json:"digest,omitempty"`
}

func CreateSyncWebhook(repoName string, secret string, token string) (int, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/hooks", repoName)

	payload := Payload{
		Name:   "web",
		Active: true,
		Config: Config{
			Url:          fmt.Sprintf("%s/enqueue-repo-update", repoupdater.DefaultClient.URL),
			Content_type: "json",
			Secret:       secret,
			Token:        token,
			Insecure_ssl: "0",
		},
		Events: []string{
			"push",
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return -1, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return -1, err
	}
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	req.Header.Add("Authorization", fmt.Sprintf("token %s", token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return -1, err
	}

	if resp.StatusCode >= 300 {
		return -1, errors.Newf("non-2XX status code: %d", resp.StatusCode)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return -1, err
	}

	var obj Payload
	if err := json.Unmarshal(respBody, &obj); err != nil {
		return -1, err
	}

	return obj.ID, nil
}

func ListSyncWebhooks(repoName string, token string) ([]Payload, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/hooks", repoName)

	req, err := http.NewRequest("GET", url, bytes.NewBuffer([]byte("")))
	if err != nil {
		fmt.Println("making new request error:", err)
	}
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	req.Header.Add("Authorization", fmt.Sprintf("token %s", token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 300 {
		return nil, errors.Newf("non-2XX status code: %d", resp.StatusCode)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var obj []Payload
	if err := json.Unmarshal(respBody, &obj); err != nil {
		return nil, err
	}

	return obj, nil
}

func FindSyncWebhook(repoName string, token string) bool {
	payloads, err := ListSyncWebhooks(repoName, token)
	if err != nil {
		return false
	}

	for _, payload := range payloads {
		endpoint := payload.Config.Url
		parts := strings.Split(endpoint, "/")
		if parts[len(parts)-1] == "enqueue-repo-update" {
			return true
		}
	}

	return false
}

func DeleteSyncWebhook(repoName string, hookID int, token string) (bool, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/hooks/%d", repoName, hookID)

	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer([]byte("")))
	if err != nil {
		return false, err
	}
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	req.Header.Add("Authorization", fmt.Sprintf("token %s", token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}

	if resp.StatusCode >= 300 {
		return false, errors.Newf("non-2XX status code: %d", resp.StatusCode)
	}

	return true, nil
}
