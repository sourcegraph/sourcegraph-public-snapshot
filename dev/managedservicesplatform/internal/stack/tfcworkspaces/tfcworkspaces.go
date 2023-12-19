package tfcworkspaces

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/tfe/datatfeworkspace"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/tfe/notificationconfiguration"
	tfe "github.com/sourcegraph/managed-services-platform-cdktf/gen/tfe/provider"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/tfe/workspacerun"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/googlesecretsmanager"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/gsmsecret"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack/options/googleprovider"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack/options/tfcbackend"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack/options/tfeprovider"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/terraformcloud"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type CrossStackOutput struct{}

type Variables struct {
	PreviousStacks      []stack.Stack
	EnableNotifications bool
}

const StackName = "tfcworkspaces"

// NewStack creates a stack that applies additional configuration to workspaces
// post-creation, with 'sg msp tfc sync' creating and applying base workspace
// configurations.
func NewStack(stacks *stack.Set, vars Variables) (*CrossStackOutput, error) {
	scope, _, err := stacks.New(StackName,
		// googleprovider only needed for GSM access to set up tfeprovider
		googleprovider.With(googlesecretsmanager.ProjectID),
		tfeprovider.With(gsmsecret.DataConfig{
			Secret:    googlesecretsmanager.SecretTFCMSPTeamToken,
			ProjectID: googlesecretsmanager.ProjectID,
		}))
	if err != nil {
		return nil, err
	}

	// We need another provider to manage things at the org-level.
	tfeOrgProvider := tfe.NewTfeProvider(scope, pointers.Ptr("tfe-org"),
		&tfe.TfeProviderConfig{
			Alias:        pointers.Stringf("tfe-org"),
			Hostname:     pointers.Ptr(terraformcloud.Hostname),
			Organization: pointers.Ptr(terraformcloud.Organization),
			Token: &gsmsecret.Get(scope, resourceid.New("tfe-org-provider-token"), gsmsecret.DataConfig{
				Secret:    googlesecretsmanager.SecretTFCOrgToken,
				ProjectID: googlesecretsmanager.ProjectID,
			}).Value,
		})

	// Apply each workspace in sequence, with each subsequent workspace depending
	// on the previous. We should not have too many stacks, and a linear relationship
	// is easier to reason with. As we create runs, keep assigning the previous
	// run to this variable.
	var previousWorkspaceRun workspacerun.WorkspaceRun

	// First, let's prepare our own workspace.
	{
		self := stack.ExtractCurrentStack(stacks)
		id := resourceid.New(self.Name)
		workspaceName := self.Metadata[tfcbackend.MetadataKeyWorkspace]
		if workspaceName == "" {
			return nil, errors.Wrapf(err, "stack %q missing tfcbackend.MetadataKeyWorkspace",
				self.Name)
		}
		previousWorkspaceRun = configureWorkspace(scope, tfeOrgProvider, id, workspaceName, previousWorkspaceRun, workspaceConfiguration{
			enableNotifications: true,
			// We can't provision a run for ourselves.
			provisionRun: false,
		})
	}

	// Then, apply configuration for all previous stacks.
	for _, s := range vars.PreviousStacks {
		id := resourceid.New(s.Name)
		workspaceName := s.Metadata[tfcbackend.MetadataKeyWorkspace]
		if workspaceName == "" {
			return nil, errors.Wrapf(err, "stack %q missing tfcbackend.MetadataKeyWorkspace",
				s.Name)
		}
		previousWorkspaceRun = configureWorkspace(scope, tfeOrgProvider, id, workspaceName, previousWorkspaceRun, workspaceConfiguration{
			provisionRun:        true,
			enableNotifications: true,
		})
	}

	return &CrossStackOutput{}, nil
}

type workspaceConfiguration struct {
	provisionRun        bool
	enableNotifications bool
}

func configureWorkspace(
	scope cdktf.TerraformStack,
	orgProvider tfe.TfeProvider,
	id resourceid.ID,
	workspaceName string,
	previousWorkspaceRun workspacerun.WorkspaceRun,
	config workspaceConfiguration,
) workspacerun.WorkspaceRun {
	// Workspace must be provisioned by `sg msp tfc`, as we don't want to use
	// TFC itself to provision TFC workspaces (to avoid chicken-and-egg problems)
	workspace := datatfeworkspace.NewDataTfeWorkspace(scope,
		id.TerraformID("workspace"),
		&datatfeworkspace.DataTfeWorkspaceConfig{
			Name:         pointers.Ptr(workspaceName),
			Organization: pointers.Ptr("sourcegraph"),
		})

	// Configure notifications
	_ = notificationconfiguration.NewNotificationConfiguration(scope,
		id.TerraformID("notifications"),
		&notificationconfiguration.NotificationConfigurationConfig{
			Provider: orgProvider, // needs workspace management permissions

			Enabled:     pointers.Ptr(config.enableNotifications),
			Name:        pointers.Ptr("#alerts-msp-tfc"),
			WorkspaceId: workspace.Id(),

			// Send to a single Slack channel for Core Services to monitor
			DestinationType: pointers.Ptr("slack"),
			Url: &gsmsecret.Get(scope, id.Group("slack_webhook"), gsmsecret.DataConfig{
				Secret:    googlesecretsmanager.SecretTFCMSPSlackWebhook,
				ProjectID: googlesecretsmanager.ProjectID,
			}).Value,
			// Trigger options are documented here:
			// https://registry.terraform.io/providers/hashicorp/tfe/latest/docs/resources/notification_configuration#triggers
			Triggers: pointers.Ptr(pointers.Slice([]string{
				"run:errored",
			})),
		})

	// If no run is needed, we are done
	if !config.provisionRun {
		return nil
	}

	// The next workspace should depend on this one, and we should depend on
	// the previous workspace.
	var dependsOn *[]cdktf.ITerraformDependable
	if previousWorkspaceRun != nil {
		dependsOn = &[]cdktf.ITerraformDependable{previousWorkspaceRun}
	}
	return workspacerun.NewWorkspaceRun(scope,
		id.TerraformID("workspace_run"),
		&workspacerun.WorkspaceRunConfig{
			WorkspaceId: workspace.Id(),
			DependsOn:   dependsOn,

			Apply: &workspacerun.WorkspaceRunApply{
				// Automatically start the run
				ManualConfirm: pointers.Ptr(false),
				// Wait for run to complete before resource success
				WaitForRun: pointers.Ptr(true),
			},
			Destroy: &workspacerun.WorkspaceRunDestroy{
				// Automatically start the run
				ManualConfirm: pointers.Ptr(false),
				// Wait for run to complete before resource success
				WaitForRun: pointers.Ptr(true),
			},
		})
}
