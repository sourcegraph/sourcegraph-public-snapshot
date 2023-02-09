package backend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type ClaudeRequest struct {
	Prompt            string   `json:"prompt"`
	StopSequences     []string `json:"stop_sequences"`
	MaxTokensToSample int      `json:"max_tokens_to_sample"`
	Model             string   `json:"model"`
}

type CompletionResponse struct {
	Completion string `json:"completion"`
	Stop       string `json:"stop,omitempty"`
	StopReason string `json:"stop_reason,omitempty"`
	Truncated  bool   `json:"truncated"`
	Exception  string `json:"exception"`
	LogID      string `json:"log_id"`
}

const (
	HumanPrompt string = "\n\nHuman:"
	AIPrompt    string = "\n\nAssistant:"
	ClientID    string = "anthropic-typescript/0.4.0"
)

// checkLabel invokes Claude asking whether contents of a given file
func checkLabel(fileContents, label string) (bool, error) {
	posturl := "https://api.anthropic.com/v1/complete"
	filePrompt := fmt.Sprintf("%sHere is the text of a source file:\n```\n%s\n```%sOK. Adding this to the context.", HumanPrompt, string(fileContents), AIPrompt)
	prompt := fmt.Sprintf("%sDoes understanding the source file require %s expertise? Please answer with a 'Yes' or 'No'.%s", HumanPrompt, label, AIPrompt)
	request := &ClaudeRequest{
		Prompt:            filePrompt + prompt,
		StopSequences:     []string{HumanPrompt},
		MaxTokensToSample: 200,
		Model:             "claude-v1",
	}
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(request); err != nil {
		return false, errors.Wrap(err, "Cannot encode request body.")
	}
	r, err := http.NewRequest("POST", posturl, &body)
	if err != nil {
		panic(err)
	}
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return false, errors.New("ANTHROPIC_API_KEY env var missing")
	}
	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("Accept", "application/json")
	r.Header.Add("Client", ClientID)
	r.Header.Add("X-API-Key", apiKey)
	client := &http.Client{}
	res, err := client.Do(r)
	if err != nil {
		return false, errors.Wrap(err, "Error communiacting with Claude")
	}
	defer res.Body.Close()
	var response CompletionResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return false, errors.Wrap(err, "Cannot decode response from Claude")
	}
	if got, want := res.StatusCode, http.StatusOK; got != want {
		return false, errors.Newf("Response HTTP code, got %d, want, %d:\n%v", got, want, response)
	}
	completion := " " + strings.ToLower(response.Completion) + " "
	if strings.Contains(completion, " yes ") {
		return true, nil
	}
	return false, nil
}

func TestAskCody(t *testing.T) {
	fmt.Println(os.Getwd())
	path := os.Getenv("FILE_PATH")
	if path == "" {
		t.Fatalf("Please set FILE_PATH env var")
	}
	fileContents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Could not read %q: %s", path, err)
	}
	requiresSecurity, err := checkLabel(string(fileContents), "security")
	if err != nil {
		t.Fatal(err)
	}
	if requiresSecurity {
		t.Error("File requires security expertise")
	} else {
		t.Error("File does not require security expertise")
	}
}
