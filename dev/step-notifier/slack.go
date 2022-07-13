package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"text/template"
)

type SlackWebhookClient struct {
	webhook string
	http.Client
}

func NewSlackWebhookClient() *SlackWebhookClient {
	url := os.Getenv("SLACK_WEBHOOK")
	if url == "" {
		panic("SLACK_WEBHOOK cannot be empty")
	}

	return &SlackWebhookClient{
		webhook: url,
	}
}

func (c *SlackWebhookClient) sendNotification(build *Build) error {
	payload, err := createSlackJSON(build)
	if err != nil {
		return err
	}

	buf := bytes.NewBufferString(payload)
	resp, err := c.Post(c.webhook, "application/json", buf)
	if err != nil {
		log.Printf("failed to send notification for build %d: %v", *build.Number, err)
	}

	if resp.StatusCode != http.StatusOK {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("failed to read response body: %v", err)
		}
		body := string(data)
		log.Printf("failed to send notification. Status Code %d\nBody:%s", resp.StatusCode, body)
	}

	return nil
}

func createSlackJSON(build *Build) (string, error) {
	failed := make([]string, 0)
	for _, j := range build.Jobs {
		if j.ExitStatus != nil && *j.ExitStatus != 0 && !j.SoftFailed {
			failed = append(failed, *j.Name)
		}
	}
	data := struct {
		BuildNumber int
		FailedSteps []string
		Author      string
		BuildURL    string
		GrafanaURL  string
		FlakeURL    string
	}{
		*build.Number,
		failed,
		build.Author.Name,
		"something",
		"something else",
		"More",
	}
	tmplt := template.Must(template.New("SG").Parse(`
    {
        "blocks": [
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": "Build {{.BuildNumber}} failure"
            }
        },
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": "The following steps failed:{{range .FailedSteps}}\n{{.}}{{end}}"
            }
        },
        {
            "type": "actions",
            "elements": [
            {
                "type": "button",
                "text": {
                    "type": "plain_text",
                    "text": "View Build",
                    "emoji": true
                },
                "value": "{{.BuildURL}}"
            },
            {
                "type": "button",
                "text": {
                    "type": "plain_text",
                    "text": "View Grafana Logs",
                    "emoji": true
                },
                "value": "{{.GrafanaURL}}"
            },
            {
                "type": "button",
                "text": {
                    "type": "plain_text",
                    "text": "Is this a flake?",
                    "emoji": true
                },
                "value": "{{.FlakeURL}}"
            },
            ]
        }
        ]
    }
    `))

	var buf bytes.Buffer
	err := tmplt.Execute(&buf, data)
	if err != nil {
		return "", nil
	}

	return buf.String(), nil
}
