package types

import "time"

// Executor describes an executor instance that has recently connected to Sourcegraph.
type Executor struct {
	ID              int
	Hostname        string
	QueueName       string
	OS              string
	Architecture    string
	ExecutorVersion string
	SrcCliVersion   string
	GitVersion      string
	DockerVersion   string
	IgniteVersion   string
	FirstSeenAt     time.Time
	LastSeenAt      time.Time
}
