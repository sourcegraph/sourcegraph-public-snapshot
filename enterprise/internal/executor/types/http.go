package types

import (
	"github.com/sourcegraph/sourcegraph/internal/executor"
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

type HeartbeatRequest struct {
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

type HeartbeatResponse struct {
	KnownIDs  []string `json:"knownIds"`
	CancelIDs []string `json:"cancelIds"`
}
