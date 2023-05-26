package types

import (
	"encoding/json"
	"strconv"
	"time"
)

// Job describes a series of steps to perform within an executor.
type Job struct {
	// Version is used to version the shape of the Job payload, so that older
	// executors can still talk to Sourcegraph. The dequeue endpoint takes an
	// executor version to determine which maximum version said executor supports.
	Version int `json:"version,omitempty"`

	// ID is the unique identifier of a job within the source queue. Note
	// that different queues can share identifiers.
	ID int `json:"id"`

	// Queue contains the name of the source queue.
	Queue string `json:"queue,omitempty"`

	// Token is the authentication token for the specific Job.
	Token string `json:"Token"`

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

	// VirtualMachineFiles is a map from file names to content. Each entry in
	// this map will be written into the workspace prior to job execution.
	// The file paths must be relative and within the working directory.
	VirtualMachineFiles map[string]VirtualMachineFile `json:"files"`

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

	// DockerAuthConfig can optionally set the content of the docker CLI config to be used
	// when spawning containers. Used to authenticate to private registries. When set, this
	// takes precedence over a potentially configured EXECUTOR_DOCKER_AUTH_CONFIG environment
	// variable.
	DockerAuthConfig DockerAuthConfig `json:"dockerAuthConfig,omitempty"`
}

func (j Job) MarshalJSON() ([]byte, error) {
	if j.Version == 2 {
		v2 := v2Job{
			Version:             j.Version,
			ID:                  j.ID,
			Token:               j.Token,
			Queue:               j.Queue,
			RepositoryName:      j.RepositoryName,
			RepositoryDirectory: j.RepositoryDirectory,
			Commit:              j.Commit,
			FetchTags:           j.FetchTags,
			ShallowClone:        j.ShallowClone,
			SparseCheckout:      j.SparseCheckout,
			DockerSteps:         j.DockerSteps,
			CliSteps:            j.CliSteps,
			RedactedValues:      j.RedactedValues,
			DockerAuthConfig:    j.DockerAuthConfig,
		}
		v2.VirtualMachineFiles = make(map[string]v2VirtualMachineFile, len(j.VirtualMachineFiles))
		for k, v := range j.VirtualMachineFiles {
			v2.VirtualMachineFiles[k] = v2VirtualMachineFile(v)
		}
		return json.Marshal(v2)
	}
	v1 := v1Job{
		ID:                  j.ID,
		Token:               j.Token,
		Queue:               j.Queue,
		RepositoryName:      j.RepositoryName,
		RepositoryDirectory: j.RepositoryDirectory,
		Commit:              j.Commit,
		FetchTags:           j.FetchTags,
		ShallowClone:        j.ShallowClone,
		SparseCheckout:      j.SparseCheckout,
		DockerSteps:         j.DockerSteps,
		CliSteps:            j.CliSteps,
		RedactedValues:      j.RedactedValues,
	}
	v1.VirtualMachineFiles = make(map[string]v1VirtualMachineFile, len(j.VirtualMachineFiles))
	for k, v := range j.VirtualMachineFiles {
		v1.VirtualMachineFiles[k] = v1VirtualMachineFile{
			Content:    string(v.Content),
			Bucket:     v.Bucket,
			Key:        v.Key,
			ModifiedAt: v.ModifiedAt,
		}
	}
	return json.Marshal(v1)
}

func (j *Job) UnmarshalJSON(data []byte) error {
	var version versionJob
	if err := json.Unmarshal(data, &version); err != nil {
		return err
	}
	if version.Version == 2 {
		var v2 v2Job
		if err := json.Unmarshal(data, &v2); err != nil {
			return err
		}
		j.Version = v2.Version
		j.ID = v2.ID
		j.Token = v2.Token
		j.Queue = v2.Queue
		j.RepositoryName = v2.RepositoryName
		j.RepositoryDirectory = v2.RepositoryDirectory
		j.Commit = v2.Commit
		j.FetchTags = v2.FetchTags
		j.ShallowClone = v2.ShallowClone
		j.SparseCheckout = v2.SparseCheckout
		j.VirtualMachineFiles = make(map[string]VirtualMachineFile, len(v2.VirtualMachineFiles))
		for k, v := range v2.VirtualMachineFiles {
			j.VirtualMachineFiles[k] = VirtualMachineFile(v)
		}
		j.DockerSteps = v2.DockerSteps
		j.CliSteps = v2.CliSteps
		j.RedactedValues = v2.RedactedValues
		j.DockerAuthConfig = v2.DockerAuthConfig
		return nil
	}
	var v1 v1Job
	if err := json.Unmarshal(data, &v1); err != nil {
		return err
	}
	j.ID = v1.ID
	j.Token = v1.Token
	j.Queue = v1.Queue
	j.RepositoryName = v1.RepositoryName
	j.RepositoryDirectory = v1.RepositoryDirectory
	j.Commit = v1.Commit
	j.FetchTags = v1.FetchTags
	j.ShallowClone = v1.ShallowClone
	j.SparseCheckout = v1.SparseCheckout
	j.VirtualMachineFiles = make(map[string]VirtualMachineFile, len(v1.VirtualMachineFiles))
	for k, v := range v1.VirtualMachineFiles {
		j.VirtualMachineFiles[k] = VirtualMachineFile{
			Content:    []byte(v.Content),
			Bucket:     v.Bucket,
			Key:        v.Key,
			ModifiedAt: v.ModifiedAt,
		}
	}
	j.DockerSteps = v1.DockerSteps
	j.CliSteps = v1.CliSteps
	j.RedactedValues = v1.RedactedValues
	return nil
}

type versionJob struct {
	Version int `json:"version,omitempty"`
}

type v2Job struct {
	Version             int                             `json:"version,omitempty"`
	ID                  int                             `json:"id"`
	Token               string                          `json:"token"`
	Queue               string                          `json:"queue,omitempty"`
	RepositoryName      string                          `json:"repositoryName"`
	RepositoryDirectory string                          `json:"repositoryDirectory"`
	Commit              string                          `json:"commit"`
	FetchTags           bool                            `json:"fetchTags"`
	ShallowClone        bool                            `json:"shallowClone"`
	SparseCheckout      []string                        `json:"sparseCheckout"`
	VirtualMachineFiles map[string]v2VirtualMachineFile `json:"files"`
	DockerSteps         []DockerStep                    `json:"dockerSteps"`
	CliSteps            []CliStep                       `json:"cliSteps"`
	RedactedValues      map[string]string               `json:"redactedValues"`
	DockerAuthConfig    DockerAuthConfig                `json:"dockerAuthConfig,omitempty"`
}

type v1Job struct {
	ID                  int                             `json:"id"`
	Token               string                          `json:"token"`
	Queue               string                          `json:"queue,omitempty"`
	RepositoryName      string                          `json:"repositoryName"`
	RepositoryDirectory string                          `json:"repositoryDirectory"`
	Commit              string                          `json:"commit"`
	FetchTags           bool                            `json:"fetchTags"`
	ShallowClone        bool                            `json:"shallowClone"`
	SparseCheckout      []string                        `json:"sparseCheckout"`
	VirtualMachineFiles map[string]v1VirtualMachineFile `json:"files"`
	DockerSteps         []DockerStep                    `json:"dockerSteps"`
	CliSteps            []CliStep                       `json:"cliSteps"`
	RedactedValues      map[string]string               `json:"redactedValues"`
}

// VirtualMachineFile is a file that will be written to the VM. A file can contain the raw content of the file or
// specify the coordinates of the file in the file store.
// A file must either contain Content or a Bucket/Key. If neither are provided, an empty file is written.
type VirtualMachineFile struct {
	// Content is the raw content of the file. If not provided, the file is retrieved from the file store based on the
	// configured Bucket and Key. If Content, Bucket, and Key are not provided, an empty file will be written.
	Content []byte `json:"content,omitempty"`

	// Bucket is the bucket in the files store the file belongs to. A Key must also be configured.
	Bucket string `json:"bucket,omitempty"`

	// Key the key or coordinates of the files in the Bucket. The Bucket must be configured.
	Key string `json:"key,omitempty"`

	// ModifiedAt an optional field that specifies when the file was last modified.
	ModifiedAt time.Time `json:"modifiedAt,omitempty"`
}

type v2VirtualMachineFile struct {
	Content    []byte    `json:"content,omitempty"`
	Bucket     string    `json:"bucket,omitempty"`
	Key        string    `json:"key,omitempty"`
	ModifiedAt time.Time `json:"modifiedAt,omitempty"`
}

type v1VirtualMachineFile struct {
	Content    string    `json:"content,omitempty"`
	Bucket     string    `json:"bucket,omitempty"`
	Key        string    `json:"key,omitempty"`
	ModifiedAt time.Time `json:"modifiedAt,omitempty"`
}

func (j Job) RecordID() int {
	return j.ID
}

func (j Job) RecordUID() string {
	uid := strconv.Itoa(j.ID)
	// outside of multi-queue executors, jobs aren't guaranteed to have a queue specified
	if j.Queue != "" {
		uid += "-" + j.Queue
	}
	return uid
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

// CliStep is a step that runs a src-cli command.
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

// DockerAuthConfig represents a subset of the docker cli config with the necessary
// fields to make authentication work.
type DockerAuthConfig struct {
	// Auths is a map from registry URL to auth object.
	Auths DockerAuthConfigAuths `json:"auths,omitempty"`
}

// DockerAuthConfigAuths maps a registry URL to an auth object.
type DockerAuthConfigAuths map[string]DockerAuthConfigAuth

// DockerAuthConfigAuth is a single registry's auth configuration.
type DockerAuthConfigAuth struct {
	// Auth is the base64 encoded credential in the format user:password.
	Auth []byte `json:"auth"`
}
