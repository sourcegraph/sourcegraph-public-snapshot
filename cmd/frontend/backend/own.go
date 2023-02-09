package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/own/codeowners"
	codeownerspb "github.com/sourcegraph/sourcegraph/internal/own/codeowners/v1"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type DomainExpertise struct {
	Why string
	Who string
}

// OwnService gives access to code ownership data.
// At this point only data from CODEOWNERS file is presented, if available.
type OwnService interface {
	// OwnersFile returns a CODEOWNERS file from a given repository at given commit ID.
	// In the case the file cannot be found, `nil` `*codeownerspb.File` and `nil` `error` is returned.
	OwnersFile(context.Context, api.RepoName, api.CommitID) (*codeownerspb.File, error)
	LabelledExpertise(ctx context.Context, repoName api.RepoName, commitID api.CommitID, filePath string) (map[string]DomainExpertise, error)
}

var _ OwnService = ownService{}

func NewOwnService(g gitserver.Client, db database.DB) OwnService {
	return ownService{gitserverClient: g, db: db}
}

type ownService struct {
	gitserverClient gitserver.Client
	db              database.DB
}

// codeownersLocations contains the locations where CODEOWNERS file
// is expected to be found relative to the repository root directory.
// These are in line with GitHub and GitLab documentation.
// https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-code-owners
// https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-code-owners
var codeownersLocations = []string{
	"CODEOWNERS",
	".github/CODEOWNERS",
	".gitlab/CODEOWNERS",
	"docs/CODEOWNERS",
}

// OwnersFile makes a best effort attempt to return a CODEOWNERS file from one of
// the possible codeownersLocations. It returns nil if no match is found.
func (s ownService) OwnersFile(ctx context.Context, repoName api.RepoName, commitID api.CommitID) (*codeownerspb.File, error) {
	for _, path := range codeownersLocations {
		content, err := s.gitserverClient.ReadFile(
			ctx,
			authz.DefaultSubRepoPermsChecker,
			repoName,
			commitID,
			path,
		)
		if content != nil && err == nil {
			return codeowners.Parse(bytes.NewReader(content))
		} else if os.IsNotExist(err) {
			continue
		}
		return nil, err
	}
	return nil, nil
}

var expertiseLabels = []string{"security"}

// LabelledExpertise evaluates what expertise is needed for given file.
// Then it looks for people who have that expertise.
func (s ownService) LabelledExpertise(ctx context.Context, repoName api.RepoName, commitID api.CommitID, filePath string) (map[string]DomainExpertise, error) {
	fileContents, err := s.gitserverClient.ReadFile(ctx, authz.DefaultSubRepoPermsChecker, repoName, commitID, filePath)
	if err != nil {
		return nil, err
	}
	reasonByLabel := map[string]string{}
	for _, l := range expertiseLabels {
		reason, err := CheckLabel(string(fileContents), l)
		if err != nil {
			return nil, err
		}
		if reason != nil {
			reasonByLabel[l] = *reason
		}
	}
	expertByLabel := map[string]DomainExpertise{}
	for l, reason := range reasonByLabel {
		m, err := s.db.Ownerships().FetchExperts(ctx, l)
		if err != nil {
			return nil, err
		}
		var max int
		var topExpert string
		for person, score := range m {
			if score > max {
				topExpert = person
				max = score
			}
		}
		if topExpert != "" {
			expertByLabel[l] = DomainExpertise{Who: topExpert, Why: reason}
		}
	}
	return expertByLabel, nil
}

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
func CheckLabel(fileContents, label string) (*string, error) {
	posturl := "https://api.anthropic.com/v1/complete"
	filePrompt := fmt.Sprintf("%sHere is the text of a source file:\n```\n%s\n```%sOK. Adding this to the context.", HumanPrompt, string(fileContents), AIPrompt)
	prompt := fmt.Sprintf("%sDoes understanding the source file require knowing a little bit about %s? Please answer with a 'Yes' or 'No'.%s", HumanPrompt, label, AIPrompt)
	request := &ClaudeRequest{
		Prompt:            filePrompt + prompt,
		StopSequences:     []string{HumanPrompt},
		MaxTokensToSample: 200,
		Model:             "claude-v1",
	}
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(request); err != nil {
		return nil, errors.Wrap(err, "Cannot encode request body.")
	}
	r, err := http.NewRequest("POST", posturl, &body)
	if err != nil {
		panic(err)
	}
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, errors.New("ANTHROPIC_API_KEY env var missing")
	}
	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("Accept", "application/json")
	r.Header.Add("Client", ClientID)
	r.Header.Add("X-API-Key", apiKey)
	client := &http.Client{}
	res, err := client.Do(r)
	if err != nil {
		return nil, errors.Wrap(err, "Error communiacting with Claude")
	}
	defer res.Body.Close()
	var response CompletionResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, errors.Wrap(err, "Cannot decode response from Claude")
	}
	if got, want := res.StatusCode, http.StatusOK; got != want {
		return nil, errors.Newf("Response HTTP code, got %d, want, %d:\n%v", got, want, response)
	}
	fmt.Println("CLAUDE", response.Completion)
	completion := " " + strings.ToLower(response.Completion) + " "
	completion = strings.ReplaceAll(completion, ",", " ")
	completion = strings.ReplaceAll(completion, ".", " ")
	completion = strings.TrimSpace(completion)
	yesPrefix := "yes"
	i := strings.Index(completion, yesPrefix)
	if i == -1 {
		return nil, nil
	}
	c := response.Completion[i+len(yesPrefix)+1:]
	c = strings.TrimLeft(c, " ,.")
	if c != "" {
		c = strings.ToUpper(c[0:1]) + c[1:]
	}
	return &c, nil
}
