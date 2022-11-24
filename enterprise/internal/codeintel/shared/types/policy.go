package types

import "time"

type ConfigurationPolicy struct {
	ID                        int
	RepositoryID              *int
	RepositoryPatterns        *[]string
	Name                      string
	Type                      GitObjectType
	Pattern                   string
	Protected                 bool
	RetentionEnabled          bool
	RetentionDuration         *time.Duration
	RetainIntermediateCommits bool
	IndexingEnabled           bool
	IndexCommitMaxAge         *time.Duration
	IndexIntermediateCommits  bool
}

type GitObjectType string

const (
	GitObjectTypeCommit GitObjectType = "GIT_COMMIT"
	GitObjectTypeTag    GitObjectType = "GIT_TAG"
	GitObjectTypeTree   GitObjectType = "GIT_TREE"
)

type RetentionPolicyMatchCandidate struct {
	*ConfigurationPolicy
	Matched           bool
	ProtectingCommits []string
}
