// Package msp exports the 'sg msp' command for the Managed Services Platform.
package msp

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/urfave/cli/v2"
	"golang.org/x/exp/maps"

	"github.com/sourcegraph/run"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/googlesecretsmanager"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/operationdocs"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/stacks"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/stacks/cloudrun"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/stacks/iam"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/terraformcloud"
	"github.com/sourcegraph/sourcegraph/dev/sg/cloudsqlproxy"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/category"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/open"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/secrets"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/msp/example"
	msprepo "github.com/sourcegraph/sourcegraph/dev/sg/msp/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/msp/schema"
	"github.com/sourcegraph/sourcegraph/lib/cliutil/completions"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
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
					Name:  "owner",
					Usage: "Name of team owning this new service",
				},
				&cli.StringFlag{
					Name:  "name",
					Usage: "Specify a human-readable name for this service",
				},
				&cli.BoolFlag{
					Name:  "dev",
					Usage: "Generate a dev environment",
				},
				&cli.IntFlag{
					Name:  "project-id-suffix-length",
					Usage: "Length of random suffix appended to generated project IDs",
					Value: spec.DefaultSuffixLength,
				},
			},
			Before: msprepo.UseManagedServicesRepo,
			Action: func(c *cli.Context) error {
				if c.Args().Len() > 1 {
					return errors.New("exactly 1 argument allowed: the desired service ID, or no arguments to use interactive setup")
				}

				// Track if no args were provided at all to guide interactive
				// setup features
				fullyInteractive := c.Args().Len() == 0

				// Collect required inputs
				template := example.Template{
					ID:    c.Args().First(),
					Name:  c.String("name"),
					Owner: c.String("owner"),
					Dev:   c.Bool("dev"),

					ProjectIDSuffixLength: c.Int("project-id-suffix-length"),
				}
				if template.ID == "" {
					std.Out.Write("Please provide an all-lowercase, dash-delimited, machine-friendly identifier for your new service, e.g. 'my-service'.")
					ok, err := std.PromptAndScan(std.Out, "Service ID:", &template.ID)
					if err != nil {
						return err
					} else if !ok {
						return errors.New("response is required")
					}
				}
				if allServices, err := msprepo.ListServices(); err != nil {
					return errors.Wrap(err, "checking existing services")
				} else if slices.Contains(allServices, template.ID) {
					return errors.Newf("service with ID %q already exists", template.ID)
				}
				if template.Name == "" {
					std.Out.Write("Please provide a human-readable name for your new service, e.g. 'My Service'.")
					// optional, we can automatically generate one
					if _, err := std.PromptAndScan(std.Out, "Service name (optional):", &template.Name); err != nil {
						return err
					}
				}
				if template.Owner == "" {
					std.Out.Write("Please provide the name of the Opsgenie team that owns this new service - this MUST be an existing team listed in https://sourcegraph.app.opsgenie.com/teams/list")
					ok, err := std.PromptAndScan(std.Out, "Service owner:", &template.Owner)
					if err != nil {
						return err
					} else if !ok {
						return errors.New("response is required")
					}
				}
				if fullyInteractive && !c.IsSet("dev") { // ask only in interactive setup
					std.Out.Write("We are going to scaffold an initial environment for your service - do you want to start with a 'dev' environment?")
					std.Out.WriteSuggestionf("You can scaffold additional environments later using 'sg msp init-env %s'.", template.ID)
					var dev string
					ok, err := std.PromptAndScan(std.Out, "Start with a 'dev' environment (y/N):", &dev)
					if err != nil {
						return err
					} else if !ok {
						return errors.New("response is required")
					}
					template.Dev = strings.EqualFold(dev, "y")
				}

				var kind = c.String("kind")
				if fullyInteractive && !c.IsSet("kind") { // ask only in interactive setup
					std.Out.Write("MSP supports long-running services, or cron jobs.")
					ok, err := std.PromptAndScan(std.Out, "Service kind (one of: 'service', 'job'):", &kind)
					if err != nil {
						return err
					} else if !ok {
						return errors.New("response is required")
					}
				}

				var exampleSpec []byte
				switch kind {
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
					return errors.Newf("unsupported service kind: %q", kind)
				}

				outputPath := msprepo.ServiceYAMLPath(template.ID)

				_ = os.MkdirAll(filepath.Dir(outputPath), 0o755)
				if err := os.WriteFile(outputPath, exampleSpec, 0o644); err != nil {
					return err
				}

				std.Out.WriteSuccessf("Rendered %s template spec in %s",
					c.String("kind"), outputPath)

				std.Out.WriteSuggestionf("Take a look at the spec to see what you can change! "+
					"When you are done, run 'sg msp generate -all %s' to render the required manifests and assets, and open a pull request for Core Services review.",
					template.ID)
				return nil
			},
		},
		{
			Name:      "init-env",
			ArgsUsage: "<service ID> <env ID>",
			Usage:     "Add an environment to an existing Managed Services Platform service",
			Flags: []cli.Flag{
				&cli.IntFlag{
					Name:  "project-id-suffix-length",
					Usage: "Length of random suffix appended to generated project IDs",
					Value: spec.DefaultSuffixLength,
				},
			},
			Before: msprepo.UseManagedServicesRepo,
			BashComplete: completions.CompleteArgs(func() (options []string) {
				ss, _ := msprepo.ListServices()
				return ss
			}),
			Action: func(c *cli.Context) error {
				svc, err := useServiceArgument(c, false) // we're expecting a potential second argument
				if err != nil {
					// A bad argument suggests a user misunderstanding of this
					// command, so provide a hint with the error
					return errors.Wrap(err,
						"this command is for adding an environment to an existing service, did you mean to use 'sg msp init' instead?")
				}

				envID := c.Args().Get(1)
				if envID == "" {
					std.Out.Write("Please provide an all-lowercase, dash-delimited, machine-friendly identifier for your new environment, e.g. 'dev' or 'prod'.")
					ok, err := std.PromptAndScan(std.Out, "Environment ID:", &envID)
					if err != nil {
						return err
					} else if !ok {
						return errors.New("response is required")
					}
				}
				if existing := svc.GetEnvironment(envID); existing != nil {
					return errors.Newf("environment %q already exists for service %q", envID, svc.Service.ID)
				}

				envNode, err := example.NewEnvironment(example.EnvironmentTemplate{
					ServiceID:             svc.Service.ID,
					EnvironmentID:         envID,
					ProjectIDSuffixLength: c.Int("project-id-suffix-length"),
				})
				if err != nil {
					return errors.Wrap(err, "example.NewEnvironment")
				}

				specPath := msprepo.ServiceYAMLPath(svc.Service.ID)
				specData, err := os.ReadFile(specPath)
				if err != nil {
					return errors.Wrap(err, "ReadFile")
				}

				specData, err = spec.AppendEnvironment(specData, envNode)
				if err != nil {
					return errors.Wrap(err, "spec.AppendEnvironment")
				}

				if err := os.WriteFile(specPath, specData, 0o644); err != nil {
					return err
				}

				std.Out.WriteSuccessf("Initialized environment %q in %s",
					envID, specPath)
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
# generate all envs for a single service
sg msp generate -all <service>
# generate all envs across all services
sg msp generate -all
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
					Usage: "Configure updating of any values that are evaluated at generation time",
					Value: true,
				},
			},
			BashComplete: msprepo.ServicesAndEnvironmentsCompletion(),
			Action: func(c *cli.Context) error {
				var (
					generateAll    = c.Bool("all")
					stableGenerate = c.Bool("stable")
				)

				if stableGenerate {
					std.Out.WriteSuggestionf("Using stable generate - tfvars will not be updated.")
				}

				// Generate a specific service environment if '-all' is not provided
				if !generateAll {
					std.Out.WriteNoticef("Generating a specific service environment...")
					svc, env, err := useServiceAndEnvironmentArguments(c, true)
					if err != nil {
						return err
					}
					return generateTerraform(svc, generateTerraformOptions{
						targetEnv:      env.ID,
						stableGenerate: stableGenerate,
					})
				}

				// 1+ argument indicates we are generating all envs for a single service
				if c.Args().Len() > 0 {
					std.Out.WriteNoticef("Generating all environments for a specific service...")
					svc, err := useServiceArgument(c, true) // error if additional arguments are provided
					if err != nil {
						return err
					}
					return generateTerraform(svc, generateTerraformOptions{
						stableGenerate: stableGenerate,
					})
				}

				// Otherwise, generate all environments for all services
				serviceIDs, err := msprepo.ListServices()
				if err != nil {
					return errors.Wrap(err, "list services")
				}
				if len(serviceIDs) == 0 {
					return errors.New("no services found")
				}
				for _, serviceID := range serviceIDs {
					s, err := spec.Open(msprepo.ServiceYAMLPath(serviceID))
					if err != nil {
						return err
					}
					if err := generateTerraform(s, generateTerraformOptions{
						stableGenerate: stableGenerate,
					}); err != nil {
						return errors.Wrap(err, serviceID)
					}
				}
				return nil
			},
		},
		{
			Name:    "operations",
			Aliases: []string{"ops"},
			Usage:   "Generate operational reference for a service",
			Before:  msprepo.UseManagedServicesRepo,
			BashComplete: completions.CompleteArgs(func() (options []string) {
				ss, _ := msprepo.ListServices()
				return ss
			}),
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "pretty",
					Usage: "Render syntax-highlighed Markdown",
					Value: true,
				},
			},
			Action: func(c *cli.Context) error {
				svc, err := useServiceArgument(c, true)
				if err != nil {
					return err
				}

				repoRev, err := msprepo.GitRevision(c.Context)
				if err != nil {
					return errors.Wrap(err, "msprepo.GitRevision")
				}

				collectedAlerts, err := CollectAlertPolicies(svc)
				if err != nil {
					return errors.Wrap(err, "CollectAlertPolicies")
				}

				doc, err := operationdocs.Render(*svc, operationdocs.Options{
					ManagedServicesRevision: repoRev,
					AlertPolicies:           collectedAlerts,
				})
				if err != nil {
					return errors.Wrap(err, "operationdocs.Render")
				}
				if c.Bool("pretty") {
					return std.Out.WriteCode("markdown", doc)
				}
				std.Out.Write(doc)
				return nil
			},
			Subcommands: []*cli.Command{
				{
					Name:   "generate-handbook-pages",
					Usage:  "Generate operations handbook pages for all services",
					Hidden: true, // not meant for day-to-day use
					Description: `By default, we expect the 'sourcegraph/handbook' repository to be checked out adjacent to the 'sourcegraph/managed-services' repository, i.e.:
	/
	├─ managed-services/ <-- current directory
	├─ handbook/         <-- github.com/sourcegraph/handbook

The '-handbook-path' flag can also be used to specify where sourcegraph/handbook is cloned.`,
					Before: msprepo.UseManagedServicesRepo,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "handbook-path",
							Usage: "Path to the directory in which sourcegraph/handbook is cloned",
							Value: "../handbook",
							Action: func(_ *cli.Context, v string) error {
								// 'Required: true' will error out even if a default
								// value is set, so do our own validation here.
								if v == "" {
									return errors.New("cannot be empty")
								}
								return nil
							},
						},
					},
					Action: func(c *cli.Context) error {
						handbookPath := c.String("handbook-path")
						if err := isHandbookRepo(handbookPath); err != nil {
							return errors.Wrapf(err, "expecting github.com/sourcegraph/handbook at %q", handbookPath)
						}

						services, err := msprepo.ListServices()
						if err != nil {
							return err
						}

						repoRev, err := msprepo.GitRevision(c.Context)
						if err != nil {
							return errors.Wrap(err, "msprepo.GitRevision")
						}

						opts := operationdocs.Options{
							ManagedServicesRevision: repoRev,
							GenerateCommand:         strings.Join(os.Args, " "),
							Handbook:                true,
						}

						// Reset directory to ensure we don't have lingering references
						{
							dir := filepath.Join(handbookPath, operationdocs.HandbookDirectory)
							_ = os.RemoveAll(dir)
							_ = os.Mkdir(dir, os.ModePerm)
							std.Out.Writef("Reset destination directory %q.", dir)
						}

						var serviceSpecs []*spec.Spec
						for _, s := range services {
							svc, err := spec.Open(msprepo.ServiceYAMLPath(s))
							if err != nil {
								return errors.Wrapf(err, "load service %q", s)
							}
							serviceSpecs = append(serviceSpecs, svc)
							collectedAlerts, err := CollectAlertPolicies(svc)
							opts.AlertPolicies = collectedAlerts
							if err != nil {
								return errors.Wrapf(err, "%s: CollectAlertPolicies", s)
							}
							doc, err := operationdocs.Render(*svc, opts)
							if err != nil {
								return errors.Wrap(err, s)
							}
							pagePath := filepath.Join(handbookPath,
								operationdocs.ServiceHandbookPath(s))
							if err := os.WriteFile(pagePath, []byte(doc), 0o644); err != nil {
								return errors.Wrap(err, s)
							}
							std.Out.WriteNoticef("[%s]\tWrote %q", s, pagePath)
						}

						indexDoc := operationdocs.RenderIndexPage(serviceSpecs, opts)
						indexPath := filepath.Join(handbookPath,
							operationdocs.IndexPathHandbookPath())
						if err := os.WriteFile(indexPath, []byte(indexDoc), 0o644); err != nil {
							return errors.Wrap(err, "index page")
						}
						std.Out.WriteNoticef("[index]\tWrote %q", indexPath)

						std.Out.WriteSuccessf("All pages generated!")
						std.Out.WriteSuggestionf("Make sure to commit the generated changes and open a pull request in github.com/sourcegraph/handbook.")
						return nil
					},
				},
			},
		},
		{
			Name:   "logs",
			Usage:  "Quick links for logs of various MSP components",
			Before: msprepo.UseManagedServicesRepo,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "component",
					Aliases: []string{"c"},
					Value:   "service",
				},
			},
			BashComplete: msprepo.ServicesAndEnvironmentsCompletion(),
			Action: func(c *cli.Context) error {
				svc, env, err := useServiceAndEnvironmentArguments(c, true)
				if err != nil {
					return err
				}

				switch component := c.String("component"); component {
				case "service":
					std.Out.WriteNoticef("Opening link to service logs in browser...")
					return open.URL(operationdocs.ServiceLogsURL(pointers.DerefZero(svc.Service.Kind), env.ProjectID))

				default:
					return errors.Newf("unsupported -component=%s", component)
				}
			},
		},
		{
			Name:    "postgresql",
			Aliases: []string{"pg"},
			Usage:   "Interact with PostgreSQL instances provisioned by MSP",
			Before:  msprepo.UseManagedServicesRepo,
			Subcommands: []*cli.Command{
				{
					Name:  "connect",
					Usage: "Connect to the PostgreSQL instance",
					Description: `
This command runs 'cloud-sql-proxy' authenticated against the specified MSP
service environment, and provides 'psql' commands for interacting with the
database through the proxy.

If this is your first time using this command, include the '-download' flag to
install 'cloud-sql-proxy'.

By default, you will only have 'SELECT' privileges through the connection - for
full access, use the '-write-access' flag.
`,
					ArgsUsage: "<service ID> <environment ID>",
					Flags: []cli.Flag{
						&cli.IntFlag{
							Name:  "port",
							Value: 5433,
							Usage: "Port to use for the cloud-sql-proxy",
						},
						&cli.BoolFlag{
							Name:  "download",
							Usage: "Install or update the cloud-sql-proxy",
						},
						&cli.BoolFlag{
							Name:  "write-access",
							Usage: "Connect to the database with write access - by default, only select access is granted.",
						},
						// db proxy provides privileged access to the database,
						// so we want to avoid having it dangling around for too long unattended
						&cli.IntFlag{
							Name:  "session.timeout",
							Usage: "Timeout for the proxy session in seconds - 0 means no timeout",
							Value: 300,
						},
					},
					BashComplete: msprepo.ServicesAndEnvironmentsCompletion(),
					Action: func(c *cli.Context) error {
						svc, env, err := useServiceAndEnvironmentArguments(c, true)
						if err != nil {
							return err
						}
						if env.Resources.PostgreSQL == nil {
							return errors.New("no postgresql instance provisioned")
						}

						err = cloudsqlproxy.Init(c.Bool("download"))
						if err != nil {
							return err
						}

						secretStore, err := secrets.FromContext(c.Context)
						if err != nil {
							return err
						}

						var serviceAccountEmail string
						if c.Bool("write-access") {
							// Use the workload identity if all access is requested
							serviceAccountEmail, err = secretStore.GetExternal(c.Context, secrets.ExternalSecret{
								Name:    stacks.OutputSecretID(iam.StackName, iam.OutputCloudRunServiceAccount),
								Project: env.ProjectID,
							})
							if err != nil {
								return errors.Wrap(err, "find IAM output")
							}
							std.Out.WriteAlertf("Preparing a connection with write access - proceed with caution!")
						} else {
							// Otherwise, use the operator access account which
							// is a bit more limited.
							serviceAccountEmail, err = secretStore.GetExternal(c.Context, secrets.ExternalSecret{
								Name:    stacks.OutputSecretID(iam.StackName, iam.OutputOperatorServiceAccount),
								Project: env.ProjectID,
							})
							if err != nil {
								return errors.Wrap(err, "find IAM output")
							}
							std.Out.WriteSuggestionf("Preparing a connection with read-only access - for write access, use the '-write-access' flag.")
						}

						connectionName, err := secretStore.GetExternal(c.Context, secrets.ExternalSecret{
							Name:    stacks.OutputSecretID(cloudrun.StackName, cloudrun.OutputCloudSQLConnectionName),
							Project: env.ProjectID,
						})
						if err != nil {
							return errors.Wrap(err, "find Cloud Run output")
						}

						proxyPort := c.Int("port")
						proxy, err := cloudsqlproxy.NewCloudSQLProxy(
							connectionName,
							serviceAccountEmail,
							proxyPort,
							svc.Service.GetGoLink(env.ID))
						if err != nil {
							return err
						}

						for _, db := range env.Resources.PostgreSQL.Databases {
							std.Out.WriteNoticef("Use this command to connect to database %q:", db)

							saUsername := strings.ReplaceAll(serviceAccountEmail,
								".gserviceaccount.com", "")
							if err := std.Out.WriteCode("bash",
								fmt.Sprintf(`psql -U %s -d %s -h 127.0.0.1 -p %d`,
									saUsername,
									db,
									proxyPort)); err != nil {
								return errors.Wrapf(err, "write command for db %q", db)
							}
						}

						// Run proxy until stopped
						return proxy.Start(c.Context, c.Int("session.timeout"))
					},
				},
			},
		},
		{
			Name:    "terraform-cloud",
			Aliases: []string{"tfc"},
			Usage:   "Manage Terraform Cloud workspaces for a service",
			Before:  msprepo.UseManagedServicesRepo,
			Subcommands: []*cli.Command{
				{
					Name:        "view",
					Usage:       "View MSP Terraform Cloud workspaces",
					Description: "You may need to request access to the workspaces - see https://handbook.sourcegraph.com/departments/engineering/teams/core-services/managed-services/platform/#terraform-cloud",
					UsageText: `
# View all workspaces for all MSP services
sg msp tfc view

# View all workspaces for all environments for a MSP service
sg msp tfc view <service>

# View all workspaces for a specific MSP service environment
sg msp tfc view <service> <environment>
`,
					ArgsUsage:    "[service ID] [environment ID]",
					BashComplete: msprepo.ServicesAndEnvironmentsCompletion(),
					Action: func(c *cli.Context) error {
						if c.Args().Len() == 0 {
							std.Out.WriteNoticef("Opening link to all MSP Terraform Cloud workspaces in browser...")
							return open.URL(fmt.Sprintf("https://app.terraform.io/app/sourcegraph/workspaces?tag=%s",
								terraformcloud.MSPWorkspaceTag))
						}

						service, err := useServiceArgument(c, false)
						if err != nil {
							return err
						}
						if c.Args().Len() == 1 {
							std.Out.WriteNoticef("Opening link to service Terraform Cloud workspaces in browser...")
							return open.URL(fmt.Sprintf("https://app.terraform.io/app/sourcegraph/workspaces?tag=%s",
								terraformcloud.ServiceWorkspaceTag(service.Service)))
						}

						env := service.GetEnvironment(c.Args().Get(1))
						if env == nil {
							return errors.Wrapf(err, "environment %q not found", c.Args().Get(1))
						}
						std.Out.WriteNoticef("Opening link to service environment Terraform Cloud workspaces in browser...")
						return open.URL(fmt.Sprintf("https://app.terraform.io/app/sourcegraph/workspaces?tag=%s",
							terraformcloud.EnvironmentWorkspaceTag(service.Service, *env)))
					},
				},
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
						secretStore, err := secrets.FromContext(c.Context)
						if err != nil {
							return err
						}
						tfcAccessToken, err := secretStore.GetExternal(c.Context, secrets.ExternalSecret{
							Name:    googlesecretsmanager.SecretTFCOrgToken,
							Project: googlesecretsmanager.SharedSecretsProjectID,
						})
						if err != nil {
							return errors.Wrap(err, "get AccessToken")
						}
						tfcOAuthClient, err := secretStore.GetExternal(c.Context, secrets.ExternalSecret{
							Name:    googlesecretsmanager.SecretTFCOAuthClientID,
							Project: googlesecretsmanager.SharedSecretsProjectID,
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

						// If we are not syncing all environments for a service,
						// then we are syncing a specific service environment.
						if !c.Bool("all") {
							std.Out.WriteNoticef("Syncing a specific service environment...")
							svc, env, err := useServiceAndEnvironmentArguments(c, true)
							if err != nil {
								return err
							}
							return syncEnvironmentWorkspaces(c, tfcClient, svc.Service, *env)
						}

						// Otherwise, we are syncing all environments for a service.
						std.Out.WriteNoticef("Syncing all environments for a specific service ...")
						svc, err := useServiceArgument(c, true)
						if err != nil {
							return err
						}
						for _, env := range svc.Environments {
							if err := syncEnvironmentWorkspaces(c, tfcClient, svc.Service, env); err != nil {
								return errors.Wrapf(err, "sync env %q", env.ID)
							}
						}

						return nil
					},
				},
				{
					Name:      "graph",
					Usage:     "EXPERIMENTAL: Graph the core resources within a Terraform workspace",
					ArgsUsage: "<service ID> <environment ID> <stack ID>",
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:  "dot",
							Usage: "Dump dot graph configuration instead of rendering the image with 'dot'",
						},
					},
					BashComplete: msprepo.ServicesAndEnvironmentsCompletion(
						func(cli.Args) (options []string) {
							return managedservicesplatform.StackNames()
						},
					),
					Action: func(c *cli.Context) error {
						service, env, err := useServiceAndEnvironmentArguments(c, false)
						if err != nil {
							return err
						}

						stack := c.Args().Get(2)
						if stack == "" {
							return errors.New("third argument <stack ID> is required")
						}

						dotgraph, err := msprepo.TerraformGraph(c.Context, service.Service.ID, env.ID, stack)
						if err != nil {
							return err
						}

						if c.Bool("dot") {
							std.Out.Write(dotgraph)
							return nil
						}

						output := fmt.Sprintf("./%s-%s.%s.png", service.Service.ID, env.ID, stack)
						f, err := os.OpenFile(output, os.O_RDWR|os.O_CREATE, 0o644)
						if err != nil {
							return errors.Wrapf(err, "open %q", output)
						}
						defer f.Close()
						if err := run.Cmd(c.Context, "dot -Tpng").
							Input(strings.NewReader(dotgraph + "\n")).
							Environ(os.Environ()).
							Run().
							Stream(f); err != nil {
							return err
						}
						std.Out.WriteSuccessf("Graph rendered in %q", output)
						return nil
					},
				},
			},
		},
		{
			Name:   "fleet",
			Usage:  "Summarize aspects of the MSP fleet",
			Before: msprepo.UseManagedServicesRepo,
			Action: func(c *cli.Context) error {
				services, err := msprepo.ListServices()
				if err != nil {
					return err
				}

				var environmentCount int
				categories := make(map[spec.EnvironmentCategory]int)
				teams := make(map[string]int)
				for _, s := range services {
					svc, err := spec.Open(msprepo.ServiceYAMLPath(s))
					if err != nil {
						return err
					}
					for _, t := range svc.Service.Owners {
						teams[t] += 1
					}
					for _, e := range svc.Environments {
						environmentCount += 1
						categories[e.Category] += 1
					}
				}

				teamNames := maps.Keys(teams)
				sort.Strings(teamNames)
				summary := fmt.Sprintf(`Managed Services Platform fleet summary:

- %d services
- %d teams (%s)
- %d environments
`, len(services), len(teams), strings.Join(teamNames, ", "), environmentCount)
				for category, count := range categories {
					summary += fmt.Sprintf("\t- %s environments: %d\n", category, count)
				}
				return std.Out.WriteMarkdown(summary)
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
					if err := os.WriteFile(output, jsonSchema, 0o644); err != nil {
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
