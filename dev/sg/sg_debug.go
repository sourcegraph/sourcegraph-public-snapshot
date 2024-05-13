package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/category"
	"github.com/sourcegraph/sourcegraph/internal/completions/client/awsbedrock"
	"github.com/sourcegraph/sourcegraph/internal/completions/tokenusage"
	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

var debugBedrockCommand = &cli.Command{
	Name:        "debug-bedrock",
	Usage:       "performs some debug invocations of bedrock models",
	Description: ``,
	Category:    category.Util,
	Action:      runDebugBedrockCommand,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "endpoint",
			Aliases:  []string{"e"},
			Usage:    "The endpoint",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "access-token",
			Aliases:  []string{"t"},
			Usage:    "key:secret",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "model",
			Aliases:  []string{"m"},
			Usage:    "model_arn",
			Required: true,
		},
	},
}

func runDebugBedrockCommand(cmd *cli.Context) error {
	endpoint := cmd.String("endpoint")
	accessToken := cmd.String("access-token")
	model := cmd.String("model")
	tokenManager := tokenusage.NewManager()

	client := awsbedrock.NewClient(httpcli.UncachedExternalDoer, endpoint, accessToken, *tokenManager)

	// resp, err := mockClient.Complete(context.Background(), types.CompletionsFeatureChat, types.CompletionsVersionLegacy, types.CompletionRequestParameters{}, logger)
	logger := log.Scoped("bedrock")

	logger.Info("Trying non-streaming complete")

	messages := []types.Message{
		{Speaker: "human", Text: "Hi how are you doing?"},
		// /complete prompts can have human messages without an assistant response. These should
		// be ignored.
		{Speaker: "assistant", Text: "I am doing good how are you."},
		{Speaker: "human", Text: "I am also good. Can you help me with something?"},
	}

	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	err := client.Stream(ctx, types.CompletionsFeatureChat, types.CompletionsV1, types.CompletionRequestParameters{
		Model:    model,
		Messages: messages,
	}, func(event types.CompletionResponse) error {
		logger.Info("Got completion event")
		completionJson, _ := json.MarshalIndent(event, "", "  ")
		logger.Info(string(completionJson))
		return nil
	}, logger)
	if err != nil {
		logger.Error("Error completing request")
		logger.Error(err.Error())
	}
	logger.Info("Completed request")

	return nil
}
