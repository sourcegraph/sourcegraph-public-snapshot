package apiclient

import "github.com/sourcegraph/sourcegraph/internal/workerutil"

// Job describes a series of steps to perform within an executor.
type Job struct {
	// ID is the unique identifier of a job within the source queue. Note
	// that different queues can share identifiers.
	ID int `json:"id"`

	// RepositoryName is the name of the repository to be cloned into the
	// workspace prior to job execution.
	RepositoryName string `json:"repositoryName"`

	// Commit is the revhash that should be checked out prior to job execution.
	Commit string `json:"commit"`

	// VirtualMachineFiles is a map from file names to content. Each entry in
	// this map will be written into the workspace prior to job execution.
	VirtualMachineFiles map[string]string `json:"files"`

	// DockerSteps describe a series of docker run commands to be invoked in the
	// workspace. This may be done inside or outside of a Firecracker virtual
	// machine.
	DockerSteps []DockerStep `json:"dockerSteps"`

	// CliSteps describe a series of src commands to be invoked in the workspace.
	// These run after all docker commands have been completed successfully. This
	// may be done inside or outside of a Firecracker virtual machine.
	CliSteps []CliStep `json:"cliSteps"`
}

func (j Job) RecordID() int {
	return j.ID
}

type DockerStep struct {
	// Image specifies the docker image.
	Image string `json:"image"`

	// Commands specifies the arguments supplied to the end of a docker run command.
	Commands []string `json:"commands"`

	// Dir specifies the target working directory.
	Dir string `json:"dir"`

	// Env specifies a set of NAME=value pairs to supply to the docker command.
	Env []string `json:"env"`
}

type CliStep struct {
	// Commands specifies the arguments supplied to the src command.
	Commands []string `json:"command"`

	// Dir specifies the target working directory.
	Dir string `json:"dir"`

	// Env specifies a set of NAME=value pairs to supply to the src command.
	Env []string `json:"env"`
}

type DequeueRequest struct {
	ExecutorName string `json:"executorName"`
}

type AddExecutionLogEntryRequest struct {
	ExecutorName string `json:"executorName"`
	JobID        int    `json:"jobId"`
	workerutil.ExecutionLogEntry
}

type MarkCompleteRequest struct {
	ExecutorName string `json:"executorName"`
	JobID        int    `json:"jobId"`
}

type MarkErroredRequest struct {
	ExecutorName string `json:"executorName"`
	JobID        int    `json:"jobId"`
	ErrorMessage string `json:"errorMessage"`
}

type HeartbeatRequest struct {
	ExecutorName string `json:"executorName"`
	JobIDs       []int  `json:"jobIds"`
}
