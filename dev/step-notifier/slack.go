package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"text/template"
	"time"
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

type GrafanaQuery struct {
	RefId string `json:"refId"`
	Expr  string `json:"expr"`
}

type GrafanaRange struct {
	From int64 `json:"From"`
	To   int64 `json:"To"`
}

type GrafanaPayload struct {
	DataSource string         `json:"datasource"`
	Queries    []GrafanaQuery `json:"queries"`
	Range      GrafanaRange   `json:"range"`
}

func grafanaURLFor(build *Build) string {
	base, _ := url.Parse("http://sourcegraph.grafana.net/explore")
	queryData := struct {
		Build int
	}{
		Build: *build.Number,
	}
	tmplt := template.Must(template.New("Expression").Parse(`{app="buildkite", build="{{.Build}}", branch="main", state="failed"} # to search the whole build remove job here!
    |~ "(?i)failed|panic" # this is a case insensitive regular expression, feel free to unleash your regex-fu!
    `))

	var buf bytes.Buffer
	if err := tmplt.Execute(&buf, queryData); err != nil {
		log.Printf("ERR failed to execute Expression template: %v", err)
	}
	var expression = buf.String()

	log.Println(expression)

	buf.Reset()
	begin := time.Now().Add(-(1 * time.Hour)).UnixNano()
	end := time.Now().Add(5 * time.Minute).UnixNano()

	data := GrafanaPayload{
		DataSource: "grafanacloud-sourcegraph-logs",
		Queries: []GrafanaQuery{
			{
				RefId: "A",
				Expr:  expression,
			},
		},
		Range: GrafanaRange{
			From: begin,
			To:   end,
		},
	}

	result, err := json.Marshal(data)
	if err != nil {
		log.Printf("failed to marshall GrafanaPayload: %v", err)
	}

	log.Println(string(result))
	var values = make(url.Values)
	values.Add("orgId", "1")
	values.Add("left", string(result))

	base.RawQuery = values.Encode()

	return base.String()
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
		*build.WebURL,
		grafanaURLFor(build),
		"TBA",
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
