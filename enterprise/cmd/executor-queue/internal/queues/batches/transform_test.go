package batches

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor-queue/internal/config"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	apiclient "github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
)

func TestTransformRecord(t *testing.T) {
	testBatchSpec := `batchSpec: yeah`
	index := &btypes.BatchSpecExecution{
		ID:        42,
		BatchSpec: testBatchSpec,
	}
	config := &Config{
		Shared: &config.SharedConfig{
			FrontendURL:      "https://test.io",
			FrontendUsername: "test*",
			FrontendPassword: "hunter2",
		},
	}

	job, err := transformRecord(index, config)
	if err != nil {
		t.Fatalf("unexpected error transforming record: %s", err)
	}

	expected := apiclient.Job{
		ID:                  42,
		VirtualMachineFiles: map[string]string{"spec.yml": testBatchSpec},
		CliSteps: []apiclient.CliStep{
			{
				Commands: []string{
					"batch", "preview", "-f", "spec.yml",
				},
				Dir: ".",
				Env: []string{"SRC_ENDPOINT=https://test%2A:hunter2@test.io"},
			},
		},
		RedactedValues: map[string]string{
			"https://test%2A:hunter2@test.io": "https://USERNAME_REMOVED:PASSWORD_REMOVED@test.io",
			"test*":                           "USERNAME_REMOVED",
			"hunter2":                         "PASSWORD_REMOVED",
		},
	}
	if diff := cmp.Diff(expected, job); diff != "" {
		t.Errorf("unexpected job (-want +got):\n%s", diff)
	}
}
