package store

import "time"

type IncompleteDatapoint struct {
	Reason IncompleteReason
	RepoId *int
	Time   time.Time
}

type IncompleteReason string

const (
	ReasonTimeout IncompleteReason = "timeout"
)
