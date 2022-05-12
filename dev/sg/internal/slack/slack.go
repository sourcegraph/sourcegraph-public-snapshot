package slack

import (
	"context"

	"github.com/slack-go/slack"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/open"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/secrets"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var out = std.Out

type slackToken struct {
	Token string `json:"token"`
}

func NewClient(ctx context.Context) (*slack.Client, error) {
	token, err := retrieveToken(ctx)
	if err != nil {
		return nil, err
	}
	return slack.New(token), nil
}

// retrieveToken obtains a token either from the cached configuration or by asking the user for it.
func retrieveToken(ctx context.Context) (string, error) {
	sec, err := secrets.FromContext(ctx)
	if err != nil {
		return "", err
	}
	tok := slackToken{}
	err = sec.Get("slack", &tok)
	if errors.Is(err, secrets.ErrSecretNotFound) {
		str, err := getTokenFromUser()
		if err != nil {
			return "", nil
		}
		if err := sec.PutAndSave("slack", slackToken{Token: str}); err != nil {
			return "", err
		}
		return str, nil
	}
	if err != nil {
		return "", err
	}
	return tok.Token, nil
}

// getTokenFromUser prompts the user for a slack OAuth token.
func getTokenFromUser() (string, error) {
	out.WriteLine(output.Linef(output.EmojiLightbulb, output.StylePending, `Please copy the content of "SG Slack Integration" from the "Shared" 1Password vault`))
	return open.Prompt("Paste your token here:")
}
