package gqltest

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/inconshreveable/log15"
	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

var Client *gqltestutil.Client

var (
	long = flag.Bool("long", false, "Enable the integration tests to run. Required flag, otherwise tests are skipped.")

	BaseURL  = flag.String("base-url", "http://127.0.0.1:7080", "The base URL of the Sourcegraph instance")
	Email    = flag.String("email", "gqltest@sourcegraph.com", "The email of the admin user")
	Username = flag.String("username", "gqltest-admin", "The username of the admin user")
	Password = flag.String("password", "supersecurepassword", "The password of the admin user")

	GithubToken           = flag.String("github-token", os.Getenv("GITHUB_TOKEN"), "The GitHub personal access token that will be used to authenticate a GitHub external service")
	AwsAccessKeyID        = flag.String("aws-access-key-id", os.Getenv("AWS_ACCESS_KEY_ID"), "The AWS access key ID")
	AwsSecretAccessKey    = flag.String("aws-secret-access-key", os.Getenv("AWS_SECRET_ACCESS_KEY"), "The AWS secret access key")
	AwsCodeCommitUsername = flag.String("aws-code-commit-username", os.Getenv("AWS_CODE_COMMIT_USERNAME"), "The AWS code commit username")
	AwsCodeCommitPassword = flag.String("aws-code-commit-password", os.Getenv("AWS_CODE_COMMIT_PASSWORD"), "The AWS code commit password")
	BbsURL                = flag.String("bbs-url", os.Getenv("BITBUCKET_SERVER_URL"), "The Bitbucket Server URL")
	BbsToken              = flag.String("bbs-token", os.Getenv("BITBUCKET_SERVER_TOKEN"), "The Bitbucket Server token")
	BbsUsername           = flag.String("bbs-username", os.Getenv("BITBUCKET_SERVER_USERNAME"), "The Bitbucket Server username")
	AzureDevOpsUsername   = flag.String("azure-devops-username", os.Getenv("AZURE_DEVOPS_USERNAME"), "The Azure DevOps username")
	AzureDevOpsToken      = flag.String("azure-devops-token", os.Getenv("AZURE_DEVOPS_TOKEN"), "The Azure DevOps personal access token")
	PerforcePort          = flag.String("perforce-port", os.Getenv("PERFORCE_PORT"), "The URL of the Perforce server")
	PerforceUser          = flag.String("perforce-user", os.Getenv("PERFORCE_USER"), "The username required to access the Perforce server")
	PerforcePassword      = flag.String("perforce-password", os.Getenv("PERFORCE_PASSWORD"), "The password required to access the Perforce server")
)

func TestMain(m *testing.M) {
	flag.Parse()

	if !*long {
		fmt.Println("SKIP: skipping gqltest since -long is not specified.")
		return
	}

	*BaseURL = strings.TrimSuffix(*BaseURL, "/")

	// Make it possible to use a different token on non-default branches
	// so we don't break builds on the default branch.
	mockGitHubToken := os.Getenv("MOCK_GITHUB_TOKEN")
	if mockGitHubToken != "" {
		*GithubToken = mockGitHubToken
	}

	needsSiteInit, resp, err := gqltestutil.NeedsSiteInit(*BaseURL)
	if resp != "" && os.Getenv("BUILDKITE") == "true" {
		log.Println("server response: ", resp)
	}
	if err != nil {
		log.Fatal("Failed to check if site needs init:", err)
	}

	Client, err = gqltestutil.NewClient(*BaseURL)
	if err != nil {
		log.Fatal("Failed to create gql client: ", err)
	}
	if needsSiteInit {
		if err := Client.SiteAdminInit(*Email, *Username, *Password); err != nil {
			log.Fatal("Failed to create site admin: ", err)
		}
		log.Println("Site admin has been created:", *Username)
	} else {
		if err := Client.SignIn(*Email, *Password); err != nil {
			log.Fatal("Failed to sign in:", err)
		}
		log.Println("Site admin authenticated:", *Username)
	}

	licenseKey := os.Getenv("SOURCEGRAPH_LICENSE_KEY")
	if licenseKey != "" {
		siteConfig, lastID, err := Client.SiteConfiguration()
		if err != nil {
			log.Fatal("Failed to get site configuration:", err)
		}

		err = func() error {
			// Update site configuration to set up a test license key if the instance doesn't have one yet.
			if siteConfig.LicenseKey != "" {
				return nil
			}

			siteConfig.LicenseKey = licenseKey
			err = Client.UpdateSiteConfiguration(siteConfig, lastID)
			if err != nil {
				return errors.Wrap(err, "update site configuration")
			}

			// Verify the provided license is valid, retry because the configuration update
			// endpoint is eventually consistent.
			err = gqltestutil.Retry(5*time.Second, func() error {
				ps, err := Client.ProductSubscription()
				if err != nil {
					return errors.Wrap(err, "get product subscription")
				}

				if ps.License == nil {
					return gqltestutil.ErrContinueRetry
				}
				return nil
			})
			if err != nil {
				return errors.Wrap(err, "verify license")
			}
			return nil
		}()
		if err != nil {
			log.Fatal("Failed to update license:", err)
		}
		log.Println("License key added and verified")
	}

	if !testing.Verbose() {
		log15.Root().SetHandler(log15.DiscardHandler())
	}
	os.Exit(m.Run())
}

func MustMarshalJSONString(v any) string {
	str, err := jsoniter.MarshalToString(v)
	if err != nil {
		panic(err)
	}
	return str
}

func RemoveTestUserAfterTest(t *testing.T, userID string) {
	t.Helper()
	t.Cleanup(func() {
		if err := Client.DeleteUser(userID, true); err != nil {
			t.Fatal(err)
		}
	})
}

func RemoveExternalServiceAfterTest(t *testing.T, esID string) {
	t.Helper()
	t.Cleanup(func() {
		err := Client.DeleteExternalService(esID, false)
		if err != nil {
			t.Fatal(err)
		}
	})
}


func EnableSubRepoPermissions(t *testing.T) {
	t.Helper()
	t.Log("Enabling sub-repo permissions")

	reset, err := Client.ModifySiteConfiguration(func(siteConfig *schema.SiteConfiguration) {
		if siteConfig.ExperimentalFeatures == nil {
			siteConfig.ExperimentalFeatures = &schema.ExperimentalFeatures{}
		}
		siteConfig.ExperimentalFeatures.Perforce = "enabled"
		siteConfig.ExperimentalFeatures.SubRepoPermissions = &schema.SubRepoPermissions{Enabled: true}
	})
	require.NoError(t, err)
	if reset != nil {
		t.Cleanup(func() {
			require.NoError(t, reset())
		})
	}
}

func EnableStructuralSearch(t *testing.T) {
	t.Helper()
	t.Log("Enabling structural search")

	reset, err := Client.ModifySiteConfiguration(func(siteConfig *schema.SiteConfiguration) {
		if siteConfig.ExperimentalFeatures == nil {
			siteConfig.ExperimentalFeatures = &schema.ExperimentalFeatures{}
		}
		siteConfig.ExperimentalFeatures.StructuralSearch = "enabled"
	})
	require.NoError(t, err)
	if reset != nil {
		t.Cleanup(func() {
			require.NoError(t, reset())
		})
	}
}
