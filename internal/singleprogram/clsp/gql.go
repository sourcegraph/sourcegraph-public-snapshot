package clsp

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type CompletionsArgs struct {
	Input CompletionsInput
	Fast  bool
}

type Message struct {
	Speaker string `json:"speaker"`
	Text    string `json:"text"`
}

type CompletionsInput struct {
	Messages          []Message `json:"messages"`
	Temperature       float64   `json:"temperature"`
	MaxTokensToSample int32     `json:"maxTokensToSample"`
	TopK              int32     `json:"topK"`
	TopP              int32     `json:"topP"`
}

func codyCompletions(ctx context.Context, human, assistant string, fast bool) (string, error) {
	queryInput := CompletionsInput{
		Messages: []Message{
			{
				Speaker: "HUMAN",
				Text:    human,
			},
			{
				Speaker: "ASSISTANT",
				Text:    assistant,
			},
		},
		Temperature:       0.5,
		MaxTokensToSample: 300,
		TopK:              -1,
		TopP:              -1,
	}

	queryPayload, err := json.Marshal(struct {
		OperationName string      `json:"operationName"`
		Variables     interface{} `json:"variables"`
		Query         string      `json:"query"`
	}{
		OperationName: "Completions",
		Query:         "query Completions($input: CompletionsInput!, $fast: Boolean!) { completions(input: $input, fast: $fast) }",
		Variables:     map[string]interface{}{"input": queryInput, "fast": fast},
	})
	if err != nil {
		return "", errors.Wrap(err, "Marshal")
	}

	url, err := gqlURL("Completions")
	if err != nil {
		return "", err
	}
	cli := httpcli.InternalDoer
	payload := bytes.NewReader(queryPayload)
	req, err := http.NewRequestWithContext(ctx, "POST", url, payload)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "token "+conf.Get().App.DotcomAuthToken)
	resp, err := cli.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.Newf("request failed with status: %v", resp.StatusCode)
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "ReadAll")
	}

	var v struct {
		Data struct {
			Completions string
		}
		Errors []any
	}
	if err := json.Unmarshal(respBody, &v); err != nil {
		return "", errors.Wrap(err, "Unmarshal")
	}
	if len(v.Errors) > 0 {
		return "", errors.Errorf("graphql: errors: %v", v.Errors)
	}
	return assistant + v.Data.Completions, nil
}

// gqlURL returns the frontend's internal GraphQL API URL, with the given ?queryName parameter
// which is used to keep track of the source and type of GraphQL queries.
func gqlURL(queryName string) (string, error) {
	u, err := url.Parse("https://sourcegraph.com")
	if err != nil {
		return "", err
	}
	u.Path = "/.api/graphql"
	u.RawQuery = queryName
	return u.String(), nil
}
