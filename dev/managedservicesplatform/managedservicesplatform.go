// Package managedservicesplatform manages infrastructure-as-code using CDKTF
// for Managed Services Platform (MSP) services.
package managedservicesplatform

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack/cloudrun"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack/iam"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack/options/terraformversion"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack/options/tfcbackend"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack/project"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/terraform"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
)

type TerraformCloudOptions struct {
	// Enabled will render all stacks to use a Terraform CLoud workspace as its
	// Terraform state backend with the following format as the workspace name
	// for each stack:
	//
	//  msp-${svc.id}-${env.name}-${stackName}
	//
	// If false, a local backend will be used.
	Enabled bool

	AccessToken string
}

type GCPOptions struct {
	// ProjectID can override the generated project ID, if provided. Otherwise,
	// one is generated using the naming scheme:
	//
	//   ${svc.id}-${env.id}
	//
	// This is useful when importing projects.
	ProjectID *string

	ParentFolderID   string
	BillingAccountID string
}

// Renderer takes MSP service specifications
type Renderer struct {
	// OutputDir is the target directory for generated CDKTF assets.
	OutputDir string
	// TFC declares Terraform-Cloud-specific configuration for rendered CDKTF
	// components.
	TFC TerraformCloudOptions
	// GCPOptions declares GCP-specific configuration for rendered CDKTF components.
	GCP GCPOptions
}

// RenderEnvironment sets up a CDKTF application comprised of stacks that define
// the infrastructure required to deploy an environment as specified.
func (r *Renderer) RenderEnvironment(
	svc spec.ServiceSpec,
	build spec.BuildSpec,
	env spec.EnvironmentSpec,
) (*CDKTF, error) {
	terraformVersion := terraform.Version
	stackSetOptions := []stack.NewStackOption{
		// Enforce Terraform versions on all stacks
		terraformversion.With(terraformVersion),
	}
	if r.TFC.Enabled {
		// Use a Terraform Cloud backend on all stacks
		stackSetOptions = append(stackSetOptions,
			tfcbackend.With(tfcbackend.Config{
				Workspace: func(stackName string) string {
					return fmt.Sprintf("msp-%s-%s-%s", svc.ID, env.ID, stackName)
				},
			}))
	}

	var (
		projectID = pointers.Deref(r.GCP.ProjectID, fmt.Sprintf("%s-%s", svc.ID, env.ID))
		stacks    = stack.NewSet(r.OutputDir, stackSetOptions...)
	)

	// Render all required CDKTF stacks for this environment
	projectOutput, err := project.NewStack(stacks, project.Variables{
		ProjectID:        projectID,
		Name:             pointers.Deref(svc.Name, svc.ID),
		ParentFolderID:   r.GCP.ParentFolderID,
		BillingAccountID: r.GCP.BillingAccountID,
		Labels: map[string]string{
			"service":     svc.ID,
			"environment": env.ID,
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
		app:              stack.ExtractApp(stacks),
		stacks:           stack.ExtractStacks(stacks),
		terraformVersion: terraformVersion,
	}, nil
}
