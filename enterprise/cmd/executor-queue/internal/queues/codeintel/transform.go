package codeintel

import (
	"fmt"
	"net/url"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/apiworker/apiclient"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
)

const defaultOutfile = "dump.lsif"
const uploadRoute = "/.executors/lsif/upload"

func transformRecord(index store.Index, config *Config) (apiclient.Job, error) {
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
			Commands: index.IndexerArgs,
			Dir:      index.Root,
			Env:      nil,
		})
	}

	srcEndpoint, err := makeURL(config.FrontendURL, config.FrontendUsername, config.FrontendPassword)
	if err != nil {
		return apiclient.Job{}, err
	}

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
				},
				Dir: index.Root,
				Env: []string{
					fmt.Sprintf("SRC_ENDPOINT=%s", srcEndpoint),
				},
			},
		},
	}, nil
}

func makeURL(base, username, password string) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}

	u.User = url.UserPassword(username, password)
	return u.String(), nil
}
