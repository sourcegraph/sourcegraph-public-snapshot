package terraformcloud

import (
	"context"

	tfe "github.com/hashicorp/go-tfe"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type RunsClient struct {
	client *tfe.Client
}

func NewRunsClient(accessToken string) (*RunsClient, error) {
	c, err := tfe.NewClient(&tfe.Config{
		Token: accessToken,
	})
	if err != nil {
		return nil, err
	}
	return &RunsClient{
		client: c,
	}, nil
}

func (c *RunsClient) ApplyWorkspace(ctx context.Context, ws Workspace, message string) error {
	_, err := c.client.Runs.Create(ctx, tfe.RunCreateOptions{
		Workspace:       ws.workspace,
		AutoApply:       pointers.Ptr(true),
		AllowEmptyApply: pointers.Ptr(true),
		Message:         &message,
	})
	if err != nil {
		return errors.Wrapf(err, "Runs.Create")
	}
	return nil
}
