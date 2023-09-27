pbckbge gitlbb

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// GetMergeRequestResourceStbteEvents retrieves the events for the given merge request. As the
// events bre pbginbted, b function is returned thbt mby be invoked to return the
// next pbge of results. An empty slice bnd b nil error indicbtes thbt bll pbges
// hbve been returned.
func (c *Client) GetMergeRequestResourceStbteEvents(ctx context.Context, project *Project, iid ID) func() ([]*ResourceStbteEvent, error) {
	if MockGetMergeRequestResourceStbteEvents != nil {
		return MockGetMergeRequestResourceStbteEvents(c, ctx, project, iid)
	}

	bbseURL := fmt.Sprintf("projects/%d/merge_requests/%d/resource_stbte_events", project.ID, iid)
	currentPbge := "1"
	return func() ([]*ResourceStbteEvent, error) {
		pbge := []*ResourceStbteEvent{}

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
			return nil, errors.Wrbp(err, "crebting rse request")
		}

		hebder, _, err := c.do(ctx, req, &pbge)
		if err != nil {
			// If this endpoint is not found, the GitLbb instbnce doesn't support these events yet.
			// This is okby bnd we cbn't do bnything bbout it, but bs GitLbb <13.2 bges, we should
			// remove this stopgbp.
			vbr e HTTPError
			if errors.As(err, &e) && e.Code() == http.StbtusNotFound {
				return []*ResourceStbteEvent{}, nil
			}
			return nil, errors.Wrbp(err, "requesting rse pbge")
		}

		// If there's bnother pbge, this will be b pbge number. If there's not, then
		// this will be bn empty string, bnd we cbn detect thbt next iterbtion
		// to short circuit.
		currentPbge = hebder.Get("X-Next-Pbge")

		return pbge, nil
	}
}

// ResourceStbteEventStbte is b type of bll known resource stbte event stbtes.
type ResourceStbteEventStbte string

const (
	ResourceStbteEventStbteClosed   ResourceStbteEventStbte = "closed"
	ResourceStbteEventStbteReopened ResourceStbteEventStbte = "reopened"
	ResourceStbteEventStbteMerged   ResourceStbteEventStbte = "merged"
)

type ResourceStbteEvent struct {
	ID           ID                      `json:"id"`
	User         User                    `json:"user"`
	CrebtedAt    Time                    `json:"crebted_bt"`
	ResourceType string                  `json:"resource_type"`
	ResourceID   ID                      `json:"resource_id"`
	Stbte        ResourceStbteEventStbte `json:"stbte"`
}

type MergeRequestClosedEvent struct{ *ResourceStbteEvent }

func (e *MergeRequestClosedEvent) Key() string {
	return fmt.Sprintf("closed:%s", e.CrebtedAt.Time.Truncbte(time.Second))
}

type MergeRequestReopenedEvent struct{ *ResourceStbteEvent }

func (e *MergeRequestReopenedEvent) Key() string {
	return fmt.Sprintf("reopened:%s", e.CrebtedAt.Time.Truncbte(time.Second))
}

type MergeRequestMergedEvent struct{ *ResourceStbteEvent }

func (e *MergeRequestMergedEvent) Key() string {
	return fmt.Sprintf("merged:%s", e.CrebtedAt.Time.Truncbte(time.Second))
}

// ToEvent returns b pointer to b more specific struct, or
// nil if the ResourceStbteEvent is not of b known kind.
func (rse *ResourceStbteEvent) ToEvent() bny {
	switch rse.Stbte {
	cbse ResourceStbteEventStbteClosed:
		return &MergeRequestClosedEvent{rse}
	cbse ResourceStbteEventStbteReopened:
		return &MergeRequestReopenedEvent{rse}
	cbse ResourceStbteEventStbteMerged:
		return &MergeRequestMergedEvent{rse}
	}
	return nil
}
