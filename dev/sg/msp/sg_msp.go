//go:build msp
// +build msp

package msp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/googlesecretsmanager"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/terraformcloud"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/secrets"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/msp/example"
	msprepo "github.com/sourcegraph/sourcegraph/dev/sg/msp/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/msp/schema"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

// This file is only built when '-tags=msp' is passed to go build while 'sg msp'
// is experimental, as the introduction of this command currently increases the
// the binary size of 'sg' by ~20%.
//
// To install a variant of 'sg' with 'sg msp' enabled, run:
//
//   go build -tags=msp -o=./sg ./dev/sg && ./sg install -f -p=false
//
// To work with msp in VS Code, add the following to your VS Code configuration:
//
//  "gopls": {
//     "build.buildFlags": ["-tags=msp"]
//  }

func init() {
	// Override no-op implementation with our real implementation.
	Command.Hidden = false
	Command.Action = nil
	// Trim description to just be the command description
	Command.Description = commandDescription
	// All 'sg msp ...' subcommands
	Command.Subcommands = []*cli.Command{
		{
			Name:      "init",
			ArgsUsage: "<service ID>",
			Usage:     "Initialize a template Managed Services Platform service spec",
			UsageText: `
sg msp init -owner core-services -name "MSP Example Service" msp-example
`,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "kind",
					Usage: "Kind of service (one of: 'service', 'job')",
					Value: "service",
				},
				&cli.StringFlag{
					Name:     "owner",
					Usage:    "Name of team owning this new service",
					Required: true,
				},
				&cli.StringFlag{
					Name:  "name",
					Usage: "Specify a human-readable name for this service",
				},
				&cli.BoolFlag{
					Name:  "dev",
					Usage: "Generate a dev environment as the initial environment",
				},
			},
			Before: msprepo.UseManagedServicesRepo,
			Action: func(c *cli.Context) error {
				if c.Args().Len() != 1 {
					return errors.New("exactly 1 argument required: service ID")
				}

				template := example.Template{
					ID:    c.Args().First(),
					Name:  c.String("name"),
					Owner: c.String("owner"),
					Dev:   c.Bool("dev"),
				}

				var exampleSpec []byte
				switch c.String("kind") {
				case "service":
					var err error
					exampleSpec, err = example.NewService(template)
					if err != nil {
						return errors.Wrap(err, "example.NewService")
					}
				case "job":
					var err error
					exampleSpec, err = example.NewJob(template)
					if err != nil {
						return errors.Wrap(err, "example.NewJob")
					}
				default:
					return errors.Newf("unsupported service kind: %q", c.String("kind"))
				}

				outputPath := msprepo.ServiceYAMLPath(c.Args().First())

				_ = os.MkdirAll(filepath.Dir(outputPath), 0755)
				if err := os.WriteFile(outputPath, exampleSpec, 0644); err != nil {
					return err
				}

				std.Out.WriteSuccessf("Rendered %s template spec in %s",
					c.String("kind"), outputPath)
				return nil
			},
		},
		{
			Name:      "generate",
			Aliases:   []string{"gen"},
			ArgsUsage: "<service ID>",
			Usage:     "Generate Terraform assets for a Managed Services Platform service spec.",
			Description: `Optionally use '-all' to sync all environments for a service.

Supports completions on services and environments.`,
			UsageText: `
# generate single env for a single service
sg msp generate <service> <env>
# generate all envs across all services
sg msp generate -all
# generate all envs for a single service
sg msp generate -all <service>
			`,
			Before: msprepo.UseManagedServicesRepo,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "all",
					Usage: "Generate infrastructure stacks for all services, or all envs for a service if service ID is provided",
					Value: false,
				},
				&cli.BoolFlag{
					Name:  "stable",
					Usage: "Disable updating of any values that are evaluated at generation time",
					Value: false,
				},
				&cli.BoolFlag{
					Name:  "tfc",
					Usage: "Generate infrastructure stacks with Terraform Cloud backends",
					Value: true,
				},
			},
			BashComplete: msprepo.ServicesAndEnvironmentsCompletion(),
			Action: func(c *cli.Context) error {
				var (
					generateAll    = c.Bool("all")
					stableGenerate = c.Bool("stable")
					useTFC         = c.Bool("tfc")
				)

				if stableGenerate {
					std.Out.WriteSuggestionf("Using stable generate - tfvars will not be updated.")
				}

				// Generate specific service
				if serviceID := c.Args().First(); serviceID == "" && !generateAll {
					return errors.New("first argument service ID is required without the '-all' flag")
				} else if serviceID != "" {
					targetEnv := c.Args().Get(1)
					if targetEnv == "" && !generateAll {
						return errors.New("second argument environment ID is required without the '-all' flag")
					}

					return generateTerraform(serviceID, generateTerraformOptions{
						targetEnv:      targetEnv,
						stableGenerate: stableGenerate,
						useTFC:         useTFC,
					})
				}

				// Generate all services
				serviceIDs, err := msprepo.ListServices()
				if err != nil {
					return errors.Wrap(err, "list services")
				}
				if len(serviceIDs) == 0 {
					return errors.New("no services found")
				}
				for _, serviceID := range serviceIDs {
					if err := generateTerraform(serviceID, generateTerraformOptions{
						stableGenerate: stableGenerate,
						useTFC:         useTFC,
					}); err != nil {
						return errors.Wrap(err, serviceID)
					}
				}
				return nil
			},
		},
		{
			Name:    "terraform-cloud",
			Aliases: []string{"tfc"},
			Usage:   "Manage Terraform Cloud workspaces for a service",
			Before:  msprepo.UseManagedServicesRepo,
			Subcommands: []*cli.Command{
				{
					Name:  "sync",
					Usage: "Create or update all required Terraform Cloud workspaces for an environment",
					Description: `Optionally use '-all' to sync all environments for a service.

Supports completions on services and environments.`,
					ArgsUsage: "<service ID> [environment ID]",
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:  "all",
							Usage: "Generate Terraform Cloud workspaces for all environments",
							Value: false,
						},
						&cli.StringFlag{
							Name:  "workspace-run-mode",
							Usage: "One of 'vcs', 'cli'",
							Value: "vcs",
						},
						&cli.BoolFlag{
							Name:  "delete",
							Usage: "Delete workspaces and projects - does NOT apply a teardown run",
							Value: false,
						},
						&cli.BoolFlag{
							Name:  "apply-created",
							Usage: "Apply newly created workspaces - only valid in 'vcs' mode",
							Value: true,
						},
						&cli.BoolFlag{
							Name:  "apply-updated",
							Usage: "Apply updated workspaces - only valid in 'vcs' mode",
							Value: false,
						},
					},
					BashComplete: msprepo.ServicesAndEnvironmentsCompletion(),
					Action: func(c *cli.Context) error {
						serviceID := c.Args().First()
						if serviceID == "" {
							return errors.New("argument service is required")
						}
						serviceSpecPath := msprepo.ServiceYAMLPath(serviceID)

						service, err := spec.Open(serviceSpecPath)
						if err != nil {
							return err
						}

						secretStore, err := secrets.FromContext(c.Context)
						if err != nil {
							return err
						}
						tfcAccessToken, err := secretStore.GetExternal(c.Context, secrets.ExternalSecret{
							Name:    googlesecretsmanager.SecretTFCAccessToken,
							Project: googlesecretsmanager.ProjectID,
						})
						if err != nil {
							return errors.Wrap(err, "get AccessToken")
						}
						tfcOAuthClient, err := secretStore.GetExternal(c.Context, secrets.ExternalSecret{
							Name:    googlesecretsmanager.SecretTFCOAuthClientID,
							Project: googlesecretsmanager.ProjectID,
						})
						if err != nil {
							return errors.Wrap(err, "get TFC OAuth client ID")
						}

						tfcClient, err := terraformcloud.NewClient(tfcAccessToken, tfcOAuthClient,
							terraformcloud.WorkspaceConfig{
								RunMode: terraformcloud.WorkspaceRunMode(c.String("workspace-run-mode")),
							})
						if err != nil {
							return errors.Wrap(err, "init Terraform Cloud client")
						}

						if targetEnv := c.Args().Get(1); targetEnv != "" {
							env := service.GetEnvironment(targetEnv)
							if env == nil {
								return errors.Newf("environment %q not found in service spec", targetEnv)
							}

							if err := syncEnvironmentWorkspaces(c, tfcClient, service.Service, service.Build, *env, service.Monitoring); err != nil {
								return errors.Wrapf(err, "sync env %q", env.ID)
							}
						} else {
							if targetEnv == "" && !c.Bool("all") {
								return errors.New("second argument environment ID is required without the '-all' flag")
							}
							for _, env := range service.Environments {
								if err := syncEnvironmentWorkspaces(c, tfcClient, service.Service, service.Build, env, service.Monitoring); err != nil {
									return errors.Wrapf(err, "sync env %q", env.ID)
								}
							}
						}

						return nil
					},
				},
			},
		},
		{
			Name:  "schema",
			Usage: "Generate JSON schema definition for service specification",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "output",
					Aliases: []string{"o"},
					Usage:   "Output path for generated schema",
				},
			},
			Action: func(c *cli.Context) error {
				jsonSchema, err := schema.Render()
				if err != nil {
					return err
				}
				if output := c.String("output"); output != "" {
					_ = os.Remove(output)
					if err := os.WriteFile(output, jsonSchema, 0644); err != nil {
						return err
					}
					std.Out.WriteSuccessf("Rendered service spec JSON schema in %s", output)
					return nil
				}
				// Otherwise render it for reader
				return std.Out.WriteCode("json", string(jsonSchema))
			},
		},
	}
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
		TFC: managedservicesplatform.TerraformCloudOptions{},
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
		summary.WriteString(fmt.Sprintf("- %s: %s", ws.Name(), ws.URL()))
		if ws.Created {
			if ws.RunMode == terraformcloud.WorkspaceRunModeVCS && c.Bool("apply-created") {
				if err := tfc.ApplyWorkspace(c.Context, ws, "sg msp: apply newly created workspace"); err != nil {
					// not fatal, just add error to summary
					summary.WriteString(fmt.Sprintf(" (created, failed to run: %s)",
						err.Error()))
				} else {
					summary.WriteString(" (created, running)")
				}
			} else {
				summary.WriteString(" (created)")
			}
		} else {
			if ws.RunMode == terraformcloud.WorkspaceRunModeVCS && c.Bool("apply-updated") {
				if err := tfc.ApplyWorkspace(c.Context, ws, "sg msp: apply updated workspace"); err != nil {
					// not fatal, just add error to summary
					summary.WriteString(fmt.Sprintf(" (updated, failed to run: %s)",
						err.Error()))
				} else {
					summary.WriteString(" (updated, running)")
				}
			} else {
				summary.WriteString(" (updated)")
			}
		}
		summary.WriteString("\n")
	}
	std.Out.WriteMarkdown(summary.String())
	return nil
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
		cdktf, err := renderer.RenderEnvironment(service.Service, service.Build, env, service.Monitoring)
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
