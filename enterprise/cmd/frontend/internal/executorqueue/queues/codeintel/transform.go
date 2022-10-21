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
	// Create a transformer over the commands that will be invoked inside of a docker container.
	// This replaces string literals such as `$VM_MEM` with the actual resource capacity of the
	// VM that will run this job.
	injectResourceCapacities := makeResourceCapacityInjector(resourceMetadata)

	dockerSteps := make([]apiclient.DockerStep, 0, len(index.DockerSteps)+2)
	for i, dockerStep := range index.DockerSteps {
		dockerSteps = append(dockerSteps, apiclient.DockerStep{
			Key:      fmt.Sprintf("pre-index.%d", i),
			Image:    dockerStep.Image,
			Commands: injectResourceCapacities(dockerStep.Commands),
			Dir:      dockerStep.Root,
			Env:      nil,
		})
	}

	if index.Indexer != "" {
		dockerSteps = append(dockerSteps, apiclient.DockerStep{
			Key:   "indexer",
			Image: index.Indexer,
			// Ensure we do string replacement BEFORE shellquoting, otherwise we'll end up
			// escaping the `$` at the beginning of our replacement tokens, but not replace
			// the escape.
			Commands: append(injectResourceCapacities(index.LocalSteps), shellquote.Join(injectResourceCapacities(index.IndexerArgs)...)),
			Dir:      index.Root,
			Env:      nil,
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

func makeResourceCapacityInjector(resourceMetadata handler.ResourceMetadata) func([]string) []string {
	replacements := []string{}
	addBytesValuesReplacement := func(value, defaultValue, prefix string) {
		if value == "" {
			value = defaultValue
		}

		var gb, mb, human string
		if parsed, _ := datasize.ParseString(value); parsed.Bytes() != 0 {
			gb = strconv.Itoa(int(parsed.GBytes()))
			mb = strconv.Itoa(int(parsed.MBytes()))
			human = parsed.HumanReadable()
		}

		replacements = append(
			replacements,
			fmt.Sprintf("$%s_GB", prefix), gb,
			fmt.Sprintf("$%s_MB", prefix), mb,
			// N.B.: Ensure substring of longer keys come later
			fmt.Sprintf("$%s", prefix), human,
		)
	}

	cpus := resourceMetadata.NumCPUs
	if cpus != 0 {
		replacements = append(replacements, "$VM_CPUS", strconv.Itoa(cpus))
	}
	addBytesValuesReplacement(resourceMetadata.Memory, defaultMemory, "VM_MEM")
	addBytesValuesReplacement(resourceMetadata.DiskSpace, defaultDiskSpace, "VM_DISK")

	replacer := strings.NewReplacer(replacements...)

	return func(vs []string) []string {
		for i, v := range vs {
			vs[i] = replacer.Replace(v)
		}

		return vs
	}
}

func makeAuthHeaderValue(token string) string {
	return fmt.Sprintf("%s %s", schemeExecutorToken, token)
}
