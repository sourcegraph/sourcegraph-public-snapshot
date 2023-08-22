//go:build msp
// +build msp

package msp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/secrets"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/internal/pointer"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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
	Command.Subcommands = []*cli.Command{
		{
			Name:        "init",
			ArgsUsage:   "<service ID>",
			Description: "Initialize a template Managed Services Platform service spec",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "output",
					Aliases: []string{"o"},
					Usage:   "Output directory for generated spec file",
				},
			},
			Action: func(c *cli.Context) error {
				if c.Args().Len() != 1 {
					return errors.New("exactly 1 argument required: service ID")
				}
				exampleSpec, err := (spec.Spec{
					Service: spec.ServiceSpec{
						ID: c.Args().First(),
					},
					Build: spec.BuildSpec{
						Image: "index.docker.io/sourcegraph/" + c.Args().First(),
					},
					Environments: []spec.EnvironmentSpec{
						{
							ID: "dev",
							Deploy: spec.EnvironmentDeploySpec{
								Type: "manual",
							},
							Domain: spec.EnvironmentDomainSpec{
								Type: "cloudflare",
								Cloudflare: &spec.EnvironmentDomainCloudflareSpec{
									Subdomain: c.Args().First(),
									Zone:      "sgdev.org",
									Required:  false,
								},
							},
							Instances: spec.EnvironmentInstancesSpec{
								Resources: spec.EnvironmentInstancesResourcesSpec{
									CPU:    1,
									Memory: "512Mi",
								},
								Scaling: spec.EnvironmentInstancesScalingSpec{
									MaxCount: pointer.Value(1),
								},
							},
							Resources: spec.EnvironmentResourcesSpec{
								Redis: &spec.EnvironmentResourceRedisSpec{},
								// TODO: Not implemented.
								// BigQueryTable: &spec.EnvironmentResourceBigQueryTableSpec{
								// 	Region:  "us-central1",
								// 	TableID: "events",
								// 	Schema: []spec.EnvironmentResourceBigQuerySchemaColumn{
								// 		{
								// 			Name:        "name",
								// 			Type:        "STRING",
								// 			Mode:        "REQUIRED",
								// 			Description: "The name of the event",
								// 		},
								// 		// TODO
								// 	},
								// },
							},
							Env: map[string]string{
								"SRC_LOG_LEVEL": "info",
							},
							SecretEnv: map[string]string{
								"SUPER_SEKRET_VAR": "SUPER_SEKRET_VAR",
							},
						},
					},
				}).MarshalYAML()
				if err != nil {
					return err
				}
				output := filepath.Join(
					c.String("output"), fmt.Sprintf("%s.service.yaml", c.Args().First()))
				if err := os.WriteFile(output, exampleSpec, os.ModePerm); err != nil {
					return err
				}

				std.Out.WriteSuccessf("Rendered template spec in %s", output)
				return nil
			},
		},
		{
			Name:        "generate",
			ArgsUsage:   "<service spec file> <environment ID>",
			Description: "Generate Terraform assets for a Managed Services Platform service spec.",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "output",
					Aliases:  []string{"o"},
					Usage:    "Output directory for generated Terraform assets",
					Required: true,
				},
				&cli.BoolFlag{
					Name:  "tfc",
					Usage: "Generate infrastructure stacks with Terraform Cloud backends",
					Value: false, // TODO default to true
				},
				&cli.BoolFlag{
					Name:  "gcp",
					Usage: "Generate infrastructure stacks on real GCP configuration",
					Value: false, // TODO default to true
				},
			},
			Action: func(c *cli.Context) error {
				if c.Args().Len() != 2 {
					return errors.New("exactly 2 arguments required: service spec file and environment ID")
				}

				// Load specification
				serviceSpecData, err := os.ReadFile(c.Args().First())
				if err != nil {
					return err
				}
				serviceSpec, err := spec.Parse(serviceSpecData)
				if err != nil {
					return err
				}
				deployEnv := serviceSpec.GetEnvironment(c.Args().Get(1))
				if deployEnv == nil {
					return errors.Newf("environment %q not found in service spec", c.Args().Get(1))
				}

				// Collect shared configuration
				secretStore, err := secrets.FromContext(c.Context)
				if err != nil {
					return err
				}
				tfcOptions, err := mspLoadTFCConfig(c.Context, secretStore, c.Bool("tfc"))
				if err != nil {
					return errors.Wrap(err, "load TFC config")
				}
				gcpOptions, err := mspLoadGCPConfig(c.Context, secretStore, c.Bool("gcp"))
				if err != nil {
					return errors.Wrap(err, "load GCP config")
				}

				// Render assets
				renderer := managedservicesplatform.Renderer{
					OutputDir: c.String("output"),
					GCP:       *gcpOptions,
					TFC:       *tfcOptions,
				}
				cdktf, err := renderer.RenderEnvironment(serviceSpec.Service, serviceSpec.Build, *deployEnv)
				if err != nil {
					return err
				}
				return cdktf.Synthesize()
			},
		},
	}
}

// mspLoadGCPConfig retrieves secret GCP configuration for Managed Services Platform.
func mspLoadGCPConfig(ctx context.Context, s *secrets.Store, enabled bool) (*managedservicesplatform.GCPOptions, error) {
	if !enabled {
		return &managedservicesplatform.GCPOptions{
			ParentFolderID:   "EXAMPLE",
			BillingAccountID: "EXAMPLE",
		}, nil
	}

	// TODO
	parentFolderID, err := s.GetExternal(ctx, secrets.ExternalSecret{
		Project: "",
		Name:    "",
	})
	if err != nil {
		return nil, errors.Wrap(err, "get ParentFolderID")
	}

	// TODO
	billingAccountID, err := s.GetExternal(ctx, secrets.ExternalSecret{
		Project: "",
		Name:    "",
	})
	if err != nil {
		return nil, errors.Wrap(err, "get BillingAccountID")
	}

	return &managedservicesplatform.GCPOptions{
		ParentFolderID:   parentFolderID,
		BillingAccountID: billingAccountID,
	}, nil
}

// mspLoadTFCConfig retrieves secret TFC configuration for Managed Services Platform.
func mspLoadTFCConfig(ctx context.Context, s *secrets.Store, enabled bool) (*managedservicesplatform.TerraformCloudOptions, error) {
	if !enabled {
		return &managedservicesplatform.TerraformCloudOptions{
			Enabled: false,
		}, nil
	}

	// TODO
	accessToken, err := s.GetExternal(ctx, secrets.ExternalSecret{
		Project: "",
		Name:    "",
	})
	if err != nil {
		return nil, errors.Wrap(err, "get AccessToken")
	}

	return &managedservicesplatform.TerraformCloudOptions{
		Enabled:     true,
		AccessToken: accessToken,
	}, nil
}
