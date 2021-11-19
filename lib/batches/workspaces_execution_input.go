package batches

type WorkspacesExecutionInput struct {
	RawSpec   string    `json:"rawSpec"`
	Workspace Workspace `json:"workspace"`
}

type Workspace struct {
	Repository         WorkspaceRepo   `json:"repository"`
	Branch             WorkspaceBranch `json:"branch"`
	Path               string          `json:"path"`
	OnlyFetchWorkspace bool            `json:"onlyFetchWorkspace"`
	Steps              []Step          `json:"steps"`
	SearchResultPaths  []string        `json:"searchResultPaths"`
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
