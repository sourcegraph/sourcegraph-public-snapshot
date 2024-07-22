package codeintel

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/exp/maps"

	"github.com/c2h5oh/datasize"
	"github.com/kballard/go-shellquote"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/executorqueue/handler"
	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	apiclient "github.com/sourcegraph/sourcegraph/internal/executor/types"
)

const (
	defaultOutfile      = "dump.lsif"
	uploadRoute         = "/.executors/lsif/upload"
	schemeExecutorToken = "token-executor"
)

// accessLogTransformer sets the approriate fields on the executor secret access log entry
// for auto-indexing access
type accessLogTransformer struct {
	database.ExecutorSecretAccessLogCreator
}

func (e *accessLogTransformer) Create(ctx context.Context, log *database.ExecutorSecretAccessLog) error {
	log.MachineUser = "codeintel-autoindexing"
	log.UserID = nil
	return e.ExecutorSecretAccessLogCreator.Create(ctx, log)
}

func transformRecord(ctx context.Context, db database.DB, autoIndexJob uploadsshared.AutoIndexJob, resourceMetadata handler.ResourceMetadata, accessToken string) (apiclient.Job, error) {
	resourceEnvironment := makeResourceEnvironment(resourceMetadata)

	var secrets []*database.ExecutorSecret
	var err error
	if len(autoIndexJob.RequestedEnvVars) > 0 {
		secretsStore := db.ExecutorSecrets(keyring.Default().ExecutorSecretKey)
		secrets, _, err = secretsStore.List(ctx, database.ExecutorSecretScopeCodeIntel, database.ExecutorSecretsListOpts{
			// Note: No namespace set, codeintel secrets are only available in the global namespace for now.
			Keys: autoIndexJob.RequestedEnvVars,
		})
		if err != nil {
			return apiclient.Job{}, err
		}
	}

	// And build the env vars from the secrets.
	secretEnvVars := make([]string, len(secrets))
	redactedEnvVars := make(map[string]string, len(secrets))
	secretStore := &accessLogTransformer{db.ExecutorSecretAccessLogs()}
	for i, secret := range secrets {
		// Get the secret value. This also creates an access log entry in the
		// name of the user.
		val, err := secret.Value(ctx, secretStore)
		if err != nil {
			return apiclient.Job{}, err
		}

		secretEnvVars[i] = fmt.Sprintf("%s=%s", secret.Key, val)
		// We redact secret values as ${{ secrets.NAME }}.
		redactedEnvVars[val] = fmt.Sprintf("${{ secrets.%s }}", secret.Key)
	}

	envVars := append(resourceEnvironment, secretEnvVars...)

	dockerSteps := make([]apiclient.DockerStep, 0, len(autoIndexJob.DockerSteps)+2)
	for i, dockerStep := range autoIndexJob.DockerSteps {
		dockerSteps = append(dockerSteps, apiclient.DockerStep{
			Key:      fmt.Sprintf("pre-index.%d", i),
			Image:    dockerStep.Image,
			Commands: dockerStep.Commands,
			Dir:      dockerStep.Root,
			Env:      envVars,
		})
	}

	if autoIndexJob.Indexer != "" {
		dockerSteps = append(dockerSteps, apiclient.DockerStep{
			Key:      "indexer",
			Image:    autoIndexJob.Indexer,
			Commands: append(autoIndexJob.LocalSteps, shellquote.Join(autoIndexJob.IndexerArgs...)),
			Dir:      autoIndexJob.Root,
			Env:      envVars,
		})
	}

	frontendURL := conf.ExecutorsFrontendURL()
	authorizationHeader := makeAuthHeaderValue(accessToken)
	redactedAuthorizationHeader := makeAuthHeaderValue("REDACTED")
	srcCliImage := fmt.Sprintf("%s:%s", conf.ExecutorsSrcCLIImage(), conf.ExecutorsSrcCLIImageTag())

	root := autoIndexJob.Root
	if root == "" {
		root = "."
	}

	outfile := autoIndexJob.Outfile
	if outfile == "" {
		outfile = defaultOutfile
	}

	// TODO: Temporary workaround. LSIF-go needs tags, but they make git fetching slower.
	fetchTags := strings.HasPrefix(autoIndexJob.Indexer, conf.ExecutorsLsifGoImage())

	dockerSteps = append(dockerSteps, apiclient.DockerStep{
		Key:   "upload",
		Image: srcCliImage,
		Commands: []string{
			shellquote.Join(
				"src",
				"code-intel",
				"upload",
				"-no-progress",
				"-repo", autoIndexJob.RepositoryName,
				"-commit", autoIndexJob.Commit,
				"-root", root,
				"-upload-route", uploadRoute,
				"-file", outfile,
				"-associated-index-id", strconv.Itoa(autoIndexJob.ID),
			),
		},
		Dir: autoIndexJob.Root,
		Env: []string{
			fmt.Sprintf("SRC_ENDPOINT=%s", frontendURL),
			fmt.Sprintf("SRC_HEADER_AUTHORIZATION=%s", authorizationHeader),
		},
	})

	allRedactedValues := map[string]string{
		// 🚨 SECURITY: Catch leak of authorization header.
		authorizationHeader: redactedAuthorizationHeader,

		// 🚨 SECURITY: Catch uses of fragments pulled from auth header to
		// construct another target (in src-cli). We only pass the
		// Authorization header to src-cli, which we trust not to ship the
		// values to a third party, but not to trust to ensure the values
		// are absent from the command's stdout or stderr streams.
		accessToken: "PASSWORD_REMOVED",
	}
	// 🚨 SECURITY: Catch uses of executor secrets from the executor secret store
	maps.Copy(allRedactedValues, redactedEnvVars)

	aj := apiclient.Job{
		ID:             autoIndexJob.ID,
		Commit:         autoIndexJob.Commit,
		RepositoryName: autoIndexJob.RepositoryName,
		ShallowClone:   true,
		FetchTags:      fetchTags,
		DockerSteps:    dockerSteps,
		RedactedValues: allRedactedValues,
	}

	// Append docker auth config.
	esStore := db.ExecutorSecrets(keyring.Default().ExecutorSecretKey)
	secrets, _, err = esStore.List(ctx, database.ExecutorSecretScopeCodeIntel, database.ExecutorSecretsListOpts{
		// Codeintel only has a global namespace for now.
		NamespaceUserID: 0,
		NamespaceOrgID:  0,
		Keys:            []string{"DOCKER_AUTH_CONFIG"},
	})
	if err != nil {
		return apiclient.Job{}, err
	}
	if len(secrets) == 1 {
		val, err := secrets[0].Value(ctx, secretStore)
		if err != nil {
			return apiclient.Job{}, err
		}
		if err := json.Unmarshal([]byte(val), &aj.DockerAuthConfig); err != nil {
			return aj, err
		}
	}

	return aj, nil
}

const (
	defaultMemory    = "12G"
	defaultDiskSpace = "20G"
)

func makeResourceEnvironment(resourceMetadata handler.ResourceMetadata) []string {
	env := []string{}
	addBytesValuesVariables := func(value, defaultValue, prefix string) {
		if value == "" {
			value = defaultValue
		}

		if parsed, _ := datasize.ParseString(value); parsed.Bytes() != 0 {
			env = append(
				env,
				fmt.Sprintf("%s=%s", prefix, parsed.HumanReadable()),
				fmt.Sprintf("%s_GB=%d", prefix, int(parsed.GBytes())),
				fmt.Sprintf("%s_MB=%d", prefix, int(parsed.MBytes())),
			)
		}
	}

	if cpus := resourceMetadata.NumCPUs; cpus != 0 {
		env = append(env, fmt.Sprintf("VM_CPUS=%d", cpus))
	}
	addBytesValuesVariables(resourceMetadata.Memory, defaultMemory, "VM_MEM")
	addBytesValuesVariables(resourceMetadata.DiskSpace, defaultDiskSpace, "VM_DISK")

	return env
}

func makeAuthHeaderValue(token string) string {
	return fmt.Sprintf("%s %s", schemeExecutorToken, token)
}
