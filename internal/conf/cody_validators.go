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

	if completionsConf.Provider == "" {
		problems = append(problems, "completions.provider is required")
	}

	if completionsConf.Enabled != nil && q.SiteConfig().CodyEnabled == nil {
		problems = append(problems, "completions.enabled has been renamed to cody.enabled, please migrate")
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
		problems = append(problems, "embeddings.provider is required")
	}

	minimumIntervalString := embeddingsConf.MinimumInterval
	_, err := time.ParseDuration(minimumIntervalString)
	if err != nil && minimumIntervalString != "" {
		problems = append(problems, fmt.Sprintf("Could not parse \"embeddings.minimumInterval: %s\". %s", minimumIntervalString, err))
	}

	if len(problems) > 0 {
		return NewSiteProblems(problems...)
	}

	return nil
}
