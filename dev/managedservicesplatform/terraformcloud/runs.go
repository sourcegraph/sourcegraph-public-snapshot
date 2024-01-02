package terraformcloud

import (
	"context"
	"fmt"

	tfe "github.com/hashicorp/go-tfe"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type RunsClient struct {
	client *tfe.Client
	org    string
}

func NewRunsClient(accessToken string) (*RunsClient, error) {
	c, err := tfe.NewClient(&tfe.Config{
		Token: accessToken,
	})
	if err != nil {
		return nil, err
	}
	return &RunsClient{
		org:    Organization,
		client: c,
	}, nil
}

type Outputs []*tfe.StateVersionOutput

func (o Outputs) Find(name string) (*tfe.StateVersionOutput, error) {
	// In stacks we prefix all output IDs with 'output-'.
	key := fmt.Sprintf("output-%s", name)
	for _, output := range o {
		if output.Name == key {
			return output, nil
		}
	}
	return nil, errors.Newf("output %q not found, available: %+v", key, o.Names())
}

func (o Outputs) Names() []string {
	var names []string
	for _, output := range o {
		names = append(names, output.Name)
	}
	return names
}

func (c *RunsClient) GetOutputs(ctx context.Context, workspaceName string) (Outputs, error) {
	ws, err := c.client.Workspaces.Read(ctx, c.org, workspaceName)
	if err != nil {
		return nil, errors.Wrapf(err, "get workspace %q", workspaceName)
	}
	outputs, err := c.client.StateVersionOutputs.ReadCurrent(ctx, ws.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "get current outputs for workspace ID %q", ws.ID)
	}
	// We  don't need pagination for now, we have very few outputs
	return outputs.Items, nil
}
