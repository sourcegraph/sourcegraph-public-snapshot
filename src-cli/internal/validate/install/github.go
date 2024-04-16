package install

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/sourcegraph/src-cli/internal/api"
	"github.com/sourcegraph/src-cli/internal/validate"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const GITHUB = "GITHUB"

func validateGithub(ctx context.Context, client api.Client, config *ValidationSpec) (func(), error) {
	// validate external service
	log.Printf("%s validating external service", validate.EmojiFingerPointRight)

	srvId, err := addGithubExternalService(ctx, client, config.ExternalService)
	if err != nil {
		return nil, err
	}

	log.Printf("%s external service %s is being added", validate.HourglassEmoji, config.ExternalService.DisplayName)

	cleanupFunc := func() {
		if srvId != "" && config.ExternalService.DeleteWhenDone {
			_ = removeExternalService(ctx, client, srvId)
			log.Printf("%s external service %s has been removed", validate.SuccessEmoji, config.ExternalService.DisplayName)
		}
	}

	log.Printf("%s cloning repository", validate.HourglassEmoji)

	repo := fmt.Sprintf("github.com/%s", config.ExternalService.Config.GitHub.Repos[0])
	cloned, err := repoCloneTimeout(ctx, client, repo, config.ExternalService)
	if err != nil {
		return nil, err
	}
	if !cloned {
		return nil, errors.Newf("%s validate failed, repo did not clone\n", validate.FailureEmoji)
	}

	log.Printf("%s repositry successfully cloned", validate.SuccessEmoji)

	return cleanupFunc, nil
}

func addGithubExternalService(ctx context.Context, client api.Client, srv ExternalService) (string, error) {
	if srv.Config.GitHub.Token == "" {
		return "", errors.Newf("%s  failed to read `SRC_GITHUB_TOKEN` environment variable", validate.WarningSign)
	}

	cfg, err := json.Marshal(srv.Config.GitHub)
	if err != nil {
		return "", errors.Wrap(err, "addExternalService failed")
	}

	q := clientQuery{
		opName: "AddExternalService",
		query: `mutation AddExternalService($kind: ExternalServiceKind!, $displayName: String!, $config: String!) {
				addExternalService(input:{
					kind:$kind,
					displayName:$displayName,
					config: $config
		  		})
		  		{
					id
		  		}
		}`,
		variables: jsonVars{
			"kind":        GITHUB,
			"displayName": srv.DisplayName,
			"config":      string(cfg),
		},
	}

	var result struct {
		AddExternalService struct {
			ID string `json:"id"`
		} `json:"addExternalService"`
	}

	ok, err := client.NewRequest(q.query, q.variables).Do(ctx, &result)
	if err != nil {
		return "", errors.Wrap(err, "addExternalService failed")
	}
	if !ok {
		return "", errors.New("addExternalService failed, no data to unmarshal")
	}

	return result.AddExternalService.ID, nil
}
