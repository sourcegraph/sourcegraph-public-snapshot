package repos

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/defaults"
	"github.com/aws/aws-sdk-go-v2/aws/endpoints"
	"github.com/sourcegraph/sourcegraph/pkg/conf/reposource"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/awscodecommit"
	"github.com/sourcegraph/sourcegraph/pkg/httpcli"
	"github.com/sourcegraph/sourcegraph/pkg/jsonc"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// An AWSCodeCommitSource yields repositories from a single AWS Code Commit
// connection configured in Sourcegraph via the external services
// configuration.
type AWSCodeCommitSource struct {
	svc    *ExternalService
	config *schema.AWSCodeCommitConnection

	awsConfig    aws.Config
	awsPartition endpoints.Partition // "aws", "aws-cn", "aws-us-gov"
	awsRegion    endpoints.Region
	client       *awscodecommit.Client

	exclude map[string]bool
}

// NewAWSCodeCommitSource returns a new AWSCodeCommitSource from the given external service.
func NewAWSCodeCommitSource(svc *ExternalService, cf *httpcli.Factory) (*AWSCodeCommitSource, error) {
	var c schema.AWSCodeCommitConnection
	if err := jsonc.Unmarshal(svc.Config, &c); err != nil {
		return nil, fmt.Errorf("external service id=%d config error: %s", svc.ID, err)
	}
	return newAWSCodeCommitSource(svc, &c, cf)
}

func newAWSCodeCommitSource(svc *ExternalService, c *schema.AWSCodeCommitConnection, cf *httpcli.Factory) (*AWSCodeCommitSource, error) {
	awsConfig := defaults.Config()
	awsConfig.Region = c.Region
	awsConfig.Credentials = aws.StaticCredentialsProvider{
		Value: aws.Credentials{
			AccessKeyID:     c.AccessKeyID,
			SecretAccessKey: c.SecretAccessKey,
			Source:          "sourcegraph-site-configuration",
		},
	}

	if cf == nil {
		cf = NewHTTPClientFactory()
	}

	cli, err := cf.Client(func(c *http.Client) error {
		*c = *awsConfig.HTTPClient
		return nil
	})
	if err != nil {
		return nil, err
	}
	awsConfig.HTTPClient = cli

	exclude := make(map[string]bool, len(c.Exclude))
	for _, r := range c.Exclude {
		if r.Name != "" {
			exclude[r.Name] = true
		}

		if r.Id != "" {
			exclude[r.Id] = true
		}
	}

	s := &AWSCodeCommitSource{
		svc:       svc,
		config:    c,
		awsConfig: awsConfig,
		exclude:   exclude,
		client:    awscodecommit.NewClient(awsConfig),
	}

	var ok bool
	s.awsPartition, ok = endpoints.DefaultPartitions().ForRegion(c.Region)
	if ok {
		s.awsRegion, ok = s.awsPartition.Regions()[c.Region]
	}
	if !ok {
		return nil, fmt.Errorf("unrecognized AWS region name: %q", c.Region)
	}

	return s, nil
}

// ListRepos returns all AWS Code Commit repositories accessible to all
// connections configured in Sourcegraph via the external services
// configuration.
func (s *AWSCodeCommitSource) ListRepos(ctx context.Context, results chan *SourceResult) {
	s.listAllRepositories(ctx, results)
}

// ExternalServices returns a singleton slice containing the external service.
func (s *AWSCodeCommitSource) ExternalServices() ExternalServices {
	return ExternalServices{s.svc}
}

func (s *AWSCodeCommitSource) makeRepo(r *awscodecommit.Repository) (*Repo, error) {
	urn := s.svc.URN()
	cloneURL := s.authenticatedRemoteURL(r)
	serviceID := awscodecommit.ServiceID(s.awsPartition, s.awsRegion, r.AccountID)

	return &Repo{
		Name:         string(reposource.AWSRepoName(s.config.RepositoryPathPattern, r.Name)),
		URI:          string(reposource.AWSRepoName("", r.Name)),
		ExternalRepo: awscodecommit.ExternalRepoSpec(r, serviceID),
		Description:  r.Description,
		Enabled:      true,
		Sources: map[string]*SourceInfo{
			urn: {
				ID:       urn,
				CloneURL: cloneURL,
			},
		},
		Metadata: r,
	}, nil
}

// authenticatedRemoteURL returns the repository's Git remote URL with the
// configured AWS CodeCommit Git credentials inserted in the URL userinfo, for
// repositories needing authentication.
func (s *AWSCodeCommitSource) authenticatedRemoteURL(repo *awscodecommit.Repository) string {
	u, err := url.Parse(repo.HTTPCloneURL)
	if err != nil {
		log15.Warn("Error adding authentication to AWS CodeCommit repository Git remote URL.", "url", repo.HTTPCloneURL, "error", err)
		return repo.HTTPCloneURL
	}

	username := s.config.GitCredentials.Username
	password := s.config.GitCredentials.Password

	u.User = url.UserPassword(username, password)
	return u.String()
}

func (s *AWSCodeCommitSource) listAllRepositories(ctx context.Context, results chan *SourceResult) {
	var nextToken string
	for {
		batch, token, err := s.client.ListRepositories(ctx, nextToken)
		if err != nil {
			results <- &SourceResult{Source: s, Err: err}
			return
		}

		for _, r := range batch {
			if !s.excludes(r) {
				repo, err := s.makeRepo(r)
				if err != nil {
					results <- &SourceResult{Source: s, Err: err}
					return
				}
				results <- &SourceResult{Source: s, Repo: repo}
			}
		}

		if len(batch) == 0 || token == "" {
			break // last page
		}

		nextToken = token
	}
}

func (s *AWSCodeCommitSource) excludes(r *awscodecommit.Repository) bool {
	return s.exclude[r.Name] || s.exclude[r.ID]
}
