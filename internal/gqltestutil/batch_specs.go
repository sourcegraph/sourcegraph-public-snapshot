package gqltestutil

import (
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (c *Client) CreateEmptyBatchChange(namespace, name string) (string, error) {
	const query = `
	mutation CreateEmptyBatchChange($namespace: ID!, $name: String!) {
		createEmptyBatchChange(namespace: $namespace, name: $name) {
			id
		}
	}
	`
	variables := map[string]any{
		"namespace": namespace,
		"name":      name,
	}
	var resp struct {
		Data struct {
			CreateEmptyBatchChange struct {
				ID string `json:"id"`
			} `json:"createEmptyBatchChange"`
		} `json:"data"`
	}
	err := c.GraphQL("", query, variables, &resp)
	if err != nil {
		return "", errors.Wrap(err, "request GraphQL")
	}

	return resp.Data.CreateEmptyBatchChange.ID, nil
}

func (c *Client) CreateBatchSpecFromRaw(batchChange, namespace, batchSpec string) (string, error) {
	const query = `
	mutation CreateBatchSpecFromRaw($namespace: ID!, $batchChange: ID!, $batchSpec: String!) {
		createBatchSpecFromRaw(namespace: $namespace, batchChange: $batchChange, batchSpec: $batchSpec) {
			id
		}
	}
	`
	variables := map[string]any{
		"namespace":   namespace,
		"batchChange": batchChange,
		"batchSpec":   batchSpec,
	}
	var resp struct {
		Data struct {
			CreateBatchSpecFromRaw struct {
				ID string `json:"id"`
			} `json:"createBatchSpecFromRaw"`
		} `json:"data"`
	}
	err := c.GraphQL("", query, variables, &resp)
	if err != nil {
		return "", errors.Wrap(err, "request GraphQL")
	}

	return resp.Data.CreateBatchSpecFromRaw.ID, nil
}

func (c *Client) GetBatchSpecWorkspaceResolutionStatus(batchSpec string) (string, error) {
	const query = `
	query GetBatchSpecWorkspaceResolutionStatus($batchSpec: ID!) {
		node(id: $batchSpec) {
			__typename
			... on BatchSpec {
				workspaceResolution {
					state
				}
			}
		}
	}
	`
	variables := map[string]any{
		"batchSpec": batchSpec,
	}
	var resp struct {
		Data struct {
			Node struct {
				Typename            string `json:"__typename"`
				WorkspaceResolution struct {
					State string `json:"state"`
				} `json:"workspaceResolution"`
			} `json:"node"`
		} `json:"data"`
	}
	err := c.GraphQL("", query, variables, &resp)
	if err != nil {
		return "", errors.Wrap(err, "request GraphQL")
	}

	return resp.Data.Node.WorkspaceResolution.State, nil
}

func (c *Client) ExecuteBatchSpec(batchSpec string, noCache bool) error {
	const query = `
	mutation ExecuteBatchSpec($batchSpec: ID!, $noCache: Boolean!) {
		executeBatchSpec(batchSpec: $batchSpec, noCache: $noCache) {
			id
		}
	}
	`
	variables := map[string]any{
		"batchSpec": batchSpec,
		"noCache":   noCache,
	}
	err := c.GraphQL("", query, variables, nil)
	if err != nil {
		return errors.Wrap(err, "request GraphQL")
	}

	return nil
}

func (c *Client) GetBatchSpecState(batchSpec string) (string, string, error) {
	const query = `
	query GetBatchSpecState($batchSpec: ID!) {
		node(id: $batchSpec) {
			__typename
			... on BatchSpec {
				state
				failureMessage
			}
		}
	}
	`
	variables := map[string]any{
		"batchSpec": batchSpec,
	}
	var resp struct {
		Data struct {
			Node struct {
				Typename       string `json:"__typename"`
				State          string `json:"state"`
				FailureMessage string `json:"failureMessage"`
			} `json:"node"`
		} `json:"data"`
	}
	err := c.GraphQL("", query, variables, &resp)
	if err != nil {
		return "", "", errors.Wrap(err, "request GraphQL")
	}

	return resp.Data.Node.State, resp.Data.Node.FailureMessage, nil
}

const getBatchSpecDeep = `
query GetBatchSpecDeep($id: ID!) {
	node(id: $id) {
	  ... on BatchSpec {
		id
		autoApplyEnabled
		state
		changesetSpecs {
		  totalCount
		  nodes {
			id
			type
			... on VisibleChangesetSpec {
			  description {
				... on GitBranchChangesetDescription {
				  baseRepository {
					name
				  }
				  baseRef
				  baseRev
				  headRef
				  title
				  body
				  commits {
					message
					subject
					body
					author {
					  name
					  email
					}
				  }
				  diff {
					fileDiffs {
					  rawDiff
					}
				  }
				}
			  }
			  forkTarget {
				namespace
			  }
			}
		  }
		}
		createdAt
		startedAt
		finishedAt
		namespace {
		  id
		}
		workspaceResolution {
		  workspaces {
			totalCount
			stats {
			  errored
			  completed
			  processing
			  queued
			  ignored
			}
			nodes {
			  onlyFetchWorkspace
			  ignored
			  unsupported
			  cachedResultFound
			  stepCacheResultCount
			  queuedAt
			  startedAt
			  finishedAt
			  state
			  placeInQueue
			  placeInGlobalQueue
			  diffStat {
				added
				deleted
			  }
			  ... on VisibleBatchSpecWorkspace {
				repository {
				  name
				}
				branch {
				  name
				}
				path
				stages {
				  setup {
					key
					command
					startTime
					exitCode
					out
					durationMilliseconds
				  }
				  srcExec {
					key
					command
					startTime
					exitCode
					out
					durationMilliseconds
				  }
				  teardown {
					key
					command
					startTime
					exitCode
					out
					durationMilliseconds
				  }
				}
				steps {
				  number
				  run
				  container
				  ifCondition
				  cachedResultFound
				  skipped
				  outputLines {
					nodes
					totalCount
				  }
				  startedAt
				  finishedAt
				  exitCode
				  environment {
					name
					value
				  }
				  outputVariables {
					name
					value
				  }
				  diffStat {
					added
					deleted
				  }
				  diff {
					fileDiffs {
					  rawDiff
					}
				  }
				}
				searchResultPaths
				failureMessage
				changesetSpecs {
				  id
				}
				executor {
				  hostname
				  queueName
				  active
				}
			  }
			}
		  }
		}
		expiresAt
		failureMessage
		source
		files {
		  totalCount
		  nodes {
			path
			name
		  }
		}
	  }
	}
}
`

type BatchSpecDeep struct {
	ID                  string `json:"id"`
	AutoApplyEnabled    bool   `json:"autoApplyEnabled"`
	State               string `json:"state"`
	ChangesetSpecs      BatchSpecChangesetSpecs
	CreatedAt           string
	StartedAt           string
	FinishedAt          string
	Namespace           Namespace
	WorkspaceResolution WorkspaceResolution
	ExpiresAt           string
	FailureMessage      string
	Source              string
	Files               BatchSpecFiles
}

type BatchSpecFiles struct {
	TotalCount int
	Nodes      []BatchSpecFile
}

type BatchSpecFile struct {
	Path string
	Name string
}

type WorkspaceResolution struct {
	Workspaces WorkspaceResolutionWorkspaces
}

type WorkspaceResolutionWorkspaces struct {
	TotalCount int
	Stats      WorkspaceResolutionWorkspacesStats
	Nodes      []BatchSpecWorkspace
}

type BatchSpecWorkspace struct {
	OnlyFetchWorkspace   bool
	Ignored              bool
	Unsupported          bool
	CachedResultFound    bool
	StepCacheResultCount int
	QueuedAt             string
	StartedAt            string
	FinishedAt           string
	State                string
	PlaceInQueue         int
	PlaceInGlobalQueue   int
	DiffStat             DiffStat
	Repository           ChangesetRepository
	Branch               WorkspaceBranch
	Path                 string
	SearchResultPaths    []string
	FailureMessage       string
	ChangesetSpecs       []WorkspaceChangesetSpec
	Stages               BatchSpecWorkspaceStages
	Steps                []BatchSpecWorkspaceStep
	Executor             Executor
}

type WorkspaceOutputLines struct {
	Nodes      []string
	TotalCount int
}

type BatchSpecWorkspaceStep struct {
	Number            int
	Run               string
	Container         string
	IfCondition       string
	CachedResultFound bool
	Skipped           bool
	OutputLines       WorkspaceOutputLines
	StartedAt         string
	FinishedAt        string
	ExitCode          int
	Environment       []WorkspaceEnvironmentVariable
	OutputVariables   []WorkspaceOutputVariable
	DiffStat          DiffStat
	Diff              ChangesetSpecDiffs
}

type WorkspaceEnvironmentVariable struct {
	Name  string
	Value string
}
type WorkspaceOutputVariable struct {
	Name  string
	Value string
}

type BatchSpecWorkspaceStages struct {
	Setup    []ExecutionLogEntry
	SrcExec  []ExecutionLogEntry
	Teardown []ExecutionLogEntry
}

type ExecutionLogEntry struct {
	Key                  string
	Command              []string
	StartTime            string
	ExitCode             int
	Out                  string
	DurationMilliseconds int
}

type Executor struct {
	Hostname  string
	QueueName string
	Active    bool
}

type WorkspaceChangesetSpec struct {
	ID string
}

type WorkspaceBranch struct {
	Name string
}

type DiffStat struct {
	Added   int
	Deleted int
}

type WorkspaceResolutionWorkspacesStats struct {
	Errored    int
	Completed  int
	Processing int
	Queued     int
	Ignored    int
}

type Namespace struct {
	ID string
}

type BatchSpecChangesetSpecs struct {
	TotalCount int
	Nodes      []ChangesetSpec
}

type ChangesetSpec struct {
	ID          string
	Type        string
	Description ChangesetSpecDescription
	ForkTarget  ChangesetForkTarget
}

type ChangesetForkTarget struct {
	Namespace string
}

type ChangesetSpecDescription struct {
	BaseRepository ChangesetRepository
	BaseRef        string
	BaseRev        string
	HeadRef        string
	Title          string
	Body           string
	Commits        []ChangesetSpecCommit
	Diffs          ChangesetSpecDiffs
}

type ChangesetSpecDiffs struct {
	FileDiffs ChangesetSpecFileDiffs
}

type ChangesetSpecFileDiffs struct {
	RawDiff string
}

type ChangesetSpecCommit struct {
	Message string
	Subject string
	Body    string
	Author  ChangesetSpecCommitAuthor
}

type ChangesetSpecCommitAuthor struct {
	Name  string
	Email string
}

type ChangesetRepository struct {
	Name string
}

func (c *Client) GetBatchSpecDeep(batchSpec string) (*BatchSpecDeep, error) {
	variables := map[string]any{
		"id": batchSpec,
	}
	var resp struct {
		Data struct {
			Node *BatchSpecDeep `json:"node"`
		} `json:"data"`
	}
	err := c.GraphQL("", getBatchSpecDeep, variables, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "request GraphQL")
	}

	return resp.Data.Node, nil
}
