package install

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/sourcegraph/src-cli/internal/api"
	"github.com/sourcegraph/src-cli/internal/validate"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type ExternalService struct {
	// Type of code host, e.g. GITHUB.
	Kind string `yaml:"kind"`

	// Display name of external service, e.g. sourcegraph-test.
	DisplayName string `yaml:"displayName"`

	// Configuration for code host.
	Config Config `yaml:"config"`

	// Maximum retry attempts when cloning test repositories. Defaults to 5 retries.
	MaxRetries int `yaml:"maxRetries"`

	// Retry timeout in seconds. Defaults to 5 seconds
	RetryTimeoutSeconds int `yaml:"retryTimeoutSeconds"`

	// Delete code host when test is done. Defaults to true.
	DeleteWhenDone bool `yaml:"deleteWhenDone"`
}

// Config for different types of code hosts.
type Config struct {
	GitHub GitHub `yaml:"gitHub"`
}

// GitHub configuration parameters.
type GitHub struct {
	// URL used to access your GitHub instance, e.g. https://github.com.
	URL string `yaml:"url" json:"url"`

	// Auth token used to authenticate to GitHub instance. This should be provided via env var SRC_GITHUB_TOKEN.
	Token string `yaml:"token" json:"token"`

	// List of organizations.
	Orgs []string `yaml:"orgs" json:"orgs"`

	// List of repositories to pull.
	Repos []string `yaml:"repos" json:"repos"`
}

type Insight struct {
	Title      string           `yaml:"title"`
	DataSeries []map[string]any `yaml:"dataSeries"`
}

type ValidationSpec struct {
	// Search queries used for validation testing, e.g. "repo:^github\\.com/gorilla/mux$ Router".
	SearchQuery []string `yaml:"searchQuery"`

	// External Service configuration.
	ExternalService ExternalService `yaml:"externalService"`

	// Insight used for validation testing.
	Insight Insight `yaml:"insight"`
}

// DefaultConfig returns a default configuration to be used for testing.
func DefaultConfig() *ValidationSpec {
	return &ValidationSpec{
		SearchQuery: []string{
			"repo:^github.com/sourcegraph/src-cli$ config",
			"repo:^github.com/sourcegraph/src-cli$@4.0.0 config",
			"repo:^github.com/sourcegraph/src-cli$ type:symbol config",
		},
		ExternalService: ExternalService{
			Kind:        "GITHUB",
			DisplayName: "sourcegraph-test",
			Config: Config{
				GitHub: GitHub{
					URL:   "https://github.com",
					Token: "",
					Orgs:  []string{},
					Repos: []string{"sourcegraph/src-cli"},
				},
			},
			MaxRetries:          5,
			RetryTimeoutSeconds: 5,
			DeleteWhenDone:      true,
		},
		Insight: Insight{
			Title: "test insight",
			DataSeries: []map[string]any{
				{
					"query":           "lang:javascript",
					"label":           "javascript",
					"repositoryScope": "",
					"lineColor":       "#6495ED",
					"timeScopeUnit":   "MONTH",
					"timeScopeValue":  1,
				},
				{
					"query":           "lang:typescript",
					"label":           "typescript",
					"lineColor":       "#DE3163",
					"repositoryScope": "",
					"timeScopeUnit":   "MONTH",
					"timeScopeValue":  1,
				},
			},
		},
	}
}

type jsonVars map[string]interface{}

type clientQuery struct {
	opName    string
	query     string
	variables jsonVars
}

// LoadYamlConfig will unmarshal a YAML configuration file into a ValidationSpec.
func LoadYamlConfig(userConfig []byte) (*ValidationSpec, error) {
	var config ValidationSpec
	if err := yaml.Unmarshal(userConfig, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// LoadJsonConfig will unmarshal a JSON configuration file into a ValidationSpec.
func LoadJsonConfig(userConfig []byte) (*ValidationSpec, error) {
	var config ValidationSpec
	if err := json.Unmarshal(userConfig, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// Validate runs a series of validation checks such as cloning a repository, running search queries, and
// creating insights, based on the configuration provided.
func Validate(ctx context.Context, client api.Client, config *ValidationSpec) error {
	log.Printf("%s validating external service", validate.EmojiFingerPointRight)

	if config.ExternalService.DisplayName != "" {
		srvID, err := addExternalService(ctx, client, config.ExternalService)
		if err != nil {
			return err
		}

		log.Printf("%s external service %s is being added", validate.HourglassEmoji, config.ExternalService.DisplayName)

		defer func() {
			if srvID != "" && config.ExternalService.DeleteWhenDone {
				_ = removeExternalService(ctx, client, srvID)
				log.Printf("%s external service %s has been removed", validate.SuccessEmoji, config.ExternalService.DisplayName)
			}
		}()
	}

	log.Printf("%s cloning repository", validate.HourglassEmoji)

	cloned, err := repoCloneTimeout(ctx, client, config.ExternalService)
	if err != nil {
		return err //TODO make sure errors are wrapped once
	}
	if !cloned {
		return errors.Newf("%s validate failed, repo did not clone\n", validate.FailureEmoji)
	}

	log.Printf("%s repositry successfully cloned", validate.SuccessEmoji)

	log.Printf("%s validating search queries", validate.EmojiFingerPointRight)

	if config.SearchQuery != nil {
		for i := 0; i < len(config.SearchQuery); i++ {
			matchCount, err := searchMatchCount(ctx, client, config.SearchQuery[i])
			if err != nil {
				return err
			}
			if matchCount == 0 {
				return errors.Newf("validate failed, search query %s returned no results", config.SearchQuery[i])
			}
			log.Printf("%s search query '%s' was successful", validate.SuccessEmoji, config.SearchQuery[i])
		}
	}

	log.Printf("%s validating code insight", validate.EmojiFingerPointRight)

	if config.Insight.Title != "" {
		log.Printf("%s insight %s is being added", validate.HourglassEmoji, config.Insight.Title)

		insightId, err := createInsight(ctx, client, config.Insight)
		if err != nil {
			return err
		}

		log.Printf("%s insight successfully added", validate.SuccessEmoji)

		defer func() {
			if insightId != "" {
				_ = removeInsight(ctx, client, insightId)
				log.Printf("%s insight %s has been removed", validate.SuccessEmoji, config.Insight.Title)

			}
		}()
	}

	return nil
}

func addExternalService(ctx context.Context, client api.Client, srv ExternalService) (string, error) {
	config, err := json.Marshal(srv.Config.GitHub)
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
			"kind":        srv.Kind,
			"displayName": srv.DisplayName,
			"config":      string(config),
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

func repoCloneTimeout(ctx context.Context, client api.Client, srv ExternalService) (bool, error) {
	// construct repo string for query
	var name strings.Builder

	name.WriteString("github.com/")
	name.WriteString(srv.Config.GitHub.Repos[0])

	for i := 0; i < srv.MaxRetries; i++ {
		repos, err := listClonedRepos(ctx, client, []string{name.String()})
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
		query: `query ListRepos($names: [String!]) {
			  repositories(
				names: $names
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

func createInsight(ctx context.Context, client api.Client, insight Insight) (string, error) {
	var dataSeries []map[string]interface{}

	for _, ds := range insight.DataSeries {
		var series = map[string]interface{}{
			"query": ds["query"],
			"options": map[string]interface{}{
				"label":     ds["label"],
				"lineColor": ds["lineColor"],
			},
			"repositoryScope": map[string]interface{}{
				"repositories": ds["repositoryScope"],
			},
			"timeScope": map[string]interface{}{
				"stepInterval": map[string]interface{}{
					"unit":  ds["timeScopeUnit"],
					"value": ds["timeScopeValue"],
				},
			},
		}

		dataSeries = append(dataSeries, series)
	}

	q := clientQuery{
		opName: "CreateLineChartSearchInsight",
		query: `mutation CreateLineChartSearchInsight($input: LineChartSearchInsightInput!) {
			createLineChartSearchInsight(input: $input) {
	  			view {
					id
	  			}
			}
		}`,
		variables: jsonVars{
			"input": map[string]interface{}{
				"options":    map[string]interface{}{"title": insight.Title},
				"dataSeries": dataSeries,
			},
		},
	}

	var result struct {
		CreateLineChartSearchInsight struct {
			View struct {
				ID string `json:"id"`
			} `json:"view"`
		} `json:"createLineChartSearchInsight"`
	}

	ok, err := client.NewRequest(q.query, q.variables).Do(ctx, &result)
	if err != nil {
		return "", errors.Wrap(err, "createInsight failed")
	}
	if !ok {
		return "", errors.New("createInsight failed, no data to unmarshal")
	}

	return result.CreateLineChartSearchInsight.View.ID, nil
}

func removeInsight(ctx context.Context, client api.Client, insightId string) error {
	q := clientQuery{
		opName: "DeleteInsightView",
		query: `mutation DeleteInsightView ($id: ID!) {
			deleteInsightView(id: $id){
				alwaysNil
			}
		}`,
		variables: jsonVars{
			"id": insightId,
		},
	}

	var result struct {
		Data struct {
			DeleteInsightView struct {
				AlwaysNil string `json:"alwaysNil"`
			} `json:"deleteInsightView"`
		} `json:"data"`
	}

	ok, err := client.NewRequest(q.query, q.variables).Do(ctx, &result)
	if err != nil {
		return errors.Wrap(err, "removeInsight failed")
	}
	if !ok {
		return errors.New("removeInsight failed, no data to unmarshal")
	}

	return nil
}
