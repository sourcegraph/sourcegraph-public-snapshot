package msp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/terraformcloud"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	msprepo "github.com/sourcegraph/sourcegraph/dev/sg/msp/repo"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

// useServiceArgument retrieves the service spec corresponding to the first
// argument.
func useServiceArgument(c *cli.Context) (*spec.Spec, error) {
	serviceID := c.Args().First()
	if serviceID == "" {
		return nil, errors.New("argument service is required")
	}
	serviceSpecPath := msprepo.ServiceYAMLPath(serviceID)

	return spec.Open(serviceSpecPath)
}

func syncEnvironmentWorkspaces(c *cli.Context, tfc *terraformcloud.Client, service spec.ServiceSpec, build spec.BuildSpec, env spec.EnvironmentSpec, monitoring spec.MonitoringSpec) error {
	if os.TempDir() == "" {
		return errors.New("no temp dir available")
	}

	renderer := &managedservicesplatform.Renderer{
		// Even though we're not synthesizing we still
		// need an output dir or CDKTF will not work
		OutputDir: filepath.Join(os.TempDir(), fmt.Sprintf("msp-tfc-%s-%s-%d",
			service.ID, env.ID, time.Now().Unix())),
		GCP: managedservicesplatform.GCPOptions{},
		TFC: managedservicesplatform.TerraformCloudOptions{
			Enabled: true, // required to generate all workspaces
		},
	}
	defer os.RemoveAll(renderer.OutputDir)

	renderPending := std.Out.Pending(output.Styledf(output.StylePending,
		"[%s] Rendering required Terraform Cloud workspaces for environment %q",
		service.ID, env.ID))
	cdktf, err := renderer.RenderEnvironment(service, build, env, monitoring)
	if err != nil {
		return err
	}
	renderPending.Destroy() // We need to destroy this pending so we can prompt on deletion.

	if c.Bool("delete") {
		std.Out.Promptf("[%s] Deleting workspaces for environment %q - are you sure? (y/N) ",
			service.ID, env.ID)
		var input string
		if _, err := fmt.Scan(&input); err != nil {
			return err
		}
		if input != "y" {
			return errors.New("aborting")
		}

		pending := std.Out.Pending(output.Styledf(output.StylePending,
			"[%s] Deleting Terraform Cloud workspaces for environment %q", service.ID, env.ID))
		if errs := tfc.DeleteWorkspaces(c.Context, service, env, cdktf.Stacks()); len(errs) > 0 {
			for _, err := range errs {
				std.Out.WriteWarningf(err.Error())
			}
			return errors.New("some errors occurred when deleting workspaces")
		}
		pending.Complete(output.Styledf(output.StyleSuccess,
			"[%s] Deleting Terraform Cloud workspaces for environment %q", service.ID, env.ID))

		return nil // exit early for deletion, we are done
	}

	pending := std.Out.Pending(output.Styledf(output.StylePending,
		"[%s] Synchronizing Terraform Cloud workspaces for environment %q", service.ID, env.ID))
	workspaces, err := tfc.SyncWorkspaces(c.Context, service, env, cdktf.Stacks())
	if err != nil {
		return errors.Wrap(err, "sync Terraform Cloud workspace")
	}
	pending.Complete(output.Styledf(output.StyleSuccess,
		"[%s] Synchronized Terraform Cloud workspaces for environment %q", service.ID, env.ID))

	var summary strings.Builder
	for _, ws := range workspaces {
		summary.WriteString(fmt.Sprintf("- %s: %s", ws.Name, ws.URL()))
		if ws.Created {
			summary.WriteString(" (created)")
		} else {
			summary.WriteString(" (updated)")
		}
		summary.WriteString("\n")
	}
	return std.Out.WriteMarkdown(summary.String())
}

type generateTerraformOptions struct {
	// targetEnv generates the specified env only, otherwise generates all
	targetEnv string
	// stableGenerate disables updating of any values that are evaluated at
	// generation time
	stableGenerate bool
	// useTFC enables Terraform Cloud integration
	useTFC bool
}

func generateTerraform(serviceID string, opts generateTerraformOptions) error {
	serviceSpecPath := msprepo.ServiceYAMLPath(serviceID)

	service, err := spec.Open(serviceSpecPath)
	if err != nil {
		return err
	}

	var envs []spec.EnvironmentSpec
	if opts.targetEnv != "" {
		deployEnv := service.GetEnvironment(opts.targetEnv)
		if deployEnv == nil {
			return errors.Newf("environment %q not found in service spec", opts.targetEnv)
		}
		envs = append(envs, *deployEnv)
	} else {
		envs = service.Environments
	}

	for _, env := range envs {
		env := env

		pending := std.Out.Pending(output.Styledf(output.StylePending,
			"[%s] Preparing Terraform for environment %q", serviceID, env.ID))
		renderer := managedservicesplatform.Renderer{
			OutputDir: filepath.Join(filepath.Dir(serviceSpecPath), "terraform", env.ID),
			GCP:       managedservicesplatform.GCPOptions{},
			TFC: managedservicesplatform.TerraformCloudOptions{
				Enabled: opts.useTFC,
			},
			StableGenerate: opts.stableGenerate,
		}

		// CDKTF needs the output dir to exist ahead of time, even for
		// rendering. If it doesn't exist yet, create it
		if f, err := os.Lstat(renderer.OutputDir); err != nil {
			if !os.IsNotExist(err) {
				return errors.Wrap(err, "check output directory")
			}
			if err := os.MkdirAll(renderer.OutputDir, 0755); err != nil {
				return errors.Wrap(err, "prepare output directory")
			}
		} else if !f.IsDir() {
			return errors.Newf("output directory %q is not a directory", renderer.OutputDir)
		}

		// Render environment
		cdktf, err := renderer.RenderEnvironment(service.Service, service.Build, env, *service.Monitoring)
		if err != nil {
			return err
		}

		pending.Updatef("[%s] Generating Terraform assets in %q for environment %q...",
			serviceID, renderer.OutputDir, env.ID)
		if err := cdktf.Synthesize(); err != nil {
			return err
		}
		pending.Complete(output.Styledf(output.StyleSuccess,
			"[%s] Terraform assets generated in %q!", serviceID, renderer.OutputDir))
	}

	return nil
}
