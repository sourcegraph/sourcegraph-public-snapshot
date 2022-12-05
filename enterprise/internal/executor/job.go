package executor

import (
	"encoding/json"
	"fmt"
	"time"
)

type VersionedJob struct {
	Job
}

type Job interface {
	Version() int
	RecordID() int
}

func (v VersionedJob) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.Job)
}

func (v *VersionedJob) UnmarshalJSON(bytes []byte) error {
	var ver unmarshalVersion
	if err := json.Unmarshal(bytes, &ver); err != nil {
		return err
	}
	switch ver.Version {
	case 0:
		fallthrough
	case 1:
		var v1 V1Job
		if err := json.Unmarshal(bytes, &v1); err != nil {
			return err
		}
		v1.JobVersion = 1
		v.Job = v1
	case 2:
		var v2 V2Job
		if err := json.Unmarshal(bytes, &v2); err != nil {
			return err
		}
		v2.JobVersion = 2
		v.Job = v2
	default:
		return fmt.Errorf("unknown job version %d", ver.Version)
	}
	return nil
}

type unmarshalVersion struct {
	Version int `json:"version"`
}

type BaseJob struct {
	// Version is used to version the shape of the Job payload, so that older
	// executors can still talk to Sourcegraph. The dequeue endpoint takes an
	// executor version to determine which maximum version said executor supports.
	JobVersion int `json:"version,omitempty"`
	// ID is the unique identifier of a job within the source queue. Note
	// that different queues can share identifiers.
	ID int `json:"id"`
	// RepositoryName is the name of the repository to be cloned into the
	// workspace prior to job execution.
	RepositoryName string `json:"repositoryName"`
	// RepositoryDirectory is the relative path to which the repo is cloned. If
	// unset, defaults to the workspace root.
	RepositoryDirectory string `json:"repositoryDirectory"`
	// Commit is the revhash that should be checked out prior to job execution.
	Commit string `json:"commit"`
	// FetchTags, when true also fetches tags from the remote.
	FetchTags bool `json:"fetchTags"`
	// ShallowClone, when true speeds up repo cloning by fetching only the target commit
	// and no tags.
	ShallowClone bool `json:"shallowClone"`
	// SparseCheckout denotes the path patterns to check out. This can be used to fetch
	// only a part of a repository.
	SparseCheckout []string `json:"sparseCheckout"`
	// DockerSteps describe a series of docker run commands to be invoked in the
	// workspace. This may be done inside or outside a Firecracker virtual
	// machine.
	DockerSteps []DockerStep `json:"dockerSteps"`
	// CliSteps describe a series of src commands to be invoked in the workspace.
	// These run after all docker commands have been completed successfully. This
	// may be done inside or outside a Firecracker virtual machine.
	CliSteps []CliStep `json:"cliSteps"`
	// RedactedValues is a map from strings to replace to their replacement in the command
	// output before sending it to the underlying job store. This should contain all worker
	// environment variables, as well as secret values passed along with the dequeued job
	// payload, which may be sensitive (e.g. shared API tokens, URLs with credentials).
	RedactedValues map[string]string `json:"redactedValues"`
}

func (b BaseJob) Version() int {
	return b.JobVersion
}

func (b BaseJob) RecordID() int {
	return b.ID
}

type V1Job struct {
	BaseJob
	VirtualMachineFiles map[string]V1VirtualMachineFile `json:"files"`
}

type V2Job struct {
	BaseJob
	VirtualMachineFiles map[string]V2VirtualMachineFile `json:"files"`
}

// BaseVirtualMachineFile is a file that will be written to the VM. A file can contain the raw content of the file or
// specify the coordinates of the file in the file store.
// A file must either contain Content or a Bucket/Key. If neither are provided, an empty file is written.
type BaseVirtualMachineFile struct {
	// Bucket is the bucket in the files store the file belongs to. A Key must also be configured.
	Bucket string `json:"bucket,omitempty"`
	// Key the key or coordinates of the files in the Bucket. The Bucket must be configured.
	Key string `json:"key,omitempty"`
	// ModifiedAt an optional field that specifies when the file was last modified.
	ModifiedAt time.Time `json:"modifiedAt,omitempty"`
}

type V1VirtualMachineFile struct {
	BaseVirtualMachineFile
	Content string `json:"content,omitempty"`
}

type V2VirtualMachineFile struct {
	BaseVirtualMachineFile
	Content []byte `json:"content,omitempty"`
}

type DockerStep struct {
	// Key is a unique identifier of the step. It can be used to retrieve the
	// associated log entry.
	Key string `json:"key,omitempty"`
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
	// Key is a unique identifier of the step. It can be used to retrieve the
	// associated log entry.
	Key string `json:"key,omitempty"`
	// Commands specifies the arguments supplied to the src command.
	Commands []string `json:"command"`
	// Dir specifies the target working directory.
	Dir string `json:"dir"`
	// Env specifies a set of NAME=value pairs to supply to the src command.
	Env []string `json:"env"`
}
