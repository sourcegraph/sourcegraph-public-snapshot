package main

import (
	"bytes"
	"encoding/json"
	"github.com/sourcegraph/log"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"text/template"
	"time"
)

type SlackWebhookClient struct {
	webhook string
	logger  log.Logger
	http.Client
}

func NewSlackWebhookClient(logger log.Logger) *SlackWebhookClient {
	url := os.Getenv("SLACK_WEBHOOK")
	if url == "" {
		panic("SLACK_WEBHOOK cannot be empty")
	}

	return &SlackWebhookClient{
		logger:  logger.Scoped("slack", "client which interacts with Slack's webhook API"),
		webhook: url,
	}
}

func (c *SlackWebhookClient) sendNotification(build *Build) error {
	c.logger.Debug("creating slack json", log.Int("buildNumber", *build.Number))
	payload, err := createSlackJSON(c.logger, build)
	if err != nil {
		return err
	}

	buf := bytes.NewBufferString(payload)
	c.logger.Debug("sending notification", log.Int("buildNumber", *build.Number))
	resp, err := c.Post(c.webhook, "application/json", buf)
	if err != nil {
		c.logger.Error("failed to send notification", log.Int("buildNumber", *build.Number), log.Error(err))
	}

	if resp.StatusCode != http.StatusOK {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			c.logger.Error("failed to read body of response for failed notification", log.Int("buildNumber", *build.Number), log.Int("status", resp.StatusCode), log.Error(err))
		}
		body := string(data)
		c.logger.Error("error response received for notification", log.Int("buildNumber", *build.Number), log.Int("status", resp.StatusCode), log.String("body", body))
		return err
	}

	c.logger.Info("notification posted", log.Int("buildNumber", *build.Number))
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

func grafanaURLFor(logger log.Logger, build *Build) string {
	logger = logger.Scoped("grafana", "generates grafana links")
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
		logger.Error("failed to execute template", log.String("template", "Expression"), log.Error(err))
	}
	var expression = buf.String()

	logger.Debug("---->", log.String("expression", expression))

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
		logger.Error("failed to marshall GrafanaPayload", log.Error(err))
	}

	logger.Debug("---->", log.String("GrafanaPayload", (string(result))))
	var values = make(url.Values)
	values.Add("orgId", "1")
	values.Add("left", string(result))

	base.RawQuery = values.Encode()
	logger.Debug("---->", log.String("url params", base.RawQuery))

	return base.String()
}

func createSlackJSON(logger log.Logger, build *Build) (string, error) {
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
		grafanaURLFor(logger, build),
		"TBA",
	}
	tmplt := template.Must(template.New("SG").Parse(`
    {
        "blocks": [
        {
            "type": "header",
            "text": {
                "type": "plain_test",
                "text": "Build {{.BuildNumber}} failure"
            }
        },
        {
            "type": "divider"
        },
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": "The following steps failed:"
            }
        },
        {
            "type": "context",
            "elements": [{
                "type": "section",
                "text": {
                    "type": "mrkdwn",
                    "text": "{{range .FailedSteps}}{{.}}\n{{end}}"
                }
            }],
        },
        {
            "type": "divider"
        },
        {
            "type": "actions",
            "elements": [
            {
                "type": "button",
                "style": "primary",
                "text": {
                    "type": "plain_text",
                    "text": "View Build",
                    "emoji": true
                },
                "url": "{{.BuildURL}}"
            },
            {
                "type": "button",
                "style": "default",
                "text": {
                    "type": "plain_text",
                    "text": "View Grafana Logs",
                    "emoji": true
                },
                "url": "{{.GrafanaURL}}"
            },
            {
                "type": "button",
                "style": "danger",
                "text": {
                    "type": "plain_text",
                    "text": "Is this a flake?",
                    "emoji": true
                },
                "url": "{{.FlakeURL}}"
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
