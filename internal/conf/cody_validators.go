package conf

import (
	"fmt"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
)

func init() {
	ContributeValidator(completionsConfigValidator)
	ContributeValidator(embeddingsConfigValidator)
}

const bedrockArnMessageTemplate = "completions.%s is invalid. Provisioned Capacity IDs must be formatted like \"model_id/provisioned_capacity_arn\".\nFor example \"anthropic.claude-instant-v1/%s\""

type modelId struct {
	value string
	field string
}

func completionsConfigValidator(q conftypes.SiteConfigQuerier) Problems {
	problems := []string{}
	completionsConf := q.SiteConfig().Completions
	if completionsConf == nil {
		return nil
	}

	if completionsConf.Enabled != nil && q.SiteConfig().CodyEnabled == nil {
		problems = append(problems, "'completions.enabled' has been superceded by 'cody.enabled', please migrate to the new configuration.")
	}

	allModelIds := []modelId{
		{value: completionsConf.ChatModel, field: "chatModel"},
		{value: completionsConf.FastChatModel, field: "fastChatModel"},
		{value: completionsConf.CompletionModel, field: "completionModel"},
	}
	var modelIdsToCheck []modelId
	for _, modelId := range allModelIds {
		if modelId.value != "" {
			modelIdsToCheck = append(modelIdsToCheck, modelId)
		}
	}

	//check for bedrock ARNs
	if completionsConf.Provider == string(conftypes.CompletionsProviderNameAWSBedrock) {
		for _, modelId := range modelIdsToCheck {
			if strings.HasPrefix(modelId.value, "arn:aws:") {
				problems = append(problems, fmt.Sprintf(bedrockArnMessageTemplate, modelId.field, modelId.value))
			}
		}
	}

	if len(problems) > 0 {
		return NewSiteProblems(problems...)
	}

	return nil
}

func embeddingsConfigValidator(q conftypes.SiteConfigQuerier) Problems {
	problems := []string{}
	embeddingsConf := q.SiteConfig().Embeddings
	if embeddingsConf == nil {
		return nil
	}

	if embeddingsConf.Provider == "" {
		if embeddingsConf.AccessToken != "" {
			problems = append(problems, "Because \"embeddings.accessToken\" is set, \"embeddings.provider\" is required")
		}
	}

	minimumIntervalString := embeddingsConf.MinimumInterval
	_, err := time.ParseDuration(minimumIntervalString)
	if err != nil && minimumIntervalString != "" {
		problems = append(problems, fmt.Sprintf("Could not parse \"embeddings.minimumInterval: %s\". %s", minimumIntervalString, err))
	}

	if evaluatedConfig := GetEmbeddingsConfig(q.SiteConfig()); evaluatedConfig != nil {
		if evaluatedConfig.Dimensions <= 0 {
			problems = append(problems, "Could not set a default \"embeddings.dimensions\", please configure one manually")
		}
	}

	if len(problems) > 0 {
		return NewSiteProblems(problems...)
	}

	return nil
}
