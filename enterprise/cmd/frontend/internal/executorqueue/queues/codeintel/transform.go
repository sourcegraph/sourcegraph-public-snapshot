package codeintel

import (
	"fmt"
	"strconv"
	"strings"

	apiclient "github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	store "github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/conf"
)

const defaultOutfile = "dump.lsif"
const uploadRoute = "/.executors/lsif/upload"
const schemeExecutorToken = "token-executor"

func transformRecord(index store.Index, accessToken string) (apiclient.Job, error) {
	dockerSteps := make([]apiclient.DockerStep, 0, len(index.DockerSteps)+2)
	for _, dockerStep := range index.DockerSteps {
		dockerSteps = append(dockerSteps, apiclient.DockerStep{
			Image:    dockerStep.Image,
			Commands: dockerStep.Commands,
			Dir:      dockerStep.Root,
			Env:      nil,
		})
	}

	if index.Indexer != "" {
		dockerSteps = append(dockerSteps, apiclient.DockerStep{
			Image:    index.Indexer,
			Commands: append(index.LocalSteps, strings.Join(index.IndexerArgs, " ")),
			Dir:      index.Root,
			Env:      nil,
		})
	}

	frontendURL := conf.Get().ExternalURL

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

	return apiclient.Job{
		ID:             index.ID,
		Commit:         index.Commit,
		RepositoryName: index.RepositoryName,
		DockerSteps:    dockerSteps,
		CliSteps: []apiclient.CliStep{
			{
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
