package repos

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/defaults"
	"github.com/aws/aws-sdk-go-v2/aws/endpoints"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/atomicvalue"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/conf/reposource"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/awscodecommit"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

var awsCodeCommitConnections = atomicvalue.New()

func init() {
	awsCodeCommitConnections.Set(func() interface{} {
		return []*awsCodeCommitConnection{}
	})

	go func() {
		t := time.NewTicker(configWatchInterval)
		var lastConfig []*schema.AWSCodeCommitConnection
		for range t.C {
			config, err := conf.AWSCodeCommitConfigs(context.Background())
			if err != nil {
				log15.Error("unable to fetch AWS CodeCommit configs", "err", err)
				continue
			}

			if reflect.DeepEqual(config, lastConfig) {
				continue
			}
			lastConfig = config

			var conns []*awsCodeCommitConnection
			for _, c := range config {
				conn, err := newAWSCodeCommitConnection(c)
				if err != nil {
					log15.Error("Error processing configured AWS CodeCommit connection. Skipping it.", "region", c.Region, "error", err)
					continue
				}
				conns = append(conns, conn)
			}

			awsCodeCommitConnections.Set(func() interface{} {
				return conns
			})

			awsCodeCommitRepositorySyncWorker.restart()
		}
	}()
}

// GetAWSCodeCommitRepositoryMock is set by tests that need to mock GetAWSCodeCommitRepository.
var GetAWSCodeCommitRepositoryMock func(args protocol.RepoLookupArgs) (repo *protocol.RepoInfo, authoritative bool, err error)

// GetAWSCodeCommitRepository queries a configured AWS CodeCommit connection endpoint for
// information about the specified repository.
//
// If args.Repo refers to a repository that is not known to be on a configured AWS CodeCommit
// connection's host, it returns authoritative == false.
func GetAWSCodeCommitRepository(ctx context.Context, args protocol.RepoLookupArgs) (repo *protocol.RepoInfo, authoritative bool, err error) {
	if GetAWSCodeCommitRepositoryMock != nil {
		return GetAWSCodeCommitRepositoryMock(args)
	}

	log15.Debug("GetAWSCodeCommitRepository", "repo", args.Repo, "externalRepo", args.ExternalRepo)

	if args.ExternalRepo != nil && args.ExternalRepo.ServiceType == awscodecommit.ServiceType {
		// Look up by external repository spec.
		var err error
		for _, conn := range awsCodeCommitConnections.Get().([]*awsCodeCommitConnection) {
			var serviceID string
			serviceID, err = conn.getServiceID()
			if serviceID != "" && args.ExternalRepo.ServiceID == serviceID {
				ccrepo, err := conn.client.GetRepository(ctx, args.ExternalRepo.ID)
				if ccrepo != nil {
					remoteURL, err := conn.authenticatedRemoteURL(ccrepo)
					if err != nil {
						return nil, true, errors.Wrap(err, "authenticatedRemoteURL")
					}
					webURL := fmt.Sprintf("https://%s.console.aws.amazon.com/codecommit/home#/repository/%s", conn.awsRegion.ID(), ccrepo.Name)
					repo = &protocol.RepoInfo{
						Name:         awsCodeCommitRepositoryToRepoPath(conn, ccrepo),
						ExternalRepo: awscodecommit.ExternalRepoSpec(ccrepo, serviceID),
						Description:  ccrepo.Description,
						VCS:          protocol.VCSInfo{URL: remoteURL},
						Links: &protocol.RepoLinks{
							Root:   webURL,
							Tree:   webURL + "/browse/{rev}/--/{path}",
							Blob:   webURL + "/browse/{rev}/--/{path}",
							Commit: webURL + "/commit/{commit}",
						},
					}
				}
				return repo, true, errors.Wrap(err, "GetRepository")
			}
		}
		return nil, true, errors.Wrap(err, "getServiceID")
	}

	// Unlike other code hosts (e.g., GitHub and GitLab), looking up by repository name is not
	// supported because it's far less likely to be useful for AWS CodeCommit, which usually has a
	// more limited universe of repositories.
	return nil, false, nil
}

var awsCodeCommitRepositorySyncWorker = &worker{
	work: func(ctx context.Context, shutdown chan struct{}) {
		awsCodeCommitConnections := awsCodeCommitConnections.Get().([]*awsCodeCommitConnection)
		if len(awsCodeCommitConnections) == 0 {
			return
		}
		for _, c := range awsCodeCommitConnections {
			go func(c *awsCodeCommitConnection) {
				// Hit the AWS API to determine our account ID (which is a fixed value but not derivable
				// from the values in the Sourcegraph site config). Be robust to the API being
				// unreachable when we start up.
				const retryInterval = 20 * time.Second
				for {
					_, err := c.tryPopulateAWSAccountID()
					if err == nil {
						break
					}
					log15.Error("Unable to reach AWS CodeCommit API to determine AWS account ID.", "region", c.config.Region, "error", err, "retryInterval", retryInterval)
					select {
					case <-shutdown:
						return
					case <-time.After(retryInterval):
					}
				}

				for {
					updateAWSCodeCommitRepositories(ctx, c)
					awsCodeCommitUpdateTime.WithLabelValues(c.awsAccountID).Set(float64(time.Now().Unix()))
					select {
					case <-shutdown:
						return
					case <-time.After(GetUpdateInterval()):
					}
				}
			}(c)
		}
	},
}

// RunAWSCodeCommitRepositorySyncWorker runs the worker that syncs repositories from the configured AWSCodeCommit and AWSCodeCommit
// Enterprise instances to Sourcegraph.
func RunAWSCodeCommitRepositorySyncWorker(ctx context.Context) {
	awsCodeCommitRepositorySyncWorker.start(ctx)
}

func awsCodeCommitRepositoryToRepoPath(conn *awsCodeCommitConnection, repo *awscodecommit.Repository) api.RepoName {
	return reposource.AWSRepoName(conn.config.RepositoryPathPattern, repo.Name)
}

// updateAWSCodeCommitRepositories ensures that all provided repositories have been added and updated on Sourcegraph.
func updateAWSCodeCommitRepositories(ctx context.Context, conn *awsCodeCommitConnection) {
	repos := conn.listAllRepositories(ctx)

	repoChan := make(chan repoCreateOrUpdateRequest)
	defer close(repoChan)
	go createEnableUpdateRepos(ctx, fmt.Sprintf("aws:%s", conn.config.AccessKeyID), repoChan)
	for repo := range repos {
		// log15.Debug("awscodecommit sync: create/enable/update repo", "repo", repo.Name)
		remoteURL, err := conn.authenticatedRemoteURL(repo)
		if err != nil {
			log15.Error("Error generating remote URL for AWS CodeCommit repository. Skipping.", "repo", repo.ARN, "error", err)
			continue
		}
		repoChan <- repoCreateOrUpdateRequest{
			RepoCreateOrUpdateRequest: api.RepoCreateOrUpdateRequest{
				RepoName:     awsCodeCommitRepositoryToRepoPath(conn, repo),
				ExternalRepo: awscodecommit.ExternalRepoSpec(repo, awscodecommit.ServiceID(conn.awsPartition, conn.awsRegion, repo.AccountID)),
				Description:  repo.Description,
				Enabled:      conn.config.InitialRepositoryEnablement,
			},
			URL: remoteURL,
		}
	}
}

func newAWSCodeCommitConnection(config *schema.AWSCodeCommitConnection) (*awsCodeCommitConnection, error) {
	awsConfig := defaults.Config()
	awsConfig.Region = config.Region
	awsConfig.Credentials = aws.StaticCredentialsProvider{
		Value: aws.Credentials{
			AccessKeyID:     config.AccessKeyID,
			SecretAccessKey: config.SecretAccessKey,
			Source:          "sourcegraph-site-configuration",
		},
	}
	conn := &awsCodeCommitConnection{
		config:    config,
		awsConfig: awsConfig,
	}
	conn.client = awscodecommit.NewClient(conn.awsConfig)

	var ok bool
	conn.awsPartition, ok = endpoints.DefaultPartitions().ForRegion(config.Region)
	if ok {
		conn.awsRegion, ok = conn.awsPartition.Regions()[config.Region]
	}
	if !ok {
		return nil, fmt.Errorf("unrecognized AWS region name: %q", config.Region)
	}

	return conn, nil
}

type awsCodeCommitConnection struct {
	config       *schema.AWSCodeCommitConnection
	awsConfig    aws.Config
	awsPartition endpoints.Partition // "aws", "aws-cn", "aws-us-gov"
	awsRegion    endpoints.Region
	client       *awscodecommit.Client

	mu           sync.Mutex
	awsAccountID string
}

func (c *awsCodeCommitConnection) getServiceID() (string, error) {
	awsAccountID, err := c.tryPopulateAWSAccountID()
	if err != nil {
		return "", err
	}
	if awsAccountID == "" {
		return "", nil
	}
	return awscodecommit.ServiceID(c.awsPartition, c.awsRegion, c.awsAccountID), nil
}

func (c *awsCodeCommitConnection) tryPopulateAWSAccountID() (string, error) {
	c.mu.Lock()
	awsAccountID := c.awsAccountID
	c.mu.Unlock()
	if awsAccountID != "" {
		return awsAccountID, nil
	}

	result, err := sts.New(c.awsConfig).GetCallerIdentityRequest(&sts.GetCallerIdentityInput{}).Send()
	if err != nil {
		return "", err
	}
	if result.Account == nil {
		return "", errors.New("AWS STS GetCallerIdentity unexpectedly returned empty AWS account ID")
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.awsAccountID = *result.Account
	return c.awsAccountID, nil
}

// authenticatedRemoteURL returns the repository's Git remote URL with the configured AWS CodeCommit
// credentials inserted in the URL userinfo, for repositories needing authentication.
func (c *awsCodeCommitConnection) authenticatedRemoteURL(repo *awscodecommit.Repository) (string, error) {
	// Mimic what `aws codecommit credential-helper` does (to create Git credentials). See
	// https://github.com/aws/aws-cli/blob/2e3fb985e21968abb09bba5bf439245fccb02a9f/awscli/customizations/codecommit.py.
	u, err := url.Parse(repo.HTTPCloneURL)
	if err != nil {
		return "", err
	}

	cred, err := c.awsConfig.Credentials.Retrieve()
	if err != nil {
		return "", err
	}

	// Need to reimplement some of the AWS v4 signing because the Go SDK does not expose the ability
	// to sign a specific canonical request string (it always adds headers like X-Amz-... that must
	// not exist when creating credentials for AWS CodeCommit git cloning).
	const (
		authHeaderPrefix = "AWS4-HMAC-SHA256"
		serviceName      = "codecommit"
		shortTimeFormat  = "20060102"
		timeFormat       = "20060102T150405"
	)
	signTime := time.Now().UTC()
	formattedShortTime := signTime.Format(shortTimeFormat)
	canonicalRequest := fmt.Sprintf("GIT\n%s\n\nhost:%s\n\nhost\n", u.Path, u.Host)
	// fmt.Printf("=============\nCanonicalRequest:\n%s\n", canonicalRequest)
	stringToSign := strings.Join([]string{
		authHeaderPrefix,
		signTime.Format(timeFormat),
		strings.Join([]string{formattedShortTime, c.awsRegion.ID(), serviceName, "aws4_request"}, "/"),
		hex.EncodeToString(makeSHA256([]byte(canonicalRequest))),
	}, "\n")
	// fmt.Printf("=============\nStringToSign:\n%s\n", stringToSign)
	date := makeHMAC([]byte("AWS4"+cred.SecretAccessKey), []byte(formattedShortTime))
	region := makeHMAC(date, []byte(c.awsRegion.ID()))
	service := makeHMAC(region, []byte(serviceName))
	credentials := makeHMAC(service, []byte("aws4_request"))
	signature := hex.EncodeToString(makeHMAC(credentials, []byte(stringToSign)))

	password := signTime.Format(timeFormat) + "Z" + signature
	username := c.config.AccessKeyID

	u.User = url.UserPassword(username, password)
	return u.String(), nil
}

func makeHMAC(key []byte, data []byte) []byte {
	hash := hmac.New(sha256.New, key)
	hash.Write(data)
	return hash.Sum(nil)
}

func makeSHA256(data []byte) []byte {
	hash := sha256.New()
	hash.Write(data)
	return hash.Sum(nil)
}

func (c *awsCodeCommitConnection) listAllRepositories(ctx context.Context) <-chan *awscodecommit.Repository {
	ch := make(chan *awscodecommit.Repository, awscodecommit.MaxMetadataBatch)
	go func() {
		defer close(ch)
		var nextToken string
		for {
			repos, token, err := c.client.ListRepositories(ctx, nextToken)
			if err != nil {
				log15.Error("Error listing AWS CodeCommit repositories", "error", err)
				return
			}
			for _, r := range repos {
				ch <- r
			}
			if len(repos) == 0 || token == "" {
				return // last page
			}
			nextToken = token
		}
	}()
	return ch
}
