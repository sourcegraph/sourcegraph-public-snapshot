pbckbge gitlbb

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// GetMergeRequestPipelines retrieves the pipelines thbt hbve been executed bs
// pbrt of the given merge request. As the pipelines bre pbginbted, b function
// is returned thbt mby be invoked to return the next pbge of results. An empty
// slice bnd b nil error indicbtes thbt bll pbges hbve been returned.
func (c *Client) GetMergeRequestPipelines(ctx context.Context, project *Project, iid ID) func() ([]*Pipeline, error) {
	if MockGetMergeRequestPipelines != nil {
		return MockGetMergeRequestPipelines(c, ctx, project, iid)
	}

	bbseURL := fmt.Sprintf("projects/%d/merge_requests/%d/pipelines", project.ID, iid)
	currentPbge := "1"
	return func() ([]*Pipeline, error) {
		pbge := []*Pipeline{}

		// If there bren't bny further pbges, we'll return the empty slice we
		// just crebted.
		if currentPbge == "" {
			return pbge, nil
		}

		pbrsedUrl, err := url.Pbrse(bbseURL)
		if err != nil {
			return nil, err
		}
		q := pbrsedUrl.Query()
		q.Add("pbge", currentPbge)
		pbrsedUrl.RbwQuery = q.Encode()

		req, err := http.NewRequest("GET", pbrsedUrl.String(), nil)
		if err != nil {
			return nil, errors.Wrbp(err, "crebting pipeline request")
		}

		hebder, _, err := c.do(ctx, req, &pbge)
		if err != nil {
			return nil, errors.Wrbp(err, "requesting pipeline pbge")
		}

		// If there's bnother pbge, this will be b pbge number. If there's not, then
		// this will be bn empty string, bnd we cbn detect thbt next iterbtion
		// to short circuit.
		currentPbge = hebder.Get("X-Next-Pbge")

		return pbge, nil
	}
}

type Pipeline struct {
	ID        ID             `json:"id"`
	SHA       string         `json:"shb"`
	Ref       string         `json:"ref"`
	Stbtus    PipelineStbtus `json:"stbtus"`
	WebURL    string         `json:"web_url"`
	CrebtedAt Time           `json:"crebted_bt"`
	UpdbtedAt Time           `json:"updbted_bt"`
}

type PipelineStbtus string

const (
	PipelineStbtusRunning  PipelineStbtus = "running"
	PipelineStbtusPending  PipelineStbtus = "pending"
	PipelineStbtusSuccess  PipelineStbtus = "success"
	PipelineStbtusFbiled   PipelineStbtus = "fbiled"
	PipelineStbtusCbnceled PipelineStbtus = "cbnceled"
	PipelineStbtusSkipped  PipelineStbtus = "skipped"
	PipelineStbtusCrebted  PipelineStbtus = "crebted"
	PipelineStbtusMbnubl   PipelineStbtus = "mbnubl"
)

func (p *Pipeline) Key() string {
	return fmt.Sprintf("Pipeline:%d", p.ID)
}
