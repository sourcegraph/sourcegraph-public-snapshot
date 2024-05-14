package install

import (
	"context"
	"fmt"
	"time"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/src-cli/internal/api"
	"github.com/sourcegraph/sourcegraph/src-cli/internal/validate"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type jsonVars map[string]interface{}

type clientQuery struct {
	opName    string
	query     string
	variables jsonVars
}

// Validate runs a series of validation checks such as cloning a repository, running search queries, and
// creating insights, based on the configuration provided.
func Validate(ctx context.Context, client api.Client, config *ValidationSpec) error {
	switch config.ExternalService.Kind {
	case GITHUB:
		cleanup, err := validateGithub(ctx, client, config)
		if err != nil {
			return err
		}
		defer cleanup()
	}

	logger := log.Scoped("ValidateGithubInstall")

	// run search queries
	if config.SearchQuery != nil {
		logger.Info(fmt.Sprintf("%s validating search queries", validate.EmojiFingerPointRight))

		for i := 0; i < len(config.SearchQuery); i++ {
			matchCount, err := searchMatchCount(ctx, client, config.SearchQuery[i])
			if err != nil {
				return err
			}
			if matchCount == 0 {
				return errors.Newf("validate failed, search query %s returned no results", config.SearchQuery[i])
			}
			logger.Info(fmt.Sprintf("%s search query '%s' was successful", validate.SuccessEmoji, config.SearchQuery[i]))
		}
	}

	// run executor queries
	if config.Executor.Enabled {
		logger.Info(fmt.Sprintf("%s validating executor connections", validate.EmojiFingerPointRight))

		executorQuery := `query executors($query: String, $active: Boolean, $first: Int, $after: String) {
						executors(query: $query, active: $active, first: $first, after: $after){
							totalCount
						}
					}`
		executorVars := map[string]interface{}{
			"query":  "",
			"active": true,
			"first":  100,
			"after":  "",
		}

		totalCount, err := checkExecutors(ctx, client, executorQuery, executorVars)
		if err != nil {
			return err
		}
		if totalCount == 0 {
			logger.Error(fmt.Sprintf("%s validation failed, 0 executors found", validate.FlashingLightEmoji))
		}
		if totalCount >= 1 {
			logger.Info(fmt.Sprintf("%s executors found, %d executor(s) connected to Sourcegraph instance", validate.SuccessEmoji, totalCount))
		}
	}

	if config.Smtp.Enabled {
		logger.Info(fmt.Sprintf("%s validating smtp connection", validate.EmojiFingerPointRight))

		smtpQuery := `mutation sendTestEmail($to: String!) {
			sendTestEmail(to: $to)
		  }`
		smtpVars := map[string]interface{}{
			"to": config.Smtp.To,
		}

		result, err := checkSmtp(ctx, client, smtpQuery, smtpVars)
		if err != nil {
			return err
		}
		logger.Info(fmt.Sprintf("%s '%s'", validate.SuccessEmoji, result))
	}

	if config.Insight.Title != "" {
		logger.Info(fmt.Sprintf("%s validating code insight", validate.EmojiFingerPointRight))

		logger.Info(fmt.Sprintf("%s insight %s is being added", validate.HourglassEmoji, config.Insight.Title))

		insightId, err := createInsight(ctx, client, config.Insight)
		if err != nil {
			return err
		}

		logger.Info(fmt.Sprintf("%s insight successfully added", validate.SuccessEmoji))

		defer func() {
			if insightId != "" && config.Insight.DeleteWhenDone {
				_ = removeInsight(ctx, client, insightId)
				logger.Info(fmt.Sprintf("%s insight %s has been removed", validate.SuccessEmoji, config.Insight.Title))

			}
		}()
	}

	return nil
}

func checkExecutors(ctx context.Context, client api.Client, query string, variables map[string]interface{}) (int, error) {
	q := clientQuery{
		opName:    "CheckExecutorConnection",
		query:     query,
		variables: variables,
	}

	var result struct {
		Executor struct {
			TotalCount int `json:"totalCount"`
		} `json:"executors"`
	}

	ok, err := client.NewRequest(q.query, q.variables).Do(ctx, &result)
	if err != nil {
		return -1, errors.Wrap(err, "checkExecutors failed")
	}
	if !ok {
		return -1, errors.New("checkExecutors failed, no data to unmarshal")
	}

	return result.Executor.TotalCount, nil
}

func removeExternalService(ctx context.Context, client api.Client, id string) error {
	q := clientQuery{
		opName: "DeleteExternalService",
		query: `mutation DeleteExternalService($id: ID!) {
					deleteExternalService(externalService: $id){
					alwaysNil
					}
				}`,
		variables: jsonVars{
			"id": id,
		},
	}

	var result struct{}

	ok, err := client.NewRequest(q.query, q.variables).Do(ctx, &result)
	if err != nil {
		return errors.Wrap(err, "removeExternalService failed")
	}
	if !ok {
		return errors.New("removeExternalService failed, no data to unmarshal")
	}
	return nil
}

func searchMatchCount(ctx context.Context, client api.Client, searchExpr string) (int, error) {
	q := clientQuery{
		opName: "SearchMatchCount",
		query: `query ($query: String!) {
					search(query: $query, version: V2, patternType:literal){
						results {
							matchCount
						}
					}
				}`,
		variables: jsonVars{
			"query": searchExpr,
		},
	}

	var result struct {
		Search struct {
			Results struct {
				MatchCount int `json:"matchCount"`
			} `json:"results"`
		} `json:"search"`
	}

	ok, err := client.NewRequest(q.query, q.variables).Do(ctx, &result)
	if err != nil {
		return 0, errors.Wrap(err, "searchMatchCount failed")
	}
	if !ok {
		return 0, errors.New("searchMatchCount failed, no data to unmarshal")
	}

	return result.Search.Results.MatchCount, nil
}

func checkSmtp(ctx context.Context, client api.Client, query string, variables map[string]interface{}) (string, error) {
	q := clientQuery{
		opName:    "CheckSmtpConfig",
		query:     query,
		variables: variables,
	}

	var result struct {
		SendTestEmail string `json:"sendTestEmail"`
	}

	ok, err := client.NewRequest(q.query, q.variables).Do(ctx, &result)
	if err != nil {
		return "", errors.Wrap(err, "sendTestEmail failed")
	}
	if !ok {
		return "", errors.New("sendTestEmail failed, no data to unmarshal")
	}
	return result.SendTestEmail, nil
}

func repoCloneTimeout(ctx context.Context, client api.Client, repo string, srv ExternalService) (bool, error) {
	for i := 0; i < srv.MaxRetries; i++ {
		repos, err := listClonedRepos(ctx, client, []string{repo})
		if err != nil {
			return false, err
		}
		if len(repos) >= 1 {
			return true, nil
		}
		time.Sleep(time.Second * time.Duration(srv.RetryTimeoutSeconds))
	}
	return false, nil
}

func listClonedRepos(ctx context.Context, client api.Client, names []string) ([]string, error) {
	q := clientQuery{
		opName: "ListRepos",
		query: `query ListRepos($names: [String!], $first: Int) {
			  repositories(
				names: $names
				first: $first
			  ) {
				nodes {
				  name
				  mirrorInfo {
					 cloned
				  }
				}
			  }
			}`,
		variables: jsonVars{
			"names": names,
			"first": 5,
		},
	}

	var result struct {
		Repositories struct {
			Nodes []struct {
				Name       string `json:"name"`
				MirrorInfo struct {
					Cloned bool `json:"cloned"`
				} `json:"mirrorInfo"`
			} `json:"nodes"`
		} `json:"repositories"`
	}

	ok, err := client.NewRequest(q.query, q.variables).Do(ctx, &result)
	if err != nil {
		return nil, errors.Wrap(err, "listClonedRepos failed")
	}
	if !ok {
		return nil, errors.New("listClonedRepos failed, no data to unmarshal")
	}

	nodeNames := make([]string, 0, len(result.Repositories.Nodes))
	for _, node := range result.Repositories.Nodes {
		if node.MirrorInfo.Cloned {
			nodeNames = append(nodeNames, node.Name)
		}
	}

	return nodeNames, nil
}
