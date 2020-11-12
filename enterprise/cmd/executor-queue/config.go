package main

import (
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/apiworker/apiserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
)

const port = 3191

type Config struct {
	env.BaseConfig

	MaximumNumTransactions     int
	JobRequeueDelay            time.Duration
	JobCleanupInterval         time.Duration
	MaximumNumMissedHeartbeats int
}

func (c *Config) Load() {
	c.MaximumNumTransactions = c.GetInt("EXECUTOR_QUEUE_MAXIMUM_NUM_TRANSACTIONS", "10", "Number of jobs that can be processing at one time.")
	c.JobRequeueDelay = c.GetInterval("EXECUTOR_QUEUE_JOB_REQUEUE_DELAY", "1m", "The requeue delay of jobs assigned to an unreachable executor.")
	c.JobCleanupInterval = c.GetInterval("EXECUTOR_QUEUE_JOB_CLEANUP_INTERVAL", "10s", "Interval between cleanup runs.")
	c.MaximumNumMissedHeartbeats = c.GetInt("EXECUTOR_QUEUE_MAXIMUM_NUM_MISSED_HEARTBEATS", "5", "The number of heartbeats an executor must miss to be considered unreachable.")
}

func (c *Config) ServerOptions(queueOptions map[string]apiserver.QueueOptions) apiserver.Options {
	return apiserver.Options{
		Port:                   port,
		QueueOptions:           queueOptions,
		MaximumNumTransactions: c.MaximumNumTransactions,
		RequeueDelay:           c.JobRequeueDelay,
		UnreportedMaxAge:       c.JobCleanupInterval * time.Duration(c.MaximumNumMissedHeartbeats),
		DeathThreshold:         c.JobCleanupInterval * time.Duration(c.MaximumNumMissedHeartbeats),
		CleanupInterval:        c.JobCleanupInterval,
	}
}
