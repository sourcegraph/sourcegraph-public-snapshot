package batches

import (
	"fmt"
	"net/url"

	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	apiclient "github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
)

func transformRecord(exec *btypes.BatchSpecExecution, config *Config) (apiclient.Job, error) {
	srcEndpoint, err := makeURL(config.FrontendURL, config.FrontendUsername, config.FrontendPassword)
	if err != nil {
		return apiclient.Job{}, err
	}

	redactedSrcEndpoint, err := makeURL(config.FrontendURL, "USERNAME_REMOVED", "PASSWORD_REMOVED")
	if err != nil {
		return apiclient.Job{}, err
	}

	return apiclient.Job{
		ID: int(exec.ID),
		// Commit:         exec.Commit,
		// RepositoryName: exec.RepositoryName,
		DockerSteps:         []apiclient.DockerStep{},
		VirtualMachineFiles: map[string]string{"spec.yml": exec.BatchSpec},
		CliSteps: []apiclient.CliStep{
			{
				Commands: []string{
					"batch", "preview",
					"-f", "spec.yml",
				},
				Dir: ".",
				Env: []string{
					fmt.Sprintf("SRC_ENDPOINT=%s", srcEndpoint),
					"SRC_ACCESS_TOKEN=3f2bbf0202e227379f7fddda2bf6a6ada8e7833b",
				},
			},
		},
		RedactedValues: map[string]string{
			// 🚨 SECURITY: Catch leak of upload endpoint. This is necessary in addition
			// to the below in case the username or password contains illegal URL characters,
			// which are then urlencoded and are not replaceable via byte comparison.
			srcEndpoint: redactedSrcEndpoint,

			// 🚨 SECURITY: Catch uses of fragments pulled from URL to construct another target
			// (in src-cli). We only pass the constructed URL to src-cli, which we trust not to
			// ship the values to a third party, but not to trust to ensure the values are absent
			// from the command's stdout or stderr streams.
			config.FrontendUsername: "USERNAME_REMOVED",
			config.FrontendPassword: "PASSWORD_REMOVED",
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
