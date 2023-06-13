package batches

import (
	"github.com/sourcegraph/sourcegraph/lib/batches/execution"
	"github.com/sourcegraph/sourcegraph/lib/batches/template"
)

type WorkspacesExecutionInput struct {
	BatchChangeAttributes template.BatchChangeAttributes
	Repository            WorkspaceRepo   `json:"repository"`
	Branch                WorkspaceBranch `json:"branch"`
	Path                  string          `json:"path"`
	OnlyFetchWorkspace    bool            `json:"onlyFetchWorkspace"`
	Steps                 []Step          `json:"steps"`
	SearchResultPaths     []string        `json:"searchResultPaths"`
	// CachedStepResultFound is only required for V1 executions.
	// TODO: Remove me once V2 is the only execution format.
	CachedStepResultFound bool `json:"cachedStepResultFound"`
	// CachedStepResult is only required for V1 executions.
	// TODO: Remove me once V2 is the only execution format.
	CachedStepResult execution.AfterStepResult `json:"cachedStepResult,omitempty"`
	// SkippedSteps determines which steps are skipped in the execution.
	SkippedSteps map[int]struct{} `json:"skippedSteps"`
}

type WorkspaceRepo struct {
	// ID is the GraphQL ID of the repository.
	ID   string `json:"id"`
	Name string `json:"name"`
}

type WorkspaceBranch struct {
	Name   string `json:"name"`
	Target Commit `json:"target"`
}

type Commit struct {
	OID string `json:"oid"`
}
