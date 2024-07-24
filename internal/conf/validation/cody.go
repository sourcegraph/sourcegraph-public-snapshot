package validation

import (
	"fmt"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
)

func init() {
	contributeWarning(func(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
		if c.SiteConfig().CodyRestrictUsersFeatureFlag != nil {
			problems = append(problems, conf.NewSiteProblem("cody.restrictUsersFeatureFlag has been deprecated. Please remove it from your site config and use cody.permissions instead: https://sourcegraph.com/docs/cody/overview/enable-cody-enterprise#enable-cody-only-for-some-users"))
		}
		return
	})
	conf.ContributeValidator(completionsConfigValidator)
	conf.ContributeValidator(embeddingsConfigValidator)
}

const bedrockArnMessageTemplate = "completions.%s is invalid. Provisioned Capacity IDs must be formatted like \"model_id/provisioned_capacity_arn\".\nFor example \"anthropic.claude-instant-v1/%s\""

func completionsConfigValidator(q conftypes.SiteConfigQuerier) conf.Problems {
	problems := []string{}
	completionsConf := q.SiteConfig().Completions
	if completionsConf == nil {
		return nil
	}

	if completionsConf.Enabled != nil && q.SiteConfig().CodyEnabled == nil {
		problems = append(problems, "'completions.enabled' has been superceded by 'cody.enabled', please migrate to the new configuration.")
	}

	if completionsConf.Provider == "azure-openai" && (completionsConf.AzureChatModel == "" || completionsConf.AzureCompletionModel == "") {
		problems = append(problems, `when using azure-openai provider its mandatory to set both completions.azureChatModel and completions.azureCompletionModel for proper LLM Token usage`)
	}

	// Check for bedrock Provisioned Capacity ARNs which should instead be
	// formatted like:
	// "anthropic.claude-v2/arn:aws:bedrock:us-west-2:012345678901:provisioned-model/xxxxxxxx"
	if completionsConf.Provider == string(conftypes.CompletionsProviderNameAWSBedrock) {
		type modelID struct {
			value string
			field string
		}
		allModelIds := []modelID{
			{value: completionsConf.ChatModel, field: "chatModel"},
			{value: completionsConf.FastChatModel, field: "fastChatModel"},
			{value: completionsConf.CompletionModel, field: "completionModel"},
		}
		var modelIdsToCheck []modelID
		for _, modelId := range allModelIds {
			if modelId.value != "" {
				modelIdsToCheck = append(modelIdsToCheck, modelId)
			}
		}

		for _, modelId := range modelIdsToCheck {
			// When using provisioned capacity we expect an admin would just put the ARN
			// here directly, but we need both the model AND the ARN. Hence the check.
			if strings.HasPrefix(modelId.value, "arn:aws:") {
				problems = append(problems, fmt.Sprintf(bedrockArnMessageTemplate, modelId.field, modelId.value))
			}
		}
	}

	if len(problems) > 0 {
		return conf.NewSiteProblems(problems...)
	}

	return nil
}

func embeddingsConfigValidator(q conftypes.SiteConfigQuerier) conf.Problems {
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

	if evaluatedConfig := conf.GetEmbeddingsConfig(q.SiteConfig()); evaluatedConfig != nil {
		if evaluatedConfig.Dimensions <= 0 {
			problems = append(problems, "Could not set a default \"embeddings.dimensions\", please configure one manually")
		}
	}

	if len(problems) > 0 {
		return conf.NewSiteProblems(problems...)
	}

	return nil
}
