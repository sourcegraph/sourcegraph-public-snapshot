package conf

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/pkg/conf/confdefaults"
	"github.com/sourcegraph/sourcegraph/pkg/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/pkg/jsonc"
	"github.com/sourcegraph/sourcegraph/schema"
	"github.com/xeipuuv/gojsonschema"
)

// ignoreLegacyKubernetesFields is the set of field names for which validation errors should be
// ignored. The validation errors occur only because deploy-sourcegraph config merged site config
// and Kubernetes cluster-specific config. This is deprecated. Until we have transitioned fully, we
// suppress validation errors on these fields.
var ignoreLegacyKubernetesFields = map[string]struct{}{
	"alertmanagerConfig":    {},
	"alertmanagerURL":       {},
	"authProxyIP":           {},
	"authProxyPassword":     {},
	"deploymentOverrides":   {},
	"gitoliteIP":            {},
	"gitserverCount":        {},
	"gitserverDiskSize":     {},
	"gitserverSSH":          {},
	"httpNodePort":          {},
	"httpsNodePort":         {},
	"indexedSearchDiskSize": {},
	"langGo":                {},
	"langJava":              {},
	"langJavaScript":        {},
	"langPHP":               {},
	"langPython":            {},
	"langSwift":             {},
	"langTypeScript":        {},
	"namespace":             {},
	"nodeSSDPath":           {},
	"phabricatorIP":         {},
	"prometheus":            {},
	"pyPIIP":                {},
	"rbac":                  {},
	"storageClass":          {},
	"useAlertManager":       {},
}

// Validate validates the configuration against the JSON Schema and other
// custom validation checks.
func Validate(input conftypes.RawUnified) (problems []string, err error) {
	criticalProblems, err := doValidate(input.Critical, schema.CriticalSchemaJSON)
	if err != nil {
		return nil, err
	}
	problems = append(problems, criticalProblems...)

	siteProblems, err := doValidate(input.Site, schema.SiteSchemaJSON)
	if err != nil {
		return nil, err
	}
	problems = append(problems, siteProblems...)

	customProblems, err := validateCustomRaw(conftypes.RawUnified{
		Critical: string(jsonc.Normalize(input.Critical)),
		Site:     string(jsonc.Normalize(input.Site)),
	})
	if err != nil {
		return nil, err
	}
	problems = append(problems, customProblems...)
	return problems, nil
}

// ValidateSite is like Validate, except it only validates the site configuration.
func ValidateSite(input string) (problems []string, err error) {
	raw := Raw()
	raw.Site = input
	return Validate(raw)
}

func doValidate(inputStr, schema string) (problems []string, err error) {
	input := []byte(jsonc.Normalize(inputStr))

	res, err := validate([]byte(schema), input)
	if err != nil {
		return nil, err
	}
	problems = make([]string, 0, len(res.Errors()))
	for _, e := range res.Errors() {
		if _, ok := ignoreLegacyKubernetesFields[e.Field()]; ok {
			continue
		}

		var keyPath string
		if c := e.Context(); c != nil {
			keyPath = strings.TrimPrefix(e.Context().String("."), "(root).")
		} else {
			keyPath = e.Field()
		}

		problems = append(problems, fmt.Sprintf("%s: %s", keyPath, e.Description()))
	}
	return problems, nil
}

func validate(schema, input []byte) (*gojsonschema.Result, error) {
	if len(input) > 0 {
		// HACK: Remove the "settings" field from site config because
		// github.com/xeipuuv/gojsonschema has a bug where $ref'd schemas do not always get
		// loaded. When https://github.com/xeipuuv/gojsonschema/pull/196 is merged, it will probably
		// be fixed. This means that the backend config validation will not validate settings, but
		// that is OK because specifying settings here is discouraged anyway.
		var v map[string]interface{}
		if err := json.Unmarshal(input, &v); err != nil {
			return nil, err
		}
		delete(v, "settings")
		var err error
		input, err = json.Marshal(v)
		if err != nil {
			return nil, err
		}
	}

	s, err := gojsonschema.NewSchema(jsonLoader{gojsonschema.NewBytesLoader(schema)})
	if err != nil {
		return nil, err
	}
	return s.Validate(gojsonschema.NewBytesLoader(input))
}

type jsonLoader struct {
	gojsonschema.JSONLoader
}

func (l jsonLoader) LoaderFactory() gojsonschema.JSONLoaderFactory {
	return &jsonLoaderFactory{}
}

type jsonLoaderFactory struct{}

func (f jsonLoaderFactory) New(source string) gojsonschema.JSONLoader {
	switch source {
	case "settings.schema.json":
		return gojsonschema.NewStringLoader(schema.SettingsSchemaJSON)
	case "site.schema.json":
		return gojsonschema.NewStringLoader(schema.SiteSchemaJSON)
	case "critical.schema.json":
		return gojsonschema.NewStringLoader(schema.CriticalSchemaJSON)
	}
	return nil
}

// MustValidateDefaults should be called after all custom validators have been
// registered. It will panic if any of the default deployment configurations
// are invalid.
func MustValidateDefaults() {
	mustValidate("DevAndTesting", confdefaults.DevAndTesting)
	mustValidate("DockerContainer", confdefaults.DockerContainer)
	mustValidate("Cluster", confdefaults.Cluster)
}

// mustValidate panics if the configuration does not pass validation.
func mustValidate(name string, cfg conftypes.RawUnified) conftypes.RawUnified {
	problems, err := Validate(cfg)
	if err != nil {
		panic(fmt.Sprintf("Error with %q: %s", name, err))
	}
	if len(problems) > 0 {
		panic(fmt.Sprintf("conf: problems with default configuration for %q:\n  %s", name, strings.Join(problems, "\n  ")))
	}
	return cfg
}
