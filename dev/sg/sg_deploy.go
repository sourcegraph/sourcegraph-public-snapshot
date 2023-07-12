package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var deployCommand = &cli.Command{
	Name:        "deploy",
	Usage:       `Generate a Kubernetes manifest for a Sourcegraph deployment`,
	Description: `Internal deployments live in the sourcegraph/infra repository.`,
	UsageText: `
sg deploy --values <path to values file>
`,
	Category: CategoryDev,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "values",
			Usage:    "perform the RFC action on the private RFC drive",
			Required: true,
		},
	},
	Action: func(c *cli.Context) error {
		err := generateManifest(c.String("values"))
		if err != nil {
			return errors.Wrap(err, "generate manifest")
		}
		return nil
	}}

type Envvar struct {
	name  string
	value string
}

type Port struct {
	name  string
	value int
}
type Values struct {
	Name    string
	Envvars []struct {
		Name  string `yaml:"name"`
		Value string `yaml:"value"`
	}
	Image    string
	Replicas int
	Ports    []struct {
		Name  string
		Value int
	}
}

func generateManifest(configFile string) error {

	var values Values
	v, err := os.ReadFile(configFile)
	if err != nil {
		return errors.Wrap(err, "read Values file")
	}

	err = yaml.Unmarshal(v, &values)
	if err != nil {
		return errors.Wrap(err, "reading config file")
	}
	fmt.Println(values.Envvars)
	return nil
}
