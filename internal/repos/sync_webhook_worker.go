package repos

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

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
	Config Config   `json:"config"`
	Events []string `json:"events"`
	Active bool     `json:"active"`
}

func CreateSyncWebhook(repoURL string) (interface{}, error) {
	// fmt.Println("Creating webhook...")

	// u := "https://api.github.com/repos/susantoscott/Task-Tracker/hooks"
	parts := strings.Split(repoURL, "/")
	serviceID := parts[0]
	owner := parts[1]
	repoName := parts[2]
	url := fmt.Sprintf("https://api.%s/repos/%s/%s/hooks", serviceID, owner, repoName)
	fmt.Println("Url:", url)
	payload := Payload{
		Name:   "web",
		Active: true,
		Config: Config{
			Url:          "https://b80d-101-128-118-134777.ap.ngrok.io/webhooks",
			Content_type: "json",
			Secret:       "secret",
			Insecure_ssl: "0",
			Token:        "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",
			Digest:       "",
		},
		Events: []string{
			"push",
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	req.Header.Add("Authorization", "token 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")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	fmt.Println("RespBody:", string(respBody))
	PrettyPrint(respBody)

	if resp.StatusCode >= 300 {
		fmt.Println("STATUS CODE:", resp.StatusCode)
		return nil, errors.Newf("non-200 status code, %s", err)
	}

	var obj Payload
	if err := json.Unmarshal(respBody, &obj); err != nil {
		return nil, err
	}

	return obj, nil
}

func ListWebhooks() interface{} {
	fmt.Println("Listing webhooks...")

	url := "https://api.github.com/repos/susantoscott/Task-Tracker/hooks"
	req, err := http.NewRequest("GET", url, bytes.NewBuffer([]byte("")))
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

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("readall error:", err)
	}

	var obj []Payload
	if err := json.Unmarshal(respBody, &obj); err != nil {
		fmt.Println("unmarshal error:", err)
	}

	return obj[0]
}

func PrettyPrint(data interface{}) string {
	val, err := json.MarshalIndent(data, "", "   ")
	if err != nil {
		fmt.Println("pretty printing error:", err)
		return ""
	}
	return string(val)
}
