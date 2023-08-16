package main

import (
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform"
	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/spec"
)

var managedServicesPlatformCommand = &cli.Command{
	Name:        "managed-services-platform",
	Aliases:     []string{"msp"},
	Usage:       "",
	Description: ``,
	Category:    CategoryCompany,
	Subcommands: []*cli.Command{
		{
			Name: "generate",
			Action: func(c *cli.Context) error {
				r := managedservicesplatform.Renderer{
					// TODO populate from secrets
					OutputDir: "./tmp",
				}
				s := spec.Spec{
					Service: spec.ServiceSpec{ID: "cody-gateway"},
					Environments: []spec.EnvironmentSpec{
						{Name: "prod", Deploy: spec.EnvironmentDeploySpec{Type: "manual"}},
					},
				}
				cdktf, err := r.RenderEnvironment(s.Service, s.Build, *s.GetEnvironment("prod"))
				if err != nil {
					return err
				}
				return cdktf.Synthesize()
			},
		},
	},
}
