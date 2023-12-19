// Package msp exports the 'sg msp' command for the Managed Services Platform.
package msp

import (
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/googlesecretsmanager"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/terraformcloud"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/category"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/secrets"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/msp/example"
	msprepo "github.com/sourcegraph/sourcegraph/dev/sg/msp/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/msp/schema"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Command is the 'sg msp' toolchain for the Managed Services Platform:
// https://handbook.sourcegraph.com/departments/engineering/teams/core-services/managed-services/platform
var Command = &cli.Command{
	Name:    "managed-services-platform",
	Aliases: []string{"msp"},
	Usage:   "EXPERIMENTAL: Generate and manage services deployed on the Sourcegraph Managed Services Platform",
	Description: `WARNING: MSP is currently still an experimental project.
To learm more, refer to go/rfc-msp and go/msp (https://handbook.sourcegraph.com/departments/engineering/teams/core-services/managed-services/platform)`,
	UsageText: `
# Create a service specification
sg msp init $SERVICE

# Provision Terraform Cloud workspaces
sg msp tfc sync $SERVICE $ENVIRONMENT

# Generate Terraform manifests
sg msp generate $SERVICE $ENVIRONMENT
`,
	Category: category.Company,
	Subcommands: []*cli.Command{
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
			Usage:     "Generate Terraform assets for a Managed Services Platform service spec",
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
							Name:    googlesecretsmanager.SecretTFCOrgToken,
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

							if err := syncEnvironmentWorkspaces(c, tfcClient, service.Service, service.Build, *env, *service.Monitoring); err != nil {
								return errors.Wrapf(err, "sync env %q", env.ID)
							}
						} else {
							if targetEnv == "" && !c.Bool("all") {
								return errors.New("second argument environment ID is required without the '-all' flag")
							}
							for _, env := range service.Environments {
								if err := syncEnvironmentWorkspaces(c, tfcClient, service.Service, service.Build, env, *service.Monitoring); err != nil {
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
	},
}
