package slack

import (
	"context"
	"os"

	"github.com/slack-go/slack"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/secrets"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type slackToken struct {
	Token string `json:"token"`
}

func NewClient(ctx context.Context, out *std.Output) (*slack.Client, error) {
	token, err := retrieveToken(ctx, out)
	if err != nil {
		return nil, err
	}
	return slack.New(token), nil
}

// retrieveToken obtains a token either from the cached configuration or by asking the user for it.
func retrieveToken(ctx context.Context, out *std.Output) (string, error) {
	sec, err := secrets.FromContext(ctx)
	if err != nil {
		return "", err
	}
	tok := slackToken{}
	err = sec.Get("slack", &tok)
	if errors.Is(err, secrets.ErrSecretNotFound) {
		str, err := out.PromptPasswordf(os.Stdin, `Please copy the content of "SG Slack Integration" from the "Shared" 1Password vault:`)
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
