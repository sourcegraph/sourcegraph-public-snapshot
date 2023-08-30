package terraformcloud

import (
	"context"
	"errors"
	"fmt"

	tfe "github.com/hashicorp/go-tfe"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/terraform"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// WorkspaceName is a fixed format for the Terraform Cloud workspace for a given
// service environment's stack:
//
//	msp-${svc.id}-${env.id}-${stackName}
func WorkspaceName(svc spec.ServiceSpec, env spec.EnvironmentSpec, stackName string) string {
	return fmt.Sprintf("msp-%s-%s-%s", svc.ID, env.ID, stackName)
}

const (
	// VCSRepo is the repository that is expected to house Terraform assets.
	VCSRepo = "sourcegraph/managed-services"
)

type Client struct {
	client *tfe.Client
	org    string
}

func New(organization, accessToken string) (*Client, error) {
	c, err := tfe.NewClient(&tfe.Config{
		Token: accessToken,
	})
	if err != nil {
		return nil, err
	}
	return &Client{
		client: c,
	}, nil
}

func (c *Client) SyncWorkspaces(ctx context.Context, svc spec.ServiceSpec, env spec.EnvironmentSpec, stacks []string) error {
	// TODO shared secret TFC_OAUTH_CLIENT_ID
	oauthClient, err := c.client.OAuthClients.Read(ctx, "")
	if err != nil {
		return err
	}

	for _, s := range stacks {
		workspaceTags := []*tfe.Tag{
			{Name: "msp"},
			{Name: fmt.Sprintf("msp-service-%s", svc.ID)},
			{Name: fmt.Sprintf("msp-env-%s-%s", svc.ID, env.ID)},
		}

		workspaceName := WorkspaceName(svc, env, s)
		if _, err := c.client.Workspaces.Read(ctx, c.org, workspaceName); err != nil {
			if !errors.Is(err, tfe.ErrResourceNotFound) {
				return err
			}

			if _, err := c.client.Workspaces.Create(ctx, c.org, tfe.WorkspaceCreateOptions{
				Name: pointers.Ptr(workspaceName),
				Tags: workspaceTags,

				// Workspaces options below - keep up to date with the Update
				// code path.
				VCSRepo: &tfe.VCSRepoOptions{
					OAuthTokenID: &oauthClient.ID,
					Identifier:   pointers.Ptr(VCSRepo),
					Branch:       pointers.Ptr("main"),
				},
				TriggerPrefixes:  []string{},
				ExecutionMode:    pointers.Ptr("remote"),
				TerraformVersion: pointers.Ptr(terraform.Version),
				AutoApply:        pointers.Ptr(true),
			}); err != nil {
				return err
			}
		} else {
			// Forcibly update the workspace to match our expected configuration
			if _, err := c.client.Workspaces.Update(ctx, c.org, workspaceName, tfe.WorkspaceUpdateOptions{
				// Keep up to date with the Create code path.
				VCSRepo: &tfe.VCSRepoOptions{
					OAuthTokenID: &oauthClient.ID,
					Identifier:   pointers.Ptr(VCSRepo),
					Branch:       pointers.Ptr("main"),
				},
				TriggerPrefixes:  []string{},
				ExecutionMode:    pointers.Ptr("remote"),
				TerraformVersion: pointers.Ptr(terraform.Version),
				AutoApply:        pointers.Ptr(true),
			}); err != nil {
				return err
			}
		}

		// TODO https://github.com/sourcegraph/infrastructure/blob/main/modules/tfcworkspace/workspace.tf
	}

	return nil
}
