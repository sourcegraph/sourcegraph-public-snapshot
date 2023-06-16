package conf

import (
	"fmt"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
)

func init() {
	ContributeValidator(completionsConfigValidator)
	ContributeValidator(embeddingsConfigValidator)
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
