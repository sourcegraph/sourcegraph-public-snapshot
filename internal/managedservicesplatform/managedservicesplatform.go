package managedservicesplatform

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/internal/stack/options/terraformversion"
	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/internal/stack/options/tfcbackend"
	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/internal/terraform"

	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/internal/stack"
	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/internal/stack/cloudrun"
	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/internal/stack/iam"
	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/internal/stack/project"

	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/spec"
)

type GCPOptions struct {
	ParentFolderID   string
	BillingAccountID string
}

type Renderer struct {
	// OutputDir is the target directory for generated CDKTF assets.
	OutputDir string
	// GCPOptions declares GCP-specific configuration for rendered CDKTF components.
	GCPOptions
}

// RenderEnvironment sets up a CDKTF application comprised of stacks that define
// the infrastructure required to deploy an environment as specified.
func (r *Renderer) RenderEnvironment(
	svc spec.ServiceSpec,
	build spec.BuildSpec,
	env spec.EnvironmentSpec,
) (*CDKTF, error) {
	var (
		projectID = fmt.Sprintf("%s-%s", svc.ID, env.Name)
		stacks    = stack.NewSet(r.OutputDir,
			// Enforce Terraform versions on all stacks
			terraformversion.With(terraform.Version),
			// Use a Terraform Cloud backend on all stacks
			tfcbackend.With(tfcbackend.Config{
				Workspace: func(stackName string) string {
					return fmt.Sprintf("msp-%s-%s", projectID, stackName)
				},
			}))
	)

	// Render all required CDKTF stacks for this environment
	projectOutput, err := project.NewStack(stacks, project.Variables{
		ProjectID:        projectID,
		Name:             svc.Name,
		ParentFolderID:   r.ParentFolderID,
		BillingAccountID: r.BillingAccountID,
		Labels: map[string]string{
			"service":     svc.ID,
			"environment": env.Name,
			"msp":         "true",
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create project stack")
	}
	if _, err := iam.NewStack(stacks, iam.Variables{
		Project: projectOutput.Project,
	}); err != nil {
		return nil, errors.Wrap(err, "failed to create iam stack")
	}
	if _, err = cloudrun.NewStack(stacks, cloudrun.Variables{
		Project:     projectOutput.Project,
		Service:     svc,
		Image:       build.Image,
		Environment: env,
	}); err != nil {
		return nil, errors.Wrap(err, "failed to create cloudrun stack")
	}

	// Return CDKTF representation for caller to synthesize
	return &CDKTF{
		app:    stack.ExtractApp(stacks),
		stacks: stack.ExtractStacks(stacks),
	}, nil
}
