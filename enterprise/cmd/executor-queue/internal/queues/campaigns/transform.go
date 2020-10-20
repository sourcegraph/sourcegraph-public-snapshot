package campaigns

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/apiworker/apiclient"
)

// TODO - make these configurable or part of the job payload
const namespace = "eric"
const srcEndpoint = "http://127.0.0.1:3080"
const srcAccessToken = "bda6886bdfab5090e1be5c7cc891a8a6b473c2c0"

func transformRecord(job CampaignApplyJob) (apiclient.Job, error) {
	cliSteps := []apiclient.CliStep{
		{
			Commands: []string{
				"campaigns",
				"apply",
				"-f",
				"spec.yaml",
				"-namespace", namespace,
			},
			Env: []string{
				fmt.Sprintf("SRC_ENDPOINT=%s", srcEndpoint),
				fmt.Sprintf("SRC_ACCESS_TOKEN=%s", srcAccessToken),
			},
		},
	}

	return apiclient.Job{
		ID:                  job.ID,
		VirtualMachineFiles: map[string]string{"spec.yaml": job.Spec},
		CliSteps:            cliSteps,
	}, nil
}
