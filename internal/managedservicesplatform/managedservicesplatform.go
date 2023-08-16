package managedservicesplatform

import (
	"fmt"

	"github.com/hashicorp/terraform-cdk-go/cdktf"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/internal/resource/aspect"
	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/internal/resource/tfcbackend"

	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/internal/stack"
	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/internal/stack/cloudrun"
	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/internal/stack/iam"
	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/internal/stack/project"

	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/internal/terraform"
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

func (r *Renderer) RenderEnvironment(svc spec.ServiceSpec, build spec.BuildSpec, env spec.EnvironmentSpec) (*CDKTF, error) {
	var (
		stacks    = stack.NewSet(r.OutputDir)
		projectID = fmt.Sprintf("%s-%s", svc.ID, env.Name)
	)

	// Render all required CDKTF components
	projectOutput, err := project.NewStack(stacks, project.Variables{
		ProjectID:        projectID,
		Name:             projectID, // TODO
		ParentFolderID:   r.ParentFolderID,
		BillingAccountID: r.BillingAccountID,
		Labels: map[string]string{
			"service":     svc.ID,
			"environment": env.Name,
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create project stack")
	}
	_, err = iam.NewStack(stacks, iam.Variables{
		Project: projectOutput.Project,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create iam stack")
	}
	_, err = cloudrun.NewStack(stacks, cloudrun.Variables{
		Project:     projectOutput.Project,
		Service:     svc,
		Image:       build.Image,
		Environment: env,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cloudrun stack")
	}

	// Apply any required post-processing to our stacks
	for _, s := range stacks.GetStacks() {
		// Apply aspect enforcing Terraform version
		cdktf.Aspects_Of(s.Stack).Add(&aspect.EnforceTerraformVersion{
			TerraformVersion: terraform.Version,
		})

		// Configure a TFC state backend
		_ = tfcbackend.New(s.Stack, tfcbackend.Config{
			Workspace: fmt.Sprintf("msp-%s-%s", projectID, s.Name),
		})
	}

	// Return CDKTF representation for caller to synthesize
	return &CDKTF{stacks: stacks}, nil
}
