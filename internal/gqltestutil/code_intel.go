package gqltestutil

import (
	"context"
	"sync"
	"time"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/conc/pool"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type IndexingJob struct {
	ID     graphql.ID
	RepoID graphql.ID
}

type AutoIndexJobMap map[string]IndexingJob

// TriggerAutoIndexing enqueues auto-indexing jobs for the provided repos
func (c *Client) TriggerAutoIndexing(repos ...string) (AutoIndexJobMap, error) {
	const query = `
query GetRepoIds($repoCount: Int!, $repos: [String!]!) {
    repositories(first: $repoCount, names: $repos) {
                nodes {
                        id
						name
                }
        }
}
`
	variables := map[string]any{
		"repoCount": len(repos),
		"repos":     repos,
	}
	var resp struct {
		Data struct {
			NewRepositoryConnection struct {
				Nodes []struct {
					ID   graphql.ID `json:"id"`
					Name string     `json:"name"`
				} `json:"nodes"`
			} `json:"repositories"`
		} `json:"data"`
	}

	err := c.GraphQL("", query, variables, &resp)
	if err != nil {
		return AutoIndexJobMap{}, errors.Wrap(err, "request GraphQL")
	}

	mapping := map[string]IndexingJob{}
	for _, repo := range resp.Data.NewRepositoryConnection.Nodes {
		const mutation = `
mutation AutoIndexRepos($repoID: ID!) {
    queueAutoIndexJobsForRepo(repository: $repoID) {
		id
	}
}
`
		variables := map[string]any{
			"repoID": repo.ID,
		}
		var resp struct {
			Data struct {
				QueueAutoIndexJobsForRepo []struct {
					ID graphql.ID `json:"id"`
				} `json:"queueAutoIndexJobsForRepo"`
			} `json:"data"`
		}
		err := c.GraphQL("", mutation, variables, &resp)
		if err != nil {
			return AutoIndexJobMap{}, errors.Wrapf(err, "failed to queue auto-indexing job for repo: %v", repo.Name)
		}
		if len(resp.Data.QueueAutoIndexJobsForRepo) != 1 {
			return AutoIndexJobMap{}, errors.Newf("unexpected number of auto-indexing jobs: %v", len(resp.Data.QueueAutoIndexJobsForRepo))
		}
		mapping[repo.Name] = IndexingJob{
			ID:     resp.Data.QueueAutoIndexJobsForRepo[0].ID,
			RepoID: repo.ID,
		}
	}
	return mapping, nil
}

type JobState struct {
	State        string `json:"state"`
	PlaceInQueue int    `json:"placeInQueue"`
}

type JobStateMap map[graphql.ID]JobState

func (c *Client) WaitForAutoIndexingJobsToComplete(jobMap AutoIndexJobMap, timeout time.Duration) (JobStateMap, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), timeout)
	defer cancelFunc()

	mtx := sync.Mutex{}
	jobStateMap := JobStateMap{}

	workPool := pool.New().WithErrors().WithContext(ctx)
	for repoName, jobInfo := range jobMap {
		jobId := jobInfo.ID
		workPool.Go(func(ctx context.Context) error {
			for {
				if err := ctx.Err(); err != nil {
					return err
				}
				const query = `
query GetJobById($jobID: ID!) {
	node(id: $jobID) {
		... on PreciseIndex {
			state
			placeInQueue
		}
	}
}
`
				variables := map[string]any{"jobID": jobId}
				var resp struct {
					Data struct {
						Node JobState `json:"node"`
					} `json:"data"`
				}
				err := c.GraphQL("", query, variables, &resp)
				if err != nil {
					return errors.Wrapf(err, "when requesting index status for repo: %v, jobID: %v", repoName, jobId)
				}
				mtx.Lock()
				jobStateMap[jobId] = resp.Data.Node
				mtx.Unlock()
				if resp.Data.Node.State == "COMPLETED" {
					return nil
				}
				time.Sleep(100 * time.Millisecond)
			}
		})
	}
	if err := workPool.Wait(); err != nil {
		return jobStateMap, errors.Wrap(err, "error in work pool")
	}
	return jobStateMap, nil
}
