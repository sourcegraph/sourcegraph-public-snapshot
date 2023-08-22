//go:build msp
// +build msp

package msp

import (
	"context"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/secrets"
	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform"
	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/internal/pointer"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// This file is only built when '-tags=msp' is passed to go build while 'sg msp'
// is experimental, as the introduction of this command significantly introduces
// the binary size of 'sg'.
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
			Name: "generate",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "output",
					Usage:    "Output directory for generated Terraform assets",
					Required: true, // TODO have a default
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

				renderer := managedservicesplatform.Renderer{
					OutputDir: c.String("output"),
					GCP:       *gcpOptions,
					TFC:       *tfcOptions,
				}

				// This is just an example spec for now emulating Cody Gateway
				// infrastructure.
				// TODO: load from file.
				exampleSpec := spec.Spec{
					Service: spec.ServiceSpec{
						ID: "cody-gateway",
					},
					Build: spec.BuildSpec{
						Image: "index.docker.io/sourcegraph/cody-gateway",
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
									Subdomain: "cody-gateway",
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
								BigQueryTable: &spec.EnvironmentResourceBigQueryTableSpec{
									Region:  "us-central1",
									TableID: "events",
									Schema: []spec.EnvironmentResourceBigQuerySchemaColumn{
										{
											Name:        "name",
											Type:        "STRING",
											Mode:        "REQUIRED",
											Description: "The name of the event",
										},
										// TODO
									},
								},
							},
							Env: map[string]string{
								"SRC_LOG_LEVEL": "debug",
							},
							SecretEnv: map[string]string{
								"CODY_GATEWAY_ANTHROPIC_ACCESS_TOKEN": "CODY_GATEWAY_ANTHROPIC_ACCESS_TOKEN",
							},
						},
					},
				}

				cdktf, err := renderer.RenderEnvironment(exampleSpec.Service, exampleSpec.Build, *exampleSpec.GetEnvironment("prod"))
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
