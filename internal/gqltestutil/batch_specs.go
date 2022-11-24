package gqltestutil

import (
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (c *Client) CreateEmptyBatchChange(namespace, name string) (string, error) {
	const query = `
	mutation CreateEmptyBatchChange($namespace: ID!, $name: ID!) {
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

func (c *Client) GetBatchSpecState(batchSpec string) (string, error) {
	const query = `
	query GetBatchSpecState($batchSpec: ID!) {
		node(id: $batchSpec) {
			__typename
			... on BatchSpec {
				state
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
				Typename string `json:"__typename"`
				State    string `json:"state"`
			} `json:"node"`
		} `json:"data"`
	}
	err := c.GraphQL("", query, variables, &resp)
	if err != nil {
		return "", errors.Wrap(err, "request GraphQL")
	}

	return resp.Data.Node.State, nil
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
				  outputLines
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
				queuedAt
				startedAt
				finishedAt
				failureMessage
				state
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
	ID               string `json:"id"`
	AutoApplyEnabled bool   `json:"autoApplyEnabled"`
	State            string `json:"state"`
}

func (c *Client) GetBatchSpecDeep(batchSpec string) (*BatchSpecDeep, error) {
	variables := map[string]any{
		"batchSpec": batchSpec,
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
