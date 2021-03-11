package repos

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/defaults"
	"github.com/aws/aws-sdk-go-v2/aws/endpoints"
	"github.com/inconshreveable/log15"
	"golang.org/x/net/http2"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/awscodecommit"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

// An AWSCodeCommitSource yields repositories from a single AWS Code Commit
// connection configured in Sourcegraph via the external services
// configuration.
type AWSCodeCommitSource struct {
	svc    *types.ExternalService
	config *schema.AWSCodeCommitConnection

	awsConfig    aws.Config
	awsPartition endpoints.Partition // "aws", "aws-cn", "aws-us-gov"
	awsRegion    endpoints.Region
	client       *awscodecommit.Client

	exclude excludeFunc
}

// NewAWSCodeCommitSource returns a new AWSCodeCommitSource from the given external service.
func NewAWSCodeCommitSource(svc *types.ExternalService, cf *httpcli.Factory) (*AWSCodeCommitSource, error) {
	var c schema.AWSCodeCommitConnection
	if err := jsonc.Unmarshal(svc.Config, &c); err != nil {
		return nil, fmt.Errorf("external service id=%d config error: %s", svc.ID, err)
	}
	return newAWSCodeCommitSource(svc, &c, cf)
}

func newAWSCodeCommitSource(svc *types.ExternalService, c *schema.AWSCodeCommitConnection, cf *httpcli.Factory) (*AWSCodeCommitSource, error) {
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
		cf = httpcli.NewExternalHTTPClientFactory()
	}

	cli, err := cf.Doer(func(c *http.Client) error {
		tr := aws.NewBuildableHTTPClient().GetTransport()
		if err := http2.ConfigureTransport(tr); err != nil {
			return err
		}
		c.Transport = tr
		wrapWithoutRedirect(c)

		return nil
	})
	if err != nil {
		return nil, err
	}
	awsConfig.HTTPClient = cli

	var eb excludeBuilder
	for _, r := range c.Exclude {
		eb.Exact(r.Name)
		eb.Exact(r.Id)
	}
	exclude, err := eb.Build()
	if err != nil {
		return nil, err
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
func (s *AWSCodeCommitSource) ListRepos(ctx context.Context, results chan SourceResult) {
	s.listAllRepositories(ctx, results)
}

// ExternalServices returns a singleton slice containing the external service.
func (s *AWSCodeCommitSource) ExternalServices() types.ExternalServices {
	return types.ExternalServices{s.svc}
}

func (s *AWSCodeCommitSource) makeRepo(r *awscodecommit.Repository) (*types.Repo, error) {
	urn := s.svc.URN()
	cloneURL := s.authenticatedRemoteURL(r)
	serviceID := awscodecommit.ServiceID(s.awsPartition, s.awsRegion, r.AccountID)

	return &types.Repo{
		Name:         reposource.AWSRepoName(s.config.RepositoryPathPattern, r.Name),
		URI:          string(reposource.AWSRepoName("", r.Name)),
		ExternalRepo: awscodecommit.ExternalRepoSpec(r, serviceID),
		Description:  r.Description,
		Sources: map[string]*types.SourceInfo{
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

func (s *AWSCodeCommitSource) listAllRepositories(ctx context.Context, results chan SourceResult) {
	var nextToken string
	for {
		batch, token, err := s.client.ListRepositories(ctx, nextToken)
		if err != nil {
			results <- SourceResult{Source: s, Err: err}
			return
		}

		for _, r := range batch {
			if !s.excludes(r) {
				repo, err := s.makeRepo(r)
				if err != nil {
					results <- SourceResult{Source: s, Err: err}
					return
				}
				results <- SourceResult{Source: s, Repo: repo}
			}
		}

		if len(batch) == 0 || token == "" {
			break // last page
		}

		nextToken = token
	}
}

func (s *AWSCodeCommitSource) excludes(r *awscodecommit.Repository) bool {
	return s.exclude(r.Name) || s.exclude(r.ID)
}

// The code below is copied from
// github.com/aws/aws-sdk-go-v2@v0.11.0/aws/client.go so we use the same HTTP
// client that AWS wants to use, but fits into our HTTP factory
// pattern. Additionally we change wrapWithoutRedirect to mutate c instead of
// returning a copy.
func wrapWithoutRedirect(c *http.Client) {
	tr := c.Transport
	if tr == nil {
		tr = http.DefaultTransport
	}

	c.CheckRedirect = limitedRedirect
	c.Transport = stubBadHTTPRedirectTransport{
		tr: tr,
	}
}

func limitedRedirect(r *http.Request, via []*http.Request) error {
	// Request.Response, in CheckRedirect is the response that is triggering
	// the redirect.
	resp := r.Response
	if r.URL.String() == stubBadHTTPRedirectLocation {
		resp.Header.Del(stubBadHTTPRedirectLocation)
		return http.ErrUseLastResponse
	}

	switch resp.StatusCode {
	case 307, 308:
		// Only allow 307 and 308 redirects as they preserve the method.
		return nil
	}

	return http.ErrUseLastResponse
}

type stubBadHTTPRedirectTransport struct {
	tr http.RoundTripper
}

const stubBadHTTPRedirectLocation = `https://amazonaws.com/badhttpredirectlocation`

func (t stubBadHTTPRedirectTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	resp, err := t.tr.RoundTrip(r)
	if err != nil {
		return resp, err
	}

	// TODO S3 is the only known service to return 301 without location header.
	// consider moving this to a S3 customization.
	switch resp.StatusCode {
	case 301, 302:
		if v := resp.Header.Get("Location"); len(v) == 0 {
			resp.Header.Set("Location", stubBadHTTPRedirectLocation)
		}
	}

	return resp, err
}

// UnwrappableTransport signals that this transport can't be wrapped. In
// particular this means we won't respect global external
// settings. https://github.com/sourcegraph/sourcegraph/issues/71 and
// https://github.com/sourcegraph/sourcegraph/issues/7738
func (stubBadHTTPRedirectTransport) UnwrappableTransport() {}
