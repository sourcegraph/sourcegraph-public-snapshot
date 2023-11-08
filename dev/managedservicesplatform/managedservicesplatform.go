// Package managedservicesplatform manages infrastructure-as-code using CDKTF
// for Managed Services Platform (MSP) services.
package managedservicesplatform

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack/cloudrun"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack/options/terraformversion"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack/options/tfcbackend"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack/project"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/terraform"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/terraformcloud"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
)

type TerraformCloudOptions struct {
	// Enabled will render all stacks to use a Terraform CLoud workspace as its
	// Terraform state backend with the following format as the workspace name
	// for each stack:
	//
	//  msp-${svc.id}-${env.id}-${stackName}
	//
	// If false, a local backend will be used.
	Enabled bool
}

type GCPOptions struct{}

// Renderer takes MSP service specifications
type Renderer struct {
	// OutputDir is the target directory for generated CDKTF assets.
	OutputDir string
	// TFC declares Terraform-Cloud-specific configuration for rendered CDKTF
	// components.
	TFC TerraformCloudOptions
	// GCPOptions declares GCP-specific configuration for rendered CDKTF components.
	GCP GCPOptions
	// StableGenerate, if true, is propagated to stacks to indicate that any values
	// populated at generation time should not be regenerated.
	StableGenerate bool
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
					return terraformcloud.WorkspaceName(svc, env, stackName)
				},
			}))
	}

	var (
		projectIDPrefix = fmt.Sprintf("%s-%s", svc.ID, env.ID)
		stacks          = stack.NewSet(r.OutputDir, stackSetOptions...)
	)

	// Render all required CDKTF stacks for this environment
	projectOutput, err := project.NewStack(stacks, project.Variables{
		ProjectIDPrefix:       projectIDPrefix,
		ProjectIDSuffixLength: svc.ProjectIDSuffixLength,

		DisplayName: fmt.Sprintf("%s - %s",
			pointers.Deref(svc.Name, svc.ID), env.ID),

		Category: env.Category,
		Labels: map[string]string{
			"service":     svc.ID,
			"environment": env.ID,
			"msp":         "true",
		},
		Services: func() []string {
			if svc.IAM != nil && len(svc.IAM.Services) > 0 {
				return svc.IAM.Services
			}
			return nil
		}(),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create project stack")
	}
	if _, err := cloudrun.NewStack(stacks, cloudrun.Variables{
		ProjectID:             *projectOutput.Project.ProjectId(),
		CloudRunIdentityEmail: *projectOutput.CloudRunIdentity.Email(),

		Service:     svc,
		Image:       build.Image,
		Environment: env,

		StableGenerate: r.StableGenerate,
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
