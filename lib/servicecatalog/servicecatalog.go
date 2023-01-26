package servicecatalogue

import (
	_ "embed"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"gopkg.in/yaml.v3"
)

//go:embed service-catalog.yaml
var rawCatalog string

type Service struct {
	Consumers []string `yaml:"consumers" json:"consumers"`
}

type Catalog struct {
	ProtectedServices map[string]Service `yaml:"protected_services" json:"protected_services"`
}

func Get() (Catalog, error) {
	var c Catalog
	if err := yaml.Unmarshal([]byte(rawCatalog), &c); err != nil {
		return c, errors.Wrap(err, "'service-catalog.yaml' is invalid")
	}
	return c, nil
}
