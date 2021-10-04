package slack

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/slack-go/slack"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/secrets"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var (
	ErrUserNotFound                = errors.New("User not found")
	out             *output.Output = stdout.Out
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
	out.WriteLine(output.Linef("", output.StyleWarning, "Please find the Slack OAuth Token in the 1Password vault named 'TODO'"))
	fmt.Printf("Paste it here: ")
	var token string
	if _, err := fmt.Scan(&token); err != nil {
		return "", err
	}
	return token, nil
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
		return "", errors.Wrapf(err, "cannot find nickname '%s'", nick)
	}
	loc, err := time.LoadLocation(u.TZ)
	if err != nil {
		return "", err
	}
	str := fmt.Sprintf("%s's current time is %s", nick, time.Now().In(loc).Format(time.RFC822))
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
		return "", errors.Wrapf(err, "cannot find nickname '%s'", nick)
	}
	p, err := api.GetUserProfile(&slack.GetUserProfileParameters{
		UserID:        u.ID,
		IncludeLabels: true,
	})
	if err != nil {
		panic(err)
	}
	for _, v := range p.FieldsMap() {
		if v.Label == "Handbook link" {
			return v.Value, nil
		}
	}
	return "", ErrUserNotFound
}

// findUserByNickname searches for a user by its nickname, e.g. what we type in Slack after a '@' character.
// TODO would be great to have some "did you mean" and use Levenshtein distance or something else to return
// a list of possible matches.
func findUserByNickname(users []slack.User, nickname string) *slack.User {
	nickname = strings.ToLower(nickname)
	for _, u := range users {
		if strings.ToLower(u.Profile.DisplayName) == nickname || strings.ToLower(u.Profile.RealName) == nickname {
			return &u
		}
	}
	return nil
}
