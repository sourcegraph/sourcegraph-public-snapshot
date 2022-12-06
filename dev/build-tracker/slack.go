package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v41/github"
	"github.com/slack-go/slack"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/dev/team"
)

const JobShowLimit = 10

type NotificationClient struct {
	slack   slack.Client
	team    team.TeammateResolver
	logger  log.Logger
	channel string
}

func NewNotificationClient(logger log.Logger, slackToken, githubToken, channel string) *NotificationClient {
	debug := os.Getenv("BUILD_TRACKER_SLACK_DEBUG") == "1"
	slack := slack.New(slackToken, slack.OptionDebug(debug))

	httpClient := http.Client{
		Timeout: 5 * time.Second,
	}
	githubClient := github.NewClient(&httpClient)
	teamResolver := team.NewTeammateResolver(githubClient, slack)

	return &NotificationClient{
		logger:  logger.Scoped("notificationClient", "client which interacts with Slack and Github to send notifications"),
		slack:   *slack,
		team:    teamResolver,
		channel: channel,
	}
}

func (c *NotificationClient) getTeammateForBuild(build *Build) (*team.Teammate, error) {
	return c.team.ResolveByCommitAuthor(context.Background(), "sourcegraph", "sourcegraph", build.commit())
}

func (c *NotificationClient) sendFailedBuild(build *Build) error {
	logger := c.logger.With(log.Int("buildNumber", build.number()), log.String("channel", c.channel))
	logger.Debug("creating slack json")

	blocks, err := c.createMessageBlocks(logger, build)
	if err != nil {
		return err
	}

	logger.Debug("sending notification")
	msgOptBlocks := slack.MsgOptionBlocks(blocks...)
	_, _, err = c.slack.PostMessage(c.channel, msgOptBlocks)
	if err != nil {
		logger.Error("failed to post message", log.Error(err))
		return err
	}

	logger.Info("notification posted")
	return nil
}

func commitLink(msg, commit string) string {
	repo := "http://github.com/sourcegraph/sourcegraph"
	sgURL := fmt.Sprintf("%s/commit/%s", repo, commit)
	return fmt.Sprintf("<%s|%s>", sgURL, msg)
}

func slackMention(teammate *team.Teammate) string {
	if teammate.SlackID == "" {
		return fmt.Sprintf("%s (%s) - We could not locate your Slack ID. Please check that your information in the Handbook team.yml file is correct", teammate.Name, teammate.Email)
	}
	return fmt.Sprintf("<@%s>", teammate.SlackID)
}

func (c *NotificationClient) createMessageBlocks(logger log.Logger, build *Build) ([]slack.Block, error) {
	msg, _, _ := strings.Cut(build.message(), "\n")
	msg += fmt.Sprintf(" (%s)", build.commit()[:7])

	failedSection := fmt.Sprintf("> %s\n\n", commitLink(msg, build.commit()))

	// create a bulleted list of all the failed jobs
	//
	// if there are more than JobShowLimit of failed jobs, we cannot print all of it
	// since the message will to big and slack will reject the message with "invalid_blocks"
	failedJobs := build.failedJobs()
	jobSection := "*Failed jobs:*\n\n"
	if len(failedJobs) > JobShowLimit {
		jobSection = fmt.Sprintf("* %d Failed jobs (showing %d):*\n\n", len(failedJobs), JobShowLimit)
	}
	logger.Info("failed job count on build", log.Int("failedJobs", len(failedJobs)))
	for i := 0; i < JobShowLimit && i < len(failedJobs); i++ {
		j := failedJobs[i]
		jobSection += fmt.Sprintf("• %s", *j.Name)
		if j.WebURL != "" {
			jobSection += fmt.Sprintf(" - <%s|logs>", j.WebURL)
		}
		jobSection += "\n"
	}

	failedSection += jobSection

	logger.Debug("getting teammate information using commit", log.String("commit", build.commit()))
	teammate, err := c.getTeammateForBuild(build)
	var author string
	if err != nil {
		c.logger.Error("failed to find teammate", log.Error(err))
		// the error has some guidance on how to fix it so that teammate resolver can figure out who you are from the commit!
		// so we set author here to that msg, so that the message can be conveyed to the person in slack
		author = err.Error()
	} else {
		logger.Debug("teammate found", log.Object("teammate",
			log.String("slackID", teammate.SlackID),
			log.String("key", teammate.Key),
			log.String("email", teammate.Email),
			log.String("handbook", teammate.HandbookLink),
			log.String("slackName", teammate.SlackName),
			log.String("github", teammate.GitHub),
		))
		author = slackMention(teammate)
	}

	blocks := []slack.Block{
		slack.NewHeaderBlock(
			slack.NewTextBlockObject(slack.PlainTextType, generateSlackHeader(build), true, false),
		),
		slack.NewSectionBlock(&slack.TextBlockObject{Type: slack.MarkdownType, Text: failedSection}, nil, nil),
		slack.NewSectionBlock(
			nil,
			[]*slack.TextBlockObject{
				{Type: slack.MarkdownType, Text: fmt.Sprintf("*Author:* %s", author)},
				{Type: slack.MarkdownType, Text: fmt.Sprintf("*Pipeline:* %s", build.Pipeline.name())},
			},
			nil,
		),
		slack.NewActionBlock(
			"",
			[]slack.BlockElement{
				&slack.ButtonBlockElement{
					Type:  slack.METButton,
					Style: slack.StylePrimary,
					URL:   *build.WebURL,
					Text:  &slack.TextBlockObject{Type: slack.PlainTextType, Text: "Go to build"},
				},
				&slack.ButtonBlockElement{
					Type: slack.METButton,
					URL:  "https://www.loom.com/share/58cedf44d44c45a292f650ddd3547337",
					Text: &slack.TextBlockObject{Type: slack.PlainTextType, Text: "Is this a flake?"},
				},
			}...,
		),

		&slack.DividerBlock{Type: slack.MBTDivider},

		slack.NewSectionBlock(
			&slack.TextBlockObject{
				Type: slack.MarkdownType,
				Text: `:books: *More information on flakes*
• <https://docs.sourcegraph.com/dev/background-information/ci#flakes|How to disable flakey tests>
• <https://docs.sourcegraph.com/dev/how-to/testing#assessing-flaky-client-steps|Recognizing flakey client steps and how to fix them>

_Disable flakes on sight and save your fellow teammate some time!_`,
			},
			nil,
			nil,
		),
	}

	return blocks, nil
}

func generateSlackHeader(build *Build) string {
	header := fmt.Sprintf(":red_circle: Build %d failed", build.number())
	switch build.ConsecutiveFailure {
	case 0, 1: // no suffix
	case 2:
		header += " (2nd failure)"
	case 3:
		header += " (:exclamation: 3rd failure)"
	default:
		header += fmt.Sprintf(" (:bangbang: %dth failure)", build.ConsecutiveFailure)
	}
	return header
}
