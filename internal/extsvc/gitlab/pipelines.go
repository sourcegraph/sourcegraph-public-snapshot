package gitlab

import (
	"context"
	"fmt"
)

// GetMergeRequestPipelines retrieves the pipelines that have been executed as
// part of the given merge request. As the pipelines are paginated, a function
// is returned that may be invoked to return the next page of results. An empty
// slice and a nil error indicates that all pages have been returned.
func (c *Client) GetMergeRequestPipelines(ctx context.Context, project *Project, iid ID) func() ([]*Pipeline, error) {
	if MockGetMergeRequestPipelines != nil {
		return MockGetMergeRequestPipelines(c, ctx, project, iid)
	}

	pr := c.newPaginatedResult("GET", fmt.Sprintf("projects/%d/merge_requests/%d/pipelines", project.ID, iid), func() interface{} { return []*Pipeline{} })
	return func() ([]*Pipeline, error) {
		page, err := pr.next(ctx)
		return page.([]*Pipeline), err
	}
}

type Pipeline struct {
	ID     ID             `json:"id"`
	SHA    string         `json:"sha"`
	Ref    string         `json:"ref"`
	Status PipelineStatus `json:"status"`
}

type PipelineStatus string

const (
	PipelineStatusRunning  PipelineStatus = "running"
	PipelineStatusPending  PipelineStatus = "pending"
	PipelineStatusSuccess  PipelineStatus = "success"
	PipelineStatusFailed   PipelineStatus = "failed"
	PipelineStatusCanceled PipelineStatus = "canceled"
	PipelineStatusSkipped  PipelineStatus = "skipped"
	PipelineStatusCreated  PipelineStatus = "created"
	PipelineStatusManual   PipelineStatus = "manual"
)

func (p *Pipeline) Key() string {
	return fmt.Sprintf("Pipeline:%d", p.ID)
}
