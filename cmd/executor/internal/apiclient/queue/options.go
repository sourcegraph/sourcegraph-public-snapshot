package queue

import (
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/apiclient"
)

type Options struct {
	// ExecutorName is a unique identifier for the requesting executor.
	ExecutorName string

	// QueueName is the name of the queue being processed. Only one of QueueName and QueueNames can be set.
	QueueName string

	// QueueNames are the names of the queues being processed. Only one of QueueNames and QueueName can be set.
	QueueNames []string

	// BaseClientOptions are the underlying HTTP client options.
	BaseClientOptions apiclient.BaseClientOptions

	// TelemetryOptions captures additional parameters sent in heartbeat requests.
	TelemetryOptions TelemetryOptions

	// ResourceOptions inform the frontend how large of a VM the job will be executed in.
	// This can be used to replace magic variables in the job payload indicating how much
	// the task should be able to comfortably consume.
	ResourceOptions ResourceOptions
}

type ResourceOptions struct {
	// NumCPUs is the number of virtual CPUs a job can safely utilize.
	NumCPUs int

	// Memory is the maximum amount of memory a job can safely utilize.
	Memory string

	// DiskSpace is the maximum amount of disk a job can safely utilize.
	DiskSpace string
}

type TelemetryOptions struct {
	OS              string
	Architecture    string
	DockerVersion   string
	ExecutorVersion string
	GitVersion      string
	IgniteVersion   string
	SrcCliVersion   string
}
