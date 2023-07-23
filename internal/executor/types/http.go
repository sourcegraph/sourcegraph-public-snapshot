package types

import (
	"encoding/json"
	"strconv"

	"github.com/sourcegraph/sourcegraph/internal/executor"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type DequeueRequest struct {
	Queues       []string `json:"queues,omitempty"`
	ExecutorName string   `json:"executorName"`
	Version      string   `json:"version"`
	NumCPUs      int      `json:"numCPUs,omitempty"`
	Memory       string   `json:"memory,omitempty"`
	DiskSpace    string   `json:"diskSpace,omitempty"`
}

type JobOperationRequest struct {
	ExecutorName string `json:"executorName"`
	JobID        int    `json:"jobId"`
}

type AddExecutionLogEntryRequest struct {
	JobOperationRequest
	executor.ExecutionLogEntry
}

type UpdateExecutionLogEntryRequest struct {
	JobOperationRequest
	EntryID int `json:"entryId"`
	executor.ExecutionLogEntry
}

type MarkCompleteRequest struct {
	JobOperationRequest
}

type MarkErroredRequest struct {
	JobOperationRequest
	ErrorMessage string `json:"errorMessage"`
}

type QueueJobIDs struct {
	QueueName string   `json:"queueName"`
	JobIDs    []string `json:"jobIds"`
}

// HeartbeatRequest is the payload sent by executors to the executor service to indicate that they are still alive.
type HeartbeatRequest struct {
	// TODO: This field is set to become unnecessary in Sourcegraph 5.2.
	Version ExecutorAPIVersion `json:"version"`

	ExecutorName string `json:"executorName"`

	JobIDs []string `json:"jobIds,omitempty"`
	// Used by multi-queue executors. One of (JobIDsByQueue and QueueNames) or JobIDs must be set.
	JobIDsByQueue []QueueJobIDs `json:"jobIdsByQueue,omitempty"`
	QueueNames    []string      `json:"queueNames,omitempty"`

	// Telemetry data.
	OS              string `json:"os"`
	Architecture    string `json:"architecture"`
	DockerVersion   string `json:"dockerVersion"`
	ExecutorVersion string `json:"executorVersion"`
	GitVersion      string `json:"gitVersion"`
	IgniteVersion   string `json:"igniteVersion"`
	SrcCliVersion   string `json:"srcCliVersion"`

	PrometheusMetrics string `json:"prometheusMetrics"`
}

// HeartbeatRequestV1 is the payload sent by executors to the executor service to indicate that they are still alive.
// Job IDs are ints instead of strings to support backwards compatibility.
// TODO: Remove this in Sourcegraph 5.2
type HeartbeatRequestV1 struct {
	// TODO: This field is set to become unnecessary in Sourcegraph 5.2.
	Version ExecutorAPIVersion `json:"version"`

	ExecutorName string `json:"executorName"`
	JobIDs       []int  `json:"jobIds"`

	// Telemetry data.
	OS              string `json:"os"`
	Architecture    string `json:"architecture"`
	DockerVersion   string `json:"dockerVersion"`
	ExecutorVersion string `json:"executorVersion"`
	GitVersion      string `json:"gitVersion"`
	IgniteVersion   string `json:"igniteVersion"`
	SrcCliVersion   string `json:"srcCliVersion"`

	PrometheusMetrics string `json:"prometheusMetrics"`
}

type heartbeatRequestUnmarshaller struct {
	// TODO: This field is set to become unnecessary in Sourcegraph 5.2.
	Version ExecutorAPIVersion `json:"version"`

	ExecutorName  string        `json:"executorName"`
	JobIDs        []any         `json:"jobIds"`
	JobIDsByQueue []QueueJobIDs `json:"jobIdsByQueue"`
	QueueNames    []string      `json:"queueNames"`

	// Telemetry data.
	OS              string `json:"os"`
	Architecture    string `json:"architecture"`
	DockerVersion   string `json:"dockerVersion"`
	ExecutorVersion string `json:"executorVersion"`
	GitVersion      string `json:"gitVersion"`
	IgniteVersion   string `json:"igniteVersion"`
	SrcCliVersion   string `json:"srcCliVersion"`

	PrometheusMetrics string `json:"prometheusMetrics"`
}

// TODO: This field is set to become unnecessary in Sourcegraph 5.2.
type ExecutorAPIVersion string

const (
	ExecutorAPIVersion2 ExecutorAPIVersion = "V2"
)

// UnmarshalJSON is a custom unmarshaler for HeartbeatRequest that allows for backwards compatibility when job IDs are
// ints instead of strings.
// TODO: Remove this in Sourcegraph 5.2
func (h *HeartbeatRequest) UnmarshalJSON(b []byte) error {
	var req heartbeatRequestUnmarshaller
	if err := json.Unmarshal(b, &req); err != nil {
		return err
	}
	h.Version = req.Version
	h.JobIDsByQueue = req.JobIDsByQueue
	h.QueueNames = req.QueueNames
	h.ExecutorName = req.ExecutorName
	h.OS = req.OS
	h.Architecture = req.Architecture
	h.DockerVersion = req.DockerVersion
	h.ExecutorVersion = req.ExecutorVersion
	h.GitVersion = req.GitVersion
	h.IgniteVersion = req.IgniteVersion
	h.SrcCliVersion = req.SrcCliVersion
	h.PrometheusMetrics = req.PrometheusMetrics

	for _, id := range req.JobIDs {
		switch jobId := id.(type) {
		case int:
			h.JobIDs = append(h.JobIDs, strconv.Itoa(jobId))
		case float32:
			h.JobIDs = append(h.JobIDs, strconv.FormatFloat(float64(jobId), 'f', -1, 32))
		case float64:
			h.JobIDs = append(h.JobIDs, strconv.FormatFloat(jobId, 'f', -1, 64))
		case string:
			h.JobIDs = append(h.JobIDs, jobId)
		default:
			return errors.Newf("unknown type for job ID: %T", id)
		}
	}
	return nil
}

type HeartbeatResponse struct {
	KnownIDs  []string `json:"knownIds"`
	CancelIDs []string `json:"cancelIds"`
}

type heartbeatResponseUnmarshaller struct {
	KnownIDs  []any `json:"knownIds"`
	CancelIDs []any `json:"cancelIds"`
}

// UnmarshalJSON is a custom unmarshaler for HeartbeatResponse that allows for backwards compatibility when IDs are
// ints instead of strings.
// TODO: Remove this in Sourcegraph 5.2
func (h *HeartbeatResponse) UnmarshalJSON(b []byte) error {
	var res heartbeatResponseUnmarshaller
	if err := json.Unmarshal(b, &res); err != nil {
		return err
	}

	for _, id := range res.KnownIDs {
		switch knownId := id.(type) {
		case int:
			h.KnownIDs = append(h.KnownIDs, strconv.Itoa(knownId))
		case float32:
			h.KnownIDs = append(h.KnownIDs, strconv.FormatFloat(float64(knownId), 'f', -1, 32))
		case float64:
			h.KnownIDs = append(h.KnownIDs, strconv.FormatFloat(knownId, 'f', -1, 64))
		case string:
			h.KnownIDs = append(h.KnownIDs, knownId)
		default:
			return errors.Newf("unknown type for known ID: %T", id)
		}
	}

	for _, id := range res.CancelIDs {
		switch cancelId := id.(type) {
		case int:
			h.CancelIDs = append(h.CancelIDs, strconv.Itoa(cancelId))
		case float32:
			h.CancelIDs = append(h.CancelIDs, strconv.FormatFloat(float64(cancelId), 'f', -1, 32))
		case float64:
			h.CancelIDs = append(h.CancelIDs, strconv.FormatFloat(cancelId, 'f', -1, 64))
		case string:
			h.CancelIDs = append(h.CancelIDs, cancelId)
		default:
			return errors.Newf("unknown type for cancel ID: %T", id)
		}
	}
	return nil
}
