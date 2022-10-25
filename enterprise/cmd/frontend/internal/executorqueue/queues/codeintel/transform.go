package codeintel

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/c2h5oh/datasize"
	"github.com/kballard/go-shellquote"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/handler"
	apiclient "github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/conf"
)

const defaultOutfile = "dump.lsif"
const uploadRoute = "/.executors/lsif/upload"
const schemeExecutorToken = "token-executor"

func transformRecord(index types.Index, resourceMetadata handler.ResourceMetadata, accessToken string) (apiclient.Job, error) {
	resourceEnvironment := makeResourceEnvironment(resourceMetadata)

	dockerSteps := make([]apiclient.DockerStep, 0, len(index.DockerSteps)+2)
	for i, dockerStep := range index.DockerSteps {
		dockerSteps = append(dockerSteps, apiclient.DockerStep{
			Key:      fmt.Sprintf("pre-index.%d", i),
			Image:    dockerStep.Image,
			Commands: dockerStep.Commands,
			Dir:      dockerStep.Root,
			Env:      resourceEnvironment,
		})
	}

	if index.Indexer != "" {
		dockerSteps = append(dockerSteps, apiclient.DockerStep{
			Key:      "indexer",
			Image:    index.Indexer,
			Commands: append(index.LocalSteps, shellquote.Join(index.IndexerArgs...)),
			Dir:      index.Root,
			Env:      resourceEnvironment,
		})
	}

	frontendURL := conf.ExecutorsFrontendURL()
	authorizationHeader := makeAuthHeaderValue(accessToken)
	redactedAuthorizationHeader := makeAuthHeaderValue("REDACTED")
	srcCliImage := fmt.Sprintf("%s:%s", conf.ExecutorsSrcCLIImage(), conf.ExecutorsSrcCLIImageTag())

	root := index.Root
	if root == "" {
		root = "."
	}

	outfile := index.Outfile
	if outfile == "" {
		outfile = defaultOutfile
	}

	fetchTags := false
	// TODO: Temporary workaround. LSIF-go needs tags, but they make git fetching slower.
	if strings.HasPrefix(index.Indexer, "sourcegraph/lsif-go") {
		fetchTags = true
	}

	dockerSteps = append(dockerSteps, apiclient.DockerStep{
		Key:   "upload",
		Image: srcCliImage,
		Commands: []string{
			shellquote.Join(
				"src",
				"lsif",
				"upload",
				"-no-progress",
				"-repo", index.RepositoryName,
				"-commit", index.Commit,
				"-root", root,
				"-upload-route", uploadRoute,
				"-file", outfile,
				"-associated-index-id", strconv.Itoa(index.ID),
			),
		},
		Dir: index.Root,
		Env: []string{
			fmt.Sprintf("SRC_ENDPOINT=%s", frontendURL),
			fmt.Sprintf("SRC_HEADER_AUTHORIZATION=%s", authorizationHeader),
		},
	})

	return apiclient.Job{
		ID:             index.ID,
		Commit:         index.Commit,
		RepositoryName: index.RepositoryName,
		ShallowClone:   true,
		FetchTags:      fetchTags,
		DockerSteps:    dockerSteps,
		RedactedValues: map[string]string{
			// 🚨 SECURITY: Catch leak of authorization header.
			authorizationHeader: redactedAuthorizationHeader,

			// 🚨 SECURITY: Catch uses of fragments pulled from auth header to
			// construct another target (in src-cli). We only pass the
			// Authorization header to src-cli, which we trust not to ship the
			// values to a third party, but not to trust to ensure the values
			// are absent from the command's stdout or stderr streams.
			accessToken: "PASSWORD_REMOVED",
		},
	}, nil
}

const defaultMemory = "12G"
const defaultDiskSpace = "20G"

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
