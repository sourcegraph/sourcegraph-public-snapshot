package codeintel

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kballard/go-shellquote"

	apiclient "github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/conf"
)

const defaultOutfile = "dump.lsif"
const uploadRoute = "/.executors/lsif/upload"
const schemeExecutorToken = "token-executor"

func transformRecord(index types.Index, accessToken string) (apiclient.Job, error) {
	dockerSteps := make([]apiclient.DockerStep, 0, len(index.DockerSteps)+2)
	for i, dockerStep := range index.DockerSteps {
		dockerSteps = append(dockerSteps, apiclient.DockerStep{
			Key:      fmt.Sprintf("pre-index.%d", i),
			Image:    dockerStep.Image,
			Commands: dockerStep.Commands,
			Dir:      dockerStep.Root,
			Env:      nil,
		})
	}

	if index.Indexer != "" {
		dockerSteps = append(dockerSteps, apiclient.DockerStep{
			Key:      "indexer",
			Image:    index.Indexer,
			Commands: append(index.LocalSteps, shellquote.Join(index.IndexerArgs...)),
			Dir:      index.Root,
			Env:      nil,
		})
	}

	frontendURL := conf.ExecutorsFrontendURL()
	authorizationHeader := makeAuthHeaderValue(accessToken)
	redactedAuthorizationHeader := makeAuthHeaderValue("REDACTED")

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

	return apiclient.Job{
		ID:             index.ID,
		Commit:         index.Commit,
		RepositoryName: index.RepositoryName,
		ShallowClone:   true,
		FetchTags:      fetchTags,
		DockerSteps:    dockerSteps,
		CliSteps: []apiclient.CliStep{
			{
				Key: "upload",
				Commands: []string{
					"lsif", "upload",
					"-no-progress",
					"-repo", index.RepositoryName,
					"-commit", index.Commit,
					"-root", root,
					"-upload-route", uploadRoute,
					"-file", outfile,
					"-associated-index-id", strconv.Itoa(index.ID),
				},
				Dir: index.Root,
				Env: []string{
					fmt.Sprintf("SRC_ENDPOINT=%s", frontendURL),
					fmt.Sprintf("SRC_HEADER_AUTHORIZATION=%s", authorizationHeader),
				},
			},
		},
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

func makeAuthHeaderValue(token string) string {
	return fmt.Sprintf("%s %s", schemeExecutorToken, token)
}
