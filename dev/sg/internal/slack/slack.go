package slack

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/slack-go/slack"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/open"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/secrets"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var (
	out *output.Output = stdout.Out
)

type slackToken struct {
	Token string `json:"token"`
}

// retrieveToken obtains a token either from the cached configuration or by asking the user for it.
func retrieveToken(ctx context.Context) (string, error) {
	sec := secrets.FromContext(ctx)
	tok := slackToken{}
	err := sec.Get("slack", &tok)
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

// QueryUserCurrentTime returns a given sourcegrapher current time, in its own timezone.
func QueryUserCurrentTime(ctx context.Context, nick string) (string, error) {
	token, err := retrieveToken(ctx)
	if err != nil {
		return "", err
	}
	return queryUserCurrentTime(token, nick)
}

func queryUserCurrentTime(token, nick string) (string, error) {
	// api := slack.New(token, slack.OptionDebug(true))
	api := slack.New(token)
	users, err := api.GetUsers()
	if err != nil {
		return "", err
	}
	u := findUserByNickname(users, nick)
	if u == nil {
		return "", fmt.Errorf("cannot find user with nickname '%s'", nick)
	}
	loc, err := time.LoadLocation(u.TZ)
	if err != nil {
		return "", err
	}
	t := time.Now().In(loc)
	t2 := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.Local)
	diff := t2.Sub(t) / time.Hour

	str := fmt.Sprintf("%s's current time is %s (%dh from your local time)", u.Profile.RealName, t.Format(time.RFC822), diff)
	return str, nil
}

// QueryUserHandbook returns a link to a given sourcegrapher handbook profile.
func QueryUserHandbook(ctx context.Context, nick string) (string, error) {
	token, err := retrieveToken(ctx)
	if err != nil {
		return "", err
	}
	return queryUserHandbook(token, nick)
}

func queryUserHandbook(token, nick string) (string, error) {
	// api := slack.New(token, slack.OptionDebug(true))
	api := slack.New(token)
	users, err := api.GetUsers()
	if err != nil {
		return "", err
	}
	u := findUserByNickname(users, nick)
	if u == nil {
		return "", fmt.Errorf("cannot find user with nickname '%s'", nick)
	}
	p, err := api.GetUserProfile(&slack.GetUserProfileParameters{
		UserID:        u.ID,
		IncludeLabels: true,
	})
	if err != nil {
		return "", err
	}
	for _, v := range p.FieldsMap() {
		if v.Label == "Handbook link" {
			return v.Value, nil
		}
	}
	return "", fmt.Errorf("no handbook link found for %s", nick)
}

// findUserByNickname searches for a user by its nickname, e.g. what we type in Slack after a '@' character.
func findUserByNickname(users []slack.User, nickname string) *slack.User {
	nickname = strings.ToLower(nickname)
	nickname = strings.TrimPrefix(nickname, "@")
	for _, u := range users {
		if strings.ToLower(u.Profile.DisplayName) == nickname || strings.ToLower(u.Profile.RealName) == nickname {
			return &u
		}
	}

	// No user found, try to guess now
	candidates := []slack.User{}
	for _, u := range users {
		if strings.HasPrefix(strings.ToLower(u.Profile.RealName), nickname) {
			candidates = append(candidates, u)
		}
	}
	if len(candidates) == 1 {
		return &candidates[0]
	}
	return nil
}
