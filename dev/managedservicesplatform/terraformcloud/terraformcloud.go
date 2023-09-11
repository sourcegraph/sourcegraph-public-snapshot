package terraformcloud

import (
	"context"

	"fmt"

	tfe "github.com/hashicorp/go-tfe"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/terraform"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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
	// Organization is our default Terraform Cloud organization.
	Organization = "sourcegraph"
	// VCSRepo is the repository that is expected to house Managed Services
	// Platform Terraform assets.
	VCSRepo = "sourcegraph/managed-services"
)

type Client struct {
	client           *tfe.Client
	org              string
	vcsOAuthClientID string
}

func NewClient(accessToken, vcsOAuthClientID string) (*Client, error) {
	c, err := tfe.NewClient(&tfe.Config{
		Token: accessToken,
	})
	if err != nil {
		return nil, err
	}
	return &Client{
		org:              Organization,
		client:           c,
		vcsOAuthClientID: vcsOAuthClientID,
	}, nil
}

// workspaceOptions is a union between tfe.WorkspaceCreateOptions and
// tfe.WorkspaceUpdateOptions
type workspaceOptions struct {
	Name    *string
	Project *tfe.Project
	VCSRepo *tfe.VCSRepoOptions

	TriggerPrefixes  []string
	WorkingDirectory *string

	ExecutionMode     *string
	TerraformVersion  *string
	AutoApply         *bool
	GlobalRemoteState *bool
}

// AsCreate should be kept up to date with AsUpdate.
func (c workspaceOptions) AsCreate(tags []*tfe.Tag) tfe.WorkspaceCreateOptions {
	return tfe.WorkspaceCreateOptions{
		// Tags cannot be set in update
		Tags: tags,

		Name:    c.Name,
		Project: c.Project,
		VCSRepo: c.VCSRepo,

		WorkingDirectory: c.WorkingDirectory,
		TriggerPrefixes:  c.TriggerPrefixes,

		ExecutionMode:     c.ExecutionMode,
		TerraformVersion:  c.TerraformVersion,
		AutoApply:         c.AutoApply,
		GlobalRemoteState: c.GlobalRemoteState,
	}
}

// AsCreate should be kept up to date with the Update code path.
func (c workspaceOptions) AsUpdate() tfe.WorkspaceUpdateOptions {
	return tfe.WorkspaceUpdateOptions{
		// Tags cannot be set in update

		Name:    c.Name,
		Project: c.Project,
		VCSRepo: c.VCSRepo,

		WorkingDirectory: c.WorkingDirectory,
		TriggerPrefixes:  c.TriggerPrefixes,

		ExecutionMode:     c.ExecutionMode,
		TerraformVersion:  c.TerraformVersion,
		AutoApply:         c.AutoApply,
		GlobalRemoteState: c.GlobalRemoteState,
	}
}

// SyncWorkspaces is a bit like the Terraform Cloud Terraform provider. We do
// this directly instead of using the provider to avoid the chicken-and-egg
// problem of, if Terraform Cloud workspaces provision our resourcs, who provisions
// our Terraform Cloud workspace?
func (c *Client) SyncWorkspaces(ctx context.Context, svc spec.ServiceSpec, env spec.EnvironmentSpec, stacks []string) error {
	oauthClient, err := c.client.OAuthClients.Read(ctx, c.vcsOAuthClientID)
	if err != nil {
		return err
	}

	// Set up project for workspaces to be in
	tfcProjectName := fmt.Sprintf("msp-%s-%s", svc.ID, env.ID)
	var tfcProject *tfe.Project
	if projects, err := c.client.Projects.List(ctx, c.org, &tfe.ProjectListOptions{
		Name: tfcProjectName,
	}); err != nil {
		return err
	} else {
		for _, p := range projects.Items {
			if p.Name == tfcProjectName {
				tfcProject = p
				break
			}
		}
	}
	if tfcProject == nil {
		tfcProject, err = c.client.Projects.Create(ctx, c.org, tfe.ProjectCreateOptions{
			Name: tfcProjectName,
		})
		if err != nil {
			return err
		}
	}

	for _, s := range stacks {
		workspaceName := WorkspaceName(svc, env, s)
		workspaceDir := fmt.Sprintf("services/%s/terraform/%s/stacks/%s/", svc.ID, env.ID, s)
		wantWorkspaceOptions := workspaceOptions{
			Name:    &workspaceName,
			Project: tfcProject,
			VCSRepo: &tfe.VCSRepoOptions{
				OAuthTokenID: &oauthClient.ID,
				Identifier:   pointers.Ptr(VCSRepo),
				Branch:       pointers.Ptr("main"),
			},

			WorkingDirectory: pointers.Ptr(workspaceDir),
			TriggerPrefixes:  []string{workspaceDir},

			ExecutionMode:    pointers.Ptr("remote"),
			TerraformVersion: pointers.Ptr(terraform.Version),
			AutoApply:        pointers.Ptr(true),
		}
		// HACK: make project output available globally so that other stacks
		// can reference the generated, randomized ID.
		if s == "project" {
			wantWorkspaceOptions.GlobalRemoteState = pointers.Ptr(true)
		}

		wantWorkspaceTags := []*tfe.Tag{
			{Name: "msp"},
			{Name: fmt.Sprintf("msp-service-%s", svc.ID)},
			{Name: fmt.Sprintf("msp-env-%s-%s", svc.ID, env.ID)},
		}

		if existingWorkspace, err := c.client.Workspaces.Read(ctx, c.org, workspaceName); err != nil {
			if !errors.Is(err, tfe.ErrResourceNotFound) {
				return err
			}

			if _, err := c.client.Workspaces.Create(ctx, c.org,
				wantWorkspaceOptions.AsCreate(wantWorkspaceTags)); err != nil {
				return errors.Wrap(err, "workspaces.Create")
			}
		} else {
			// Forcibly update the workspace to match our expected configuration
			if _, err := c.client.Workspaces.Update(ctx, c.org, workspaceName,
				wantWorkspaceOptions.AsUpdate()); err != nil {
				return errors.Wrap(err, "workspaces.Update")
			}

			// Sync tags separately, as Update does not allow us to do this
			foundTags := make(map[string]struct{})
			for _, t := range existingWorkspace.Tags {
				foundTags[t.Name] = struct{}{}
			}
			addTags := tfe.WorkspaceAddTagsOptions{}
			for _, t := range wantWorkspaceTags {
				t := t
				if _, ok := foundTags[t.Name]; !ok {
					addTags.Tags = append(addTags.Tags, t)
				}
			}
			if len(addTags.Tags) > 0 {
				if err := c.client.Workspaces.AddTags(ctx, existingWorkspace.ID, addTags); err != nil {
					return errors.Wrap(err, "workspaces.AddTags")
				}
			}
		}

		// TODO https://github.com/sourcegraph/infrastructure/blob/main/modules/tfcworkspace/workspace.tf
	}

	return nil
}
