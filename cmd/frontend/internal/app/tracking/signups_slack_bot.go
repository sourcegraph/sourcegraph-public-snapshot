package tracking

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"github.com/sourcegraph/go-github/github"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/gcstracker"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/hubspot"
)

// SlackWebhookURL is the Slack endpoint that receives signup messages
// and publishes them to #bot-signups
// https://sourcegraph.slack.com/services/B4UUZAXQQ
var SlackWebhookURL = env.Get("SLACK_SIGNUPS_BOT_HOOK", "", "Webhook for posting signup notifications to the Slack #bot-signups channel.")

type slackPayload struct {
	Attachments []*slackAttachment `json:"attachments"`
}
type slackAttachment struct {
	Fallback  string        `json:"fallback"`
	Color     string        `json:"color"`
	Title     string        `json:"title"`
	ThumbURL  string        `json:"thumb_url"`
	Fields    []*slackField `json:"fields"`
	Timestamp int64         `json:"ts"`
}
type slackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

func notifySlackOnSignup(actor *actor.Actor, hubSpotProps *hubspot.ContactProperties, response *hubspot.ContactResponse, tos *gcstracker.TrackedObjects) error {
	if SlackWebhookURL == "" {
		return errors.New("Slack Webhook URL not defined")
	}

	color := "good"
	if !hubSpotProps.IsPrivateCodeUser {
		color = "danger"
	}

	links := ""
	if response != nil {
		links = fmt.Sprintf("<%s|View on GitHub>, <%s|View on Looker>, <%s|View on HubSpot>", hubSpotProps.GitHubLink, hubSpotProps.LookerLink, fmt.Sprintf("https://app.hubspot.com/contacts/2762526/contact/%v", response.VID))
	} else {
		links = fmt.Sprintf("<%s|View on GitHub>, <%s|View on Looker>", hubSpotProps.GitHubLink, hubSpotProps.LookerLink)
	}

	payload := &slackPayload{
		Attachments: []*slackAttachment{
			&slackAttachment{
				Fallback: fmt.Sprintf("%s just signed up!", hubSpotProps.GitHubName),
				Title:    fmt.Sprintf("%s just signed up!", hubSpotProps.GitHubName),
				Color:    color,
				ThumbURL: actor.AvatarURL,
				Fields: []*slackField{
					&slackField{
						Title: "User login",
						Value: hubSpotProps.UserID,
						Short: true,
					},
					&slackField{
						Title: "GitHub Name",
						Value: hubSpotProps.GitHubName,
						Short: true,
					},
					&slackField{
						Title: "GitHub Email",
						Value: actor.Email,
						Short: true,
					},
					&slackField{
						Title: "GitHub Company",
						Value: hubSpotProps.GitHubCompany,
						Short: true,
					},
					&slackField{
						Title: "GitHub Location",
						Value: hubSpotProps.GitHubLocation,
						Short: true,
					},
					&slackField{
						Title: "Private code user?",
						Value: fmt.Sprintf("%v", hubSpotProps.IsPrivateCodeUser),
						Short: true,
					},
					&slackField{
						Title: "User profile links",
						Value: links,
						Short: false,
					},
				},
			},
		},
	}

	orgsList := make([]string, 0)
	for _, obj := range tos.Objects {
		if orgCtx, ok := obj.Ctx.(*gcstracker.OrgWithDetailsContext); ok {
			orgsList = append(orgsList, fmt.Sprintf("%s (<https://github.com/%s|GitHub>, <https://sourcegraph.looker.com/dashboards/11?Org%%20Name=%s|Looker>)", orgCtx.OrgName, orgCtx.OrgName, orgCtx.OrgName))
		} else {
			break
		}
	}

	if len(orgsList) > 0 {
		payload.Attachments[0].Fields = append(payload.Attachments[0].Fields, &slackField{
			Title: "GitHub organizations",
			Value: strings.Join(orgsList, ", "),
			Short: false,
		})
	}

	return postToSlack(payload)
}

func notifySlackOnAppInstall(senderLogin string, actorGitHubLink string, actorLookerLink string, org *github.User, orgGitHubLink string) error {
	if SlackWebhookURL == "" {
		return errors.New("Slack Webhook URL not defined")
	}
	color := "good"
	links := fmt.Sprintf("<%s|View user on GitHub>, <%s|View user on Looker>, <%s|View org on GitHub>", actorGitHubLink, actorLookerLink, orgGitHubLink)

	payload := &slackPayload{
		Attachments: []*slackAttachment{
			&slackAttachment{
				Fallback: fmt.Sprintf("%s just installed Sourcegraph on their org %s!", senderLogin, *org.Login),
				Title:    fmt.Sprintf("%s just installed Sourcegraph on their org %s!", senderLogin, *org.Login),
				Color:    color,
				ThumbURL: *org.AvatarURL,
				Fields: []*slackField{
					&slackField{
						Title: "User login",
						Value: senderLogin,
						Short: true,
					},
					&slackField{
						Title: "Org name",
						Value: *org.Login,
						Short: true,
					},
					&slackField{
						Title: "Links",
						Value: links,
						Short: false,
					},
				},
			},
		},
	}

	return postToSlack(payload)
}

func postToSlack(payload *slackPayload) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "tracking.postToSlack")
	}

	req, err := http.NewRequest("POST", SlackWebhookURL, strings.NewReader(string(payloadJSON)))
	if err != nil {
		return errors.Wrap(err, "tracking.postToSlack")
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "tracking.postToSlack")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(resp.Body)
		return errors.Wrap(fmt.Errorf("Code %v: %s", resp.StatusCode, buf.String()), "tracking.postToSlack")
	}
	return nil
}
