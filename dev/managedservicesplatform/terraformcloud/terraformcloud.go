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

// WorkspaceName is a fixed format for the Terraform Cloud project housing all
// the workspaces for a given service environment:
//
//	msp-${svc.id}-${env.id}
func ProjectName(svc spec.ServiceSpec, env spec.EnvironmentSpec) string {
	return fmt.Sprintf("msp-%s-%s", svc.ID, env.ID)
}

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

	workspaceConfig WorkspaceConfig
}

type WorkspaceRunMode string

const (
	WorkspaceRunModeVCS WorkspaceRunMode = "vcs"
	WorkspaceRunModeCLI WorkspaceRunMode = "cli"
)

type WorkspaceConfig struct {
	RunMode WorkspaceRunMode
}

func NewClient(accessToken, vcsOAuthClientID string, cfg WorkspaceConfig) (*Client, error) {
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
		workspaceConfig:  cfg,
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

type Workspace struct {
	workspace *tfe.Workspace
	RunMode   WorkspaceRunMode
	Created   bool
}

func (w Workspace) Name() string { return w.workspace.Name }

func (w Workspace) ID() string { return w.workspace.ID }

func (w Workspace) URL() string {
	return fmt.Sprintf("https://app.terraform.io/app/sourcegraph/workspaces/%s", w.Name())
}

// SyncWorkspaces is a bit like the Terraform Cloud Terraform provider. We do
// this directly instead of using the provider to avoid the chicken-and-egg
// problem of, if Terraform Cloud workspaces provision our resourcs, who provisions
// our Terraform Cloud workspace?
func (c *Client) SyncWorkspaces(ctx context.Context, svc spec.ServiceSpec, env spec.EnvironmentSpec, stacks []string) ([]Workspace, error) {
	// Load preconfigured OAuth to GitHub if we are using VCS mode
	var oauthClient *tfe.OAuthClient
	if c.workspaceConfig.RunMode == WorkspaceRunModeVCS {
		var err error
		oauthClient, err = c.client.OAuthClients.Read(ctx, c.vcsOAuthClientID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get OAuth client for VCS mode")
		}
		if len(oauthClient.OAuthTokens) == 0 {
			return nil, errors.Wrapf(err, "OAuth client %q has no tokens, cannot use VCS mode", *oauthClient.Name)
		}
	}

	// Set up project for workspaces to be in
	tfcProjectName := ProjectName(svc, env)
	var tfcProject *tfe.Project
	if projects, err := c.client.Projects.List(ctx, c.org, &tfe.ProjectListOptions{
		Name: tfcProjectName,
	}); err != nil {
		return nil, errors.Wrap(err, "Projects.List")
	} else {
		for _, p := range projects.Items {
			if p.Name == tfcProjectName {
				tfcProject = p
				break
			}
		}
	}
	if tfcProject == nil {
		var err error
		tfcProject, err = c.client.Projects.Create(ctx, c.org, tfe.ProjectCreateOptions{
			Name: tfcProjectName,
		})
		if err != nil {
			return nil, errors.Wrap(err, "Projects.Create")
		}
	}

	// Grant access to project for Core Services
	if resp, err := c.client.TeamProjectAccess.List(ctx, tfe.TeamProjectAccessListOptions{
		ProjectID: tfcProject.ID,
	}); err != nil {
		return nil, errors.Wrap(err, "TeamProjectAccess.List")
	} else {
		for _, team := range []struct {
			name                 string // only for reference
			terraformCloudTeamID string
		}{
			{name: "Core Services", terraformCloudTeamID: "team-gGtVVgtNRaCnkhKp"},
			// Operators should use Entitle to request access to this team to
			// get access to workspaces, if they aren't in Core Services
			{name: "Managed Services Platform Operators", terraformCloudTeamID: "team-Wdejc42bWrRonQEY"},
		} {
			if err := c.ensureAccessForTeam(ctx, tfcProject, resp, team.terraformCloudTeamID); err != nil {
				return nil, errors.Wrapf(err, "ensure access for %q Terraform Cloud team %q",
					team.name, team.terraformCloudTeamID)
			}
		}
	}

	var workspaces []Workspace
	for _, s := range stacks {
		workspaceName := WorkspaceName(svc, env, s)
		workspaceDir := fmt.Sprintf("services/%s/terraform/%s/stacks/%s/", svc.ID, env.ID, s)
		wantWorkspaceOptions := workspaceOptions{
			Name:    &workspaceName,
			Project: tfcProject,

			ExecutionMode:    pointers.Ptr("remote"),
			TerraformVersion: pointers.Ptr(terraform.Version),
			AutoApply:        pointers.Ptr(true),

			// Allow all stacks to reference each other.
			GlobalRemoteState: pointers.Ptr(true),
		}
		switch c.workspaceConfig.RunMode {
		case WorkspaceRunModeVCS:
			// In VCS mode, TFC needs to be configured with the deployment repo
			// and provide the relative path to the root of the stack
			wantWorkspaceOptions.WorkingDirectory = pointers.Ptr(workspaceDir)
			wantWorkspaceOptions.VCSRepo = &tfe.VCSRepoOptions{
				OAuthTokenID: &oauthClient.OAuthTokens[len(oauthClient.OAuthTokens)-1].ID,
				Identifier:   pointers.Ptr(VCSRepo),
				Branch:       pointers.Ptr("main"),
			}
			wantWorkspaceOptions.TriggerPrefixes = []string{workspaceDir}
		case WorkspaceRunModeCLI:
			// In CLI, `terraform` runs will upload the content of current working directory
			// to TFC, hence we need to remove all VCS and working directory override
			wantWorkspaceOptions.VCSRepo = nil
			wantWorkspaceOptions.WorkingDirectory = nil
		default:
			return nil, errors.Errorf("invalid WorkspaceRunModeVCS %q", c.workspaceConfig.RunMode)
		}

		wantWorkspaceTags := []*tfe.Tag{
			{Name: "msp"},
			{Name: fmt.Sprintf("msp-service-%s", svc.ID)},
			{Name: fmt.Sprintf("msp-env-%s-%s", svc.ID, env.ID)},
		}

		if existingWorkspace, err := c.client.Workspaces.Read(ctx, c.org, workspaceName); err != nil {
			if !errors.Is(err, tfe.ErrResourceNotFound) {
				return nil, errors.Wrap(err, "failed to check if workspace exists")
			}

			createdWorkspace, err := c.client.Workspaces.Create(ctx, c.org,
				wantWorkspaceOptions.AsCreate(wantWorkspaceTags))
			if err != nil {
				return nil, errors.Wrap(err, "workspaces.Create")
			}

			workspaces = append(workspaces, Workspace{
				workspace: createdWorkspace,
				RunMode:   c.workspaceConfig.RunMode,
				Created:   true,
			})
		} else {
			workspaces = append(workspaces, Workspace{
				workspace: existingWorkspace,
				RunMode:   c.workspaceConfig.RunMode,
				Created:   false,
			})

			// VCSRepo must be removed by explicitly using the API - update
			// doesn't remove it - if we want to remove the connection.
			if existingWorkspace.VCSRepo != nil && wantWorkspaceOptions.VCSRepo == nil {
				if _, err := c.client.Workspaces.RemoveVCSConnection(ctx, c.org, workspaceName); err != nil {
					return nil, errors.Wrap(err, "failed to remove VCS connection")
				}
			}

			// Forcibly update the workspace to match our expected configuration
			if _, err := c.client.Workspaces.Update(ctx, c.org, workspaceName,
				wantWorkspaceOptions.AsUpdate()); err != nil {
				return nil, errors.Wrap(err, "workspaces.Update")
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
					return nil, errors.Wrap(err, "workspaces.AddTags")
				}
			}
		}

		// TODO backups https://github.com/sourcegraph/infrastructure/blob/main/modules/tfcworkspace/workspace.tf
	}

	return workspaces, nil
}

func (c *Client) DeleteWorkspaces(ctx context.Context, svc spec.ServiceSpec, env spec.EnvironmentSpec, stacks []string) []error {
	var errs []error
	for _, s := range stacks {
		workspaceName := WorkspaceName(svc, env, s)
		if err := c.client.Workspaces.Delete(ctx, c.org, workspaceName); err != nil {
			errs = append(errs, errors.Wrapf(err, "workspaces.Delete %q", workspaceName))
		}
	}

	projectName := ProjectName(svc, env)
	projects, err := c.client.Projects.List(ctx, c.org, &tfe.ProjectListOptions{
		Name: projectName,
	})
	if err != nil {
		errs = append(errs, errors.Wrap(err, "Project.List"))
		return errs
	}
	for _, p := range projects.Items {
		if p.Name == projectName {
			if err := c.client.Projects.Delete(ctx, p.ID); err != nil {
				errs = append(errs, errors.Wrapf(err, "projects.Delete %q (%s)", projectName, p.ID))
			}
		}
	}

	return errs
}

func (c *Client) ApplyWorkspace(ctx context.Context, ws Workspace, message string) error {
	_, err := c.client.Runs.Create(ctx, tfe.RunCreateOptions{
		Workspace: ws.workspace,
		AutoApply: pointers.Ptr(true),
		Message:   &message,
	})
	if err != nil {
		return errors.Wrapf(err, "Runs.Create")
	}
	return nil
}

func (c *Client) ensureAccessForTeam(ctx context.Context, project *tfe.Project, currentTeams *tfe.TeamProjectAccessList, teamID string) error {
	var existingAccessID string
	for _, a := range currentTeams.Items {
		if a.Team.ID == teamID {
			existingAccessID = a.ID
		}
	}
	if existingAccessID != "" {
		_, err := c.client.TeamProjectAccess.Update(ctx, existingAccessID, tfe.TeamProjectAccessUpdateOptions{
			Access: pointers.Ptr(tfe.TeamProjectAccessWrite),
		})
		if err != nil {
			return errors.Wrap(err, "TeamAccess.Update")
		}
	} else {
		_, err := c.client.TeamProjectAccess.Add(ctx, tfe.TeamProjectAccessAddOptions{
			Project: &tfe.Project{ID: project.ID},
			Team:    &tfe.Team{ID: teamID},
			Access:  tfe.TeamProjectAccessWrite,
		})
		if err != nil {
			return errors.Wrap(err, "TeamAccess.Add")
		}
	}

	return nil
}
